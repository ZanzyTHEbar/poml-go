package poml

import (
	"testing"
)

func TestParticipleTrace(t *testing.T) {
	input := "<poml><role>Assistant</role><task>Say hi</task></poml>"
	trace, err := ParseWithTrace(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v\ntrace:\n%s", err, trace)
	}
	t.Logf("participle trace:\n%s", trace)
}
