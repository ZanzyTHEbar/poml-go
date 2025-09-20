# POML Go SDK (native)

This package provides a native Go SDK for POML (Prompt Orchestration Markup Language).

This is an incremental implementation. The initial release implements a minimal native
reader/writer API so Go programs can load POML content and obtain an intermediate
representation (IR). The full parser and runtime will be implemented iteratively.

Quick start

```go
package main

import (
  "fmt"
  "os"

  "github.com/ZanzyTHEbar/poml/sdk"
)

func main() {
  content := "<poml><task>Hello</task></poml>"
  ir, err := poml.Read(content, &poml.ReaderOptions{Trim: true}, nil, nil, "")
  if err != nil {
    fmt.Fprintln(os.Stderr, "Read error:", err)
    os.Exit(1)
  }

  out, err := poml.Write(ir, &poml.WriteOptions{SpeakerMode: true})
  if err != nil {
    fmt.Fprintln(os.Stderr, "Write error:", err)
    os.Exit(2)
  }

  fmt.Printf("Output: %+v\n", out)
}
```

CLI

We provide a small CLI in `go/poml/cmd/pomlcli` that reads a POML file (or stdin) and renders it with the native reader/writer.

Build:

```
cd go/poml
go build ./cmd/pomlcli
```

Usage:

```
# read from file
./pomlcli -in examples/101_explain_character.poml

# read from stdin and output speaker messages
cat examples/101_explain_character.poml | ./pomlcli -speaker
```

Development notes

- This is an incremental, native Go SDK. It currently supports a small subset of POML features (parsing, templating, basic runtime constructs like `<if>` and `<for>`, and simple writer outputs).
- The `Read` function includes naive frontmatter detection; a proper YAML frontmatter parser will be added next.
- The ad-hoc parser is the current stable baseline; a `participle` grammar implementation exists and will replace the ad-hoc parser when it reaches parity.

Contributions and Issues

Please open PRs against `go/poml` for additional features, tests, and documentation. Run `go test ./...` to validate changes locally.



