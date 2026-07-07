// Package validator provides validation rules for Protocol Buffer schemas.
package validator

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/models"
)

// Rule represents a validation rule.
type Rule struct {
	ID          string
	Name        string
	Description string
	Severity    string // error, warning, info
}

// ValidationResult holds the result of validation.
type ValidationResult struct {
	Valid  bool
	Errors []*ValidationIssue
	Warnings []*ValidationIssue
	Info   []*ValidationIssue
}

// ValidationIssue represents a single validation issue.
type ValidationIssue struct {
	Rule     string
	Message  string
	Path     string
	Severity string
}

// Standard validation rules.
var (
	RuleSyntaxRequired = &Rule{
		ID:          "syntax-required",
		Name:        "Syntax Required",
		Description: "Proto files must declare a syntax version",
		Severity:    "warning",
	}
	RulePackageRequired = &Rule{
		ID:          "package-required",
		Name:        "Package Required",
		Description: "Proto files should declare a package",
		Severity:    "warning",
	}
	RuleFieldNumberValid = &Rule{
		ID:          "field-number-valid",
		Name:        "Valid Field Number",
		Description: "Field numbers must be between 1 and 536,870,911",
		Severity:    "error",
	}
	RuleFieldNumberReserved = &Rule{
		ID:          "field-number-reserved",
		Name:        "Reserved Field Number",
		Description: "Field numbers 19000-19999 are reserved by protobuf",
		Severity:    "error",
	}
	RuleFieldNameCamelCase = &Rule{
		ID:          "field-name-camelcase",
		Name:        "Field Name CamelCase",
		Description: "Field names should use lower_snake_case",
		Severity:    "warning",
	}
	RuleMessageNamePascalCase = &Rule{
		ID:          "message-name-pascalcase",
		Name:        "Message Name PascalCase",
		Description: "Message names should start with uppercase (PascalCase)",
		Severity:    "warning",
	}
	RuleEnumNamePascalCase = &Rule{
		ID:          "enum-name-pascalcase",
		Name:        "Enum Name PascalCase",
		Description: "Enum names should start with uppercase (PascalCase)",
		Severity:    "warning",
	}
	RuleEnumValueScreamingSnake = &Rule{
		ID:          "enum-value-screaming",
		Name:        "Enum Value SCREAMING_SNAKE",
		Description: "Enum values should be SCREAMING_SNAKE_CASE",
		Severity:    "warning",
	}
	RuleEnumZeroValue = &Rule{
		ID:          "enum-zero-value",
		Name:        "Enum Zero Value",
		Description: "Enums should have a zero value as first entry",
		Severity:    "warning",
	}
	RuleServiceNamePascalCase = &Rule{
		ID:          "service-name-pascalcase",
		Name:        "Service Name PascalCase",
		Description: "Service names should start with uppercase (PascalCase)",
		Severity:    "warning",
	}
	RuleMethodNamePascalCase = &Rule{
		ID:          "method-name-pascalcase",
		Name:        "Method Name PascalCase",
		Description: "Method names should start with uppercase (PascalCase)",
		Severity:    "warning",
	}
	RuleMaxNestingDepth = &Rule{
		ID:          "max-nesting-depth",
		Name:        "Max Nesting Depth",
		Description: "Messages should not be nested more than 5 levels deep",
		Severity:    "warning",
	}
	RuleMaxFieldsPerMessage = &Rule{
		ID:          "max-fields-per-message",
		Name:        "Max Fields Per Message",
		Description: "Messages should not have more than 100 fields",
		Severity:    "warning",
	}
	RuleNoDuplicateFieldNumbers = &Rule{
		ID:          "no-duplicate-field-numbers",
		Name:        "No Duplicate Field Numbers",
		Description: "Messages must not have duplicate field numbers",
		Severity:    "error",
	}
	RuleMapKeyType = &Rule{
		ID:          "map-key-type",
		Name:        "Valid Map Key Type",
		Description: "Map keys must be integral types or strings",
		Severity:    "error",
	}
	RuleProto3NoRequired = &Rule{
		ID:          "proto3-no-required",
		Name:        "Proto3 No Required",
		Description: "Proto3 syntax does not support required fields",
		Severity:    "error",
	}
	RuleProto3NoGroups = &Rule{
		ID:          "proto3-no-groups",
		Name:        "Proto3 No Groups",
		Description: "Proto3 syntax does not support group fields",
		Severity:    "error",
	}
)

