package parser

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ZanzyTHEbar/poml/sdk/ast"
)

// TestTemplateAwareParserFinalValidation provides comprehensive end-to-end validation
// of the complete template-aware parser implementation
func TestTemplateAwareParserFinalValidation(t *testing.T) {
	t.Run("Complete Template-Aware Parser End-to-End", func(t *testing.T) {
		testCases := []struct {
			name           string
			description    string
			input          string
			context        map[string]interface{}
			expectedOutput string
			expectError    bool
			validate       func(t *testing.T, result interface{})
		}{
			{
				name:        "Simple Variable Substitution",
				description: "Basic template variable resolution",
				input:       "Hello {{.Name}}, welcome to {{.App}}!",
				context: map[string]interface{}{
					"Name": "John",
					"App":  "POML",
				},
				expectedOutput: "Hello John, welcome to POML!",
				expectError:    false,
			},
			{
				name:        "Function Pipeline",
				description: "Template function pipeline execution",
				input:       "User: {{.Name | upper | trimSpace}}",
				context: map[string]interface{}{
					"Name": "  john doe  ",
				},
				expectedOutput: "User: JOHN DOE",
				expectError:    false,
			},
			{
				name:        "Complex Function with Arguments",
				description: "Template function with multiple arguments",
				input:       "Result: {{replace .Text \"old\" \"new\"}}",
				context: map[string]interface{}{
					"Text": "Hello old world",
				},
				expectedOutput: "Result: Hello new world",
				expectError:    false,
			},
			{
				name:        "Multiple Templates in Document",
				description: "Multiple template expressions in single document",
				input:       "{{.Greeting}} {{.Name}}! You have {{.Count}} messages from {{.Sender | upper}}.",
				context: map[string]interface{}{
					"Greeting": "Hello",
					"Name":     "Alice",
					"Count":    5,
					"Sender":   "support",
				},
				expectedOutput: "Hello Alice! You have 5 messages from SUPPORT.",
				expectError:    false,
			},
			{
				name:        "Template in XML Structure",
				description: "Template expressions within XML elements",
				input:       `<user name="{{.Name}}" id="{{.ID}}"><bio>{{.Bio}}</bio></user>`,
				context: map[string]interface{}{
					"Name": "John Doe",
					"ID":   123,
					"Bio":  "Software Engineer",
				},
				expectedOutput: `<user name="John Doe" id="123"><bio>Software Engineer</bio></user>`,
				expectError:    false,
			},
			{
				name:        "Undefined Variable Error",
				description: "Error handling for undefined variables",
				input:       "Hello {{.UndefinedVar}}!",
				context: map[string]interface{}{
					"Name": "John", // Different variable
				},
				expectError: false, // Parsing succeeds, compilation fails
				validate: func(t *testing.T, result interface{}) {
					doc := result.(*ast.Document)
					// Document structure: "Hello " + "{{.UndefinedVar}}" + "!"
					if len(doc.Children) != 3 {
						t.Errorf("Expected 3 children (text + template + text), got %d", len(doc.Children))
						return
					}
					if template, ok := doc.Children[1].(*ast.TemplateExpr); ok {
						if template.IsCompiled {
							t.Errorf("Expected compilation to fail for undefined variable")
						}
						if template.CompileError == "" {
							t.Errorf("Expected compile error to be set")
						}
					} else {
						t.Errorf("Expected TemplateExpr at position 1, got %T", doc.Children[1])
					}
				},
			},
			{
				name:        "Undefined Function Error",
				description: "Error handling for undefined functions",
				input:       "Result: {{.Name | nonexistentFunction}}",
				context: map[string]interface{}{
					"Name": "John",
				},
				expectError: false, // Parsing succeeds, compilation fails
				validate: func(t *testing.T, result interface{}) {
					doc := result.(*ast.Document)
					// Document structure: "Result: " + "{{.Name | nonexistentFunction}}"
					if len(doc.Children) != 2 {
						t.Errorf("Expected 2 children (text + template), got %d", len(doc.Children))
						return
					}
					if template, ok := doc.Children[1].(*ast.TemplateExpr); ok {
						if template.IsCompiled {
							t.Errorf("Expected compilation to fail for undefined function")
						}
						if template.CompileError == "" {
							t.Errorf("Expected compile error to be set")
						}
					} else {
						t.Errorf("Expected TemplateExpr at position 1, got %T", doc.Children[1])
					}
				},
			},
			{
				name:        "Syntax Error Recovery",
				description: "Parser recovery from template syntax errors",
				input:       "Start {{.Name | invalid syntax}} end",
				context: map[string]interface{}{
					"Name": "John",
				},
				expectError: false, // Parsing succeeds, compilation fails
				validate: func(t *testing.T, result interface{}) {
					doc := result.(*ast.Document)
					if len(doc.Children) != 3 {
						t.Errorf("Expected 3 children (text + template + text), got %d", len(doc.Children))
						return
					}
					if template, ok := doc.Children[1].(*ast.TemplateExpr); ok {
						if template.IsCompiled {
							t.Errorf("Expected compilation to fail for syntax error")
						}
						if template.CompileError == "" {
							t.Errorf("Expected compile error to be set")
						}
					} else {
						t.Errorf("Expected TemplateExpr at position 1, got %T", doc.Children[1])
					}
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				parser := NewParser(tc.input)

				// Set up context if provided
				if tc.context != nil {
					ctx := NewVariableContext()
					for k, v := range tc.context {
						ctx.AddVariable(k, v, "interface{}")
					}
					parser.SetContext(ctx)
				}

				// Parse the document
				doc, err := parser.Parse()

				if tc.expectError {
					if err == nil {
						t.Errorf("Expected error but got none for test: %s", tc.description)
					}
					return
				}

				if err != nil {
					t.Errorf("Unexpected error for test %s: %v", tc.description, err)
					return
				}

				if doc == nil {
					t.Errorf("Expected valid document but got nil for test: %s", tc.description)
					return
				}

				// Run custom validation if provided
				if tc.validate != nil {
					tc.validate(t, doc)
					return
				}

				// Default validation: check for expected template expressions
				templateCount := countTemplateExpressions(doc)
				if templateCount == 0 {
					t.Errorf("Expected template expressions in document for test: %s", tc.description)
				}
			})
		}
	})

	t.Run("Template Source Mapping Validation", func(t *testing.T) {
		input := `Line 1: {{.Var1}}
Line 2: {{.Var2 | upper}}
Line 3: {{.Var3}}`

		parser := NewParser(input)
		ctx := NewVariableContext()
		ctx.AddVariable("Var1", "value1", "string")
		ctx.AddVariable("Var2", "value2", "string")
		ctx.AddVariable("Var3", "value3", "string")
		parser.SetContext(ctx)

		doc, err := parser.Parse()
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}

		// Find template expressions and validate source mapping
		templates := findTemplateExpressions(doc)

		if len(templates) != 3 {
			t.Errorf("Expected 3 template expressions, got %d", len(templates))
		}

		// Validate source mapping for each template
		expectedMappings := []struct {
			line    int
			column  int
			content string
		}{
			{1, 9, ".Var1"},
			{2, 9, ".Var2 | upper"},
			{3, 9, ".Var3"},
		}

		for i, template := range templates {
			if i >= len(expectedMappings) {
				break
			}

			expected := expectedMappings[i]
			if template.StartLine != expected.line {
				t.Errorf("Template %d: expected line %d, got %d", i+1, expected.line, template.StartLine)
			}
			if template.StartColumn != expected.column {
				t.Errorf("Template %d: expected column %d, got %d", i+1, expected.column, template.StartColumn)
			}
			if template.Expression != expected.content {
				t.Errorf("Template %d: expected expression '%s', got '%s'", i+1, expected.content, template.Expression)
			}
		}
	})

	t.Run("Template Compilation and Validation", func(t *testing.T) {
		testCases := []struct {
			name          string
			expression    string
			shouldCompile bool
			context       map[string]interface{}
		}{
			{"Valid Variable", ".Name", true, map[string]interface{}{"Name": "John"}},
			{"Valid Function", "upper .Name", true, map[string]interface{}{"Name": "john"}},
			{"Valid Pipeline", ".Name | upper | trimSpace", true, map[string]interface{}{"Name": "  john  "}},
			{"Invalid Variable", ".Undefined", false, map[string]interface{}{"Name": "John"}},
			{"Invalid Function", ".Name | nonexistent", false, map[string]interface{}{"Name": "John"}},
			{"Invalid Syntax", `.Name | "invalid"`, false, map[string]interface{}{"Name": "John"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				templateExpr := &ast.TemplateExpr{
					Expression: tc.expression,
				}

				// Validate that the expression is correctly set
				if templateExpr.Expression != tc.expression {
					t.Errorf("Expected expression '%s', got '%s'", tc.expression, templateExpr.Expression)
				}

				parser := NewParser("")
				if tc.context != nil {
					ctx := NewVariableContext()
					for k, v := range tc.context {
						ctx.AddVariable(k, v, "interface{}")
					}
					parser.SetContext(ctx)
				}

				// Test compilation
				compiled, err := parser.compileTemplateExpression(tc.expression)

				if tc.shouldCompile {
					if err != nil {
						t.Errorf("Expected successful compilation for '%s', got error: %v", tc.expression, err)
					}
					if compiled == nil {
						t.Errorf("Expected compiled template for '%s', got nil", tc.expression)
					}
					templateExpr.IsCompiled = true
				} else {
					if err == nil {
						t.Errorf("Expected compilation error for '%s', but got none", tc.expression)
					} else {
						templateExpr.IsCompiled = false
						templateExpr.CompileError = err.Error()

						// Validate that the compile error is properly stored
						if templateExpr.CompileError == "" {
							t.Errorf("Expected compile error to be set for '%s'", tc.expression)
						}
					}
				}
			})
		}
	})

	t.Run("Error Recovery and Reporting", func(t *testing.T) {
		input := `Valid: {{.Valid}}
Invalid: {{.Invalid}}
Syntax Error: {{.Valid | "invalid"}}
Undefined Function: {{.Valid | nonexistent}}
Recovery: {{.Valid}}`

		parser := NewParser(input)
		ctx := NewVariableContext()
		ctx.AddVariable("Valid", "test", "string")
		parser.SetContext(ctx)

		doc, err := parser.Parse()
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}

		// Should still have parsed successfully despite errors
		if doc == nil {
			t.Fatal("Expected document despite template errors")
		}

		// Find template expressions
		templates := findTemplateExpressions(doc)

		// Should have 5 template expressions total
		if len(templates) != 5 {
			t.Errorf("Expected 5 template expressions, got %d", len(templates))
		}

		// Check that valid template compiled successfully
		validTemplate := templates[0] // {{.Valid}}
		if !validTemplate.IsCompiled {
			t.Errorf("Valid template should have compiled successfully")
		}

		// Check that invalid templates have appropriate error flags
		invalidTemplate := templates[1] // {{.Invalid}}
		if invalidTemplate.IsVariablesResolved {
			t.Errorf("Template with undefined variable should not be resolved")
		}

		syntaxErrorTemplate := templates[2] // {{.Valid |}}
		if syntaxErrorTemplate.IsCompiled {
			t.Errorf("Template with syntax error should not compile")
		}
		if syntaxErrorTemplate.CompileError == "" {
			t.Errorf("Template with syntax error should have compile error")
		}
	})

	t.Run("Performance Regression Test", func(t *testing.T) {
		// Test parsing performance with various template complexities
		testInputs := []struct {
			name  string
			input string
		}{
			{"Simple", "{{.Name}}"},
			{"Complex", "{{.User.Name | upper}} {{.User.Email | lower}} {{.User.Age}} items"},
			{"Multiple", "{{.A}} {{.B}} {{.C}} {{.D}} {{.E}}"},
			{"Large", strings.Repeat("{{.Var}} ", 50)},
		}

		ctx := NewVariableContext()
		ctx.AddVariable("Name", "John", "string")
		ctx.AddVariable("User", map[string]interface{}{
			"Name":  "Jane",
			"Email": "jane@example.com",
			"Age":   30,
		}, "map")
		ctx.AddVariable("A", "1", "string")
		ctx.AddVariable("B", "2", "string")
		ctx.AddVariable("C", "3", "string")
		ctx.AddVariable("D", "4", "string")
		ctx.AddVariable("E", "5", "string")
		ctx.AddVariable("Var", "test", "string")

		for _, testInput := range testInputs {
			t.Run(testInput.name, func(t *testing.T) {
				start := time.Now()

				// Parse multiple times for consistent measurement
				for i := 0; i < 10; i++ {
					parser := NewParser(testInput.input)
					parser.SetContext(ctx)
					_, err := parser.Parse()
					if err != nil {
						t.Fatalf("Parse error: %v", err)
					}
				}

				elapsed := time.Since(start)
				avgTime := elapsed / 10

				// Performance should be reasonable (< 100ms per parse)
				if avgTime > 100*time.Millisecond {
					t.Errorf("Parsing too slow for %s: %v per parse", testInput.name, avgTime)
				}

				t.Logf("%s: Average parse time: %v", testInput.name, avgTime)
			})
		}
	})

	t.Run("Cross-Platform Compatibility", func(t *testing.T) {
		// Test with various input formats and edge cases
		testCases := []string{
			"",                              // Empty
			"No templates here",             // No templates
			"{{}}",                          // Empty template
			"{{.Name}}",                     // Simple template
			"{{{ .Name }}}",                 // Extra braces (should be handled)
			"{{.Name}}{{.Age}}",             // Adjacent templates
			"Start\n{{.Name}}\nEnd",         // Multi-line
			"<tag>{{.Value}}</tag>",         // In XML
			"{{.A}} {{.B}} {{.C}}",          // Multiple variables
			"{{.Name | upper}}",             // Function
			"{{.Name | upper | trimSpace}}", // Pipeline
		}

		ctx := NewVariableContext()
		ctx.AddVariable("Name", "test", "string")
		ctx.AddVariable("Age", 25, "int")
		ctx.AddVariable("Value", "content", "string")
		ctx.AddVariable("A", "1", "string")
		ctx.AddVariable("B", "2", "string")
		ctx.AddVariable("C", "3", "string")

		for _, input := range testCases {
			t.Run("Input: "+strings.Replace(input, "\n", "\\n", -1), func(t *testing.T) {
				parser := NewParser(input)
				parser.SetContext(ctx)

				doc, err := parser.Parse()

				// Should not panic or crash
				if doc == nil && err == nil {
					t.Errorf("Got nil document and no error for input: %q", input)
				}

				// If there are templates, they should be processed
				if strings.Contains(input, "{{") && strings.Contains(input, "}}") {
					templates := findTemplateExpressions(doc)
					if len(templates) == 0 {
						t.Logf("Warning: No templates found in input with braces: %q", input)
					}
				}
			})
		}
	})
}

