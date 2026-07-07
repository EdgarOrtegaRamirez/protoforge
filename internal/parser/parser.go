// Package parser provides a recursive descent parser for Protocol Buffer schema files.
package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/models"
)

// Parser parses .proto files into AST.
type Parser struct {
	input  string
	pos    int
	line   int
	col    int
	errors []string
}

// New creates a new parser.
func New() *Parser {
	return &Parser{line: 1, col: 1}
}

// Parse parses a .proto file string into a ProtoFile AST.
func (p *Parser) Parse(input string) (*models.ProtoFile, error) {
	p.input = input
	p.pos = 0
	p.line = 1
	p.col = 1
	p.errors = nil

	pf := &models.ProtoFile{
		Comments: make(map[string]string),
	}

	p.skipWhitespaceAndComments()

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}

		word := p.peekWord()
		if word == "" {
			break
		}

		switch word {
		case "syntax":
			syntax, err := p.parseSyntax()
			if err != nil {
				return nil, err
			}
			pf.Syntax = syntax
		case "package":
			pkg, err := p.parsePackage()
			if err != nil {
				return nil, err
			}
			pf.Package = pkg
		case "import":
			imp, err := p.parseImport()
			if err != nil {
				return nil, err
			}
			pf.Imports = append(pf.Imports, imp)
		case "option":
			opt, err := p.parseOption()
			if err != nil {
				return nil, err
			}
			pf.Options = append(pf.Options, opt)
		case "message":
			msg, err := p.parseMessage()
			if err != nil {
				return nil, err
			}
			pf.Messages = append(pf.Messages, msg)
		case "enum":
			enum, err := p.parseEnum()
			if err != nil {
				return nil, err
			}
			pf.Enums = append(pf.Enums, enum)
		case "service":
			svc, err := p.parseService()
			if err != nil {
				return nil, err
			}
			pf.Services = append(pf.Services, svc)
		case "extend":
			ext, err := p.parseExtend()
			if err != nil {
				return nil, err
			}
			pf.Extensions = append(pf.Extensions, ext...)
		case "rpc", "returns", "replaced":
			return nil, p.errorf("unexpected keyword %q at top level", word)
		default:
			// Could be a custom option or unknown construct - skip
			p.skipStatement()
		}
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors:\n%s", strings.Join(p.errors, "\n"))
	}

	return pf, nil
}

func (p *Parser) parseSyntax() (string, error) {
	p.advanceWord() // skip "syntax"
	p.expect('=')
	val, err := p.parseString()
	if err != nil {
		return "", fmt.Errorf("syntax: %w", err)
	}
	p.expect(';')
	return val, nil
}

func (p *Parser) parsePackage() (string, error) {
	p.advanceWord() // skip "package"
	pkg := p.readUntil(';')
	p.expect(';')
	return strings.TrimSpace(pkg), nil
}

func (p *Parser) parseImport() (*models.Import, error) {
	p.advanceWord() // skip "import"
	imp := &models.Import{}

	// Check for weak or public
	word := p.peekWord()
	if word == "weak" {
		imp.Weak = true
		p.advanceWord()
		word = p.peekWord()
	}
	if word == "public" {
		imp.Public = true
		p.advanceWord()
	}

	val, err := p.parseString()
	if err != nil {
		return nil, fmt.Errorf("import: %w", err)
	}
	imp.Path = val
	p.expect(';')
	return imp, nil
}

func (p *Parser) parseOption() (*models.Option, error) {
	p.advanceWord() // skip "option"
	opt := &models.Option{}

	name := p.readUntil('=')
	name = strings.TrimSpace(name)
	if strings.HasSuffix(name, ".") {
		return nil, fmt.Errorf("option name cannot end with dot: %q", name)
	}
	opt.Name = name
	opt.IsCustom = strings.Contains(name, ".")
	p.expect('=')

	// Read value - could be a string, number, or identifier
	val, err := p.parseOptionValue()
	if err != nil {
		return nil, fmt.Errorf("option %s: %w", name, err)
	}
	opt.Value = val
	p.expect(';')
	return opt, nil
}

func (p *Parser) parseOptionValue() (string, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return "", fmt.Errorf("unexpected end of input in option value")
	}

	ch := p.input[p.pos]
	if ch == '"' {
		return p.parseString()
	}
	if ch == '\'' {
		return p.parseSingleQuotedString()
	}
	if ch == '{' {
		return p.parseAggregateValue()
	}

	// Read until semicolon (handles numeric, identifier, boolean values)
	return p.readUntil(';'), nil
}

