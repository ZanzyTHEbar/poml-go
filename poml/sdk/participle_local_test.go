//go:build !no_participle_tags
// +build !no_participle_tags

package poml

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2"
)

// Test that the lexer and participle can parse a single open tag like <poml>
func TestParticipleSimpleTag(t *testing.T) {
	input := "<poml>"
	def := buildPOMLLexer()
	// Local minimal grammar using exported literal fields
	type Tag struct {
		L    string `@"<"`
		Name string `@Ident`
		R    string `@">"`
	}

	p, err := participle.Build[Tag](participle.Lexer(def), participle.Elide("WS"))
	if err != nil {
		t.Fatalf("build parser error: %v", err)
	}
	var b strings.Builder
	_, err = p.ParseString("", input, participle.Trace(&b))
	if err != nil {
		t.Fatalf("parse error: %v\ntrace:\n%s", err, b.String())
	}
}
