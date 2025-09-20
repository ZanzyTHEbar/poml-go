package parser

import (
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/poml/sdk/ast"
)

// TestTemplateParserComprehensive provides comprehensive testing of all template parsing functionality
func TestTemplateParserComprehensive(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		setup    func(p *Parser)
		validate func(t *testing.T, doc *ast.Document)
	}{
		{
			name:  "Simple variable expression",
			input: "{{.Name}}",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "John", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 1 {
					t.Errorf("Expected 1 child, got %d", len(doc.Children))
					return
				}

				template, ok := doc.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[0])
					return
				}

				if template.Expression != ".Name" {
					t.Errorf("Expected expression '.Name', got '%s'", template.Expression)
				}

				if !template.IsCompiled {
					t.Errorf("Template should be compiled successfully")
				}

				if len(template.ResolvedVariables) != 1 || template.ResolvedVariables[0] != "Name" {
					t.Errorf("Expected resolved variable 'Name', got %v", template.ResolvedVariables)
				}

				if len(template.UnresolvedVariables) != 0 {
					t.Errorf("Expected no unresolved variables, got %v", template.UnresolvedVariables)
				}
			},
		},
		{
			name:  "Function call expression",
			input: "{{upper .Name}}",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "john", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 1 {
					t.Errorf("Expected 1 child, got %d", len(doc.Children))
					return
				}

				template, ok := doc.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[0])
					return
				}

				if template.Expression != "upper .Name" {
					t.Errorf("Expected expression 'upper .Name', got '%s'", template.Expression)
				}

				if !template.IsFunctionsValidated {
					t.Errorf("Functions should be validated successfully")
				}

				if len(template.ValidatedFunctions) != 1 || template.ValidatedFunctions[0] != "upper" {
					t.Errorf("Expected validated function 'upper', got %v", template.ValidatedFunctions)
				}
			},
		},
		{
			name:  "Pipeline expression",
			input: "{{.Name | upper | trimSpace}}",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "  john  ", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 1 {
					t.Errorf("Expected 1 child, got %d", len(doc.Children))
					return
				}

				template, ok := doc.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[0])
					return
				}

				if template.Expression != ".Name | upper | trimSpace" {
					t.Errorf("Expected pipeline expression, got '%s'", template.Expression)
				}

				if !template.HasPipeline {
					t.Errorf("Template should have pipeline flag set")
				}

				if !template.IsCompiled {
					t.Errorf("Pipeline template should compile successfully")
				}
			},
		},
		{
			name:  "Complex function with arguments",
			input: "{{replace .Text \"old\" \"new\"}}",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Text", "Hello old world", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 1 {
					t.Errorf("Expected 1 child, got %d", len(doc.Children))
					return
				}

				template, ok := doc.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[0])
					return
				}

				if template.Expression != "replace .Text \"old\" \"new\"" {
					t.Errorf("Expected function with arguments, got '%s'", template.Expression)
				}
			},
		},
		{
			name:  "Template with undefined variable",
			input: "{{.Undefined}}",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "John", "string") // Different variable
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 1 {
					t.Errorf("Expected 1 child, got %d", len(doc.Children))
					return
				}

				template, ok := doc.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[0])
					return
				}

				if template.Expression != ".Undefined" {
					t.Errorf("Expected expression '.Undefined', got '%s'", template.Expression)
				}

				if template.IsVariablesResolved {
					t.Errorf("Template with undefined variable should not be resolved")
				}

				if len(template.UnresolvedVariables) != 1 || template.UnresolvedVariables[0] != "Undefined" {
					t.Errorf("Expected unresolved variable 'Undefined', got %v", template.UnresolvedVariables)
				}
			},
		},
		{
			name:  "Template with undefined function",
			input: "{{.Name | undefinedFunction}}",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "John", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 1 {
					t.Errorf("Expected 1 child, got %d", len(doc.Children))
					return
				}

				template, ok := doc.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[0])
					return
				}

				if template.Expression != ".Name | undefinedFunction" {
					t.Errorf("Expected expression with undefined function, got '%s'", template.Expression)
				}

				if template.IsFunctionsValidated {
					t.Errorf("Template with undefined function should not be validated")
				}

				if len(template.InvalidFunctions) != 1 || template.InvalidFunctions[0] != "undefinedFunction" {
					t.Errorf("Expected invalid function 'undefinedFunction', got %v", template.InvalidFunctions)
				}
			},
		},
		{
			name:  "Template mixed with text",
			input: "Hello {{.Name}}, welcome!",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "John", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 3 {
					t.Errorf("Expected 3 children (text + template + text), got %d", len(doc.Children))
					return
				}

				// Check first text node
				text1, ok := doc.Children[0].(*ast.Text)
				if !ok {
					t.Errorf("Expected Text node, got %T", doc.Children[0])
					return
				}
				if text1.Content != "Hello " {
					t.Errorf("Expected 'Hello ', got '%s'", text1.Content)
				}

				// Check template node
				template, ok := doc.Children[1].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[1])
					return
				}
				if template.Expression != ".Name" {
					t.Errorf("Expected '.Name', got '%s'", template.Expression)
				}

				// Check second text node
				text2, ok := doc.Children[2].(*ast.Text)
				if !ok {
					t.Errorf("Expected Text node, got %T", doc.Children[2])
					return
				}
				if text2.Content != ", welcome!" {
					t.Errorf("Expected ', welcome!', got '%s'", text2.Content)
				}
			},
		},
		{
			name:  "Template in XML element",
			input: "<p>Hello {{.Name}}!</p>",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "John", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 1 {
					t.Errorf("Expected 1 child (element), got %d", len(doc.Children))
					return
				}

				element, ok := doc.Children[0].(*ast.Element)
				if !ok {
					t.Errorf("Expected Element, got %T", doc.Children[0])
					return
				}

				if element.Name != "p" {
					t.Errorf("Expected element name 'p', got '%s'", element.Name)
				}

				if len(element.Children) != 2 {
					t.Errorf("Expected 2 children in element, got %d", len(element.Children))
					return
				}

				// Check template child
				template, ok := element.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr in element, got %T", element.Children[0])
					return
				}

				if template.Expression != ".Name" {
					t.Errorf("Expected '.Name' in element, got '%s'", template.Expression)
				}
			},
		},
		{
			name:  "Multiple templates in document",
			input: "{{.Greeting}} {{.Name}}, {{.Message}}!",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Greeting", "Hello", "string")
				ctx.AddVariable("Name", "John", "string")
				ctx.AddVariable("Message", "welcome", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 6 {
					t.Errorf("Expected 6 children (template + text + template + text + template + text), got %d", len(doc.Children))
					return
				}

				// Check template expressions
				template1, ok := doc.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr at position 0, got %T", doc.Children[0])
					return
				}
				if template1.Expression != ".Greeting" {
					t.Errorf("Expected '.Greeting', got '%s'", template1.Expression)
				}

				template2, ok := doc.Children[2].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr at position 2, got %T", doc.Children[2])
					return
				}
				if template2.Expression != ".Name" {
					t.Errorf("Expected '.Name', got '%s'", template2.Expression)
				}

				template3, ok := doc.Children[4].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr at position 4, got %T", doc.Children[4])
					return
				}
				if template3.Expression != ".Message" {
					t.Errorf("Expected '.Message', got '%s'", template3.Expression)
				}
			},
		},
		{
			name:  "Template with source mapping",
			input: "Name: {{.Name}}",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "John", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 2 {
					t.Errorf("Expected 2 children, got %d", len(doc.Children))
					return
				}

				template, ok := doc.Children[1].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[1])
					return
				}

				// Check source mapping
				if template.StartIndex != 6 {
					t.Errorf("Expected StartIndex 6, got %d", template.StartIndex)
				}
				if template.EndIndex != 15 {
					t.Errorf("Expected EndIndex 15, got %d", template.EndIndex)
				}
				if template.StartLine != 1 {
					t.Errorf("Expected StartLine 1, got %d", template.StartLine)
				}
				if template.StartColumn != 7 {
					t.Errorf("Expected StartColumn 7, got %d", template.StartColumn)
				}
			},
		},
		{
			name:  "Template compilation error",
			input: "{{.Name | invalid syntax}}",
			setup: func(p *Parser) {
				ctx := NewVariableContext()
				ctx.AddVariable("Name", "John", "string")
				p.SetContext(ctx)
			},
			validate: func(t *testing.T, doc *ast.Document) {
				if len(doc.Children) != 1 {
					t.Errorf("Expected 1 child, got %d", len(doc.Children))
					return
				}

				template, ok := doc.Children[0].(*ast.TemplateExpr)
				if !ok {
					t.Errorf("Expected TemplateExpr, got %T", doc.Children[0])
					return
				}

				if template.Expression != ".Name | invalid syntax" {
					t.Errorf("Expected expression with invalid syntax, got '%s'", template.Expression)
				}

				if template.IsCompiled {
					t.Errorf("Template with invalid syntax should not compile")
				}

				if template.CompileError == "" {
					t.Errorf("Template with invalid syntax should have compile error")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser(tc.input)

			// Setup context if provided
			if tc.setup != nil {
				tc.setup(parser)
			}

			doc, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Validate the result
			tc.validate(t, doc)
		})
	}
}