func (p *Parser) parseAggregateValue() (string, error) {
	p.expect('{')
	depth := 1
	start := p.pos
	for p.pos < len(p.input) && depth > 0 {
		ch := p.input[p.pos]
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
		case '"':
			p.skipString()
			continue
		}
		p.pos++
		p.col++
	}
	return p.input[start:p.pos], nil
}

func (p *Parser) parseMessage() (*models.Message, error) {
	p.advanceWord() // skip "message"
	name, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("message: %w", err)
	}

	msg := &models.Message{Name: name}

	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return nil, p.errorf("expected '{' after message name %q", name)
	}
	p.expect('{')

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == '}' {
			p.pos++
			p.col++
			break
		}

		word := p.peekWord()
		switch word {
		case "message":
			nested, err := p.parseMessage()
			if err != nil {
				return nil, err
			}
			msg.Messages = append(msg.Messages, nested)
		case "enum":
			enum, err := p.parseEnum()
			if err != nil {
				return nil, err
			}
			msg.Enums = append(msg.Enums, enum)
		case "oneof":
			oo, err := p.parseOneOf()
			if err != nil {
				return nil, err
			}
			msg.OneOfs = append(msg.OneOfs, oo)
		case "map":
			mf, err := p.parseMapField()
			if err != nil {
				return nil, err
			}
			msg.Maps = append(msg.Maps, mf)
		case "option":
			opt, err := p.parseOption()
			if err != nil {
				return nil, err
			}
			msg.Options = append(msg.Options, opt)
		case "extensions":
			exts, err := p.parseExtensionRange()
			if err != nil {
				return nil, err
			}
			msg.Extensions = append(msg.Extensions, exts...)
		case "reserved":
			p.skipReserved()
		case "optional", "required", "repeated":
			p.advanceWord() // consume label
			field, err := p.parseField(word)
			if err != nil {
				return nil, err
			}
			msg.Fields = append(msg.Fields, field)
		default:
			// Could be a type name (field without label in proto3)
			field, err := p.parseField("")
			if err != nil {
				return nil, err
			}
			msg.Fields = append(msg.Fields, field)
		}
	}

	return msg, nil
}

func (p *Parser) parseField(label string) (*models.Field, error) {
	field := &models.Field{Label: label}

	// Read type
	typeName := p.readUntilWord()
	field.Type = typeName

	// Read name
	name, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("field: %w", err)
	}
	field.Name = name

	p.expect('=')

	// Read field number
	numStr := p.readUntilDelim()
	num, err := strconv.Atoi(strings.TrimSpace(numStr))
	if err != nil {
		return nil, fmt.Errorf("invalid field number %q for field %s", numStr, name)
	}
	field.Number = num

	// Check for options or default value
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '[' {
		opts, err := p.parseFieldOptions()
		if err != nil {
			return nil, err
		}
		field.Options = opts
	}

	// Check for default value
	p.skipWhitespace()
	if p.pos < len(p.input) {
		remaining := strings.TrimSpace(p.input[p.pos:])
		if strings.HasPrefix(remaining, "default") {
			p.advanceN(7) // skip "default"
			p.skipWhitespace()
			if p.pos < len(p.input) && p.input[p.pos] == '=' {
				p.pos++
				p.col++
				p.skipWhitespace()
				defVal := p.readUntilDelim()
				field.DefaultValue = strings.TrimSpace(defVal)
			}
		}
	}

	p.expect(';')
	return field, nil
}

func (p *Parser) parseFieldOptions() ([]*models.Option, error) {
	p.expect('[')
	var opts []*models.Option

	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == ']' {
			p.pos++
			p.col++
			break
		}

		opt := &models.Option{}
		name := p.readUntil('=')
		name = strings.TrimSpace(name)
		opt.Name = name
		p.expect('=')

		val, err := p.parseOptionValue()
		if err != nil {
			return nil, err
		}
		opt.Value = val
		opts = append(opts, opt)

		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			p.col++
		}
	}
	return opts, nil
}

func (p *Parser) parseMapField() (*models.MapField, error) {
	p.advanceWord() // skip "map"
	p.expect('<')

	keyType := p.readUntilDelim()
	keyType = strings.TrimSpace(keyType)

	p.expect(',')
	valType := p.readUntil('>')
	valType = strings.TrimSpace(valType)

	p.expect('>')

	name, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("map field: %w", err)
	}

	p.expect('=')
	numStr := p.readUntilDelim()
	num, err := strconv.Atoi(strings.TrimSpace(numStr))
	if err != nil {
		return nil, fmt.Errorf("invalid map field number: %s", numStr)
	}

	mf := &models.MapField{
		KeyType:   keyType,
		ValueType: valType,
		Name:      name,
		Number:    num,
	}

	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '[' {
		opts, err := p.parseFieldOptions()
		if err != nil {
			return nil, err
		}
		mf.Options = opts
	}

	p.expect(';')
	return mf, nil
}

