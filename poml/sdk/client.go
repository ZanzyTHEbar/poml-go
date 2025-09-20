package poml

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ZanzyTHEbar/poml/sdk/ast"
	writer "github.com/ZanzyTHEbar/poml/sdk/writer"
)

// Read parses a POML string and returns an intermediate representation (IR).
// The initial native implementation performs a lightweight validation and
// returns a simplified IR for basic usage. Full parsing is planned.
func Read(element string, opts *ReaderOptions, ctx map[string]any, stylesheet map[string]any, sourcePath string) (string, error) {
	if strings.TrimSpace(element) == "" {
		return "", errors.New("empty input")
	}

	// Integrate frontmatter detection: if content begins with '---' treat as frontmatter + body
	body := element
	if strings.HasPrefix(strings.TrimSpace(element), "---") {
		// naive split: find second '---' delimiter
		parts := strings.SplitN(element, "---", 3)
		if len(parts) == 3 {
			// parts[1] is frontmatter, parts[2] is body
			// For now, preserve frontmatter in IR header
			ir := "<ir>\n<frontmatter>" + strings.TrimSpace(parts[1]) + "</frontmatter>\n<body>" + strings.TrimSpace(parts[2]) + "</body>\n</ir>"
			return ir, nil
		}
	}

	// Default IR
	ir := "<ir>" + "length=" + strconv.Itoa(len(body)) + "</ir>\n" + body
	return ir, nil
}

// Write transforms IR into RichContent or []Message depending on options.
func Write(ir string, opts *WriteOptions) (interface{}, error) {
	if ir == "" {
		return nil, errors.New("empty ir")
	}
	// If IR appears to contain POML body, attempt to parse to AST and render structured output
	var body string = ir
	if strings.HasPrefix(strings.TrimSpace(ir), "<ir>") {
		// try to extract <body>...</body> if present
		low := strings.Index(ir, "<body>")
		high := strings.Index(ir, "</body>")
		if low != -1 && high != -1 && high > low {
			body = strings.TrimSpace(ir[low+len("<body>") : high])
		}
		// if no explicit <body>, try to strip the initial <ir> header and use the remainder
		if body == ir {
			if idx := strings.Index(ir, "</ir>"); idx != -1 {
				rest := strings.TrimSpace(ir[idx+len("</ir>"):])
				if rest != "" {
					body = rest
				}
			}
		}
	}

	// try parsing body into AST
	doc, err := ParseWithParticiple(body)
	if err == nil && doc != nil {
		// render via writer
		renderOpts := &writer.RenderOptions{
			SpeakerMode:    opts != nil && opts.SpeakerMode,
			Context:        make(map[string]interface{}),
			BaseDir:        "",
			CurrentFileDir: "",
		}
		out, werr := writer.RenderAST(doc, renderOpts)
		if werr == nil {
			// convert ast.Message slice to poml.Message slice if necessary
			switch v := out.(type) {
			case []ast.Message:
				msgs := make([]Message, 0, len(v))
				for _, m := range v {
					msgs = append(msgs, Message{Speaker: Speaker(m.Speaker), Content: m.Content})
				}
				return msgs, nil
			default:
				return out, nil
			}
		}
		// fallthrough to simple behavior on writer error
	}

	// Fallback: behave like previous simple writer
	if opts != nil && opts.SpeakerMode {
		return []Message{{Speaker: SpeakerAI, Content: ir}}, nil
	}
	return RichContent{ir}, nil
}
