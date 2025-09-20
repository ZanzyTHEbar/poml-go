package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/ZanzyTHEbar/poml/sdk"
	"github.com/ZanzyTHEbar/poml/sdk/writer"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: golden <input.poml>")
		os.Exit(2)
	}
	in := args[0]
	b, err := os.ReadFile(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}
	// normalize: strip HTML comments
	clean := stripHTMLComments(string(b))
	doc, err := poml.ParseWithParticiple(clean)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}
	ctx := map[string]interface{}{}
	out, err := writer.RenderAST(doc, &writer.RenderOptions{Context: ctx})
	if err != nil {
		fmt.Fprintln(os.Stderr, "render error:", err)
		os.Exit(1)
	}
	if arr, ok := out.([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				fmt.Print(s)
			}
		}
		fmt.Println()
		return
	}
	fmt.Printf("%v\n", out)
}

func stripHTMLComments(s string) string {
	re := regexp.MustCompile(`<!--[\s\S]*?-->`)
	return re.ReplaceAllString(s, "")
}
