// ProtoForge: Protocol Buffer Schema Toolkit
//
// A comprehensive toolkit for parsing, validating, analyzing, and diffing
// Protocol Buffer (.proto) schema files.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/analyzer"
	"github.com/EdgarOrtegaRamirez/protoforge/internal/differ"
	"github.com/EdgarOrtegaRamirez/protoforge/internal/formatter"
	"github.com/EdgarOrtegaRamirez/protoforge/internal/models"
	"github.com/EdgarOrtegaRamirez/protoforge/internal/parser"
	"github.com/EdgarOrtegaRamirez/protoforge/internal/validator"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "parse":
		cmdParse(args)
	case "analyze":
		cmdAnalyze(args)
	case "validate":
		cmdValidate(args)
	case "diff":
		cmdDiff(args)
	case "format":
		cmdFormat(args)
	case "version":
		fmt.Printf("protoforge %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`ProtoForge: Protocol Buffer Schema Toolkit

Usage:
  protoforge <command> [options]

Commands:
  parse     Parse a .proto file and display its AST
  analyze   Analyze a .proto file and display statistics
  validate  Validate a .proto file against best practices
  diff      Compare two .proto files
  format    Format a .proto file
  version   Show version
  help      Show this help message

Examples:
  protoforge parse user.proto
  protoforge analyze user.proto --format json
  protoforge validate user.proto
  protoforge diff old.proto new.proto
  protoforge format user.proto --sort-fields
`)
}

func readFile(path string) (string, error) {
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}

	// Handle --file flag
	if path == "" {
		return "", fmt.Errorf("no file specified")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	return string(data), nil
}

func parseFile(path string) (*models.ProtoFile, error) {
	content, err := readFile(path)
	if err != nil {
		return nil, err
	}

	p := parser.New()
	return p.Parse(content)
}

func getFormatFlag(args []string) string {
	for i, arg := range args {
		if arg == "--format" || arg == "-f" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--format=") {
			return arg[len("--format="):]
		}
	}
	return "text"
}

func getFileArgs(args []string) []string {
	var files []string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			files = append(files, arg)
		}
	}
	return files
}

func cmdParse(args []string) {
	files := getFileArgs(args)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: protoforge parse <file.proto>")
		os.Exit(1)
	}

	pf, err := parseFile(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	format := getFormatFlag(args)
	switch format {
	case "json":
		data := map[string]interface{}{
			"syntax":   pf.Syntax,
			"package":  pf.Package,
			"imports":  pf.Imports,
			"options":  pf.Options,
			"messages": pf.Messages,
			"enums":    pf.Enums,
			"services": pf.Services,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("syntax = %q;\n", pf.Syntax)
		fmt.Printf("package %s;\n\n", pf.Package)

		if len(pf.Imports) > 0 {
			for _, imp := range pf.Imports {
				prefix := ""
				if imp.Weak {
					prefix = "weak "
				} else if imp.Public {
					prefix = "public "
				}
				fmt.Printf("import %s\"%s\";\n", prefix, imp.Path)
			}
			fmt.Println()
		}

		for _, opt := range pf.Options {
			fmt.Printf("option %s = %s;\n", opt.Name, opt.Value)
		}
		if len(pf.Options) > 0 {
			fmt.Println()
		}

		for _, msg := range pf.Messages {
			printMessage(msg, "")
		}

		for _, enum := range pf.Enums {
			printEnum(enum, "")
		}

		for _, svc := range pf.Services {
			printService(svc)
		}
	}
}

func printMessage(msg *models.Message, indent string) {
	fmt.Printf("%smessage %s {\n", indent, msg.Name)
	inner := indent + "  "

	for _, field := range msg.Fields {
		if field.Label != "" {
			fmt.Printf("%s%s %s %s = %d;\n", inner, field.Label, field.Type, field.Name, field.Number)
		} else {
			fmt.Printf("%s%s %s = %d;\n", inner, field.Type, field.Name, field.Number)
		}
	}

	for _, mf := range msg.Maps {
		fmt.Printf("%smap<%s, %s> %s = %d;\n", inner, mf.KeyType, mf.ValueType, mf.Name, mf.Number)
	}

	for _, oo := range msg.OneOfs {
		fmt.Printf("%soneof %s {\n", inner, oo.Name)
		for _, field := range oo.Fields {
			fmt.Printf("%s  %s %s = %d;\n", inner, field.Type, field.Name, field.Number)
		}
		fmt.Printf("%s}\n", inner)
	}

	for _, nested := range msg.Messages {
		printMessage(nested, inner)
	}

	for _, enum := range msg.Enums {
		printEnum(enum, inner)
	}

	fmt.Printf("%s}\n\n", indent)
}

func printEnum(enum *models.Enum, indent string) {
	fmt.Printf("%senum %s {\n", indent, enum.Name)
	inner := indent + "  "
	for _, val := range enum.Values {
		fmt.Printf("%s%s = %d;\n", inner, val.Name, val.Number)
	}
	fmt.Printf("%s}\n\n", indent)
}

func printService(svc *models.Service) {
	fmt.Printf("service %s {\n", svc.Name)
	for _, method := range svc.Methods {
		clientStream := ""
		if method.ClientStreaming {
			clientStream = "stream "
		}
		serverStream := ""
		if method.ServerStreaming {
			serverStream = "stream "
		}
		fmt.Printf("  rpc %s (%s%s) returns (%s%s);\n", method.Name, clientStream, method.InputType, serverStream, method.OutputType)
	}
	fmt.Printf("}\n\n")
}

func cmdAnalyze(args []string) {
	files := getFileArgs(args)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: protoforge analyze <file.proto>")
		os.Exit(1)
	}

	pf, err := parseFile(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	a := analyzer.Analyze(pf)
	format := getFormatFlag(args)

	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(a.JSON()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Print(a.Summary())
	}
}

func cmdValidate(args []string) {
	files := getFileArgs(args)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: protoforge validate <file.proto>")
		os.Exit(1)
	}

	pf, err := parseFile(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	vr := validator.Validate(pf)
	format := getFormatFlag(args)

	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(validator.ToJSON(vr)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Print(validator.FormatValidationResult(vr))
	}

	if !vr.Valid {
		os.Exit(1)
	}
}

func cmdDiff(args []string) {
	files := getFileArgs(args)
	if len(files) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: protoforge diff <left.proto> <right.proto>")
		os.Exit(1)
	}

	left, err := parseFile(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", files[0], err)
		os.Exit(1)
	}

	right, err := parseFile(files[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", files[1], err)
		os.Exit(1)
	}

	diff := differ.DiffProtoFiles(left, right, files[0], files[1])
	format := getFormatFlag(args)

	switch format {
	case "json":
		data := map[string]interface{}{
			"left":    files[0],
			"right":   files[1],
			"summary": diff.Summary,
			"changes": diff.Changes,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Print(diff.SummaryText())
	}

	if diff.Summary.Breaking > 0 {
		fmt.Printf("\n⚠ %d breaking change(s) detected\n", diff.Summary.Breaking)
	}
}

func cmdFormat(args []string) {
	files := getFileArgs(args)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: protoforge format <file.proto>")
		os.Exit(1)
	}

	pf, err := parseFile(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	opts := formatter.DefaultOptions()

	// Check for sort flag
	for _, arg := range args {
		if arg == "--sort-fields" || arg == "-s" {
			opts.SortFields = true
		}
	}

	// Check for indent flag
	for i, arg := range args {
		if arg == "--indent" && i+1 < len(args) {
			opts.Indent = args[i+1]
		}
	}

	output := formatter.Format(pf, opts)
	fmt.Print(output)
}
