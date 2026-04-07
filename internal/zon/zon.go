package zon

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

type ZONEncoder struct {
	mu     sync.RWMutex
	config EncoderConfig
	schema *Schema
	writer io.Writer
	depth  int
}

type EncoderConfig struct {
	Mode          EncodingMode
	IndentSize    int
	UseCompact    bool
	PreserveTypes bool
	LLMOptimized  bool
}

type EncodingMode int

const (
	ModeReadable EncodingMode = iota
	ModeCompact
	ModeLLMOptimized
)

type Schema struct {
	Fields map[string]FieldSchema
}

type FieldSchema struct {
	Type     string
	Required bool
	Default  interface{}
}

func NewZONEncoder(config EncoderConfig) *ZONEncoder {
	return &ZONEncoder{
		config: config,
		schema: &Schema{Fields: make(map[string]FieldSchema)},
	}
}

func DefaultEncoderConfig() EncoderConfig {
	return EncoderConfig{
		Mode:          ModeLLMOptimized,
		IndentSize:    2,
		UseCompact:    false,
		PreserveTypes: true,
		LLMOptimized:  true,
	}
}

func (e *ZONEncoder) Encode(value interface{}) (string, error) {
	var sb strings.Builder
	e.writer = &sb

	err := e.encodeValue(value, 0)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

func (e *ZONEncoder) encodeValue(value interface{}, depth int) error {
	e.depth = depth

	switch v := value.(type) {
	case nil:
		io.WriteString(e.writer, "null")
	case bool:
		if v {
			io.WriteString(e.writer, "T")
		} else {
			io.WriteString(e.writer, "F")
		}
	case int:
		io.WriteString(e.writer, strconv.FormatInt(int64(v), 10))
	case int64:
		io.WriteString(e.writer, strconv.FormatInt(v, 10))
	case float64:
		io.WriteString(e.writer, strconv.FormatFloat(v, 'f', -1, 64))
	case string:
		e.encodeString(v)
	case []interface{}:
		return e.encodeArray(v)
	case map[string]interface{}:
		return e.encodeObject(v)
	default:
		io.WriteString(e.writer, fmt.Sprintf("%v", value))
	}

	return nil
}

func (e *ZONEncoder) encodeString(s string) {
	needsQuotes := false
	if strings.ContainsAny(s, " \t\n") || strings.HasPrefix(s, "@") || strings.HasPrefix(s, "(") {
		needsQuotes = true
	}

	if needsQuotes {
		io.WriteString(e.writer, "\"")
		io.WriteString(e.writer, strings.ReplaceAll(s, "\"", "\\\""))
		io.WriteString(e.writer, "\"")
	} else {
		io.WriteString(e.writer, s)
	}
}

func (e *ZONEncoder) encodeArray(arr []interface{}) error {
	if len(arr) == 0 {
		io.WriteString(e.writer, "[]")
		return nil
	}

	count := len(arr)
	io.WriteString(e.writer, fmt.Sprintf("@(%d)[", count))

	for i, item := range arr {
		if i > 0 {
			io.WriteString(e.writer, ", ")
		}
		e.encodeValue(item, e.depth+1)
	}

	io.WriteString(e.writer, "]")
	return nil
}

func (e *ZONEncoder) encodeObject(obj map[string]interface{}) error {
	if len(obj) == 0 {
		io.WriteString(e.writer, "{}")
		return nil
	}

	io.WriteString(e.writer, "{\n")

	indent := strings.Repeat(" ", (e.depth+1)*e.config.IndentSize)
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	for i, key := range keys {
		if i > 0 {
			io.WriteString(e.writer, ",\n")
		}

		io.WriteString(e.writer, indent)

		if e.config.PreserveTypes {
			io.WriteString(e.writer, key)
			io.WriteString(e.writer, ": ")
		} else {
			io.WriteString(e.writer, key)
			io.WriteString(e.writer, " = ")
		}

		err := e.encodeValue(obj[key], e.depth+1)
		if err != nil {
			return err
		}
	}

	io.WriteString(e.writer, "\n")
	io.WriteString(e.writer, strings.Repeat(" ", e.depth*e.config.IndentSize))
	io.WriteString(e.writer, "}")

	return nil
}

type ZONDecoder struct {
	mu     sync.RWMutex
	config DecoderConfig
	tokens []Token
	pos    int
}

type DecoderConfig struct {
	StrictValidation bool
	AllowComments    bool
}

type Token struct {
	Type  TokenType
	Value string
}

type TokenType int

const (
	TokenNull TokenType = iota
	TokenBool
	TokenNumber
	TokenString
	TokenArrayStart
	TokenArrayEnd
	TokenObjectStart
	TokenObjectEnd
	TokenKey
	TokenCount
	TokenEOF
)

func NewZONDecoder() *ZONDecoder {
	return &ZONDecoder{
		config: DecoderConfig{StrictValidation: true},
	}
}

func (d *ZONDecoder) Decode(input string) (interface{}, error) {
	d.tokens = d.lex(input)
	d.pos = 0

	return d.parseValue()
}

func (d *ZONDecoder) lex(input string) []Token {
	tokens := []Token{}

	i := 0
	for i < len(input) {
		c := input[i]

		switch {
		case c == ' ' || c == '\t' || c == '\n':
			i++
		case c == 'T':
			tokens = append(tokens, Token{Type: TokenBool, Value: "true"})
			i++
		case c == 'F':
			tokens = append(tokens, Token{Type: TokenBool, Value: "false"})
			i++
		case c == 'n' && i+3 < len(input) && input[i:i+4] == "null":
			tokens = append(tokens, Token{Type: TokenNull, Value: "null"})
			i += 4
		case c >= '0' && c <= '9' || c == '-':
			start := i
			for i < len(input) && (input[i] >= '0' && input[i] <= '9' || input[i] == '.' || input[i] == '-') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: input[start:i]})
		case c == '"':
			i++
			var sb strings.Builder
			for i < len(input) && input[i] != '"' {
				if input[i] == '\\' && i+1 < len(input) {
					i++
				}
				sb.WriteByte(input[i])
				i++
			}
			i++
			tokens = append(tokens, Token{Type: TokenString, Value: sb.String()})
		case c == '@':
			i++
			if i < len(input) && input[i] == '(' {
				i++
				numStart := i
				for i < len(input) && input[i] >= '0' && input[i] <= '9' {
					i++
				}
				if i < len(input) && input[i] == ')' {
					i++
					tokens = append(tokens, Token{Type: TokenCount, Value: input[numStart : i-1]})
				}
			}
		case c == '[':
			tokens = append(tokens, Token{Type: TokenArrayStart})
			i++
		case c == ']':
			tokens = append(tokens, Token{Type: TokenArrayEnd})
			i++
		case c == '{':
			tokens = append(tokens, Token{Type: TokenObjectStart})
			i++
		case c == '}':
			tokens = append(tokens, Token{Type: TokenObjectEnd})
			i++
		case c == '=':
			i++
			if i < len(input) && input[i] == ' ' {
				i++
			}
		case c == ':':
			i++
			if i < len(input) && input[i] == ' ' {
				i++
			}
		default:
			start := i
			for i < len(input) && !strings.ContainsAny(string(input[i]), " \t\n{},[]=") {
				i++
			}
			if start < i {
				tokens = append(tokens, Token{Type: TokenKey, Value: input[start:i]})
			}
		}
	}

	tokens = append(tokens, Token{Type: TokenEOF})
	return tokens
}

