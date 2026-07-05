package validator

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/parser"
)

func TestValidateValidProto(t *testing.T) {
	input := `syntax = "proto3";
package test;
message User {
  string name = 1;
  int32 age = 2;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)
	if !vr.Valid {
		t.Errorf("Expected valid proto, got errors: %v", vr.Errors)
	}
}

func TestValidateMissingSyntax(t *testing.T) {
	input := `package test;
message User { string name = 1; }`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)
	// Missing syntax is a warning, not an error
	found := false
	for _, w := range vr.Warnings {
		if w.Rule == "syntax-required" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'syntax-required' warning")
	}
}

func TestValidateReservedFieldNumber(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Bad {
  string name = 19000;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)
	if vr.Valid {
		t.Error("Expected invalid proto due to reserved field number")
	}

	found := false
	for _, e := range vr.Errors {
		if e.Rule == "field-number-reserved" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'field-number-reserved' error")
	}
}

func TestValidateFieldName(t *testing.T) {
	input := `syntax = "proto3";
package test;
message User {
  string camelCase = 1;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)

	found := false
	for _, w := range vr.Warnings {
		if w.Rule == "field-name-camelcase" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'field-name-camelcase' warning")
	}
}

func TestValidateMapKeyType(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Config {
  map<double, string> bad_map = 1;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)
	if vr.Valid {
		t.Error("Expected invalid proto due to invalid map key type")
	}

	found := false
	for _, e := range vr.Errors {
		if e.Rule == "map-key-type" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'map-key-type' error")
	}
}

func TestValidateProto3NoRequired(t *testing.T) {
	input := `syntax = "proto3";
package test;
message User {
  required string name = 1;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)
	if vr.Valid {
		t.Error("Expected invalid proto due to required field in proto3")
	}

	found := false
	for _, e := range vr.Errors {
		if e.Rule == "proto3-no-required" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'proto3-no-required' error")
	}
}

func TestValidateEnumZeroValue(t *testing.T) {
	input := `syntax = "proto3";
package test;
enum Bad {
  VALUE_A = 1;
  VALUE_B = 2;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)

	found := false
	for _, w := range vr.Warnings {
		if w.Rule == "enum-zero-value" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'enum-zero-value' warning")
	}
}

func TestValidateEnumValueNaming(t *testing.T) {
	input := `syntax = "proto3";
package test;
enum Status {
  unspecified = 0;
  active = 1;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)

	found := false
	for _, w := range vr.Warnings {
		if w.Rule == "enum-value-screaming" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'enum-value-screaming' warning")
	}
}

func TestValidateMessageName(t *testing.T) {
	input := `syntax = "proto3";
package test;
message bad_name {
  string name = 1;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	vr := Validate(pf)

	found := false
	for _, w := range vr.Warnings {
		if w.Rule == "message-name-pascalcase" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'message-name-pascalcase' warning")
	}
}

func TestFormatValidationResult(t *testing.T) {
	vr := &ValidationResult{
		Valid: false,
		Errors: []*ValidationIssue{
			{Rule: "test-rule", Message: "test error", Path: "test.path", Severity: "error"},
		},
		Warnings: []*ValidationIssue{
			{Rule: "test-warn", Message: "test warning", Path: "test.path", Severity: "warning"},
		},
	}

	result := FormatValidationResult(vr)
	if result == "" {
		t.Error("FormatValidationResult should not be empty")
	}
	if len(result) < 20 {
		t.Errorf("Result too short: %s", result)
	}
}

func TestToJSON(t *testing.T) {
	vr := &ValidationResult{
		Valid: true,
	}

	data := ToJSON(vr)
	if data == nil {
		t.Fatal("ToJSON should not return nil")
	}
	if data["valid"] != true {
		t.Error("Expected valid to be true")
	}
}

func TestAllRules(t *testing.T) {
	rules := AllRules()
	if len(rules) < 10 {
		t.Errorf("Expected at least 10 rules, got %d", len(rules))
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, rule := range rules {
		if seen[rule.ID] {
			t.Errorf("Duplicate rule ID: %s", rule.ID)
		}
		seen[rule.ID] = true
	}
}
