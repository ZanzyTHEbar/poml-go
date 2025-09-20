package poml

import "github.com/ZanzyTHEbar/poml/sdk/ast"

// ReaderOptions configures the behaviour of the reader.
type ReaderOptions struct {
	Trim bool
}

// WriteOptions controls how IR is rendered to messages.
type WriteOptions struct {
	SpeakerMode bool
}

// Speaker represents a conversation participant
type Speaker = ast.Speaker

const (
	SpeakerSystem = ast.SpeakerSystem
	SpeakerHuman  = ast.SpeakerHuman
	SpeakerAI     = ast.SpeakerAI
)

// Message represents a speaker message in POML output.
type Message struct {
	Speaker Speaker     `json:"speaker"`
	Content interface{} `json:"content"`
}

// RichContent represents non-speaker structured output.
type RichContent = []interface{}

// CliResult represents output of the CLI or equivalent runtime.
type CliResult struct {
	Messages       interface{}            `json:"messages"`
	ResponseSchema map[string]interface{} `json:"responseSchema,omitempty"`
}
