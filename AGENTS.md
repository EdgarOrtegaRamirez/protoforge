# AGENTS.md

This file provides context for AI agents working on this codebase.

## Project Overview

ProtoForge is a Go CLI tool and library for processing Protocol Buffer schema files. It provides parsing, analysis, validation, diffing, and formatting capabilities.

## Architecture

- **Parser** (`internal/parser/`) - Recursive descent parser that converts `.proto` text into an AST
- **Models** (`internal/models/`) - AST node definitions (ProtoFile, Message, Field, Enum, Service, etc.)
- **Analyzer** (`internal/analyzer/`) - Analyzes parsed AST for statistics, issues, and warnings
- **Validator** (`internal/validator/`) - Validates schemas against best practices
- **Differ** (`internal/differ/`) - Compares two proto files and detects breaking changes
- **Formatter** (`internal/formatter/`) - Formats proto files with configurable options
- **Output** (`internal/output/`) - Output format handlers (text, JSON, compact)

## Key Design Decisions

1. **AST-based parsing** - All analysis uses the AST, not regex or string matching
2. **Zero external dependencies** - Only Go stdlib used
3. **Error accumulation** - Parser collects multiple errors before reporting
4. **Breaking change detection** - Differ categorizes changes as breaking vs safe

## Testing

Run all tests:
```bash
go test ./... -v
```

Run specific package:
```bash
go test ./internal/parser/ -v
go test ./internal/analyzer/ -v
go test ./internal/differ/ -v
go test ./internal/validator/ -v
```

## Common Tasks

### Adding a new validation rule

1. Add rule definition to `internal/validator/validator.go`
2. Add validation logic to the appropriate `validate*` function
3. Add tests in `internal/validator/validator_test.go`

### Adding a new diff check

1. Add comparison logic to `internal/differ/differ.go`
2. Add tests in `internal/differ/differ_test.go`

### Adding a new parser feature

1. Add AST node type to `internal/models/proto.go`
2. Add parsing logic to `internal/parser/parser.go`
3. Add tests in `internal/parser/parser_test.go`

## CLI Commands

- `protoforge parse <file>` - Parse and display AST
- `protoforge analyze <file>` - Analyze schema
- `protoforge validate <file>` - Validate schema
- `protoforge diff <left> <right>` - Compare two schemas
- `protoforge format <file>` - Format schema

## Flags

- `--format json|text|compact` - Output format
- `--sort-fields` - Sort fields by number (format command)
- `--indent <string>` - Custom indentation (format command)
