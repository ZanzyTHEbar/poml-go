package poml

import (
	"bytes"
	"fmt"

	"github.com/beevik/etree"
)

// PocFormatWithEtree takes a POML-like XML string and returns a canonicalized IR
// using etree to parse and pretty-print the structure.
func PocFormatWithEtree(input string) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(input); err != nil {
		return "", err
	}
	// canonicalize: remove extra whitespace and write with indent
	buf := &bytes.Buffer{}
	doc.Indent(2)
	if _, err := doc.WriteTo(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ExamplePocFormat() {
	src := `<poml><role>Assistant</role><task>Say hi</task></poml>`
	out, err := PocFormatWithEtree(src)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(out)
}
