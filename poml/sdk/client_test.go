package poml

import (
	"testing"
)

func TestReadAndWrite(t *testing.T) {
	ir, err := Read("<poml><task>Say hello</task></poml>", &ReaderOptions{Trim: true}, nil, nil, "")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if ir == "" {
		t.Fatalf("Empty IR returned")
	}

	out, err := Write(ir, &WriteOptions{SpeakerMode: true})
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	msgs, ok := out.([]Message)
	if !ok {
		t.Fatalf("Expected []Message, got %T", out)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}