// TestTemplateWriterIntegration tests the integration between parser and writer
func TestTemplateWriterIntegration(t *testing.T) {
	input := "Hello {{.Name}}, you have {{.Count}} messages!"
	parser := NewParser(input)
	doc, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Import the writer package to test rendering
	// This would require the writer package to be available
	// For now, we'll just verify the AST structure is correct

	if len(doc.Children) != 5 {
		t.Errorf("Expected 5 children (text + template + text + template + text), got %d", len(doc.Children))
	}

	// Verify template expressions are parsed correctly
	template1, ok := doc.Children[1].(*ast.TemplateExpr)
	if !ok {
		t.Errorf("Expected TemplateExpr at position 1, got %T", doc.Children[1])
		return
	}
	if template1.Expression != ".Name" {
		t.Errorf("Expected '.Name', got '%s'", template1.Expression)
	}

	template2, ok := doc.Children[3].(*ast.TemplateExpr)
	if !ok {
		t.Errorf("Expected TemplateExpr at position 3, got %T", doc.Children[3])
		return
	}
	if template2.Expression != ".Count" {
		t.Errorf("Expected '.Count', got '%s'", template2.Expression)
	}
}

// TestTemplateErrorHandling tests comprehensive error handling scenarios
func TestTemplateErrorHandling(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Unclosed template brace",
			input:         "{{.Name",
			expectError:   true,
			errorContains: "Unclosed template",
		},
		{
			name:          "Invalid pipe usage",
			input:         "{{| .Name}}",
			expectError:   true,
			errorContains: "cannot start with pipe",
		},
		{
			name:          "Consecutive pipes",
			input:         "{{.Name || upper}}",
			expectError:   true,
			errorContains: "consecutive pipe",
		},
		{
			name:          "Unmatched quotes",
			input:         "{{.Name | replace \"hello}}",
			expectError:   true,
			errorContains: "unmatched",
		},
		{
			name:          "Undefined function",
			input:         "{{.Name | nonexistent}}",
			expectError:   false, // This is a runtime error, not parse error
			errorContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser(tc.input)
			doc, err := parser.Parse()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected parse error, but got none")
				} else if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected parse error: %v", err)
				}
				if doc == nil {
					t.Errorf("Expected valid document, got nil")
				}
			}
		})
	}
}

// TestTemplatePerformance benchmarks template parsing performance
func BenchmarkTemplateParser(b *testing.B) {
	simpleTemplate := "{{.Name}}"
	complexTemplate := "{{.User.Name | upper}} {{.User.Age}} items for {{.User.Email | lower}}"
	mixedContent := "Hello {{.Name}}, you have {{.Count}} messages from {{.Sender | upper}}!"

	testCases := []struct {
		name  string
		input string
	}{
		{"Simple", simpleTemplate},
		{"Complex", complexTemplate},
		{"Mixed", mixedContent},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				parser := NewParser(tc.input)
				_, err := parser.Parse()
				if err != nil {
					b.Fatalf("Parse error: %v", err)
				}
			}
		})
	}
}
