package ast

import (
	"encoding/json"
	"testing"
)

func TestTemplateExpr(t *testing.T) {
	// Test TemplateExpr creation
	expr := TemplateExpr{
		Expression:   ".Name | upper",
		StartIndex:   10,
		EndIndex:     25,
		IRStartIndex: 5,
		IREndIndex:   15,
	}

	// Test Node interface compliance
	var node Node = &expr
	node.node() // Should not panic

	// Test JSON serialization
	data, err := json.Marshal(&expr)
	if err != nil {
		t.Fatalf("Failed to marshal TemplateExpr: %v", err)
	}

	// Test JSON deserialization
	var decoded TemplateExpr
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal TemplateExpr: %v", err)
	}

	// Verify fields
	if decoded.Expression != expr.Expression {
		t.Errorf("Expected expression %q, got %q", expr.Expression, decoded.Expression)
	}
	if decoded.StartIndex != expr.StartIndex {
		t.Errorf("Expected StartIndex %d, got %d", expr.StartIndex, decoded.StartIndex)
	}
	if decoded.EndIndex != expr.EndIndex {
		t.Errorf("Expected EndIndex %d, got %d", expr.EndIndex, decoded.EndIndex)
	}
	if decoded.IRStartIndex != expr.IRStartIndex {
		t.Errorf("Expected IRStartIndex %d, got %d", expr.IRStartIndex, decoded.IRStartIndex)
	}
	if decoded.IREndIndex != expr.IREndIndex {
		t.Errorf("Expected IREndIndex %d, got %d", expr.IREndIndex, decoded.IREndIndex)
	}

	// Test JSON structure
	expectedJSON := `{"expression":".Name | upper","parsed":null,"hasPipeline":false,"isCompiled":false,"compileError":"","resolvedVariables":[],"unresolvedVariables":[],"isVariablesResolved":false,"variableResolutionError":"","validatedFunctions":[],"invalidFunctions":[],"isFunctionsValidated":false,"functionValidationError":"","functionValidationDetails":{},"startIndex":10,"endIndex":25,"irStartIndex":5,"irEndIndex":15,"startLine":1,"startColumn":11,"endLine":1,"endColumn":26,"sourceText":""}`
	actualJSON := string(data)
	if actualJSON != expectedJSON {
		t.Errorf("Expected JSON %q, got %q", expectedJSON, actualJSON)
	}
}

func TestTemplateExprInDocument(t *testing.T) {
	// Test TemplateExpr as part of a document
	textNode := Text{Content: "Hello "}
	templateNode := TemplateExpr{
		Expression: ".Name",
		StartIndex: 6,
		EndIndex:   14,
	}
	textNode2 := Text{Content: "!"}

	doc := Document{
		Children: []Node{&textNode, &templateNode, &textNode2},
	}

	// Verify document structure
	if len(doc.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(doc.Children))
	}

	// Check first child (Text)
	if text, ok := doc.Children[0].(*Text); ok {
		if text.Content != "Hello " {
			t.Errorf("Expected first text content 'Hello ', got %q", text.Content)
		}
	} else {
		t.Errorf("Expected first child to be Text, got %T", doc.Children[0])
	}

	// Check second child (TemplateExpr)
	if template, ok := doc.Children[1].(*TemplateExpr); ok {
		if template.Expression != ".Name" {
			t.Errorf("Expected template expression '.Name', got %q", template.Expression)
		}
		if template.StartIndex != 6 {
			t.Errorf("Expected StartIndex 6, got %d", template.StartIndex)
		}
	} else {
		t.Errorf("Expected second child to be TemplateExpr, got %T", doc.Children[1])
	}

	// Check third child (Text)
	if text, ok := doc.Children[2].(*Text); ok {
		if text.Content != "!" {
			t.Errorf("Expected third text content '!', got %q", text.Content)
		}
	} else {
		t.Errorf("Expected third child to be Text, got %T", doc.Children[2])
	}
}
