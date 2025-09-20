package poml

import (
	"fmt"
	"testing"

	"github.com/ZanzyTHEbar/poml/sdk/ast"
)

func TestParticipleParseSimple(t *testing.T) {
	input := "<poml><role>Assistant</role><task>Say hi</task></poml>"
	doc, err := ParseWithParticiple(input)
	if err != nil {
		t.Fatalf("ParseWithParticiple error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("expected 1 root element, got %d", len(doc.Children))
	}
	root := doc.Children[0].(*ast.Element)
	if root.Name != "poml" {
		t.Fatalf("expected root 'poml', got %s", root.Name)
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children under root, got %d", len(root.Children))
	}
}

func TestParticipleParseNestedAndText(t *testing.T) {
	input := "<poml><section><title>Hello</title><para>This is <b>bold</b> text.</para></section></poml>"
	doc, err := ParseWithParticiple(input)
	if err != nil {
		t.Fatalf("ParseWithParticiple error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("expected 1 root element, got %d", len(doc.Children))
	}
	root := doc.Children[0].(*ast.Element)
	if root.Name != "poml" {
		t.Fatalf("expected root 'poml', got %s", root.Name)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child under root, got %d", len(root.Children))
	}
	section := root.Children[0].(*ast.Element)
	if section.Name != "section" {
		t.Fatalf("expected 'section', got %s", section.Name)
	}
	if len(section.Children) != 2 {
		t.Fatalf("expected 2 children in section, got %d", len(section.Children))
	}
	title := section.Children[0].(*ast.Element)
	if title.Name != "title" {
		t.Fatalf("expected 'title', got %s", title.Name)
	}
	if len(title.Children) != 1 {
		t.Fatalf("expected 1 child in title, got %d", len(title.Children))
	}
	if text, ok := title.Children[0].(*ast.Text); !ok || text.Content != "Hello" {
		t.Fatalf("expected title text 'Hello', got %#v", title.Children[0])
	}
	para := section.Children[1].(*ast.Element)
	if para.Name != "para" {
		t.Fatalf("expected 'para', got %s", para.Name)
	}
	if len(para.Children) != 3 {
		var details string
		for i, c := range para.Children {
			switch n := c.(type) {
			case *ast.Text:
				details += fmt.Sprintf("[%d] Text(%q) ", i, n.Content)
			case *ast.Element:
				details += fmt.Sprintf("[%d] Elem(%s) ", i, n.Name)
			default:
				details += fmt.Sprintf("[%d] %T ", i, c)
			}
		}
		t.Fatalf("expected 3 children in para (text,bold,text), got %d; %s", len(para.Children), details)
	}
	if text, ok := para.Children[0].(*ast.Text); !ok || text.Content != "This is " {
		t.Fatalf("expected first para text 'This is ', got %#v", para.Children[0])
	}
	bold := para.Children[1].(*ast.Element)
	if bold.Name != "b" {
		t.Fatalf("expected 'b', got %s", bold.Name)
	}
	if len(bold.Children) != 1 {
		t.Fatalf("expected 1 child in bold, got %d", len(bold.Children))
	}
	if text, ok := bold.Children[0].(*ast.Text); !ok || text.Content != "bold" {
		t.Fatalf("expected bold text 'bold', got %#v", bold.Children[0])
	}
	if text, ok := para.Children[2].(*ast.Text); !ok || text.Content != " text." {
		t.Fatalf("expected last para text ' text.', got %#v", para.Children[2])
	}
}

func TestParticipleAttributesAndSelfClosing(t *testing.T) {
	input := `<img src="/path.png" alt="An image"/>`
	doc, err := ParseWithParticiple(input)
	if err != nil {
		t.Fatalf("ParseWithParticiple error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("expected 1 root element, got %d", len(doc.Children))
	}
	img := doc.Children[0].(*ast.Element)
	if img.Name != "img" {
		t.Fatalf("expected 'img', got %s", img.Name)
	}
	if len(img.Attrs) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(img.Attrs))
	}
	if img.Attrs[0].Key != "src" || img.Attrs[0].Val != `"/path.png"` {
		t.Fatalf("unexpected attr0: %#v", img.Attrs[0])
	}
	if img.Attrs[1].Key != "alt" || img.Attrs[1].Val != `"An image"` {
		t.Fatalf("unexpected attr1: %#v", img.Attrs[1])
	}
	if len(img.Children) != 0 {
		t.Fatalf("expected no children for self-closing img, got %d", len(img.Children))
	}
}

func TestWhitespaceAndTokenization(t *testing.T) {
	input := "<doc>Lead  and  multiple   spaces</doc>"
	doc, err := ParseWithParticiple(input)
	if err != nil {
		t.Fatalf("ParseWithParticiple error: %v", err)
	}
	root := doc.Children[0].(*ast.Element)
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child text node, got %d", len(root.Children))
	}
	if text, ok := root.Children[0].(*ast.Text); !ok || text.Content != "Lead  and  multiple   spaces" {
		t.Fatalf("unexpected whitespace handling: %q", root.Children[0])
	}
}

func TestIfAndForNodes(t *testing.T) {
	input := "<poml><if cond=\"user.is_admin\">Secret</if><for var=\"i\" in=\"items\"><item>{{i}}</item></for></poml>"
	doc, err := ParseWithParticiple(input)
	if err != nil {
		t.Fatalf("ParseWithParticiple error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("expected 1 root element, got %d", len(doc.Children))
	}
	root := doc.Children[0].(*ast.Element)
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children under root, got %d", len(root.Children))
	}
	ifnode, ok := root.Children[0].(*ast.IfNode)
	if !ok {
		t.Fatalf("expected IfNode, got %T", root.Children[0])
	}
	if ifnode.Cond != `"user.is_admin"` && ifnode.Cond != "user.is_admin" {
		t.Fatalf("unexpected if cond: %s", ifnode.Cond)
	}
	fornode, ok := root.Children[1].(*ast.ForNode)
	if !ok {
		t.Fatalf("expected ForNode, got %T", root.Children[1])
	}
	if fornode.Var != `"i"` && fornode.Var != "i" {
		t.Fatalf("unexpected for var: %s", fornode.Var)
	}
	if fornode.In != `"items"` && fornode.In != "items" {
		t.Fatalf("unexpected for in: %s", fornode.In)
	}
}
