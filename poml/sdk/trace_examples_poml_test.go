package poml

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTraceExamples(t *testing.T) {
	files, err := filepath.Glob("integration/golden/more/*.poml")
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	if len(files) == 0 {
		t.Skip("no example files found")
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Logf("%s: read error: %v", f, err)
			continue
		}
		trace, err := ParseWithTrace(string(b))
		if err != nil {
			t.Logf("%s: parse error: %v\ntrace:\n%s", f, err, trace)
		} else {
			t.Logf("%s: parse ok. trace:\n%s", f, trace)
		}
	}
}
