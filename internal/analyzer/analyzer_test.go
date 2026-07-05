package analyzer

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/parser"
)

func TestAnalyzeBasic(t *testing.T) {
	input := `syntax = "proto3";
package user.v1;
import "google/protobuf/timestamp.proto";
message User {
  int64 id = 1;
  string name = 2;
  string email = 3;
}
enum Status {
  STATUS_UNSPECIFIED = 0;
  ACTIVE = 1;
  INACTIVE = 2;
}
service UserService {
  rpc GetUser (GetUserRequest) returns (User);
  rpc ListUsers (ListUsersRequest) returns (stream User);
}
message GetUserRequest { int64 id = 1; }
message ListUsersRequest { int32 page_size = 1; }`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := Analyze(pf)

	if a.Package != "user.v1" {
		t.Errorf("Expected package 'user.v1', got %q", a.Package)
	}
	if a.Stats.TotalMessages != 3 {
		t.Errorf("Expected 3 messages, got %d", a.Stats.TotalMessages)
	}
	if a.Stats.TotalEnums != 1 {
		t.Errorf("Expected 1 enum, got %d", a.Stats.TotalEnums)
	}
	if a.Stats.TotalServices != 1 {
		t.Errorf("Expected 1 service, got %d", a.Stats.TotalServices)
	}
	if a.Stats.TotalMethods != 2 {
		t.Errorf("Expected 2 methods, got %d", a.Stats.TotalMethods)
	}
	if a.Stats.TotalFields != 5 {
		t.Errorf("Expected 5 fields, got %d", a.Stats.TotalFields)
	}
	if a.Stats.TotalImports != 1 {
		t.Errorf("Expected 1 import, got %d", a.Stats.TotalImports)
	}
}

func TestAnalyzeNestedMessages(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Outer {
  message Inner {
    message Deep {
      string value = 1;
    }
    Deep deep = 1;
  }
  Inner inner = 1;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := Analyze(pf)

	if a.Stats.TotalMessages != 3 {
		t.Errorf("Expected 3 messages (outer + inner + deep), got %d", a.Stats.TotalMessages)
	}
	if a.Stats.MaxNesting != 2 {
		t.Errorf("Expected max nesting 2, got %d", a.Stats.MaxNesting)
	}
}

func TestAnalyzeIssues(t *testing.T) {
	// Missing syntax, missing package
	input := `message User {
  string name = 1;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := Analyze(pf)

	if len(a.Issues) == 0 {
		t.Error("Expected issues for missing syntax and package")
	}

	foundSyntax := false
	foundPackage := false
	for _, issue := range a.Issues {
		if issue.Rule == "missing-syntax" {
			foundSyntax = true
		}
		if issue.Rule == "missing-package" {
			foundPackage = true
		}
	}
	if !foundSyntax {
		t.Error("Expected 'missing-syntax' issue")
	}
	if !foundPackage {
		t.Error("Expected 'missing-package' issue")
	}
}

func TestAnalyzeEmptyMessage(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Empty {}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := Analyze(pf)

	foundEmpty := false
	for _, w := range a.Warnings {
		if w.Rule == "empty-message" {
			foundEmpty = true
		}
	}
	if !foundEmpty {
		t.Error("Expected 'empty-message' warning")
	}
}

func TestAnalyzeReservedFieldNumber(t *testing.T) {
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

	a := Analyze(pf)

	foundReserved := false
	for _, issue := range a.Issues {
		if issue.Rule == "reserved-field-number" {
			foundReserved = true
		}
	}
	if !foundReserved {
		t.Error("Expected 'reserved-field-number' issue")
	}
}

func TestAnalyzeNamingConvention(t *testing.T) {
	input := `syntax = "proto3";
package test;
message bad_name {
  string name = 1;
}
enum bad_enum {
  VALUE = 0;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := Analyze(pf)

	foundMsgName := false
	for _, w := range a.Warnings {
		if w.Rule == "naming-message" {
			foundMsgName = true
		}
	}
	if !foundMsgName {
		t.Error("Expected 'naming-message' warning for lowercase message name")
	}
}

func TestAnalyzeDuplicateEnumValues(t *testing.T) {
	input := `syntax = "proto3";
package test;
enum Bad {
  A = 0;
  B = 1;
  C = 1;
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := Analyze(pf)

	foundDup := false
	for _, issue := range a.Issues {
		if issue.Rule == "enum-duplicate-value" {
			foundDup = true
		}
	}
	if !foundDup {
		t.Error("Expected 'enum-duplicate-value' issue")
	}
}

func TestSortedMessages(t *testing.T) {
	a := &Analysis{
		Messages: []*MessageInfo{
			{Name: "Charlie"},
			{Name: "Alice"},
			{Name: "Bob"},
		},
	}

	sorted := a.SortedMessages()
	if sorted[0].Name != "Alice" || sorted[1].Name != "Bob" || sorted[2].Name != "Charlie" {
		t.Error("Messages not sorted correctly")
	}
}

func TestSummary(t *testing.T) {
	a := &Analysis{
		Package: "test.v1",
		Syntax:  "proto3",
		Stats: &Stats{
			TotalMessages: 2,
			TotalEnums:    1,
			TotalServices: 1,
			TotalMethods:  3,
			TotalFields:   5,
		},
		Issues: []*Issue{
			{Rule: "test-rule", Message: "test issue"},
		},
		Warnings: []*Issue{
			{Rule: "test-warn", Message: "test warning"},
		},
	}

	summary := a.Summary()
	if summary == "" {
		t.Error("Summary should not be empty")
	}
	if len(summary) < 50 {
		t.Errorf("Summary too short: %s", summary)
	}
}

func TestAnalyzeMapFields(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Config {
  map<string, string> settings = 1;
  map<int32, User> users = 2;
}
message User { string name = 1; }`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := Analyze(pf)

	if a.Stats.TotalFields != 3 {
		t.Errorf("Expected 3 fields (2 maps + 1 regular), got %d", a.Stats.TotalFields)
	}
}

func TestAnalyzeOneOfs(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Event {
  oneof payload {
    string text = 1;
    int32 number = 2;
  }
}`

	p := parser.New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := Analyze(pf)

	if a.Stats.TotalOneOfs != 1 {
		t.Errorf("Expected 1 oneof, got %d", a.Stats.TotalOneOfs)
	}
	if a.Stats.TotalFields != 2 {
		t.Errorf("Expected 2 fields in oneof, got %d", a.Stats.TotalFields)
	}
}

func TestJSON(t *testing.T) {
	a := &Analysis{
		Package: "test",
		Syntax:  "proto3",
		Stats: &Stats{
			TotalMessages: 1,
		},
	}

	data := a.JSON()
	if data == nil {
		t.Fatal("JSON should not be nil")
	}
	if data["package"] != "test" {
		t.Errorf("Expected package 'test', got %v", data["package"])
	}
}