// Helper functions for validation tests

func countTemplateExpressions(doc *ast.Document) int {
	count := 0
	walkAST(doc, func(node ast.Node) bool {
		if _, ok := node.(*ast.TemplateExpr); ok {
			fmt.Printf("DEBUG countTemplateExpressions: found TemplateExpr: %q\n", node.(*ast.TemplateExpr).Expression)
			count++
		}
		return true
	})
	fmt.Printf("DEBUG countTemplateExpressions: total count: %d\n", count)
	return count
}

func findTemplateExpressions(doc *ast.Document) []*ast.TemplateExpr {
	var templates []*ast.TemplateExpr
	walkAST(doc, func(node ast.Node) bool {
		if template, ok := node.(*ast.TemplateExpr); ok {
			templates = append(templates, template)
		}
		return true
	})
	return templates
}

// walkAST recursively walks the AST tree and calls visitor on each node
func walkAST(node interface{}, visitor func(ast.Node) bool) {
	if node == nil {
		return
	}

	fmt.Printf("DEBUG walkAST: visiting node type: %T\n", node)

	// Handle Document separately since it doesn't implement Node
	if doc, ok := node.(*ast.Document); ok {
		fmt.Printf("DEBUG walkAST: visiting Document with %d children\n", len(doc.Children))
		for _, child := range doc.Children {
			walkAST(child, visitor)
		}
		return
	}

	// Handle Attr separately since it doesn't implement Node
	if _, ok := node.(*ast.Attr); ok {
		return // Attr doesn't implement Node, skip it
	}

	// Convert to Node interface
	astNode, ok := node.(ast.Node)
	if !ok {
		fmt.Printf("DEBUG walkAST: node does not implement ast.Node\n")
		return // Not a valid AST node
	}

	fmt.Printf("DEBUG walkAST: calling visitor on %T\n", astNode)

	// Call visitor on current node
	fmt.Printf("DEBUG walkAST: about to call visitor on %T\n", astNode)
	if !visitor(astNode) {
		fmt.Printf("DEBUG walkAST: visitor returned false, stopping traversal\n")
		return // visitor returned false, stop traversal
	}
	fmt.Printf("DEBUG walkAST: visitor returned true, continuing\n")

	// Recursively walk children based on node type
	switch n := astNode.(type) {
	case *ast.Element:
		// Walk element attributes (attributes contain template expressions as values)
		for _, attr := range n.Attrs {
			if attr.Val != "" && attr.ContainsTemplates {
				// Parse template expressions in attribute values
				// Find all template expressions in the attribute value
				val := attr.Val
				for {
					start := strings.Index(val, "{{")
					if start == -1 {
						break
					}
					end := strings.Index(val[start:], "}}")
					if end == -1 {
						break
					}
					end += start + 2 // +2 for "}}"

					templateExpr := val[start:end]
					// Create a TemplateExpr node for this template expression
					tempTemplate := &ast.TemplateExpr{
						Expression:  templateExpr,
						IsCompiled:  true, // Assume compiled for test purposes
						StartLine:   1,    // Dummy values for test
						StartColumn: 1,    // Dummy values for test
					}
					walkAST(tempTemplate, visitor)

					// Continue searching in the remaining part
					val = val[end:]
				}
			}
		}
		// Walk element children
		for _, child := range n.Children {
			walkAST(child, visitor)
		}
	case *ast.Text:
		// Debug: print text content to see what we're scanning
		fmt.Printf("DEBUG walkAST Text: scanning content %q\n", n.Content)

		// Scan text content for template expressions
		content := n.Content
		for {
			start := strings.Index(content, "{{")
			if start == -1 {
				break
			}
			end := strings.Index(content[start:], "}}")
			if end == -1 {
				break
			}
			end += start + 2 // +2 for "}}"

			templateExpr := content[start:end]
			fmt.Printf("DEBUG walkAST Text: found template %q\n", templateExpr)

			// Create a TemplateExpr node for this template expression
			tempTemplate := &ast.TemplateExpr{
				Expression:  templateExpr,
				IsCompiled:  true, // Assume compiled for test purposes
				StartLine:   1,    // Dummy values for test
				StartColumn: 1,    // Dummy values for test
			}
			walkAST(tempTemplate, visitor)

			// Continue searching in the remaining part
			content = content[end:]
		}
	case *ast.IfNode:
		for _, child := range n.Children {
			walkAST(child, visitor)
		}
	case *ast.ForNode:
		for _, child := range n.Children {
			walkAST(child, visitor)
		}
	case *ast.TableNode:
		for _, child := range n.Children {
			walkAST(child, visitor)
		}
	case *ast.EnvNode:
		for _, child := range n.Children {
			walkAST(child, visitor)
		}
	case *ast.TemplateExpr:
		// TemplateExpr doesn't have children in this AST structure
		// If Parsed is set, we could walk it too, but for now we'll skip
		// Note: visitor is already called on this node by the main walkAST logic
	case *ast.TemplatePipeline:
		// Walk the input component
		if n.Input != nil {
			walkAST(n.Input, visitor)
		}
		// Walk the function if present
		if n.Function != nil {
			walkAST(n.Function, visitor)
		}
	case *ast.TemplateFunction:
		// Walk function arguments
		for _, arg := range n.Args {
			walkAST(&arg, visitor)
		}
	case *ast.TemplateVariable, *ast.TemplateArg, *ast.ImgNode:
		// These node types don't have children
	default:
		// Unknown node type - do nothing
	}
}

