package golden

import (
	"os"
	"strings"
	"testing"

	poml "github.com/ZanzyTHEbar/poml/sdk"
	"github.com/ZanzyTHEbar/poml/sdk/ast"
	writer "github.com/ZanzyTHEbar/poml/sdk/writer"
)

func TestGoldenSample1(t *testing.T) {
	inPath := "sample1.poml"
	goldPath := "sample1.golden.txt"
	b, err := os.ReadFile(inPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	doc, err := poml.ParseWithParticiple(string(b))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// Provide deterministic context for golden rendering
	ctx := map[string]interface{}{"items": []interface{}{"A", "B"}}
	out, err := writer.RenderAST(doc, &writer.RenderOptions{Context: ctx})
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	arr, ok := out.([]interface{})
	if !ok {
		t.Fatalf("unexpected output type: %T", out)
	}
	got := ""
	for _, v := range arr {
		if s, ok := v.(string); ok {
			got += s
		}
	}
	wantB, err := os.ReadFile(goldPath)
	if err != nil {
		// if golden missing, write it and fail to notify
		_ = os.WriteFile(goldPath, []byte(got), 0o644)
		t.Fatalf("golden file missing; wrote %s — please verify and re-run tests", goldPath)
	}
	want := string(wantB)
	if got != want {
		t.Fatalf("golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestGoldenWordTodos(t *testing.T) {
	testGoldenFile(t, "103_word_todos.poml", "103_word_todos.poml.golden.txt")
}

func TestGoldenFinancialAnalysis(t *testing.T) {
	testGoldenFile(t, "104_financial_analysis.poml", "104_financial_analysis.poml.golden.txt")
}

func TestGoldenBlogPost(t *testing.T) {
	testGoldenFile(t, "105_write_blog_post.poml", "105_write_blog_post.poml.golden.txt")
}

func TestGoldenExplainCharacter(t *testing.T) {
	testGoldenFile(t, "101_explain_character.poml", "101_explain_character.poml.golden.txt")
}

func TestGoldenRenderXML(t *testing.T) {
	testGoldenFile(t, "102_render_xml.poml", "102_render_xml.poml.golden.txt")
}

func TestGoldenGeneratePOML(t *testing.T) {
	testGoldenFile(t, "301_generate_poml.poml", "301_generate_poml.poml.golden.txt")
}

func TestDebugGeneratePOML(t *testing.T) {
	// Debug the parsing issue
	inPath := "301_generate_poml.poml"
	b, err := os.ReadFile(inPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}

	content := string(b)
	t.Logf("File content length: %d", len(content))
	t.Logf("First 200 chars: %s", content[:min(200, len(content))])

	// Just try to parse without rendering
	doc, err := poml.ParseWithParticiple(content)
	if err != nil {
		t.Logf("Parse error: %v", err)
		return
	}

	t.Logf("Successfully parsed document with %d children", len(doc.Children))
	for i, child := range doc.Children {
		switch c := child.(type) {
		case *ast.Element:
			t.Logf("Child %d: Element %s with %d children", i, c.Name, len(c.Children))
		case *ast.Text:
			t.Logf("Child %d: Text (length %d)", i, len(c.Content))
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestSpeakerMode(t *testing.T) {
	// Test speaker mode functionality
	inPath := "103_word_todos.poml"
	b, err := os.ReadFile(inPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}

	doc, err := poml.ParseWithParticiple(string(b))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Test speaker mode rendering
	renderOpts := &writer.RenderOptions{
		SpeakerMode:    true,
		Context:        make(map[string]interface{}),
		BaseDir:        "",
		CurrentFileDir: "",
	}

	out, err := writer.RenderAST(doc, renderOpts)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	// Check if output is speaker messages
	messages, ok := out.([]ast.Message)
	if !ok {
		t.Fatalf("expected []ast.Message, got %T", out)
	}

	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}

	// Check that speakers are valid
	for i, msg := range messages {
		found := false
		for _, valid := range ast.ValidSpeakers {
			if msg.Speaker == valid {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("message %d has invalid speaker: %s", i, msg.Speaker)
		}
	}

	t.Logf("Successfully rendered %d speaker messages", len(messages))

	// Test the full API integration
	ir := "<ir>length=100</ir>\n<poml><p>Test content</p></poml>"
	opts := &poml.WriteOptions{SpeakerMode: true}
	result, err := poml.Write(ir, opts)
	if err != nil {
		t.Fatalf("Write with speaker mode failed: %v", err)
	}

	// Check if result is []poml.Message
	messages2, ok := result.([]poml.Message)
	if !ok {
		t.Fatalf("expected []poml.Message, got %T", result)
	}

	if len(messages2) == 0 {
		t.Fatal("expected at least one message from API")
	}

	t.Logf("Full API test passed with %d messages", len(messages2))
}

func TestMultimediaSupport(t *testing.T) {
	// Test multimedia support with img elements
	pomlContent := `<poml>
<p>Hello world <img src="test.jpg" alt="test image" /></p>
</poml>`

	doc, err := poml.ParseWithParticiple(pomlContent)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	renderOpts := &writer.RenderOptions{
		Context:        make(map[string]interface{}),
		BaseDir:        "",
		CurrentFileDir: "",
	}

	out, err := writer.RenderAST(doc, renderOpts)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	// Should return RichContent array with multimedia
	content, ok := out.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", out)
	}

	if len(content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(content))
	}

	// Debug what's actually returned
	t.Logf("Content type: %T, value: %+v", content[0], content[0])

	// First item should be a RichContent array containing string and multimedia
	richContent, ok := content[0].([]interface{})
	if !ok {
		t.Fatalf("expected []interface{} for RichContent, got %T", content[0])
	}

	// Debug all items in RichContent
	for i, item := range richContent {
		t.Logf("RichContent[%d]: type=%T, value=%+v", i, item, item)
	}

	// Should contain text and multimedia object (may be more than 2 due to whitespace)
	if len(richContent) < 2 {
		t.Fatalf("expected at least 2 items in RichContent, got %d", len(richContent))
	}

	// Find the text and multimedia items
	var multimediaItem ast.ContentMultiMedia
	foundText := false
	foundMultimedia := false

	for _, item := range richContent {
		if s, ok := item.(string); ok && strings.TrimSpace(s) == "Hello world" {
			foundText = true
		} else if mm, ok := item.(ast.ContentMultiMedia); ok {
			multimediaItem = mm
			foundMultimedia = true
		}
	}

	if !foundText {
		t.Error("expected to find 'Hello world' text")
	}
	if !foundMultimedia {
		t.Error("expected to find multimedia object")
	}

	// Verify multimedia object
	if multimediaItem.Type != "image/jpeg" {
		t.Errorf("expected image/jpeg type, got %s", multimediaItem.Type)
	}
	if multimediaItem.Base64 != "placeholder_base64_data" {
		t.Errorf("expected placeholder base64 data, got %s", multimediaItem.Base64)
	}

	t.Logf("Multimedia support test passed")
}

func TestIntegrationAllFeatures(t *testing.T) {
	// Test comprehensive integration of all features
	inPath := "integration_test_all_features.poml"

	// Test with speaker mode disabled (regular output)
	doc, err := poml.ParseWithParticiple(inPath)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	renderOpts := &writer.RenderOptions{
		Context:        make(map[string]interface{}),
		BaseDir:        "",
		CurrentFileDir: "",
	}

	out, err := writer.RenderAST(doc, renderOpts)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	// Should return regular output
	content, ok := out.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", out)
	}

	if len(content) == 0 {
		t.Fatal("expected content")
	}

	t.Logf("Regular output test passed with %d content items", len(content))

	// Test with speaker mode enabled
	renderOpts.SpeakerMode = true
	speakerOut, err := writer.RenderAST(doc, renderOpts)
	if err != nil {
		t.Fatalf("speaker mode render error: %v", err)
	}

	// Should return speaker messages
	messages, ok := speakerOut.([]ast.Message)
	if !ok {
		t.Fatalf("expected []ast.Message, got %T", speakerOut)
	}

	if len(messages) == 0 {
		t.Fatal("expected speaker messages")
	}

	// Validate speaker assignments
	validSpeakers := 0
	for _, msg := range messages {
		found := false
		for _, valid := range ast.ValidSpeakers {
			if msg.Speaker == valid {
				found = true
				break
			}
		}
		if found {
			validSpeakers++
		}
	}

	if validSpeakers == 0 {
		t.Error("expected at least one valid speaker assignment")
	}

	t.Logf("Speaker mode test passed with %d messages, %d valid speakers", len(messages), validSpeakers)

	// Test full API integration
	ir := "<ir>length=100</ir>\n" + string(mustReadFile(inPath))
	opts := &poml.WriteOptions{SpeakerMode: true}
	result, err := poml.Write(ir, opts)
	if err != nil {
		t.Fatalf("API integration error: %v", err)
	}

	apiMessages, ok := result.([]poml.Message)
	if !ok {
		t.Fatalf("expected []poml.Message, got %T", result)
	}

	if len(apiMessages) == 0 {
		t.Fatal("expected messages from API")
	}

	t.Logf("Full API integration test passed with %d messages", len(apiMessages))
}

func mustReadFile(filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func testGoldenFile(t *testing.T, inPath, goldPath string) {
	t.Helper()
	b, err := os.ReadFile(inPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	doc, err := poml.ParseWithParticiple(string(b))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// Empty context for golden rendering
	ctx := map[string]interface{}{}
	out, err := writer.RenderAST(doc, &writer.RenderOptions{Context: ctx})
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	arr, ok := out.([]interface{})
	if !ok {
		t.Fatalf("unexpected output type: %T", out)
	}
	got := ""
	for _, v := range arr {
		if s, ok := v.(string); ok {
			got += s
		}
	}
	wantB, err := os.ReadFile(goldPath)
	if err != nil {
		// if golden missing, write it and fail to notify
		_ = os.WriteFile(goldPath, []byte(got), 0o644)
		t.Fatalf("golden file missing; wrote %s — please verify and re-run tests", goldPath)
	}
	want := string(wantB)
	if got != want {
		t.Fatalf("golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
