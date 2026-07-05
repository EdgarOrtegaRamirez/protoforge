// Package differ provides semantic diffing for Protocol Buffer schemas.
package differ

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/models"
)

// ChangeType represents the type of change.
type ChangeType int

const (
	ChangeAdded ChangeType = iota
	ChangeRemoved
	ChangeModified
	ChangeUnchanged
)

func (c ChangeType) String() string {
	switch c {
	case ChangeAdded:
		return "ADDED"
	case ChangeRemoved:
		return "REMOVED"
	case ChangeModified:
		return "MODIFIED"
	case ChangeUnchanged:
		return "UNCHANGED"
	default:
		return "UNKNOWN"
	}
}

// Diff represents a comparison between two proto files.
type Diff struct {
	LeftName  string
	RightName string
	Changes   []*Change
	Summary   *DiffSummary
}

// Change represents a single change between two proto files.
type Change struct {
	Type    ChangeType
	Path    string
	Message string
	Left    interface{} // The left value (nil for additions)
	Right   interface{} // The right value (nil for removals)
}

// DiffSummary holds summary statistics about the diff.
type DiffSummary struct {
	Added    int
	Removed  int
	Modified int
	Unchanged int
	Breaking int
}

// DiffProtoFiles compares two parsed proto files.
func DiffProtoFiles(left, right *models.ProtoFile, leftName, rightName string) *Diff {
	d := &Diff{
		LeftName:  leftName,
		RightName: rightName,
		Summary:   &DiffSummary{},
	}

	// Compare syntax
	d.compareSyntax(left, right)

	// Compare package
	d.comparePackage(left, right)

	// Compare options
	d.compareOptions(left.Options, right.Options, "option")

	// Compare imports
	d.compareImports(left.Imports, right.Imports)

	// Compare messages
	d.compareMessages(left.Messages, right.Messages, "")

	// Compare enums
	d.compareEnums(left.Enums, right.Enums, "")

	// Compare services
	d.compareServices(left.Services, right.Services)

	// Compute summary
	for _, change := range d.Changes {
		switch change.Type {
		case ChangeAdded:
			d.Summary.Added++
		case ChangeRemoved:
			d.Summary.Removed++
		case ChangeModified:
			d.Summary.Modified++
		case ChangeUnchanged:
			d.Summary.Unchanged++
		}
	}

	return d
}

func (d *Diff) compareSyntax(left, right *models.ProtoFile) {
	if left.Syntax != right.Syntax {
		d.Changes = append(d.Changes, &Change{
			Type:    ChangeModified,
			Path:    "syntax",
			Message: fmt.Sprintf("Syntax changed from %q to %q", left.Syntax, right.Syntax),
			Left:    left.Syntax,
			Right:   right.Syntax,
		})
	}
}

func (d *Diff) comparePackage(left, right *models.ProtoFile) {
	if left.Package != right.Package {
		d.Changes = append(d.Changes, &Change{
			Type:    ChangeModified,
			Path:    "package",
			Message: fmt.Sprintf("Package changed from %q to %q", left.Package, right.Package),
			Left:    left.Package,
			Right:   right.Package,
		})
	}
}

func (d *Diff) compareOptions(leftOpts, rightOpts []*models.Option, prefix string) {
	leftMap := make(map[string]*models.Option)
	rightMap := make(map[string]*models.Option)

	for _, opt := range leftOpts {
		leftMap[opt.Name] = opt
	}
	for _, opt := range rightOpts {
		rightMap[opt.Name] = opt
	}

	// Check for removed/modified options
	for name, leftOpt := range leftMap {
		path := prefix + "." + name
		if rightOpt, ok := rightMap[name]; ok {
			if leftOpt.Value != rightOpt.Value {
				d.Changes = append(d.Changes, &Change{
					Type:    ChangeModified,
					Path:    path,
					Message: fmt.Sprintf("Option %s changed from %q to %q", name, leftOpt.Value, rightOpt.Value),
					Left:    leftOpt.Value,
					Right:   rightOpt.Value,
				})
			} else {
				d.Changes = append(d.Changes, &Change{
					Type:    ChangeUnchanged,
					Path:    path,
					Message: fmt.Sprintf("Option %s unchanged", name),
				})
			}
		} else {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    path,
				Message: fmt.Sprintf("Option %s removed (was %q)", name, leftOpt.Value),
				Left:    leftOpt.Value,
			})
		}
	}

	// Check for added options
	for name, rightOpt := range rightMap {
		path := prefix + "." + name
		if _, ok := leftMap[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    path,
				Message: fmt.Sprintf("Option %s added with value %q", name, rightOpt.Value),
				Right:   rightOpt.Value,
			})
		}
	}
}