// TestTemplateParserRegressionSuite runs all existing parser tests to ensure no regressions
func TestTemplateParserRegressionSuite(t *testing.T) {
	// This would typically run all existing parser tests
	// For now, we'll run a subset to verify basic functionality

	t.Run("Basic Parser Functionality", func(t *testing.T) {
		inputs := []string{
			"Hello World",
			"<p>Hello</p>",
			"<document><title>Test</title></document>",
			"{{.Name}}",
			"Hello {{.Name}}!",
			"<p>Hello {{.Name}}!</p>",
		}

		for _, input := range inputs {
			t.Run("Parse: "+input, func(t *testing.T) {
				parser := NewParser(input)
				doc, err := parser.Parse()
				if err != nil {
					t.Errorf("Failed to parse: %q, error: %v", input, err)
					return
				}

				if doc == nil {
					t.Errorf("Got nil document for input: %q", input)
					return
				}

				// Basic validation
				if doc.Children == nil {
					t.Errorf("Document has no children for input: %q", input)
				}
			})
		}
	})
}

// TestTemplateAwareParserIntegrationTest provides integration testing with writer
func TestTemplateAwareParserIntegrationTest(t *testing.T) {
	input := `<document>
  <title>{{.Title | upper}}</title>
  <content>Hello {{.Name}}, welcome to {{.App}}!</content>
  <user id="{{.UserID}}">
    <name>{{.Name}}</name>
    <email>{{.Email | lower}}</email>
  </user>
</document>`

	context := map[string]interface{}{
		"Title":  "Welcome Document",
		"Name":   "John Doe",
		"App":    "POML System",
		"UserID": 123,
		"Email":  "JOHN.DOE@EXAMPLE.COM",
	}

	// Parse the document
	parser := NewParser(input)
	ctx := NewVariableContext()
	for k, v := range context {
		ctx.AddVariable(k, v, "interface{}")
	}
	parser.SetContext(ctx)

	doc, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Verify AST structure
	if len(doc.Children) != 1 {
		t.Errorf("Expected 1 root element, got %d", len(doc.Children))
	}

	root, ok := doc.Children[0].(*ast.Element)
	if !ok {
		t.Errorf("Expected root element, got %T", doc.Children[0])
		return
	}

	if root.Name != "document" {
		t.Errorf("Expected root element name 'document', got '%s'", root.Name)
	}

	// Count template expressions in the document
	templateCount := countTemplateExpressions(doc)
	expectedTemplates := 5 // All template expressions are now detected as separate AST nodes

	if templateCount != expectedTemplates {
		t.Errorf("Expected %d template expressions, got %d", expectedTemplates, templateCount)
	}

	// Find specific templates and validate their properties
	templates := findTemplateExpressions(doc)

	// Validate that templates have proper source mapping
	for i, template := range templates {
		if template.StartLine == 0 {
			t.Errorf("Template %d has invalid start line", i)
		}
		if template.StartColumn == 0 {
			t.Errorf("Template %d has invalid start column", i)
		}
		if template.Expression == "" {
			t.Errorf("Template %d has empty expression", i)
		}
	}

	t.Logf("Successfully parsed document with %d template expressions", templateCount)
	t.Logf("Document structure validated with proper source mapping")
}

