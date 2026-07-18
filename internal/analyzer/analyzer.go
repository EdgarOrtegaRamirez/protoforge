// Package analyzer provides analysis and statistics for parsed .proto files.
package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/models"
)

// Analysis holds the results of analyzing a .proto file.
type Analysis struct {
	Package      string
	Syntax       string
	Stats        *Stats
	Issues       []*Issue
	Warnings     []*Issue
	Imports      []*ImportInfo
	Messages     []*MessageInfo
	Enums        []*EnumInfo
	Services     []*ServiceInfo
	Dependencies []string
}

// Stats holds numerical statistics.
type Stats struct {
	TotalMessages   int
	TotalEnums      int
	TotalServices   int
	TotalMethods    int
	TotalFields     int
	TotalOneOfs     int
	TotalImports    int
	TotalOptions    int
	AvgFieldsPerMsg float64
	MaxNesting      int
	SyntaxVersion   string
}

// Issue represents a linting or quality issue.
type Issue struct {
	Severity string // error, warning, info
	Message  string
	Path     string // e.g., "message.User.name"
	Line     int
	Rule     string // rule identifier
}

// ImportInfo provides information about an import.
type ImportInfo struct {
	Path   string
	Weak   bool
	Public bool
}

// MessageInfo provides information about a message.
type MessageInfo struct {
	Name        string
	FieldCount  int
	NestedCount int
	OneOfCount  int
	HasMaps     bool
	IsNested    bool
	Path        string
}

// EnumInfo provides information about an enum.
type EnumInfo struct {
	Name       string
	ValueCount int
	HasZero    bool
	Path       string
}

// ServiceInfo provides information about a service.
type ServiceInfo struct {
	Name        string
	MethodCount int
	StreamCount int
	Path        string
}

// Analyze performs a comprehensive analysis of a parsed proto file.
func Analyze(pf *models.ProtoFile) *Analysis {
	a := &Analysis{
		Package: pf.Package,
		Syntax:  pf.Syntax,
		Stats:   &Stats{SyntaxVersion: pf.Syntax},
	}

	a.collectImports(pf)
	a.collectMessages(pf, "")
	a.collectEnums(pf, "")
	a.collectServices(pf, "")
	a.collectDependencies(pf)
	a.computeStats(pf)
	a.detectIssues(pf)

	return a
}

func (a *Analysis) collectImports(pf *models.ProtoFile) {
	for _, imp := range pf.Imports {
		info := &ImportInfo{
			Path:   imp.Path,
			Weak:   imp.Weak,
			Public: imp.Public,
		}
		a.Imports = append(a.Imports, info)
		a.Stats.TotalImports++
	}
}

func (a *Analysis) collectMessages(pf *models.ProtoFile, prefix string) {
	for _, msg := range pf.Messages {
		fullPath := msg.Name
		if prefix != "" {
			fullPath = prefix + "." + msg.Name
		}

		info := &MessageInfo{
			Name:        msg.Name,
			FieldCount:  len(msg.Fields),
			NestedCount: len(msg.Messages),
			OneOfCount:  len(msg.OneOfs),
			HasMaps:     len(msg.Maps) > 0,
			IsNested:    prefix != "",
			Path:        fullPath,
		}
		a.Messages = append(a.Messages, info)
		a.Stats.TotalMessages++
		a.Stats.TotalFields += len(msg.Fields) + len(msg.Maps)

		for _, oo := range msg.OneOfs {
			a.Stats.TotalOneOfs++
			a.Stats.TotalFields += len(oo.Fields)
		}

		// Recurse into nested messages
		if len(msg.Messages) > 0 {
			nestedPf := &models.ProtoFile{Messages: msg.Messages}
			a.collectMessages(nestedPf, fullPath)
		}
	}
}