func (p *Parser) parseOneOf() (*models.OneOf, error) {
	p.advanceWord() // skip "oneof"
	name, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("oneof: %w", err)
	}

	oo := &models.OneOf{Name: name}

	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return nil, p.errorf("expected '{' after oneof name %q", name)
	}
	p.expect('{')

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == '}' {
			p.pos++
			p.col++
			break
		}

		word := p.peekWord()
		if word == "option" {
			opt, err := p.parseOption()
			if err != nil {
				return nil, err
			}
			oo.Options = append(oo.Options, opt)
		} else {
			field, err := p.parseField("")
			if err != nil {
				return nil, err
			}
			oo.Fields = append(oo.Fields, field)
		}
	}

	return oo, nil
}

func (p *Parser) parseEnum() (*models.Enum, error) {
	p.advanceWord() // skip "enum"
	name, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("enum: %w", err)
	}

	enum := &models.Enum{Name: name}

	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return nil, p.errorf("expected '{' after enum name %q", name)
	}
	p.expect('{')

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == '}' {
			p.pos++
			p.col++
			break
		}

		word := p.peekWord()
		switch word {
		case "option":
			opt, err := p.parseOption()
			if err != nil {
				return nil, err
			}
			enum.Options = append(enum.Options, opt)
		case "reserved":
			p.skipReserved()
		default:
			val, err := p.parseEnumValue()
			if err != nil {
				return nil, err
			}
			enum.Values = append(enum.Values, val)
		}
	}

	return enum, nil
}

func (p *Parser) parseEnumValue() (*models.EnumValue, error) {
	name := p.readUntil('=')
	name = strings.TrimSpace(name)

	p.expect('=')
	numStr := p.readUntilDelim()
	numStr = strings.TrimSpace(numStr)

	// Handle negative numbers
	negative := false
	if strings.HasPrefix(numStr, "-") {
		negative = true
		numStr = numStr[1:]
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return nil, fmt.Errorf("invalid enum value number: %s", numStr)
	}
	if negative {
		num = -num
	}

	ev := &models.EnumValue{Name: name, Number: num}

	// Check for options
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '[' {
		opts, err := p.parseFieldOptions()
		if err != nil {
			return nil, err
		}
		ev.Options = opts
	}

	p.expect(';')
	return ev, nil
}

func (p *Parser) parseService() (*models.Service, error) {
	p.advanceWord() // skip "service"
	name, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("service: %w", err)
	}

	svc := &models.Service{Name: name}

	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return nil, p.errorf("expected '{' after service name %q", name)
	}
	p.expect('{')

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == '}' {
			p.pos++
			p.col++
			break
		}

		word := p.peekWord()
		switch word {
		case "option":
			opt, err := p.parseOption()
			if err != nil {
				return nil, err
			}
			svc.Options = append(svc.Options, opt)
		case "rpc":
			method, err := p.parseMethod()
			if err != nil {
				return nil, err
			}
			svc.Methods = append(svc.Methods, method)
		default:
			p.skipStatement()
		}
	}

	return svc, nil
}

func (p *Parser) parseMethod() (*models.Method, error) {
	p.advanceWord() // skip "rpc"
	name, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("rpc method: %w", err)
	}

	method := &models.Method{Name: name}

	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '(' {
		return nil, p.errorf("expected '(' after rpc method name %q", name)
	}
	p.expect('(')

	// Check for client streaming
	if p.peekWord() == "stream" {
		p.advanceWord()
		method.ClientStreaming = true
	}

	// Read input type until closing paren
	inputType := p.readUntil(')')
	inputType = strings.TrimSpace(inputType)
	method.InputType = inputType
	p.expect(')')

	word := p.peekWord()
	if word == "returns" {
		p.advanceWord()
	} else {
		return nil, p.errorf("expected 'returns' in rpc %s", name)
	}

	p.expect('(')

	// Check for server streaming
	if p.peekWord() == "stream" {
		p.advanceWord()
		method.ServerStreaming = true
	}

	// Read output type until closing paren
	outputType := p.readUntil(')')
	outputType = strings.TrimSpace(outputType)
	method.OutputType = outputType
	p.expect(')')

	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '{' {
		// Parse method body (streaming options, etc.)
		p.expect('{')
		for p.pos < len(p.input) {
			p.skipWhitespaceAndComments()
			if p.pos >= len(p.input) {
				break
			}
			if p.input[p.pos] == '}' {
				p.pos++
				p.col++
				break
			}
			word := p.peekWord()
			if word == "option" {
				opt, err := p.parseOption()
				if err != nil {
					return nil, err
				}
				method.Options = append(method.Options, opt)
			} else {
				p.skipStatement()
			}
		}
	} else {
		p.expect(';')
	}

	return method, nil
}