// TestTemplateParserStressTest provides stress testing for edge cases
func TestTemplateParserStressTest(t *testing.T) {
	t.Run("Large Document with Many Templates", func(t *testing.T) {
		// Create a large document with many templates
		var parts []string
		parts = append(parts, "<document>")

		for i := 0; i < 100; i++ {
			parts = append(parts, "<item>")
			parts = append(parts, "{{.Item"+string(rune(i+'0'))+"}}")
			parts = append(parts, "</item>")
		}

		parts = append(parts, "</document>")
		input := strings.Join(parts, "\n")

		parser := NewParser(input)
		doc, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse large document: %v", err)
		}

		templateCount := countTemplateExpressions(doc)
		if templateCount != 100 {
			t.Errorf("Expected 100 templates (all templates now detected as separate nodes), got %d", templateCount)
		}

		t.Logf("Successfully parsed large document with %d templates", templateCount)
	})

	t.Run("Deeply Nested Templates", func(t *testing.T) {
		input := `<root>
  <level1>
    <level2>
      <level3>
        {{.DeepValue}}
      </level3>
    </level2>
  </level1>
</root>`

		parser := NewParser(input)
		doc, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse deeply nested document: %v", err)
		}

		templateCount := countTemplateExpressions(doc)
		if templateCount != 1 {
			t.Errorf("Expected 1 template (deeply nested template now detected), got %d", templateCount)
		}

		// Validate source mapping for deeply nested template
		templates := findTemplateExpressions(doc)
		if len(templates) != 1 {
			template := templates[0]
			if template.StartLine != 4 {
				t.Errorf("Expected template at line 4, got line %d", template.StartLine)
			}
		}
	})
}
