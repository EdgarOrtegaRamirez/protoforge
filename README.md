# ProtoForge

A comprehensive Protocol Buffer schema toolkit for parsing, validating, analyzing, and diffing `.proto` files.

## Features

- **Parse** - Parse `.proto` files into a structured AST (Abstract Syntax Tree)
- **Analyze** - Compute statistics, detect issues, and analyze schema structure
- **Validate** - Check schemas against best practices and common pitfalls
- **Diff** - Compare two `.proto` files and detect breaking changes
- **Format** - Pretty-print and format `.proto` files with configurable options

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/protoforge/cmd/protoforge@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/protoforge
cd protoforge
go build -o protoforge ./cmd/protoforge/
```

## Usage

### Parse a .proto file

```bash
# Display parsed AST
protoforge parse user.proto

# Output as JSON
protoforge parse user.proto --format json
```

### Analyze a .proto file

```bash
# Display statistics and analysis
protoforge analyze user.proto

# Output as JSON
protoforge analyze user.proto --format json
```

### Validate a .proto file

```bash
# Validate against best practices
protoforge validate user.proto

# Output as JSON
protoforge validate user.proto --format json
```

### Diff two .proto files

```bash
# Compare two proto files
protoforge diff old.proto new.proto

# Output as JSON
protoforge diff old.proto new.proto --format json
```

### Format a .proto file

```bash
# Pretty-print a proto file
protoforge format user.proto

# Sort fields by number
protoforge format user.proto --sort-fields

# Custom indentation
protoforge format user.proto --indent "    "
```

## Validation Rules

ProtoForge checks for:

- **Syntax declaration** - Proto files should declare a syntax version
- **Package declaration** - Proto files should have a package
- **Field numbers** - Valid range (1-536,870,911), no reserved numbers (19000-19999)
- **Field naming** - Should use lower_snake_case
- **Message naming** - Should use PascalCase
- **Enum values** - Should be SCREAMING_SNAKE_CASE with zero value
- **Service naming** - Should use PascalCase
- **Method naming** - Should use PascalCase
- **Map key types** - Must be integral types or strings
- **Proto3 constraints** - No required fields, no groups
- **Duplicate field numbers** - Detected and flagged
- **Nesting depth** - Warns on deep nesting (>5 levels)
- **Field count** - Warns on messages with >100 fields

## Breaking Change Detection

The diff command detects breaking changes:

- **Removed messages** - Breaking
- **Removed fields** - Breaking
- **Removed enum values** - Breaking
- **Removed services** - Breaking
- **Removed methods** - Breaking
- **Type changes** - Breaking
- **Field number changes** - Breaking
- **Method signature changes** - Breaking
- **Added fields** - Safe (non-breaking)
- **Added messages** - Safe (non-breaking)
- **Added enum values** - Safe (non-breaking)

## Architecture

```
protoforge/
├── cmd/protoforge/      # CLI entry point
├── internal/
│   ├── models/          # AST node definitions
│   ├── parser/          # Recursive descent parser
│   ├── analyzer/        # Schema analysis engine
│   ├── differ/          # Semantic diff engine
│   ├── validator/       # Validation rules engine
│   ├── formatter/       # Code formatter
│   └── output/          # Output format handlers
├── testdata/            # Sample .proto files for testing
└── tests/               # Integration tests
```

## Example Output

### Analysis

```
Package: user.v1
Syntax:  proto3

--- Statistics ---
Messages:  5
Enums:     1
Services:  1
Methods:   3
Fields:    13
OneOfs:    0
Imports:   1
Options:   1
Avg Fields/Msg: 2.6
Max Nesting:    1
```

### Validation

```
✓ Validation passed

3 Warning(s):
  [WARN]  [enum-zero-value] Status: Enum Status has no zero value
  [WARN]  [naming-message] bad_name: Message bad_name should start with uppercase
  [WARN]  [field-name-camelcase] User.camelCase: Field name camelCase should be lower_snake_case
```

### Diff

```
Comparing: old.proto ↔ new.proto

--- Summary ---
Added:     3
Removed:   0
Modified:  1
Unchanged: 0
Breaking:  0

--- Changes ---
[ADDED] message.Post: Message Post added
[ADDED] message.Post.title: Field title added (type string, number 1)
[ADDED] message.Post.content: Field content added (type string, number 2)
[MODIFIED] option.java_package: Option java_package changed from "com.old" to "com.new"
```

## License

MIT License - see [LICENSE](LICENSE) for details.
