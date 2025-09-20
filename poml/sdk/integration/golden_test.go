package integration

import (
	"os"
	"path/filepath"
	"testing"

	poml "github.com/ZanzyTHEbar/poml/sdk"
)

func TestBasicGoldenExamples(t *testing.T) {
	examples := []string{"examples/101_explain_character.poml", "examples/102_render_xml.poml"}
	for _, ex := range examples {
		// examples live at repository root 'examples/'
		data, err := os.ReadFile(filepath.Join("..", "..", "..", ex))
		if err != nil {
			t.Fatalf("read example %s: %v", ex, err)
		}
		ir, err := poml.Read(string(data), &poml.ReaderOptions{Trim: true}, nil, nil, ex)
		if err != nil {
			t.Fatalf("read failed for %s: %v", ex, err)
		}
		if ir == "" {
			t.Fatalf("empty IR for %s", ex)
		}
	}
}
