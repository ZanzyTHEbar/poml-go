package golden

import (
	"testing"

	poml "github.com/ZanzyTHEbar/poml/sdk"
	writer "github.com/ZanzyTHEbar/poml/sdk/writer"
)

func TestSimpleEnv(t *testing.T) {
	// Test basic env wrapper functionality
	pomlContent := `<poml><env presentation="multimedia"><p>Test content</p></env></poml>`

	doc, err := poml.ParseWithParticiple(pomlContent)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	renderOpts := &writer.RenderOptions{}
	out, err := writer.RenderAST(doc, renderOpts)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	t.Logf("Output type: %T", out)
	t.Logf("Output: %+v", out)
}

func TestSimpleImg(t *testing.T) {
	// Test basic img with position
	pomlContent := `<poml><p><img src="test.jpg" position="top" alt="test" /></p></poml>`

	doc, err := poml.ParseWithParticiple(pomlContent)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	renderOpts := &writer.RenderOptions{}
	out, err := writer.RenderAST(doc, renderOpts)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	t.Logf("Output type: %T", out)
	t.Logf("Output: %+v", out)
}
