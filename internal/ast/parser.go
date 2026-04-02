// Package ast provides language-specific AST parsing for context-aware compression.
package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Language represents a supported programming language.
type Language int

const (
	LangUnknown Language = iota
	LangGo
	LangRust
	LangPython
	LangJavaScript
	LangTypeScript
)

// ParseResult contains the parsed AST and extracted information.
type ParseResult struct {
	Language  Language
	FilePath  string
	Package   string
	Imports   []Import
	Types     []TypeDef
	Functions []Function
	Variables []Variable
	Constants []Constant
	Comments  []Comment
	RawAST    interface{} // language-specific AST
}

// Import represents an import statement.
type Import struct {
	Path   string
	Alias  string
	Line   int
	IsUsed bool
}

// TypeDef represents a type definition.
type TypeDef struct {
	Name    string
	Kind    string // struct, interface, enum, etc.
	Line    int
	Fields  []Field
	Methods []string
	Doc     string
}

// Field represents a struct field.
type Field struct {
	Name string
	Type string
	Line int
	Tags map[string]string
	Doc  string
}

// Function represents a function or method.
type Function struct {
	Name       string
	Receiver   string // for methods
	Params     []Param
	Returns    []string
	Line       int
	EndLine    int
	Doc        string
	IsExported bool
	BodySize   int // lines of code
}

// Param represents a function parameter.
type Param struct {
	Name string
	Type string
}

// Variable represents a variable declaration.
type Variable struct {
	Name  string
	Type  string
	Line  int
	Value string
}

// Constant represents a constant declaration.
type Constant struct {
	Name  string
	Type  string
	Line  int
	Value string
}

// Comment represents a code comment.
type Comment struct {
	Text string
	Line int
	Kind string // line, block, doc
}

// Parser parses source code into AST.
type Parser struct {
	language Language
}

// NewParser creates a new AST parser.
func NewParser(lang Language) *Parser {
	return &Parser{language: lang}
}

// DetectLanguage detects the language from file extension.
func DetectLanguage(filename string) Language {
	if strings.HasSuffix(filename, ".go") {
		return LangGo
	}
	if strings.HasSuffix(filename, ".rs") {
		return LangRust
	}
	if strings.HasSuffix(filename, ".py") {
		return LangPython
	}
	if strings.HasSuffix(filename, ".js") || strings.HasSuffix(filename, ".jsx") {
		return LangJavaScript
	}
	if strings.HasSuffix(filename, ".ts") || strings.HasSuffix(filename, ".tsx") {
		return LangTypeScript
	}
	return LangUnknown
}

// Parse parses source code and returns the AST result.
func (p *Parser) Parse(filename string, src []byte) (*ParseResult, error) {
	switch p.language {
	case LangGo:
		return p.parseGo(filename, src)
	default:
		return nil, nil // Return nil for unsupported languages
	}
}

// parseGo parses Go source code.
func (p *Parser) parseGo(filename string, src []byte) (*ParseResult, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	result := &ParseResult{
		Language: LangGo,
		FilePath: filename,
		Package:  file.Name.Name,
		RawAST:   file,
	}

	// Extract imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		result.Imports = append(result.Imports, Import{
			Path:   importPath,
			Alias:  alias,
			Line:   fset.Position(imp.Pos()).Line,
			IsUsed: false, // Will be determined by analysis
		})
	}

	// Extract declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			p.extractGenDecl(d, fset, result)
		case *ast.FuncDecl:
			p.extractFuncDecl(d, fset, result)
		}
	}

	// Extract comments
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			result.Comments = append(result.Comments, Comment{
				Text: c.Text,
				Line: fset.Position(c.Pos()).Line,
				Kind: classifyComment(c.Text),
			})
		}
	}

	return result, nil
}

func (p *Parser) extractGenDecl(d *ast.GenDecl, fset *token.FileSet, result *ParseResult) {
	switch d.Tok {
	case token.TYPE:
		for _, spec := range d.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			typeDef := TypeDef{
				Name: ts.Name.Name,
				Line: fset.Position(ts.Pos()).Line,
			}
			if d.Doc != nil {
				typeDef.Doc = d.Doc.Text()
			}

			switch t := ts.Type.(type) {
			case *ast.StructType:
				typeDef.Kind = "struct"
				for _, field := range t.Fields.List {
					f := Field{
						Line: fset.Position(field.Pos()).Line,
					}
					if len(field.Names) > 0 {
						f.Name = field.Names[0].Name
					}
					if field.Type != nil {
						f.Type = exprToString(field.Type)
					}
					if field.Tag != nil {
						f.Tags = parseTags(field.Tag.Value)
					}
					typeDef.Fields = append(typeDef.Fields, f)
				}
			case *ast.InterfaceType:
				typeDef.Kind = "interface"
			case *ast.Ident:
				typeDef.Kind = t.Name
			}

			result.Types = append(result.Types, typeDef)
		}

	case token.VAR:
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				v := Variable{
					Name: name.Name,
					Line: fset.Position(name.Pos()).Line,
				}
				if vs.Type != nil {
					v.Type = exprToString(vs.Type)
				}
				if i < len(vs.Values) {
					v.Value = exprToString(vs.Values[i])
				}
				result.Variables = append(result.Variables, v)
			}
		}

	case token.CONST:
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				c := Constant{
					Name: name.Name,
					Line: fset.Position(name.Pos()).Line,
				}
				if vs.Type != nil {
					c.Type = exprToString(vs.Type)
				}
				if i < len(vs.Values) {
					c.Value = exprToString(vs.Values[i])
				}
				result.Constants = append(result.Constants, c)
			}
		}
	}
}

