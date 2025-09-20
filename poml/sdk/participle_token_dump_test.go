package poml

import (
	"fmt"
	"strings"
	"testing"
)

func TestDumpTokens(t *testing.T) {
	input := "<poml><role>Assistant</role><task>Say hi</task></poml>"
	def := buildPOMLLexer()
	lex, err := def.Lex("", strings.NewReader(input))
	if err != nil {
		t.Fatalf("lex init error: %v", err)
	}
	for {
		tok, err := lex.Next()
		if err != nil {
			t.Fatalf("lex error: %v", err)
		}
		if tok.EOF() {
			break
		}
		// token type names may vary; print numeric type and raw value and type string
		t.Logf("TOK: %v (%s) -> %q", tok.Type, fmt.Sprintf("%v", tok.Type), tok.Value)
	}
}
