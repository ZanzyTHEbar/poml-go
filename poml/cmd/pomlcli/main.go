package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ZanzyTHEbar/poml/sdk"
)

func main() {
	inPath := flag.String("in", "", "input POML file (defaults to stdin)")
	speaker := flag.Bool("speaker", false, "output as speaker messages")
	flag.Parse()

	var data []byte
	var err error
	if *inPath == "" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
			os.Exit(2)
		}
	} else {
		data, err = os.ReadFile(*inPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read file: %v\n", err)
			os.Exit(2)
		}
	}

	ir, err := poml.Read(string(data), &poml.ReaderOptions{Trim: true}, nil, nil, *inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(2)
	}

	out, err := poml.Write(ir, &poml.WriteOptions{SpeakerMode: *speaker})
	if err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(2)
	}

	// print JSON-like output
	switch v := out.(type) {
	case []poml.Message:
		for _, m := range v {
			fmt.Printf("%s: %s\n", m.Speaker, m.Content)
		}
	default:
		fmt.Println(v)
	}
}
