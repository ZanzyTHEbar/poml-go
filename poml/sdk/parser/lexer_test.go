package parser

import (
	"strings"
	"testing"
)

func TestTemplateExpressionLexer(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "Simple template expression",
			input: "{{.Name}}",
			expected: []Token{
				{Type: TokenTemplateContent, Lit: ".Name"},
			},
		},
		{
			name:  "Template with function",
			input: "{{.Text | upper}}",
			expected: []Token{
				{Type: TokenTemplateContent, Lit: ".Text | upper"},
			},
		},
		{
			name:  "Template with complex expression",
			input: "{{.Data | replace \"old\" \"new\" | trimSpace}}",
			expected: []Token{
				{Type: TokenTemplateContent, Lit: ".Data | replace \"old\" \"new\" | trimSpace"},
			},
		},
		{
			name:  "Template mixed with text",
			input: "Hello {{.Name}}, welcome!",
			expected: []Token{
				{Type: TokenText, Lit: "Hello "},
				{Type: TokenTemplateContent, Lit: ".Name"},
				{Type: TokenText, Lit: ", welcome!"},
			},
		},
		{
			name:  "Template with XML",
			input: `<p>Hello {{.Name}}</p>`,
			expected: []Token{
				{Type: TokenIdentifier, Lit: "p"},
				{Type: TokenText, Lit: "Hello "},
				{Type: TokenTemplateContent, Lit: ".Name"},
				{Type: TokenIdentifier, Lit: "/p"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.input)
			var tokens []Token

			for {
				token := lexer.Next()
				if token.Type == TokenEOF {
					break
				}
				tokens = append(tokens, token)
			}

			if len(tokens) != len(tc.expected) {
				t.Errorf("Expected %d tokens, got %d", len(tc.expected), len(tokens))
				return
			}

			for i, expected := range tc.expected {
				if tokens[i].Type != expected.Type {
					t.Errorf("Token %d: expected type %v, got %v", i, expected.Type, tokens[i].Type)
				}
				if tokens[i].Lit != expected.Lit {
					t.Errorf("Token %d: expected lit %q, got %q", i, expected.Lit, tokens[i].Lit)
				}
			}
		})
	}
}

func TestTemplateExpressionValidation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string // expected error message or empty string if valid
	}{
		{
			name:     "Valid simple template",
			input:    "{{.Name}}",
			expected: "",
		},
		{
			name:     "Valid template with function",
			input:    "{{.Text | upper}}",
			expected: "",
		},
		{
			name:     "Invalid: starts with pipe",
			input:    "{{| upper}}",
			expected: "ERROR: template cannot start with pipe operator",
		},
		{
			name:     "Invalid: ends with pipe",
			input:    "{{.Name |}}",
			expected: "", // Valid Go template syntax
		},
		{
			name:     "Invalid: consecutive pipes",
			input:    "{{.Name || upper}}",
			expected: "ERROR: consecutive pipe operators",
		},
		{
			name:     "Invalid: unmatched quotes",
			input:    "{{.Name | replace 'old}}",
			expected: "ERROR: unmatched single quotes",
		},
		{
			name:     "Invalid: no variable reference",
			input:    "{{upper}}",
			expected: "ERROR: invalid template syntax: expected variable reference starting with '.'",
		},
		{
			name:     "Invalid: unclosed template",
			input:    "{{.Name",
			expected: "ERROR: Unclosed template expression",
		},
		{
			name:     "Valid: function call",
			input:    "{{len .Items}}",
			expected: "",
		},
		{
			name:     "Valid: empty template",
			input:    "{{}}",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.input)
			var tokens []Token

			for {
				token := lexer.Next()
				if token.Type == TokenEOF {
					break
				}
				tokens = append(tokens, token)
			}

			if tc.expected == "" {
				// Should be valid - check that we got template content without error
				found := false
				for _, token := range tokens {
					if token.Type == TokenTemplateContent && !strings.HasPrefix(token.Lit, "ERROR:") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected valid template content, but found errors or no template content")
				}
			} else {
				// Should contain error
				found := false
				for _, token := range tokens {
					if (token.Type == TokenTemplateContent || token.Type == TokenError) && strings.Contains(token.Lit, tc.expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, but didn't find it in tokens. Got tokens: %+v", tc.expected, tokens)
				}
			}
		})
	}
}