func (p *Parser) parseExtend() ([]*models.Extension, error) {
	p.advanceWord() // skip "extend"
	// Skip the type reference
	p.skipWhitespace()
	p.expect('{')
	var exts []*models.Extension

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == '}' {
			p.pos++
			p.col++
			break
		}
		p.skipStatement()
	}

	return exts, nil
}

func (p *Parser) parseExtensionRange() ([]*models.Extension, error) {
	p.advanceWord() // skip "extensions"
	var exts []*models.Extension

	for {
		p.skipWhitespace()
		numStr := p.readUntilDelim()
		numStr = strings.TrimSpace(numStr)
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return nil, fmt.Errorf("invalid extension range number: %s", numStr)
		}

		ext := &models.Extension{RangeStart: num}

		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			p.col++
			continue
		}
		if p.pos < len(p.input) && p.input[p.pos] == 't' {
			// "to max" or "to N"
			p.pos++
			p.col++
			p.skipWhitespace()
			next := p.peekWord()
			if next == "max" {
				ext.RangeEnd = -1 // sentinel for max
				p.advanceWord()
			} else {
				toNum := p.readUntilDelim()
				toNum = strings.TrimSpace(toNum)
				toInt, err := strconv.Atoi(toNum)
				if err != nil {
					return nil, fmt.Errorf("invalid extension range end: %s", toNum)
				}
				ext.RangeEnd = toInt
			}
		} else {
			ext.RangeEnd = num // single number range
		}

		exts = append(exts, ext)

		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == ';' {
			p.pos++
			p.col++
			break
		}
	}

	return exts, nil
}

func (p *Parser) skipReserved() {
	p.advanceWord() // skip "reserved"
	// Skip until semicolon
	p.skipWhitespace()
	for p.pos < len(p.input) && p.input[p.pos] != ';' {
		p.pos++
		p.col++
	}
	if p.pos < len(p.input) {
		p.pos++
		p.col++
	}
}

// Helper methods

func (p *Parser) peekWord() string {
	p.skipWhitespace()
	start := p.pos
	for p.pos < len(p.input) {
		ch := rune(p.input[p.pos])
		if unicode.IsSpace(ch) || ch == '{' || ch == '}' || ch == '=' || ch == ';' || ch == '(' || ch == ')' || ch == '[' || ch == ']' || ch == ',' || ch == '<' || ch == '>' || ch == '/' || ch == '"' || ch == '\'' {
			break
		}
		p.pos++
		p.col++
	}
	word := p.input[start:p.pos]
	p.pos = start // reset
	p.col = 1     // reset col (not perfect but ok)
	return word
}

func (p *Parser) advanceWord() {
	p.skipWhitespace()
	for p.pos < len(p.input) {
		ch := rune(p.input[p.pos])
		if unicode.IsSpace(ch) || ch == '{' || ch == '}' || ch == '=' || ch == ';' || ch == '(' || ch == ')' || ch == '[' || ch == ']' || ch == ',' || ch == '<' || ch == '>' || ch == '/' || ch == '"' || ch == '\'' {
			break
		}
		p.pos++
		p.col++
	}
}

func (p *Parser) advanceN(n int) {
	for i := 0; i < n && p.pos < len(p.input); i++ {
		if p.input[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
}

func (p *Parser) expect(ch byte) {
	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != ch {
		p.errors = append(p.errors, fmt.Sprintf("line %d, col %d: expected '%c', got '%c'", p.line, p.col, ch, p.current()))
		return
	}
	if ch == '\n' {
		p.line++
		p.col = 1
	} else {
		p.col++
	}
	p.pos++
}

func (p *Parser) current() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			if ch == '\n' {
				p.line++
				p.col = 1
			} else {
				p.col++
			}
			p.pos++
		} else if ch == '/' && p.pos+1 < len(p.input) {
			if p.input[p.pos+1] == '/' {
				// Single-line comment
				for p.pos < len(p.input) && p.input[p.pos] != '\n' {
					p.pos++
					p.col++
				}
			} else if p.input[p.pos+1] == '*' {
				// Block comment
				p.pos += 2
				p.col += 2
				for p.pos < len(p.input)-1 {
					if p.input[p.pos] == '*' && p.input[p.pos+1] == '/' {
						p.pos += 2
						p.col += 2
						break
					}
					if p.input[p.pos] == '\n' {
						p.line++
						p.col = 1
					} else {
						p.col++
					}
					p.pos++
				}
			} else {
				break
			}
		} else {
			break
		}
	}
}