func (p *Parser) extractFuncDecl(d *ast.FuncDecl, fset *token.FileSet, result *ParseResult) {
	fn := Function{
		Name:       d.Name.Name,
		Line:       fset.Position(d.Pos()).Line,
		EndLine:    fset.Position(d.End()).Line,
		IsExported: ast.IsExported(d.Name.Name),
	}

	if d.Doc != nil {
		fn.Doc = d.Doc.Text()
	}

	// Receiver for methods
	if d.Recv != nil && len(d.Recv.List) > 0 {
		fn.Receiver = exprToString(d.Recv.List[0].Type)
	}

	// Parameters
	if d.Type.Params != nil {
		for _, field := range d.Type.Params.List {
			paramType := exprToString(field.Type)
			for _, name := range field.Names {
				fn.Params = append(fn.Params, Param{
					Name: name.Name,
					Type: paramType,
				})
			}
		}
	}

	// Return values
	if d.Type.Results != nil {
		for _, field := range d.Type.Results.List {
			fn.Returns = append(fn.Returns, exprToString(field.Type))
		}
	}

	// Body size
	if d.Body != nil {
		fn.BodySize = fset.Position(d.Body.End()).Line - fn.Line
	}

	result.Functions = append(result.Functions, fn)
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + exprToString(e.Elt)
		}
		return "[" + exprToString(e.Len) + "]" + exprToString(e.Elt)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.FuncType:
		return "func"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	case *ast.ChanType:
		return "chan " + exprToString(e.Value)
	case *ast.BasicLit:
		return e.Value
	default:
		return ""
	}
}

func parseTags(tag string) map[string]string {
	tags := make(map[string]string)
	// Simple tag parsing: `json:"name,omitempty"`
	tag = strings.Trim(tag, "`")
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if idx := strings.Index(part, ":"); idx > 0 {
			key := part[:idx]
			val := strings.Trim(part[idx+1:], `"`)
			tags[key] = val
		}
	}
	return tags
}

func classifyComment(text string) string {
	if strings.HasPrefix(text, "//") {
		return "line"
	}
	if strings.HasPrefix(text, "/*") {
		return "block"
	}
	return "unknown"
}

// GetOutline returns a simplified outline of the file.
func (r *ParseResult) GetOutline() string {
	var b strings.Builder

	// Package
	b.WriteString("package ")
	b.WriteString(r.Package)
	b.WriteString("\n\n")

	// Imports
	if len(r.Imports) > 0 {
		b.WriteString("// Imports: ")
		for i, imp := range r.Imports {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(imp.Path)
		}
		b.WriteString("\n\n")
	}

	// Types
	for _, t := range r.Types {
		b.WriteString("type ")
		b.WriteString(t.Name)
		b.WriteString(" ")
		b.WriteString(t.Kind)
		if t.Kind == "struct" && len(t.Fields) > 0 {
			b.WriteString(" {")
			for _, f := range t.Fields {
				b.WriteString("\n    ")
				b.WriteString(f.Name)
				b.WriteString(" ")
				b.WriteString(f.Type)
			}
			b.WriteString("\n}")
		}
		b.WriteString("\n")
	}
	if len(r.Types) > 0 {
		b.WriteString("\n")
	}

	// Functions
	for _, fn := range r.Functions {
		if fn.Receiver != "" {
			b.WriteString("func (")
			b.WriteString(fn.Receiver)
			b.WriteString(") ")
		} else {
			b.WriteString("func ")
		}
		b.WriteString(fn.Name)
		b.WriteString("(")
		for i, p := range fn.Params {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(p.Name)
			b.WriteString(" ")
			b.WriteString(p.Type)
		}
		b.WriteString(")")
		if len(fn.Returns) > 0 {
			b.WriteString(" ")
			if len(fn.Returns) > 1 {
				b.WriteString("(")
			}
			for i, r := range fn.Returns {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(r)
			}
			if len(fn.Returns) > 1 {
				b.WriteString(")")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// GetSymbols returns all symbols for code folding.
func (r *ParseResult) GetSymbols() []Symbol {
	var symbols []Symbol

	// Types
	for _, t := range r.Types {
		symbols = append(symbols, Symbol{
			Name:   t.Name,
			Kind:   "type",
			Line:   t.Line,
			Detail: t.Kind,
		})
	}

	// Functions
	for _, fn := range r.Functions {
		kind := "function"
		if fn.Receiver != "" {
			kind = "method"
		}
		symbols = append(symbols, Symbol{
			Name:     fn.Name,
			Kind:     kind,
			Line:     fn.Line,
			EndLine:  fn.EndLine,
			Receiver: fn.Receiver,
		})
	}

	// Variables
	for _, v := range r.Variables {
		symbols = append(symbols, Symbol{
			Name: v.Name,
			Kind: "variable",
			Line: v.Line,
			Type: v.Type,
		})
	}

	// Constants
	for _, c := range r.Constants {
		symbols = append(symbols, Symbol{
			Name:  c.Name,
			Kind:  "constant",
			Line:  c.Line,
			Type:  c.Type,
			Value: c.Value,
		})
	}

	return symbols
}

// Symbol represents a code symbol.
type Symbol struct {
	Name     string
	Kind     string
	Line     int
	EndLine  int
	Type     string
	Value    string
	Receiver string
	Detail   string
}
