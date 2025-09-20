//go:build no_participle_tags
// +build no_participle_tags

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
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_-]*`},
		{Name: "WS", Pattern: `\s+`},
		{Name: "Text", Pattern: `[^<>]+`},
	}
	return lexer.MustSimple(rules)
}

// Grammar types with standard struct tags for linting
type PDocument struct {
	Leading  []*PNode  `json:"leading"`
	Root     *PElement `json:"root,omitempty"`
	Trailing []*PNode  `json:"trailing"`
}

type PNode struct {
	Element *PElement `json:"element,omitempty"`
	Text    *PText    `json:"text,omitempty"`
}

type PText struct {
	Parts []string `json:"parts"`
}

type PElement struct {
	Open     *POpen   `json:"open"`
	Children []*PNode `json:"children"`
	Close    *PClose  `json:"close,omitempty"`
}

type POpen struct {
	L     string   `json:"l"`
	Name  string   `json:"name"`
	Parts []string `json:"parts"`
	R     string   `json:"r"`
}

type PClose struct {
	L    string `json:"l"`
	_    *string
	S    string `json:"s"`
	_2   *string
	Name string `json:"name"`
	_3   *string
	R    string `json:"r"`
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

// convertPNode converts a PNode to AST node.
func convertPNode(pn *PNode) ast.Node {
	if pn.Element != nil {
		return convertElement(pn.Element)
	}
	if pn.Text != nil {
		return &ast.Text{Content: strings.Join(pn.Text.Parts, "")}
	}
	return nil
}

// convertElement converts a PElement to AST element.
func convertElement(pe *PElement) *ast.Element {
	el := &ast.Element{
		Name: pe.Open.Name,
	}
	for _, attr := range pe.Open.Parts {
		if strings.Contains(attr, "=") {
			parts := strings.SplitN(attr, "=", 2)
			if len(parts) == 2 {
				el.Attrs = append(el.Attrs, ast.Attr{
					Key: strings.TrimSpace(parts[0]),
					Val: strings.Trim(parts[1], "\""),
				})
			}
		}
	}
	for _, child := range pe.Children {
		if n := convertPNode(child); n != nil {
			el.Children = append(el.Children, n)
		}
	}
	return el
}

// ParseWithTrace runs participle parse with trace enabled and returns the trace output.
func ParseWithTrace(input string) (string, error) {
	var b strings.Builder
	_, err := pomlParser.ParseString("", input, participle.Trace(&b))
	return b.String(), err
}

// ParseParticipleTest is a test function for parsing validation.
func ParseParticipleTest(input string) (*PDocument, error) {
	var b strings.Builder
	parsed, err := pomlParser.ParseString("", input, participle.Trace(&b))
	if err != nil {
		return nil, fmt.Errorf("parse error: %v\ntrace:\n%s", err, b.String())
	}
	return parsed, nil
}