func (p *Parser) skipWhitespaceAndComments() {
	p.skipWhitespace()
}

func (p *Parser) parseString() (string, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '"' {
		return "", fmt.Errorf("expected '\"' at position %d", p.pos)
	}
	p.pos++
	p.col++
	start := p.pos
	for p.pos < len(p.input) {
		if p.input[p.pos] == '\\' && p.pos+1 < len(p.input) {
			p.pos += 2
			p.col += 2
			continue
		}
		if p.input[p.pos] == '"' {
			val := p.input[start:p.pos]
			p.pos++
			p.col++
			return val, nil
		}
		if p.input[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
	return "", fmt.Errorf("unterminated string starting at position %d", start)
}

func (p *Parser) parseSingleQuotedString() (string, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '\'' {
		return "", fmt.Errorf("expected ''' at position %d", p.pos)
	}
	p.pos++
	p.col++
	start := p.pos
	for p.pos < len(p.input) {
		if p.input[p.pos] == '\\' && p.pos+1 < len(p.input) {
			p.pos += 2
			p.col += 2
			continue
		}
		if p.input[p.pos] == '\'' {
			val := p.input[start:p.pos]
			p.pos++
			p.col++
			return val, nil
		}
		if p.input[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
	return "", fmt.Errorf("unterminated string starting at position %d", start)
}

func (p *Parser) skipString() {
	if p.pos >= len(p.input) {
		return
	}
	quote := p.input[p.pos]
	p.pos++
	p.col++
	for p.pos < len(p.input) {
		if p.input[p.pos] == '\\' && p.pos+1 < len(p.input) {
			p.pos += 2
			p.col += 2
			continue
		}
		if p.input[p.pos] == quote {
			p.pos++
			p.col++
			return
		}
		if p.input[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
}

func (p *Parser) parseIdentifier() (string, error) {
	p.skipWhitespace()
	start := p.pos
	if p.pos >= len(p.input) {
		return "", fmt.Errorf("unexpected end of input, expected identifier")
	}

	ch := rune(p.input[p.pos])
	if !unicode.IsLetter(ch) && ch != '_' {
		return "", fmt.Errorf("expected identifier, got '%c' at line %d, col %d", ch, p.line, p.col)
	}

	for p.pos < len(p.input) {
		ch = rune(p.input[p.pos])
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.' {
			p.pos++
			p.col++
		} else {
			break
		}
	}

	return p.input[start:p.pos], nil
}

func (p *Parser) readUntil(delim byte) string {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == delim {
			break
		}
		if ch == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
	return strings.TrimSpace(p.input[start:p.pos])
}

func (p *Parser) readUntilWord() string {
	p.skipWhitespace()
	start := p.pos
	for p.pos < len(p.input) {
		ch := rune(p.input[p.pos])
		if unicode.IsSpace(ch) {
			break
		}
		if ch == '{' || ch == '}' || ch == '=' || ch == ';' || ch == '(' || ch == ')' || ch == '[' || ch == ']' || ch == ',' || ch == '<' || ch == '>' {
			break
		}
		p.pos++
		p.col++
	}
	return p.input[start:p.pos]
}

func (p *Parser) readUntilDelim() string {
	p.skipWhitespace()
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ';' || ch == ',' || ch == ']' || ch == ')' || ch == '>' || ch == '}' || unicode.IsSpace(rune(ch)) {
			break
		}
		p.pos++
		p.col++
	}
	return p.input[start:p.pos]
}

func (p *Parser) skipStatement() {
	depth := 0
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		switch ch {
		case '{':
			depth++
		case '}':
			if depth == 0 {
				return
			}
			depth--
		case ';':
			if depth == 0 {
				p.pos++
				p.col++
				return
			}
		case '"':
			p.skipString()
			continue
		case '\'':
			p.skipString()
			continue
		case '\n':
			p.line++
			p.col = 1
		default:
			p.col++
		}
		p.pos++
	}
}

func (p *Parser) errorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("line %d, col %d: %s", p.line, p.col, msg)
	p.errors = append(p.errors, err.Error())
	return err
}