func (d *ZONDecoder) parseValue() (interface{}, error) {
	if d.pos >= len(d.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	tok := d.tokens[d.pos]
	d.pos++

	switch tok.Type {
	case TokenNull:
		return nil, nil
	case TokenBool:
		return tok.Value == "true", nil
	case TokenNumber:
		if strings.Contains(tok.Value, ".") {
			f, err := strconv.ParseFloat(tok.Value, 64)
			if err != nil {
				return nil, err
			}
			return f, nil
		}
		i, err := strconv.ParseInt(tok.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	case TokenString:
		return tok.Value, nil
	case TokenCount:
		return nil, nil
	case TokenArrayStart:
		return d.parseArray()
	case TokenObjectStart:
		return d.parseObject()
	case TokenKey:
		return nil, nil
	default:
		return nil, nil
	}
}

func (d *ZONDecoder) parseArray() ([]interface{}, error) {
	arr := []interface{}{}

	for d.pos < len(d.tokens) {
		tok := d.tokens[d.pos]

		if tok.Type == TokenArrayEnd {
			d.pos++
			break
		}

		val, err := d.parseValue()
		if err != nil {
			return nil, err
		}
		arr = append(arr, val)

		if d.pos < len(d.tokens) && d.tokens[d.pos].Type == TokenArrayEnd {
			continue
		}
	}

	return arr, nil
}

func (d *ZONDecoder) parseObject() (map[string]interface{}, error) {
	obj := make(map[string]interface{})

	for d.pos < len(d.tokens) {
		tok := d.tokens[d.pos]

		if tok.Type == TokenObjectEnd || tok.Type == TokenEOF {
			if tok.Type == TokenObjectEnd {
				d.pos++
			}
			break
		}

		if tok.Type != TokenKey {
			d.pos++
			continue
		}

		key := tok.Value
		d.pos++

		val, err := d.parseValue()
		if err != nil {
			return nil, err
		}

		obj[key] = val
	}

	return obj, nil
}

type ZONConverter struct{}

func NewZONConverter() *ZONConverter {
	return &ZONConverter{}
}

func (c *ZONConverter) JSONToZON(jsonStr string) (string, error) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	encoder := NewZONEncoder(DefaultEncoderConfig())
	return encoder.Encode(data)
}

func (c *ZONConverter) ZONToJSON(zonStr string) (string, error) {
	decoder := NewZONDecoder()
	data, err := decoder.Decode(zonStr)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

type ZONFormatter struct {
	config FormatConfig
}

type FormatConfig struct {
	IndentSize int
	Compact    bool
	SortKeys   bool
}

func NewZONFormatter(config FormatConfig) *ZONFormatter {
	if config.IndentSize == 0 {
		config.IndentSize = 2
	}
	return &ZONFormatter{config: config}
}

func (f *ZONFormatter) Format(input string) (string, error) {
	decoder := NewZONDecoder()
	data, err := decoder.Decode(input)
	if err != nil {
		return "", err
	}

	encoder := NewZONEncoder(EncoderConfig{
		Mode:          ModeReadable,
		IndentSize:    f.config.IndentSize,
		PreserveTypes: true,
	})

	return encoder.Encode(data)
}

func (f *ZONFormatter) Minify(input string) (string, error) {
	decoder := NewZONDecoder()
	data, err := decoder.Decode(input)
	if err != nil {
		return "", err
	}

	encoder := NewZONEncoder(EncoderConfig{
		Mode:       ModeCompact,
		UseCompact: true,
	})

	return encoder.Encode(data)
}
