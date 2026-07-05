package parser

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/models"
)

func TestParseSyntax(t *testing.T) {
	input := `syntax = "proto3";`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if pf.Syntax != "proto3" {
		t.Errorf("Expected syntax 'proto3', got %q", pf.Syntax)
	}
}

func TestParsePackage(t *testing.T) {
	input := `syntax = "proto3";
package user.v1;`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if pf.Package != "user.v1" {
		t.Errorf("Expected package 'user.v1', got %q", pf.Package)
	}
}

func TestParseImport(t *testing.T) {
	input := `syntax = "proto3";
package test;
import "google/protobuf/timestamp.proto";
import public "other.proto";`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pf.Imports) != 2 {
		t.Fatalf("Expected 2 imports, got %d", len(pf.Imports))
	}
	if pf.Imports[0].Path != "google/protobuf/timestamp.proto" {
		t.Errorf("Expected first import path, got %q", pf.Imports[0].Path)
	}
	if pf.Imports[1].Public != true {
		t.Error("Expected second import to be public")
	}
}

func TestParseMessage(t *testing.T) {
	input := `syntax = "proto3";
package test;
message User {
  string name = 1;
  int32 age = 2;
  repeated string tags = 3;
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pf.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(pf.Messages))
	}
	msg := pf.Messages[0]
	if msg.Name != "User" {
		t.Errorf("Expected message name 'User', got %q", msg.Name)
	}
	if len(msg.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(msg.Fields))
	}
	if msg.Fields[0].Name != "name" || msg.Fields[0].Type != "string" || msg.Fields[0].Number != 1 {
		t.Errorf("Field 0 mismatch: %+v", msg.Fields[0])
	}
	if msg.Fields[2].Label != "repeated" {
		t.Errorf("Expected field 2 to be 'repeated', got %q", msg.Fields[2].Label)
	}
}

func TestParseNestedMessage(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Outer {
  message Inner {
    string value = 1;
  }
  Inner inner = 1;
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pf.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(pf.Messages))
	}
	if len(pf.Messages[0].Messages) != 1 {
		t.Fatalf("Expected 1 nested message, got %d", len(pf.Messages[0].Messages))
	}
	if pf.Messages[0].Messages[0].Name != "Inner" {
		t.Errorf("Expected nested message name 'Inner', got %q", pf.Messages[0].Messages[0].Name)
	}
}

func TestParseEnum(t *testing.T) {
	input := `syntax = "proto3";
package test;
enum Status {
  STATUS_UNSPECIFIED = 0;
  ACTIVE = 1;
  INACTIVE = 2;
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pf.Enums) != 1 {
		t.Fatalf("Expected 1 enum, got %d", len(pf.Enums))
	}
	enum := pf.Enums[0]
	if enum.Name != "Status" {
		t.Errorf("Expected enum name 'Status', got %q", enum.Name)
	}
	if len(enum.Values) != 3 {
		t.Fatalf("Expected 3 enum values, got %d", len(enum.Values))
	}
	if enum.Values[0].Name != "STATUS_UNSPECIFIED" || enum.Values[0].Number != 0 {
		t.Errorf("First enum value mismatch: %+v", enum.Values[0])
	}
}

func TestParseService(t *testing.T) {
	input := `syntax = "proto3";
package test;
service UserService {
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
  rpc ListUsers (ListUsersRequest) returns (stream User);
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pf.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(pf.Services))
	}
	svc := pf.Services[0]
	if svc.Name != "UserService" {
		t.Errorf("Expected service name 'UserService', got %q", svc.Name)
	}
	if len(svc.Methods) != 2 {
		t.Fatalf("Expected 2 methods, got %d", len(svc.Methods))
	}
	if svc.Methods[0].Name != "GetUser" {
		t.Errorf("Expected method name 'GetUser', got %q", svc.Methods[0].Name)
	}
	if svc.Methods[0].InputType != "GetUserRequest" {
		t.Errorf("Expected input type 'GetUserRequest', got %q", svc.Methods[0].InputType)
	}
	if svc.Methods[0].OutputType != "GetUserResponse" {
		t.Errorf("Expected output type 'GetUserResponse', got %q", svc.Methods[0].OutputType)
	}
	if svc.Methods[1].ServerStreaming != true {
		t.Error("Expected method 1 to be server streaming")
	}
}

func TestParseMapField(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Config {
  map<string, string> settings = 1;
  map<int32, User> users = 2;
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pf.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(pf.Messages))
	}
	msg := pf.Messages[0]
	if len(msg.Maps) != 2 {
		t.Fatalf("Expected 2 map fields, got %d", len(msg.Maps))
	}
	if msg.Maps[0].KeyType != "string" || msg.Maps[0].ValueType != "string" {
		t.Errorf("Map 0 type mismatch: %s, %s", msg.Maps[0].KeyType, msg.Maps[0].ValueType)
	}
	if msg.Maps[1].KeyType != "int32" || msg.Maps[1].ValueType != "User" {
		t.Errorf("Map 1 type mismatch: %s, %s", msg.Maps[1].KeyType, msg.Maps[1].ValueType)
	}
}

func TestParseOneOf(t *testing.T) {
	input := `syntax = "proto3";
package test;
message Event {
  oneof payload {
    string text = 1;
    int32 number = 2;
    bool flag = 3;
  }
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	msg := pf.Messages[0]
	if len(msg.OneOfs) != 1 {
		t.Fatalf("Expected 1 oneof, got %d", len(msg.OneOfs))
	}
	oo := msg.OneOfs[0]
	if oo.Name != "payload" {
		t.Errorf("Expected oneof name 'payload', got %q", oo.Name)
	}
	if len(oo.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(oo.Fields))
	}
}