func (d *Diff) compareImports(leftImports, rightImports []*models.Import) {
	leftMap := make(map[string]*models.Import)
	rightMap := make(map[string]*models.Import)

	for _, imp := range leftImports {
		leftMap[imp.Path] = imp
	}
	for _, imp := range rightImports {
		rightMap[imp.Path] = imp
	}

	// Removed imports
	for path, leftImp := range leftMap {
		if _, ok := rightMap[path]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    "import." + path,
				Message: fmt.Sprintf("Import %q removed", path),
				Left:    leftImp.Path,
			})
		}
	}

	// Added imports
	for path, rightImp := range rightMap {
		if _, ok := leftMap[path]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    "import." + path,
				Message: fmt.Sprintf("Import %q added", path),
				Right:   rightImp.Path,
			})
		}
	}
}

func (d *Diff) compareMessages(leftMsgs, rightMsgs []*models.Message, prefix string) {
	leftMap := make(map[string]*models.Message)
	rightMap := make(map[string]*models.Message)

	for _, msg := range leftMsgs {
		leftMap[msg.Name] = msg
	}
	for _, msg := range rightMsgs {
		rightMap[msg.Name] = msg
	}

	// Check removed messages
	for name := range leftMap {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		if _, ok := rightMap[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    "message." + path,
				Message: fmt.Sprintf("Message %s removed", name),
			})
			d.Summary.Breaking++
		}
	}

	// Check added/modified messages
	for name, rightMsg := range rightMap {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		if leftMsg, ok := leftMap[name]; ok {
			d.compareMessageFields(leftMsg, rightMsg, "message."+path)
			d.compareMessages(leftMsg.Messages, rightMsg.Messages, path)
			d.compareEnums(leftMsg.Enums, rightMsg.Enums, path)
		} else {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    "message." + path,
				Message: fmt.Sprintf("Message %s added", name),
			})
		}
	}
}

func (d *Diff) compareMessageFields(left, right *models.Message, prefix string) {
	leftFields := make(map[string]*models.Field)
	rightFields := make(map[string]*models.Field)

	for _, field := range left.Fields {
		leftFields[field.Name] = field
	}
	for _, field := range right.Fields {
		rightFields[field.Name] = field
	}

	// Removed fields
	for name, leftField := range leftFields {
		if _, ok := rightFields[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    prefix + "." + name,
				Message: fmt.Sprintf("Field %s removed (was type %s, number %d)", name, leftField.Type, leftField.Number),
			})
			d.Summary.Breaking++
		}
	}

	// Added fields
	for name, rightField := range rightFields {
		if _, ok := leftFields[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    prefix + "." + name,
				Message: fmt.Sprintf("Field %s added (type %s, number %d)", name, rightField.Type, rightField.Number),
			})
		}
	}

	// Modified fields
	for name, leftField := range leftFields {
		if rightField, ok := rightFields[name]; ok {
			path := prefix + "." + name
			if leftField.Type != rightField.Type {
				d.Changes = append(d.Changes, &Change{
					Type:    ChangeModified,
					Path:    path,
					Message: fmt.Sprintf("Field %s type changed from %s to %s", name, leftField.Type, rightField.Type),
					Left:    leftField.Type,
					Right:   rightField.Type,
				})
				d.Summary.Breaking++
			}
			if leftField.Number != rightField.Number {
				d.Changes = append(d.Changes, &Change{
					Type:    ChangeModified,
					Path:    path,
					Message: fmt.Sprintf("Field %s number changed from %d to %d", name, leftField.Number, rightField.Number),
					Left:    fmt.Sprintf("%d", leftField.Number),
					Right:   fmt.Sprintf("%d", rightField.Number),
				})
				d.Summary.Breaking++
			}
			if leftField.Label != rightField.Label {
				d.Changes = append(d.Changes, &Change{
					Type:    ChangeModified,
					Path:    path,
					Message: fmt.Sprintf("Field %s label changed from %q to %q", name, leftField.Label, rightField.Label),
					Left:    leftField.Label,
					Right:   rightField.Label,
				})
			}
		}
	}

	// Compare oneofs
	d.compareOneOfs(left.OneOfs, right.OneOfs, prefix)
}

func (d *Diff) compareOneOfs(leftOneOfs, rightOneOfs []*models.OneOf, prefix string) {
	leftMap := make(map[string]*models.OneOf)
	rightMap := make(map[string]*models.OneOf)

	for _, oo := range leftOneOfs {
		leftMap[oo.Name] = oo
	}
	for _, oo := range rightOneOfs {
		rightMap[oo.Name] = oo
	}

	for name := range leftMap {
		if _, ok := rightMap[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    prefix + ".oneof." + name,
				Message: fmt.Sprintf("OneOf %s removed", name),
			})
		}
	}

	for name := range rightMap {
		if _, ok := leftMap[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    prefix + ".oneof." + name,
				Message: fmt.Sprintf("OneOf %s added", name),
			})
		}
	}
}

func (d *Diff) compareEnums(leftEnums, rightEnums []*models.Enum, prefix string) {
	leftMap := make(map[string]*models.Enum)
	rightMap := make(map[string]*models.Enum)

	for _, enum := range leftEnums {
		leftMap[enum.Name] = enum
	}
	for _, enum := range rightEnums {
		rightMap[enum.Name] = enum
	}

	// Removed enums
	for name := range leftMap {
		if _, ok := rightMap[name]; !ok {
			path := name
			if prefix != "" {
				path = prefix + "." + name
			}
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    "enum." + path,
				Message: fmt.Sprintf("Enum %s removed", name),
			})
			d.Summary.Breaking++
		}
	}

	// Added/modified enums
	for name, rightEnum := range rightMap {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		if leftEnum, ok := leftMap[name]; ok {
			d.compareEnumValues(leftEnum, rightEnum, "enum."+path)
		} else {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    "enum." + path,
				Message: fmt.Sprintf("Enum %s added", name),
			})
		}
	}
}

