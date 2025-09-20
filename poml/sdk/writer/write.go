package writer

import (
	"fmt"

	"github.com/ZanzyTHEbar/poml/sdk/ast"
)

// Message represents a speaker message for writer output (re-exported type)
type Message = struct {
	Speaker string `json:"speaker"`
	Content string `json:"content"`
}

// Write produces either RichContent (interface{}) or []Message depending on speakerMode.
func Write(ir interface{}, speakerMode bool, opts *RenderOptions) (interface{}, error) {
	// if ir is already an AST, render it
	switch doc := ir.(type) {
	case *ast.Document:
		rc, err := RenderAST(doc, opts)
		if err != nil {
			return nil, err
		}
		if speakerMode {
			// convert each item to a Message with speaker "assistant"
			msgs := []Message{}
			if arr, ok := rc.([]interface{}); ok {
				for _, item := range arr {
					msgs = append(msgs, Message{Speaker: "assistant", Content: fmt.Sprintf("%v", item)})
				}
			} else {
				msgs = append(msgs, Message{Speaker: "assistant", Content: fmt.Sprintf("%v", rc)})
			}
			return msgs, nil
		}
		return rc, nil
	case string:
		// assume string IR is already rendered
		if speakerMode {
			return []Message{{Speaker: "assistant", Content: doc}}, nil
		}
		return doc, nil
	default:
		return nil, fmt.Errorf("unsupported IR type: %T", ir)
	}
}
