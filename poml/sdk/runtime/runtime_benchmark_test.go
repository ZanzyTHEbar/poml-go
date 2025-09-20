package runtime

import (
	"testing"
)

// BenchmarkTemplateVsExprLang compares performance of template vs expr-lang evaluation
func BenchmarkTemplateVsExprLang(b *testing.B) {
	testCases := []struct {
		name     string
		template string
		env      map[string]interface{}
	}{
		{
			name:     "Simple",
			template: "{{.Name}}",
			env:      map[string]interface{}{"Name": "test"},
		},
		{
			name:     "ComplexTemplate",
			template: "{{.Data | replace \"old\" \"new\" | trimSpace | upper}}",
			env:      map[string]interface{}{"Data": "  old data  "},
		},
		{
			name:     "RegexTemplate",
			template: "{{.Text | replaceRegex \"\\s+\" \" \" | trimSpace}}",
			env:      map[string]interface{}{"Text": "  multiple   spaces  "},
		},
		{
			name:     "MethodChainTemplate",
			template: "{{.BlogPost | replaceRegex \"^\" \"  \" | trimSpace}}",
			env:      map[string]interface{}{"BlogPost": "line1\nline2"},
		},
		{
			name:     "FallbackExpr",
			template: "{{x + y}}",
			env:      map[string]interface{}{"x": 5, "y": 3},
		},
	}

	for _, tc := range testCases {
		b.Run("Template_"+tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = EvaluateTemplateExpression(tc.template, tc.env)
			}
		})

		// Compare with expr-lang equivalent (if applicable)
		exprEquivalent := convertToExprLang(tc.template)
		if exprEquivalent != "" {
			b.Run("ExprLang_"+tc.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, _ = evaluateWithExprLang(exprEquivalent, tc.env)
				}
			})
		}
	}
}

// convertToExprLang converts template syntax to expr-lang syntax for comparison
func convertToExprLang(template string) string {
	conversions := map[string]string{
		"{{.Name}}":             "Name",
		"{{.Data | trimSpace}}": "Data.trimSpace()",
		"{{.Text | upper}}":     "Text.upper()",
		"{{.Text | replaceRegex \"\\s+\" \" \" | trimSpace}}": "Text.replace(\"/\\s+/g\", \" \").trimSpace()",
		"{{x + y}}": "x + y",
	}
	return conversions[template]
}

// evaluateWithExprLang simulates expr-lang evaluation for benchmarking
func evaluateWithExprLang(expr string, env map[string]interface{}) (interface{}, error) {
	// This is a simplified simulation of expr-lang evaluation
	// In reality, this would use the actual expr-lang library
	switch expr {
	case "Name":
		return env["Name"], nil
	case "Data.trimSpace()":
		if data, ok := env["Data"].(string); ok {
			return data, nil // Simplified, doesn't actually trim
		}
	case "Text.upper()":
		if text, ok := env["Text"].(string); ok {
			return text, nil // Simplified, doesn't actually upper
		}
	case "x + y":
		if x, ok := env["x"].(int); ok {
			if y, ok := env["y"].(int); ok {
				return x + y, nil
			}
		}
	}
	return nil, nil
}

// BenchmarkTemplateCaching tests the performance benefit of template caching
func BenchmarkTemplateCaching(b *testing.B) {
	template := "{{.Name | upper}}"
	env := map[string]interface{}{"Name": "test"}

	// First, warm up the cache
	_, _ = EvaluateTemplateExpression(template, env)

	b.Run("Cached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = EvaluateTemplateExpression(template, env)
		}
	})
}

// BenchmarkComplexExpressions tests performance with increasingly complex expressions
func BenchmarkComplexExpressions(b *testing.B) {
	complexityLevels := []struct {
		name     string
		template string
		env      map[string]interface{}
	}{
		{"Level1", "{{.Name}}", map[string]interface{}{"Name": "test"}},
		{"Level2", "{{.Name | upper}}", map[string]interface{}{"Name": "test"}},
		{"Level3", "{{.Text | trimSpace | upper}}", map[string]interface{}{"Text": "  test  "}},
		{"Level4", "{{.Text | replace \"old\" \"new\" | trimSpace | upper}}", map[string]interface{}{"Text": "  old text  "}},
	}

	for _, level := range complexityLevels {
		b.Run(level.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = EvaluateTemplateExpression(level.template, level.env)
			}
		})
	}
}

// BenchmarkMemoryAllocation compares memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	template := "{{.Data | replace \"old\" \"new\" | trimSpace | upper}}"
	env := map[string]interface{}{"Data": "  old data with some content  "}

	b.Run("TemplateEvaluation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = EvaluateTemplateExpression(template, env)
		}
	})
}

// BenchmarkConcurrentUsage tests performance under concurrent load
func BenchmarkConcurrentUsage(b *testing.B) {
	template := "{{.Name | upper}}"

	b.Run("Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			env := map[string]interface{}{"Name": "test"}
			for pb.Next() {
				_, _ = EvaluateTemplateExpression(template, env)
			}
		})
	})
}
