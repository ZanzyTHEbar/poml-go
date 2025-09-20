package writer

import (
	"testing"

	"github.com/ZanzyTHEbar/poml/sdk/ast"
)

func TestRenderAST_Template(t *testing.T) {
	// Create AST: <root>Hello {{name}}</root>
	doc := &ast.Document{}
	root := &ast.Element{Name: "root"}
	root.Children = append(root.Children, &ast.Text{Content: "Hello {{name}}"})
	doc.Children = append(doc.Children, root)

	opts := &RenderOptions{Context: map[string]interface{}{"name": "Alice"}}
	rc, err := RenderAST(doc, opts)
	if err != nil {
		t.Fatalf("RenderAST error: %v", err)
	}
	arr, ok := rc.([]interface{})
	if !ok || len(arr) != 1 {
		t.Fatalf("unexpected render result: %#v", rc)
	}
	if s, ok := arr[0].(string); !ok || s != "Hello Alice" {
		t.Fatalf("unexpected rendered content: %#v", arr[0])
	}
}

func TestWrite_SpeakerMode(t *testing.T) {
	doc := &ast.Document{}
	root := &ast.Element{Name: "root"}
	root.Children = append(root.Children, &ast.Text{Content: "Hi {{who}}"})
	doc.Children = append(doc.Children, root)

	opts := &RenderOptions{Context: map[string]interface{}{"who": "Bob"}}
	out, err := Write(doc, true, opts)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	msgs, ok := out.([]Message)
	if !ok || len(msgs) != 1 {
		t.Fatalf("unexpected messages: %#v", out)
	}
	if msgs[0].Content != "Hi Bob" {
		t.Fatalf("unexpected message content: %s", msgs[0].Content)
	}
}
