//go:build !no_participle_tags
// +build !no_participle_tags

package poml

import (
	"fmt"
	"strings"

	"github.com/ZanzyTHEbar/poml/sdk/ast"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Build a simple lexer for POML-like markup.
func buildPOMLLexer() lexer.Definition {
	rules := []lexer.SimpleRule{
		{Name: "Comment", Pattern: `<!--[\s\S]*?-->`},
		{Name: "LAngle", Pattern: `<`},
		{Name: "RAngle", Pattern: `>`},
		{Name: "Slash", Pattern: `/`},
		{Name: "Equals", Pattern: `=`},
		{Name: "String", Pattern: "\"[^\"]*\""},
		{Name: "TemplateExpr", Pattern: `\{{[^}]*\}}`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_-]*`},
		{Name: "WS", Pattern: `\s+`},
		{Name: "Text", Pattern: `[^<>{]+`},
	}
	return lexer.MustSimple(rules)
}

// Grammar types
type PDocument struct {
	Leading  []*PNode  `@@*`
	Root     *PElement `@@?`
	Trailing []*PNode  `@@*`
}

type PNode struct {
	Element *PElement `@@`
	Text    *PText    `@@`
}

type PText struct {
	Parts []string `(@Text | @TemplateExpr | @Ident | @WS)+`
}

type PElement struct {
	Open     *POpen   `@@`
	Children []*PNode `@@*`
	Close    *PClose  `@@?`
}

type POpen struct {
	L     string   `@LAngle`
	Name  string   `@Ident`
	Parts []string `(@WS | @Ident | @String | @Equals | @Slash)*`
	R     string   `@RAngle`
}

type PClose struct {
	L    string  `@LAngle`
	_    *string `@WS?F`
	S    string  `@Slash`
	_2   *string `@WS?`
	Name string  `@Ident`
	_3   *string `@WS?`
	R    string  `@RAngle`
}

var pomlParser *participle.Parser[PDocument]

func init() {
	def := buildPOMLLexer()
	p, err := participle.Build[PDocument](participle.Lexer(def), participle.Elide("Comment"))
	if err != nil {
		panic(err)
	}
	pomlParser = p
}

// ParseWithParticiple parses input using participle grammar and converts to AST.
func ParseWithParticiple(input string) (*ast.Document, error) {
	var b strings.Builder
	parsed, err := pomlParser.ParseString("", input, participle.Trace(&b))
	if err != nil {
		return nil, fmt.Errorf("parse error: %v\ntrace:\n%s", err, b.String())
	}
	out := &ast.Document{}
	for _, ln := range parsed.Leading {
		if n := convertPNode(ln); n != nil {
			out.Children = append(out.Children, n)
		}
	}
	if parsed.Root != nil {
		out.Children = append(out.Children, convertElement(parsed.Root))
	}
	for _, tn := range parsed.Trailing {
		if n := convertPNode(tn); n != nil {
			out.Children = append(out.Children, n)
		}
	}
	return out, nil
}

// ParseWithTrace runs participle parse with trace enabled and returns the trace output.
func ParseWithTrace(input string) (string, error) {
	var b strings.Builder
	_, err := pomlParser.ParseString("", input, participle.Trace(&b))
	return b.String(), err
}

func convertPNode(child *PNode) ast.Node {
	if child == nil {
		return nil
	}
	if child.Element != nil {
		return convertElement(child.Element)
	}
	if child.Text != nil {
		return convertPText(child.Text)
	}
	return nil
}

// convertPText converts PText to appropriate AST nodes, handling template expressions
func convertPText(pt *PText) ast.Node {
	if len(pt.Parts) == 0 {
		return &ast.Text{Content: ""}
	}

	// Check if any part is a template expression
	hasTemplates := false
	for _, part := range pt.Parts {
		if strings.HasPrefix(part, "{{") && strings.HasSuffix(part, "}}") {
			hasTemplates = true
			break
		}
	}

	if hasTemplates {
		// Create a text fragment container to hold split text and template nodes
		children := []ast.Node{}

		for _, part := range pt.Parts {
			if strings.HasPrefix(part, "{{") && strings.HasSuffix(part, "}}") {
				// This is a template expression - create TemplateExpr node
				templateNode := &ast.TemplateExpr{
					Expression: part,
					IsCompiled: true, // Assume valid for now
				}
				children = append(children, templateNode)
			} else if part != "" {
				// This is regular text - create Text node
				children = append(children, &ast.Text{Content: part})
			}
		}

		// Return a special container element that holds the split nodes
		// Use empty name to indicate this is a text fragment container
		container := &ast.Element{
			Name:     "", // Empty name indicates text fragment
			Children: children,
		}
		return container
	}

	// Simple case: no templates, just join the parts
	val := strings.Join(pt.Parts, "")
	return &ast.Text{Content: val}
}

// createTemplateExpr creates a TemplateExpr from a template string
func createTemplateExpr(expr string) *ast.TemplateExpr {
	return &ast.TemplateExpr{
		Expression: expr,
		IsCompiled: true, // Assume valid for now
	}
}

func convertElement(pe *PElement) ast.Node {
	// Convert attributes into a temporary slice
	attrs := []ast.Attr{}
	for i := 0; i < len(pe.Open.Parts); i++ {
		part := pe.Open.Parts[i]
		if part == "/" || strings.TrimSpace(part) == "" {
			continue
		}
		if i+2 < len(pe.Open.Parts) && pe.Open.Parts[i+1] == "=" {
			key := part
			val := pe.Open.Parts[i+2]
			i += 2
			attrs = append(attrs, ast.Attr{Key: key, Val: val})
			continue
		}
		if part == "=" {
			continue
		}
		attrs = append(attrs, ast.Attr{Key: part, Val: ""})
	}
	// If self-closing tag, return element node with attrs (no children)
	for _, p := range pe.Open.Parts {
		if p == "/" {
			e := &ast.Element{Name: pe.Open.Name, Attrs: attrs}
			return e
		}
	}
	// Build children
	children := []ast.Node{}
	for _, child := range pe.Children {
		cn := convertPNode(child)
		if cn != nil {
			children = append(children, cn)
		}
	}
	// Handle special nodes: if, for
	name := pe.Open.Name
	if name == "if" {
		cond := ""
		for _, a := range attrs {
			if a.Key == "cond" {
				cond = a.Val
				break
			}
		}
		inode := &ast.IfNode{Cond: cond, Children: children}
		return inode
	}
	if name == "for" {
		varName := ""
		inVal := ""
		for _, a := range attrs {
			if a.Key == "var" {
				varName = a.Val
			}
			if a.Key == "in" {
				inVal = a.Val
			}
		}
		fnode := &ast.ForNode{Var: varName, In: inVal, Children: children}
		return fnode
	}
	// Default element
	e := &ast.Element{Name: name, Attrs: attrs, Children: children}
	return e
}