func (a *Analysis) collectEnums(pf *models.ProtoFile, prefix string) {
	for _, enum := range pf.Enums {
		fullPath := enum.Name
		if prefix != "" {
			fullPath = prefix + "." + enum.Name
		}

		info := &EnumInfo{
			Name:       enum.Name,
			ValueCount: len(enum.Values),
			Path:       fullPath,
		}

		// Check if enum has a zero value
		for _, v := range enum.Values {
			if v.Number == 0 {
				info.HasZero = true
				break
			}
		}

		a.Enums = append(a.Enums, info)
		a.Stats.TotalEnums++
	}
}

func (a *Analysis) collectServices(pf *models.ProtoFile, prefix string) {
	for _, svc := range pf.Services {
		info := &ServiceInfo{
			Name:        svc.Name,
			MethodCount: len(svc.Methods),
			Path:        svc.Name,
		}

		for _, method := range svc.Methods {
			a.Stats.TotalMethods++
			if method.ClientStreaming || method.ServerStreaming {
				info.StreamCount++
			}
		}

		a.Services = append(a.Services, info)
		a.Stats.TotalServices++
	}
}

func (a *Analysis) collectDependencies(pf *models.ProtoFile) {
	for _, imp := range pf.Imports {
		if !imp.Weak {
			a.Dependencies = append(a.Dependencies, imp.Path)
		}
	}
}

func (a *Analysis) computeStats(pf *models.ProtoFile) {
	a.Stats.TotalOptions = len(pf.Options)

	// Count options in messages
	for _, msg := range pf.Messages {
		a.Stats.TotalOptions += len(msg.Options)
		for _, field := range msg.Fields {
			a.Stats.TotalOptions += len(field.Options)
		}
	}

	// Compute average fields per message
	if a.Stats.TotalMessages > 0 {
		a.Stats.AvgFieldsPerMsg = float64(a.Stats.TotalFields) / float64(a.Stats.TotalMessages)
	}

	// Compute max nesting depth
	a.Stats.MaxNesting = computeMaxNesting(pf.Messages, 0)
}

