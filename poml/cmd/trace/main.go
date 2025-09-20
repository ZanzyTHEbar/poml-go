package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ZanzyTHEbar/poml/sdk"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: trace <file.poml>")
		os.Exit(2)
	}
	b, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}
	trace, err := poml.ParseWithTrace(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
	}
	fmt.Println("--- TRACE OUTPUT ---")
	fmt.Println(trace)
	if err != nil {
		os.Exit(1)
	}
}
