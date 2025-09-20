package runtime

import (
	"strings"
	"testing"
)

func TestEvaluateIf(t *testing.T) {
	ctx := map[string]interface{}{"x": 5}
	ok, err := EvaluateIf("x > 3", ctx)
	if err != nil {
		t.Fatalf("EvaluateIf error: %v", err)
	}
	if !ok {
		t.Fatalf("expected true, got false")
	}
}

func TestEvaluateFor(t *testing.T) {
	ctx := map[string]interface{}{"items": []interface{}{1, 2, 3}}
	arr, err := EvaluateFor("items", ctx)
	if err != nil {
		t.Fatalf("EvaluateFor error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 items, got %d", len(arr))
	}
}

// Template expression evaluation tests
func TestTemplateExpressionEvaluation(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		env      map[string]interface{}
		expected string
	}{
		{
			name:     "Simple template",
			template: "{{.Name}}",
			env:      map[string]interface{}{"Name": "World"},
			expected: "World",
		},
		{
			name:     "Template with functions",
			template: "{{.Text | trimSpace | upper}}",
			env:      map[string]interface{}{"Text": "  hello world  "},
			expected: "HELLO WORLD",
		},
		{
			name:     "Regex replacement",
			template: "{{.Text | replaceRegex \"^\\\\s+\" \"\"}}",
			env:      map[string]interface{}{"Text": "   hello"},
			expected: "hello", // Template should successfully remove leading whitespace
		},
		{
			name:     "Complex method chain",
			template: "{{.BlogPost | replaceRegex \"^\" \"  \" | trimSpace}}",
			env:      map[string]interface{}{"BlogPost": "Hello\nWorld"},
			expected: "Hello\nWorld", // Falls back to expr-lang evaluation
		},
		{
			name:     "Multiple expressions",
			template: "Name: {{.Name}}, Age: {{.Age}}",
			env:      map[string]interface{}{"Name": "Alice", "Age": 30},
			expected: "Name: Alice, Age: 30",
		},
		{
			name:     "String functions",
			template: "{{.Text | lower | trimSpace}}",
			env:      map[string]interface{}{"Text": "  HELLO WORLD  "},
			expected: "hello world",
		},
		{
			name:     "Length function",
			template: "Length: {{len .Items}}",
			env:      map[string]interface{}{"Items": []interface{}{1, 2, 3, 4}},
			expected: "Length: 4",
		},
		{
			name:     "Has prefix/suffix",
			template: "Prefix: {{if hasPrefix .Text \"Hello\"}}true{{else}}false{{end}}, Suffix: {{if hasSuffix .Text \"World\"}}true{{else}}false{{end}}",
			env:      map[string]interface{}{"Text": "Hello World"},
			expected: "Prefix: true, Suffix: true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := EvaluateTemplateExpression(tc.template, tc.env)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// Test fallback to expr-lang for complex expressions
func TestTemplateFallback(t *testing.T) {
	// Test an expression that should fall back to expr-lang
	template := "Result: {{x + y}}"
	env := map[string]interface{}{"x": 5, "y": 3}

	result, err := EvaluateTemplateExpression(template, env)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should fall back to expr-lang and evaluate x + y
	expected := "Result: 8"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Test template caching
func TestTemplateCaching(t *testing.T) {
	template := "{{.Name | upper}}"
	env := map[string]interface{}{"Name": "test"}

	// First call should cache the result
	result1, err := EvaluateTemplateExpression(template, env)
	if err != nil {
		t.Fatalf("First call error: %v", err)
	}

	// Second call should use cached result
	result2, err := EvaluateTemplateExpression(template, env)
	if err != nil {
		t.Fatalf("Second call error: %v", err)
	}

	if result1 != result2 {
		t.Errorf("Cached result mismatch: %q != %q", result1, result2)
	}

	if result1 != "TEST" {
		t.Errorf("Expected 'TEST', got %q", result1)
	}
}

// Test error handling in templates
func TestTemplateErrorHandling(t *testing.T) {
	// Test with invalid template syntax that should fall back to expr-lang
	template := "{{.Missing}} invalid_syntax_here"
	env := map[string]interface{}{"Missing": "value"}

	result, err := EvaluateTemplateExpression(template, env)
	// Should not return error, should fall back gracefully
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should fall back to expr-lang evaluation
	// The expr-lang will try to evaluate .Missing and leave invalid_syntax_here as-is
	if !strings.Contains(result, "value") {
		t.Errorf("Expected result to contain 'value', got %q", result)
	}
}