func computeMaxNesting(messages []*models.Message, depth int) int {
	maxDepth := depth
	for _, msg := range messages {
		if len(msg.Messages) > 0 {
			d := computeMaxNesting(msg.Messages, depth+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
	}
	return maxDepth
}

func (a *Analysis) detectIssues(pf *models.ProtoFile) {
	// Check syntax declaration
	if pf.Syntax == "" {
		a.Issues = append(a.Issues, &Issue{
			Severity: "warning",
			Message:  "Missing syntax declaration",
			Rule:     "missing-syntax",
		})
	} else if pf.Syntax != "proto2" && pf.Syntax != "proto3" {
		a.Issues = append(a.Issues, &Issue{
			Severity: "error",
			Message:  fmt.Sprintf("Unknown syntax version: %s", pf.Syntax),
			Rule:     "unknown-syntax",
		})
	}

	// Check package declaration
	if pf.Package == "" {
		a.Issues = append(a.Issues, &Issue{
			Severity: "warning",
			Message:  "Missing package declaration",
			Rule:     "missing-package",
		})
	}

	// Check for empty files
	if len(pf.Messages) == 0 && len(pf.Enums) == 0 && len(pf.Services) == 0 {
		a.Issues = append(a.Issues, &Issue{
			Severity: "info",
			Message:  "File has no definitions (messages, enums, or services)",
			Rule:     "empty-file",
		})
	}

	// Check each message
	for _, msg := range pf.Messages {
		a.checkMessage(msg, pf.Package, "")
	}

	// Check each enum
	for _, enum := range pf.Enums {
		a.checkEnum(enum, pf.Package, "")
	}

	// Check each service
	for _, svc := range pf.Services {
		a.checkService(svc)
	}
}

func (a *Analysis) checkMessage(msg *models.Message, pkg string, prefix string) {
	path := msg.Name
	if prefix != "" {
		path = prefix + "." + msg.Name
	}

	// Check for fields with 0 number
	for _, field := range msg.Fields {
		if field.Number == 0 {
			a.Issues = append(a.Issues, &Issue{
				Severity: "error",
				Message:  fmt.Sprintf("Field %s has number 0", field.Name),
				Path:     path + "." + field.Name,
				Rule:     "field-number-zero",
			})
		}
		// Check for reserved field numbers (19000-19999)
		if field.Number >= 19000 && field.Number <= 19999 {
			a.Issues = append(a.Issues, &Issue{
				Severity: "error",
				Message:  fmt.Sprintf("Field %s uses reserved number %d", field.Name, field.Number),
				Path:     path + "." + field.Name,
				Rule:     "reserved-field-number",
			})
		}
	}

	// Check for messages with no fields
	if len(msg.Fields) == 0 && len(msg.OneOfs) == 0 && len(msg.Maps) == 0 {
		a.Warnings = append(a.Warnings, &Issue{
			Severity: "warning",
			Message:  fmt.Sprintf("Message %s has no fields", msg.Name),
			Path:     path,
			Rule:     "empty-message",
		})
	}

	// Check naming convention (CamelCase for messages)
	if len(msg.Name) > 0 && msg.Name[0] >= 'a' && msg.Name[0] <= 'z' {
		a.Warnings = append(a.Warnings, &Issue{
			Severity: "warning",
			Message:  fmt.Sprintf("Message name %s should start with uppercase (CamelCase)", msg.Name),
			Path:     path,
			Rule:     "naming-message",
		})
	}

	// Recurse into nested messages
	for _, nested := range msg.Messages {
		a.checkMessage(nested, pkg, path)
	}
}

func (a *Analysis) checkEnum(enum *models.Enum, pkg string, prefix string) {
	path := enum.Name
	if prefix != "" {
		path = prefix + "." + enum.Name
	}

	// Check for zero value
	if !enumHasZeroValue(enum) {
		a.Warnings = append(a.Warnings, &Issue{
			Severity: "warning",
			Message:  fmt.Sprintf("Enum %s has no zero value (first value should be %s_UNSPECIFIED = 0)", enum.Name, strings.ToUpper(enum.Name)),
			Path:     path,
			Rule:     "enum-no-zero",
		})
	}

	// Check naming convention (SCREAMING_SNAKE for enum values)
	for _, val := range enum.Values {
		if val.Name != strings.ToUpper(val.Name) {
			a.Warnings = append(a.Warnings, &Issue{
				Severity: "info",
				Message:  fmt.Sprintf("Enum value %s should be SCREAMING_SNAKE_CASE", val.Name),
				Path:     path + "." + val.Name,
				Rule:     "naming-enum-value",
			})
		}
	}

	// Check for duplicate enum values
	seen := make(map[int]string)
	for _, val := range enum.Values {
		if prev, ok := seen[val.Number]; ok {
			a.Issues = append(a.Issues, &Issue{
				Severity: "error",
				Message:  fmt.Sprintf("Duplicate enum value number %d: %s and %s", val.Number, prev, val.Name),
				Path:     path + "." + val.Name,
				Rule:     "enum-duplicate-value",
			})
		}
		seen[val.Number] = val.Name
	}
}

func enumHasZeroValue(enum *models.Enum) bool {
	for _, v := range enum.Values {
		if v.Number == 0 {
			return true
		}
	}
	return false
}

func (a *Analysis) checkService(svc *models.Service) {
	// Check naming convention (CamelCase for services)
	if len(svc.Name) > 0 && svc.Name[0] >= 'a' && svc.Name[0] <= 'z' {
		a.Warnings = append(a.Warnings, &Issue{
			Severity: "warning",
			Message:  fmt.Sprintf("Service name %s should start with uppercase (CamelCase)", svc.Name),
			Path:     svc.Name,
			Rule:     "naming-service",
		})
	}

	// Check each method
	for _, method := range svc.Methods {
		// Check naming convention (CamelCase for methods)
		if len(method.Name) > 0 && method.Name[0] >= 'a' && method.Name[0] <= 'z' {
			a.Warnings = append(a.Warnings, &Issue{
				Severity: "info",
				Message:  fmt.Sprintf("Method name %s should start with uppercase (CamelCase)", method.Name),
				Path:     svc.Name + "." + method.Name,
				Rule:     "naming-method",
			})
		}
	}
}

// Summary returns a human-readable summary of the analysis.
func (a *Analysis) Summary() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Package: %s\n", a.Package)
	fmt.Fprintf(&sb, "Syntax:  %s\n", a.Syntax)
	sb.WriteString("\n--- Statistics ---\n")
	fmt.Fprintf(&sb, "Messages:  %d\n", a.Stats.TotalMessages)
	fmt.Fprintf(&sb, "Enums:     %d\n", a.Stats.TotalEnums)
	fmt.Fprintf(&sb, "Services:  %d\n", a.Stats.TotalServices)
	fmt.Fprintf(&sb, "Methods:   %d\n", a.Stats.TotalMethods)
	fmt.Fprintf(&sb, "Fields:    %d\n", a.Stats.TotalFields)
	fmt.Fprintf(&sb, "OneOfs:    %d\n", a.Stats.TotalOneOfs)
	fmt.Fprintf(&sb, "Imports:   %d\n", a.Stats.TotalImports)
	fmt.Fprintf(&sb, "Options:   %d\n", a.Stats.TotalOptions)
	fmt.Fprintf(&sb, "Avg Fields/Msg: %.1f\n", a.Stats.AvgFieldsPerMsg)
	fmt.Fprintf(&sb, "Max Nesting:    %d\n", a.Stats.MaxNesting)

	if len(a.Issues) > 0 {
		sb.WriteString("\n--- Errors ---\n")
		for _, issue := range a.Issues {
			fmt.Fprintf(&sb, "  [%s] %s\n", issue.Rule, issue.Message)
		}
	}

	if len(a.Warnings) > 0 {
		sb.WriteString("\n--- Warnings ---\n")
		for _, warn := range a.Warnings {
			fmt.Fprintf(&sb, "  [%s] %s\n", warn.Rule, warn.Message)
		}
	}

	return sb.String()
}

// JSON returns the analysis as a map for JSON serialization.
func (a *Analysis) JSON() map[string]interface{} {
	result := map[string]interface{}{
		"package": a.Package,
		"syntax":  a.Syntax,
		"stats": map[string]interface{}{
			"total_messages":     a.Stats.TotalMessages,
			"total_enums":        a.Stats.TotalEnums,
			"total_services":     a.Stats.TotalServices,
			"total_methods":      a.Stats.TotalMethods,
			"total_fields":       a.Stats.TotalFields,
			"total_oneofs":       a.Stats.TotalOneOfs,
			"total_imports":      a.Stats.TotalImports,
			"total_options":      a.Stats.TotalOptions,
			"avg_fields_per_msg": a.Stats.AvgFieldsPerMsg,
			"max_nesting":        a.Stats.MaxNesting,
		},
	}

	if len(a.Issues) > 0 {
		issues := make([]map[string]interface{}, 0, len(a.Issues))
		for _, issue := range a.Issues {
			issues = append(issues, map[string]interface{}{
				"severity": issue.Severity,
				"message":  issue.Message,
				"path":     issue.Path,
				"rule":     issue.Rule,
			})
		}
		result["issues"] = issues
	}

	if len(a.Warnings) > 0 {
		warnings := make([]map[string]interface{}, 0, len(a.Warnings))
		for _, w := range a.Warnings {
			warnings = append(warnings, map[string]interface{}{
				"severity": w.Severity,
				"message":  w.Message,
				"path":     w.Path,
				"rule":     w.Rule,
			})
		}
		result["warnings"] = warnings
	}

	return result
}

// SortedMessages returns messages sorted by name.
func (a *Analysis) SortedMessages() []*MessageInfo {
	sorted := make([]*MessageInfo, len(a.Messages))
	copy(sorted, a.Messages)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// SortedEnums returns enums sorted by name.
func (a *Analysis) SortedEnums() []*EnumInfo {
	sorted := make([]*EnumInfo, len(a.Enums))
	copy(sorted, a.Enums)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// SortedServices returns services sorted by name.
func (a *Analysis) SortedServices() []*ServiceInfo {
	sorted := make([]*ServiceInfo, len(a.Services))
	copy(sorted, a.Services)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}