func TestParseOption(t *testing.T) {
	input := `syntax = "proto3";
package test;
option java_package = "com.example";
option java_multiple_files = true;
message User {
  option map_entry = true;
  string name = 1;
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pf.Options) != 2 {
		t.Fatalf("Expected 2 file options, got %d", len(pf.Options))
	}
	if pf.Options[0].Name != "java_package" {
		t.Errorf("Expected option name 'java_package', got %q", pf.Options[0].Name)
	}
	// parseString returns the content without quotes
	if pf.Options[0].Value != "com.example" {
		t.Errorf("Expected option value 'com.example', got %q", pf.Options[0].Value)
	}
}

func TestParseComplexProto(t *testing.T) {
	input := `syntax = "proto3";

package user.v1;

import "google/protobuf/timestamp.proto";

option java_package = "com.example.user";

message User {
  int64 id = 1;
  string name = 2;
  string email = 3;
  UserProfile profile = 4;
  repeated string tags = 5;
  map<string, string> metadata = 6;
  google.protobuf.Timestamp created_at = 7;

  message UserProfile {
    string bio = 1;
    string avatar_url = 2;
  }
}

enum UserRole {
  ROLE_UNSPECIFIED = 0;
  ROLE_ADMIN = 1;
  ROLE_USER = 2;
  ROLE_GUEST = 3;
}

service UserService {
  rpc GetUser (GetUserRequest) returns (User);
  rpc ListUsers (ListUsersRequest) returns (stream User);
  rpc UpdateUser (UpdateUserRequest) returns (User);
}

message GetUserRequest {
  int64 id = 1;
}

message ListUsersRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message UpdateUserRequest {
  User user = 1;
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Verify all components
	if pf.Syntax != "proto3" {
		t.Errorf("Expected syntax 'proto3', got %q", pf.Syntax)
	}
	if pf.Package != "user.v1" {
		t.Errorf("Expected package 'user.v1', got %q", pf.Package)
	}
	if len(pf.Imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(pf.Imports))
	}
	if len(pf.Messages) != 4 {
		t.Errorf("Expected 4 messages, got %d", len(pf.Messages))
	}
	if len(pf.Enums) != 1 {
		t.Errorf("Expected 1 enum, got %d", len(pf.Enums))
	}
	if len(pf.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(pf.Services))
	}
	if len(pf.Services[0].Methods) != 3 {
		t.Errorf("Expected 3 methods, got %d", len(pf.Services[0].Methods))
	}
}

func TestParseWithComments(t *testing.T) {
	input := `syntax = "proto3";
// This is a comment
package test;
/* Block
   comment */
message User {
  string name = 1; // inline comment
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if pf.Package != "test" {
		t.Errorf("Expected package 'test', got %q", pf.Package)
	}
	if len(pf.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(pf.Messages))
	}
}

func TestParseEmptyFile(t *testing.T) {
	input := ``
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if pf == nil {
		t.Fatal("Expected non-nil proto file")
	}
}

func TestParseSyntaxProto2(t *testing.T) {
	input := `syntax = "proto2";
package test;
message User {
  required string name = 1;
  optional int32 age = 2;
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if pf.Syntax != "proto2" {
		t.Errorf("Expected syntax 'proto2', got %q", pf.Syntax)
	}
	if len(pf.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(pf.Messages))
	}
	if pf.Messages[0].Fields[0].Label != "required" {
		t.Errorf("Expected label 'required', got %q", pf.Messages[0].Fields[0].Label)
	}
	if pf.Messages[0].Fields[1].Label != "optional" {
		t.Errorf("Expected label 'optional', got %q", pf.Messages[0].Fields[1].Label)
	}
}

func TestParseNegativeEnumValue(t *testing.T) {
	input := `syntax = "proto3";
package test;
enum ErrorCode {
  OK = 0;
  NOT_FOUND = -1;
  INTERNAL = -2;
}`
	p := New()
	pf, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if pf.Enums[0].Values[1].Number != -1 {
		t.Errorf("Expected enum value -1, got %d", pf.Enums[0].Values[1].Number)
	}
}

func TestParseMapTypeParsing(t *testing.T) {
	tests := []struct {
		input   string
		key     string
		val     string
		wantOk  bool
	}{
		{"map<string, int32>", "string", "int32", true},
		{"map<int64, User>", "int64", "User", true},
		{"map<string, map<string, int32>>", "string", "map<string, int32>", true},
		{"not_a_map", "", "", false},
	}

	for _, tt := range tests {
		key, val, ok := models.ParseMapType(tt.input)
		if ok != tt.wantOk {
			t.Errorf("ParseMapType(%q): ok = %v, want %v", tt.input, ok, tt.wantOk)
		}
		if ok && key != tt.key {
			t.Errorf("ParseMapType(%q): key = %q, want %q", tt.input, key, tt.key)
		}
		if ok && val != tt.val {
			t.Errorf("ParseMapType(%q): val = %q, want %q", tt.input, val, tt.val)
		}
	}
}