func (d *Diff) compareEnumValues(left, right *models.Enum, prefix string) {
	leftValues := make(map[string]*models.EnumValue)
	rightValues := make(map[string]*models.EnumValue)

	for _, val := range left.Values {
		leftValues[val.Name] = val
	}
	for _, val := range right.Values {
		rightValues[val.Name] = val
	}

	// Removed values
	for name, leftVal := range leftValues {
		if _, ok := rightValues[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    prefix + "." + name,
				Message: fmt.Sprintf("Enum value %s (number %d) removed", name, leftVal.Number),
			})
			d.Summary.Breaking++
		}
	}

	// Added values
	for name, rightVal := range rightValues {
		if _, ok := leftValues[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    prefix + "." + name,
				Message: fmt.Sprintf("Enum value %s (number %d) added", name, rightVal.Number),
			})
		}
	}
}

func (d *Diff) compareServices(leftServices, rightServices []*models.Service) {
	leftMap := make(map[string]*models.Service)
	rightMap := make(map[string]*models.Service)

	for _, svc := range leftServices {
		leftMap[svc.Name] = svc
	}
	for _, svc := range rightServices {
		rightMap[svc.Name] = svc
	}

	// Removed services
	for name := range leftMap {
		if _, ok := rightMap[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    "service." + name,
				Message: fmt.Sprintf("Service %s removed", name),
			})
			d.Summary.Breaking++
		}
	}

	// Added/modified services
	for name, rightSvc := range rightMap {
		if leftSvc, ok := leftMap[name]; ok {
			d.compareServiceMethods(leftSvc, rightSvc, "service."+name)
		} else {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    "service." + name,
				Message: fmt.Sprintf("Service %s added", name),
			})
		}
	}
}

func (d *Diff) compareServiceMethods(left, right *models.Service, prefix string) {
	leftMethods := make(map[string]*models.Method)
	rightMethods := make(map[string]*models.Method)

	for _, method := range left.Methods {
		leftMethods[method.Name] = method
	}
	for _, method := range right.Methods {
		rightMethods[method.Name] = method
	}

	// Removed methods
	for name := range leftMethods {
		if _, ok := rightMethods[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeRemoved,
				Path:    prefix + "." + name,
				Message: fmt.Sprintf("Method %s removed", name),
			})
			d.Summary.Breaking++
		}
	}

	// Added methods
	for name := range rightMethods {
		if _, ok := leftMethods[name]; !ok {
			d.Changes = append(d.Changes, &Change{
				Type:    ChangeAdded,
				Path:    prefix + "." + name,
				Message: fmt.Sprintf("Method %s added", name),
			})
		}
	}

	// Modified methods
	for name, leftMethod := range leftMethods {
		if rightMethod, ok := rightMethods[name]; ok {
			path := prefix + "." + name
			if leftMethod.InputType != rightMethod.InputType {
				d.Changes = append(d.Changes, &Change{
					Type:    ChangeModified,
					Path:    path + ".input",
					Message: fmt.Sprintf("Method %s input type changed from %s to %s", name, leftMethod.InputType, rightMethod.InputType),
					Left:    leftMethod.InputType,
					Right:   rightMethod.InputType,
				})
				d.Summary.Breaking++
			}
			if leftMethod.OutputType != rightMethod.OutputType {
				d.Changes = append(d.Changes, &Change{
					Type:    ChangeModified,
					Path:    path + ".output",
					Message: fmt.Sprintf("Method %s output type changed from %s to %s", name, leftMethod.OutputType, rightMethod.OutputType),
					Left:    leftMethod.OutputType,
					Right:   rightMethod.OutputType,
				})
				d.Summary.Breaking++
			}
		}
	}
}

// Summary returns a human-readable diff summary.
func (d *Diff) SummaryText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Comparing: %s ↔ %s\n", d.LeftName, d.RightName))
	sb.WriteString("\n--- Summary ---\n")
	sb.WriteString(fmt.Sprintf("Added:     %d\n", d.Summary.Added))
	sb.WriteString(fmt.Sprintf("Removed:   %d\n", d.Summary.Removed))
	sb.WriteString(fmt.Sprintf("Modified:  %d\n", d.Summary.Modified))
	sb.WriteString(fmt.Sprintf("Unchanged: %d\n", d.Summary.Unchanged))
	sb.WriteString(fmt.Sprintf("Breaking:  %d\n", d.Summary.Breaking))

	if len(d.Changes) > 0 {
		sb.WriteString("\n--- Changes ---\n")
		for _, change := range d.Changes {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", change.Type, change.Path, change.Message))
		}
	}

	return sb.String()
}
