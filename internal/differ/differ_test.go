package differ

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/parser"
)

func TestDiffIdenticalFiles(t *testing.T) {
	input := `syntax = "proto3";
package test;
message User {
  string name = 1;
  int32 age = 2;
}`

	p := parser.New()
	left, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Added != 0 {
		t.Errorf("Expected 0 added, got %d", diff.Summary.Added)
	}
	if diff.Summary.Removed != 0 {
		t.Errorf("Expected 0 removed, got %d", diff.Summary.Removed)
	}
	if diff.Summary.Modified != 0 {
		t.Errorf("Expected 0 modified, got %d", diff.Summary.Modified)
	}
}

func TestDiffAddedMessage(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}`

	rightInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}
message Post {
  string title = 1;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Added != 1 {
		t.Errorf("Expected 1 added, got %d", diff.Summary.Added)
	}
}

func TestDiffRemovedMessage(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}
message Post {
  string title = 1;
}`

	rightInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Removed != 1 {
		t.Errorf("Expected 1 removed, got %d", diff.Summary.Removed)
	}
	if diff.Summary.Breaking != 1 {
		t.Errorf("Expected 1 breaking change, got %d", diff.Summary.Breaking)
	}
}

func TestDiffModifiedField(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}`

	rightInput := `syntax = "proto3";
package test;
message User {
  int32 name = 1;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Breaking != 1 {
		t.Errorf("Expected 1 breaking change (type change), got %d", diff.Summary.Breaking)
	}
}

func TestDiffAddedField(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}`

	rightInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
  int32 age = 2;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Added != 1 {
		t.Errorf("Expected 1 added field, got %d", diff.Summary.Added)
	}
	if diff.Summary.Breaking != 0 {
		t.Errorf("Expected 0 breaking changes (adding field is safe), got %d", diff.Summary.Breaking)
	}
}

func TestDiffRemovedField(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
  int32 age = 2;
}`

	rightInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Removed != 1 {
		t.Errorf("Expected 1 removed field, got %d", diff.Summary.Removed)
	}
	if diff.Summary.Breaking != 1 {
		t.Errorf("Expected 1 breaking change (removing field is breaking), got %d", diff.Summary.Breaking)
	}
}

func TestDiffAddedEnumValue(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
enum Status {
  STATUS_UNSPECIFIED = 0;
  ACTIVE = 1;
}`

	rightInput := `syntax = "proto3";
package test;
enum Status {
  STATUS_UNSPECIFIED = 0;
  ACTIVE = 1;
  INACTIVE = 2;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Added != 1 {
		t.Errorf("Expected 1 added enum value, got %d", diff.Summary.Added)
	}
}

func TestDiffRemovedEnumValue(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
enum Status {
  STATUS_UNSPECIFIED = 0;
  ACTIVE = 1;
  INACTIVE = 2;
}`

	rightInput := `syntax = "proto3";
package test;
enum Status {
  STATUS_UNSPECIFIED = 0;
  ACTIVE = 1;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Removed != 1 {
		t.Errorf("Expected 1 removed enum value, got %d", diff.Summary.Removed)
	}
	if diff.Summary.Breaking != 1 {
		t.Errorf("Expected 1 breaking change (removing enum value is breaking), got %d", diff.Summary.Breaking)
	}
}

func TestDiffAddedService(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
service UserService {
  rpc GetUser (GetUserRequest) returns (User);
}`

	rightInput := `syntax = "proto3";
package test;
service UserService {
  rpc GetUser (GetUserRequest) returns (User);
}
service PostService {
  rpc GetPost (GetPostRequest) returns (Post);
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Added != 1 {
		t.Errorf("Expected 1 added service, got %d", diff.Summary.Added)
	}
}

func TestDiffAddedMethod(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
service UserService {
  rpc GetUser (GetUserRequest) returns (User);
}`

	rightInput := `syntax = "proto3";
package test;
service UserService {
  rpc GetUser (GetUserRequest) returns (User);
  rpc ListUsers (ListUsersRequest) returns (User);
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Added != 1 {
		t.Errorf("Expected 1 added method, got %d", diff.Summary.Added)
	}
}

func TestDiffChangedMethodSignature(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
service UserService {
  rpc GetUser (GetUserRequest) returns (User);
}`

	rightInput := `syntax = "proto3";
package test;
service UserService {
  rpc GetUser (GetUserRequest) returns (UserResponse);
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Breaking != 1 {
		t.Errorf("Expected 1 breaking change (output type change), got %d", diff.Summary.Breaking)
	}
}

func TestDiffSyntaxChange(t *testing.T) {
	leftInput := `syntax = "proto2";
package test;
message User {
  required string name = 1;
}`

	rightInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	// Syntax change + label change (required -> "")
	if diff.Summary.Modified < 1 {
		t.Errorf("Expected at least 1 modified, got %d", diff.Summary.Modified)
	}
}

func TestDiffPackageChange(t *testing.T) {
	leftInput := `syntax = "proto3";
package user.v1;
message User {
  string name = 1;
}`

	rightInput := `syntax = "proto3";
package user.v2;
message User {
  string name = 1;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Modified != 1 {
		t.Errorf("Expected 1 modified (package), got %d", diff.Summary.Modified)
	}
}

func TestSummaryText(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}`

	rightInput := `syntax = "proto3";
package test;
message User {
  string name = 1;
}
message Post {
  string title = 1;
}`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	text := diff.SummaryText()
	if text == "" {
		t.Error("SummaryText should not be empty")
	}
	if len(text) < 20 {
		t.Errorf("SummaryText too short: %s", text)
	}
}

func TestDiffAddedImport(t *testing.T) {
	leftInput := `syntax = "proto3";
package test;
message User { string name = 1; }`

	rightInput := `syntax = "proto3";
package test;
import "google/protobuf/timestamp.proto";
message User { string name = 1; }`

	p := parser.New()
	left, err := p.Parse(leftInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	right, err := p.Parse(rightInput)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	diff := DiffProtoFiles(left, right, "left.proto", "right.proto")

	if diff.Summary.Added != 1 {
		t.Errorf("Expected 1 added import, got %d", diff.Summary.Added)
	}
}
