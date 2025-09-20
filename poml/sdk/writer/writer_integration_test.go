package writer_test

import (
	"testing"

	poml "github.com/ZanzyTHEbar/poml/sdk"
	writer "github.com/ZanzyTHEbar/poml/sdk/writer"
)

func TestRenderIfAndForIntegration(t *testing.T) {
	input := "<doc><if cond=true>YES</if><for var=i in=items><item>{{i}};</item></for></doc>"
	doc, err := poml.ParseWithParticiple(input)
	if err != nil {
		t.Fatalf("ParseWithParticiple error: %v", err)
	}
	ctx := map[string]interface{}{"items": []interface{}{"A", "B"}}
	out, err := writer.RenderAST(doc, &writer.RenderOptions{Context: ctx})
	if err != nil {
		t.Fatalf("RenderAST error: %v", err)
	}
	arr, ok := out.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{} from RenderAST, got %T", out)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 top-level output, got %d", len(arr))
	}
	if s, ok := arr[0].(string); !ok || s != "YESA;B;" {
		t.Fatalf("unexpected rendered output: %v", arr[0])
	}
}
