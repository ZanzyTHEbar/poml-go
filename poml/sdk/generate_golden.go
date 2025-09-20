//go:build ignore

package main

import (
	"fmt"
	"os"

	poml "github.com/ZanzyTHEbar/poml/sdk"
	writer "github.com/ZanzyTHEbar/poml/sdk/writer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: generate_golden <input.poml>")
		os.Exit(2)
	}
	in := os.Args[1]
	b, err := os.ReadFile(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}
	doc, err := poml.ParseWithParticiple(string(b))
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}
	out, err := writer.RenderAST(doc, &writer.RenderOptions{Context: map[string]interface{}{}})
	if err != nil {
		fmt.Fprintln(os.Stderr, "render error:", err)
		os.Exit(1)
	}
	arr, ok := out.([]interface{})
	if !ok {
		fmt.Fprintln(os.Stderr, "unexpected render output type")
		os.Exit(1)
	}
	for _, v := range arr {
		if s, ok := v.(string); ok {
			fmt.Print(s)
		}
	}
	fmt.Println()
}
