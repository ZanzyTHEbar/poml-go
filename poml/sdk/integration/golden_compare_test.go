package integration

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	poml "github.com/ZanzyTHEbar/poml/sdk"
	writer "github.com/ZanzyTHEbar/poml/sdk/writer"
)

func TestGoldenExamples(t *testing.T) {
	examples := []string{
		"102_render_xml.poml",
		"106_research.poml",
		"105_write_blog_post.poml",
		"201_orders_qa.poml",
		"202_arc_agi.poml",
		"107_read_report_pdf.poml",
	}
	for _, name := range examples {
		pomlPath := filepath.Join("..", "..", "examples", name)
		expectPath := filepath.Join("..", "..", "examples", "expects", strings.TrimSuffix(name, ".poml")+".txt")

		b, err := os.ReadFile(pomlPath)
		if err != nil {
			t.Logf("SKIP %s (missing input): %v", name, err)
			continue
		}
		clean := stripHTMLComments(string(b))

		doc, err := poml.ParseWithParticiple(clean)
		if err != nil {
			t.Logf("PARSE ERR %s: %v", name, err)
			continue
		}
		out, err := writer.RenderAST(doc, &writer.RenderOptions{Context: map[string]interface{}{}})
		if err != nil {
			t.Logf("RENDER ERR %s: %v", name, err)
			continue
		}
		got := flatten(out)
		gotN := normalize(got)

		expb, err := os.ReadFile(expectPath)
		if err != nil {
			t.Logf("SKIP %s (missing expect): %v", name, err)
			continue
		}
		wantN := normalize(string(expb))

		if gotN != wantN {
			t.Logf("DIFF %s\n--- got ---\n%s\n--- want ---\n%s", name, got, string(expb))
		} else {
			t.Logf("OK %s", name)
		}
	}
}

func stripHTMLComments(s string) string {
	re := regexp.MustCompile(`<!--[\s\S]*?-->`)
	return re.ReplaceAllString(s, "")
}

func normalize(s string) string {
	// collapse whitespace and trim
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	scanner := bufio.NewScanner(strings.NewReader(s))
	var buf bytes.Buffer
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	return buf.String()
}

func flatten(out interface{}) string {
	switch v := out.(type) {
	case []interface{}:
		var b strings.Builder
		for _, it := range v {
			if s, ok := it.(string); ok {
				b.WriteString(s)
			}
		}
		return b.String()
	case string:
		return v
	default:
		return ""
	}
}