// AllRules returns all standard validation rules.
func AllRules() []*Rule {
	return []*Rule{
		RuleSyntaxRequired,
		RulePackageRequired,
		RuleFieldNumberValid,
		RuleFieldNumberReserved,
		RuleFieldNameCamelCase,
		RuleMessageNamePascalCase,
		RuleEnumNamePascalCase,
		RuleEnumValueScreamingSnake,
		RuleEnumZeroValue,
		RuleServiceNamePascalCase,
		RuleMethodNamePascalCase,
		RuleMaxNestingDepth,
		RuleMaxFieldsPerMessage,
		RuleNoDuplicateFieldNumbers,
		RuleMapKeyType,
		RuleProto3NoRequired,
		RuleProto3NoGroups,
	}
}

// Validate performs validation on a parsed proto file.
func Validate(pf *models.ProtoFile) *ValidationResult {
	result := &ValidationResult{Valid: true}

	validateSyntax(pf, result)
	validatePackage(pf, result)
	validateMessages(pf, result, pf.Syntax, "")
	validateEnums(pf, result, pf.Syntax, "")
	validateServices(pf, result)
	validateImports(pf, result)

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

func validateSyntax(pf *models.ProtoFile, result *ValidationResult) {
	if pf.Syntax == "" {
		addIssue(result, RuleSyntaxRequired, "", "No syntax declaration found")
	} else if pf.Syntax != "proto2" && pf.Syntax != "proto3" {
		addIssue(result, RuleSyntaxRequired, "", fmt.Sprintf("Unknown syntax: %s", pf.Syntax))
	}
}

func validatePackage(pf *models.ProtoFile, result *ValidationResult) {
	if pf.Package == "" {
		addIssue(result, RulePackageRequired, "", "No package declaration found")
	}
}

func validateMessages(pf *models.ProtoFile, result *ValidationResult, syntax string, prefix string) {
	for _, msg := range pf.Messages {
		path := msg.Name
		if prefix != "" {
			path = prefix + "." + msg.Name
		}

		// Check message name
		if len(msg.Name) > 0 && msg.Name[0] >= 'a' && msg.Name[0] <= 'z' {
			addIssue(result, RuleMessageNamePascalCase, path, fmt.Sprintf("Message %s should start with uppercase", msg.Name))
		}

		// Check max fields
		fieldCount := len(msg.Fields) + len(msg.Maps)
		for _, oo := range msg.OneOfs {
			fieldCount += len(oo.Fields)
		}
		if fieldCount > 100 {
			addIssue(result, RuleMaxFieldsPerMessage, path, fmt.Sprintf("Message %s has %d fields (max 100)", msg.Name, fieldCount))
		}

		// Check duplicate field numbers
		fieldNumbers := make(map[int]string)
		for _, field := range msg.Fields {
			if prev, ok := fieldNumbers[field.Number]; ok {
				addIssue(result, RuleNoDuplicateFieldNumbers, path+"."+field.Name,
					fmt.Sprintf("Duplicate field number %d: %s and %s", field.Number, prev, field.Name))
			}
			fieldNumbers[field.Number] = field.Name
		}

		// Validate each field
		for _, field := range msg.Fields {
			validateField(field, path, syntax, result)
		}

		// Validate maps
		for _, mf := range msg.Maps {
			validateMapField(mf, path, syntax, result)
		}

		// Validate oneofs
		for _, oo := range msg.OneOfs {
			for _, field := range oo.Fields {
				validateField(field, path+"."+oo.Name, syntax, result)
			}
		}

		// Recurse into nested messages
		nestedPf := &models.ProtoFile{Messages: msg.Messages}
		validateMessages(nestedPf, result, syntax, path)
	}
}

func validateField(field *models.Field, path string, syntax string, result *ValidationResult) {
	fieldPath := path + "." + field.Name

	// Check field number validity
	if field.Number < 0 || field.Number > 536870911 {
		addIssue(result, RuleFieldNumberValid, fieldPath,
			fmt.Sprintf("Field number %d is out of range (0-536,870,911)", field.Number))
	}

	// Check reserved field numbers
	if field.Number >= 19000 && field.Number <= 19999 {
		addIssue(result, RuleFieldNumberReserved, fieldPath,
			fmt.Sprintf("Field number %d is reserved by protobuf", field.Number))
	}

	// Check field name (should be lower_snake_case)
	if field.Name != strings.ToLower(field.Name) {
		addIssue(result, RuleFieldNameCamelCase, fieldPath,
			fmt.Sprintf("Field name %s should be lower_snake_case", field.Name))
	}

	// Check proto3 syntax constraints
	if syntax == "proto3" {
		if field.Label == "required" {
			addIssue(result, RuleProto3NoRequired, fieldPath,
				"Proto3 does not support required fields")
		}
	}
}

func validateMapField(mf *models.MapField, path string, syntax string, result *ValidationResult) {
	fieldPath := path + "." + mf.Name

	// Check field number
	if mf.Number < 0 || mf.Number > 536870911 {
		addIssue(result, RuleFieldNumberValid, fieldPath,
			fmt.Sprintf("Map field number %d is out of range", mf.Number))
	}
	if mf.Number >= 19000 && mf.Number <= 19999 {
		addIssue(result, RuleFieldNumberReserved, fieldPath,
			fmt.Sprintf("Map field number %d is reserved", mf.Number))
	}

	// Check map key type
	validKeyTypes := map[string]bool{
		"int32": true, "int64": true,
		"uint32": true, "uint64": true,
		"sint32": true, "sint64": true,
		"fixed32": true, "fixed64": true,
		"sfixed32": true, "sfixed64": true,
		"bool": true,
		"string": true,
	}
	if !validKeyTypes[mf.KeyType] {
		addIssue(result, RuleMapKeyType, fieldPath,
			fmt.Sprintf("Map key type %s is not valid (must be integral or string)", mf.KeyType))
	}
}

func validateEnums(pf *models.ProtoFile, result *ValidationResult, syntax string, prefix string) {
	for _, enum := range pf.Enums {
		path := enum.Name
		if prefix != "" {
			path = prefix + "." + enum.Name
		}

		// Check enum name
		if len(enum.Name) > 0 && enum.Name[0] >= 'a' && enum.Name[0] <= 'z' {
			addIssue(result, RuleEnumNamePascalCase, path,
				fmt.Sprintf("Enum %s should start with uppercase", enum.Name))
		}

		// Check zero value
		hasZero := false
		for _, val := range enum.Values {
			if val.Number == 0 {
				hasZero = true
				break
			}
		}
		if !hasZero {
			addIssue(result, RuleEnumZeroValue, path,
				fmt.Sprintf("Enum %s has no zero value", enum.Name))
		}

		// Check enum value naming
		for _, val := range enum.Values {
			valPath := path + "." + val.Name
			if val.Name != strings.ToUpper(val.Name) && !strings.HasPrefix(val.Name, strings.ToUpper(enum.Name)+"_") {
				addIssue(result, RuleEnumValueScreamingSnake, valPath,
					fmt.Sprintf("Enum value %s should be SCREAMING_SNAKE_CASE", val.Name))
			}
		}
	}
}

func validateServices(pf *models.ProtoFile, result *ValidationResult) {
	for _, svc := range pf.Services {
		// Check service name
		if len(svc.Name) > 0 && svc.Name[0] >= 'a' && svc.Name[0] <= 'z' {
			addIssue(result, RuleServiceNamePascalCase, svc.Name,
				fmt.Sprintf("Service %s should start with uppercase", svc.Name))
		}

		// Check method names
		for _, method := range svc.Methods {
			methodPath := svc.Name + "." + method.Name
			if len(method.Name) > 0 && method.Name[0] >= 'a' && method.Name[0] <= 'z' {
				addIssue(result, RuleMethodNamePascalCase, methodPath,
					fmt.Sprintf("Method %s should start with uppercase", method.Name))
			}
		}
	}
}

func validateImports(pf *models.ProtoFile, result *ValidationResult) {
	seen := make(map[string]bool)
	for _, imp := range pf.Imports {
		if seen[imp.Path] {
			addIssue(result, RulePackageRequired, "import."+imp.Path,
				fmt.Sprintf("Duplicate import: %s", imp.Path))
		}
		seen[imp.Path] = true
	}
}

func addIssue(result *ValidationResult, rule *Rule, path, message string) {
	issue := &ValidationIssue{
		Rule:     rule.ID,
		Message:  message,
		Path:     path,
		Severity: rule.Severity,
	}

	switch rule.Severity {
	case "error":
		result.Errors = append(result.Errors, issue)
	case "warning":
		result.Warnings = append(result.Warnings, issue)
	case "info":
		result.Info = append(result.Info, issue)
	}
}

// FormatValidationResult returns a human-readable string.
func FormatValidationResult(vr *ValidationResult) string {
	var sb strings.Builder

	if vr.Valid {
		sb.WriteString("✓ Validation passed\n")
	} else {
		sb.WriteString("✗ Validation failed\n")
	}

	if len(vr.Errors) > 0 {
		fmt.Fprintf(&sb, "\n%d Error(s):\n", len(vr.Errors))
		for _, err := range vr.Errors {
			fmt.Fprintf(&sb, "  [ERROR] [%s] %s: %s\n", err.Rule, err.Path, err.Message)
		}
	}

	if len(vr.Warnings) > 0 {
		fmt.Fprintf(&sb, "\n%d Warning(s):\n", len(vr.Warnings))
		for _, w := range vr.Warnings {
			fmt.Fprintf(&sb, "  [WARN]  [%s] %s: %s\n", w.Rule, w.Path, w.Message)
		}
	}

	if len(vr.Info) > 0 {
		fmt.Fprintf(&sb, "\n%d Info:\n", len(vr.Info))
		for _, info := range vr.Info {
			fmt.Fprintf(&sb, "  [INFO]  [%s] %s: %s\n", info.Rule, info.Path, info.Message)
		}
	}

	return sb.String()
}

// ToJSON returns validation results as a map.
func ToJSON(vr *ValidationResult) map[string]interface{} {
	result := map[string]interface{}{
		"valid": vr.Valid,
	}

	if len(vr.Errors) > 0 {
		errors := make([]map[string]interface{}, 0, len(vr.Errors))
		for _, e := range vr.Errors {
			errors = append(errors, map[string]interface{}{
				"rule":    e.Rule,
				"message": e.Message,
				"path":    e.Path,
			})
		}
		result["errors"] = errors
	}

	if len(vr.Warnings) > 0 {
		warnings := make([]map[string]interface{}, 0, len(vr.Warnings))
		for _, w := range vr.Warnings {
			warnings = append(warnings, map[string]interface{}{
				"rule":    w.Rule,
				"message": w.Message,
				"path":    w.Path,
			})
		}
		result["warnings"] = warnings
	}

	if len(vr.Info) > 0 {
		info := make([]map[string]interface{}, 0, len(vr.Info))
		for _, i := range vr.Info {
			info = append(info, map[string]interface{}{
				"rule":    i.Rule,
				"message": i.Message,
				"path":    i.Path,
			})
		}
		result["info"] = info
	}

	return result
}
