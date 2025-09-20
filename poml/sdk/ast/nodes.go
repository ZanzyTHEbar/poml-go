package ast

// No external imports needed for basic AST types

// Speaker represents a conversation participant
type Speaker string

const (
	SpeakerSystem Speaker = "system"
	SpeakerHuman  Speaker = "human"
	SpeakerAI     Speaker = "ai"
)

// ValidSpeakers contains all valid speaker values
var ValidSpeakers = []Speaker{SpeakerSystem, SpeakerHuman, SpeakerAI}

// ContentMultiMedia represents multimedia content like images
type ContentMultiMedia struct {
	Type     string `json:"type"`   // image/png, image/jpeg, etc.
	Base64   string `json:"base64"` // base64 encoded content
	Alt      string `json:"alt,omitempty"`
	Position string `json:"position,omitempty"` // top, bottom, or empty for default
}

// RichContent can be a string or array of strings and multimedia
type RichContent interface{}

// Message represents a speaker message with content
type Message struct {
	Speaker Speaker     `json:"speaker"`
	Content RichContent `json:"content"`
}

// SourceMapRichContent includes source mapping information
type SourceMapRichContent struct {
	StartIndex   int         `json:"startIndex"`
	EndIndex     int         `json:"endIndex"`
	IRStartIndex int         `json:"irStartIndex"`
	IREndIndex   int         `json:"irEndIndex"`
	Content      RichContent `json:"content"`
}

// SourceMapMessage includes source mapping for speaker messages
type SourceMapMessage struct {
	StartIndex   int                    `json:"startIndex"`
	EndIndex     int                    `json:"endIndex"`
	IRStartIndex int                    `json:"irStartIndex"`
	IREndIndex   int                    `json:"irEndIndex"`
	Speaker      Speaker                `json:"speaker"`
	Content      []SourceMapRichContent `json:"content"`
}

// SpeakerNode represents a segment with speaker information
type SpeakerNode struct {
	Start   int     `json:"start"`
	End     int     `json:"end"`
	Speaker Speaker `json:"speaker"`
}

// Node is a node in the POML AST.
type Node interface {
	node()
}

// Document is the root of a parsed POML file.
type Document struct {
	Children []Node
}

// Element represents a tag element like <task>...</task>
type Element struct {
	Name string
	// Attributes
	Attrs    []Attr
	Children []Node
	// Source mapping
	StartIndex   int `json:"startIndex,omitempty"`
	EndIndex     int `json:"endIndex,omitempty"`
	IRStartIndex int `json:"irStartIndex,omitempty"`
	IREndIndex   int `json:"irEndIndex,omitempty"`
}

func (e *Element) node() {}

// Text is plain text content.
type Text struct {
	Content string
}

func (t *Text) node() {}

// Attr represents an element attribute key/value pair (value is raw string including quotes)
type Attr struct {
	Key               string
	Val               string
	ContainsTemplates bool
}

// IfNode represents a conditional node: <if cond="...">...</if>
type IfNode struct {
	Cond     string
	Children []Node
	// Source mapping
	StartIndex   int `json:"startIndex,omitempty"`
	EndIndex     int `json:"endIndex,omitempty"`
	IRStartIndex int `json:"irStartIndex,omitempty"`
	IREndIndex   int `json:"irEndIndex,omitempty"`
}

func (i *IfNode) node() {}

// ForNode represents a loop node: <for var="x" in="...">...</for>
type ForNode struct {
	Var      string
	In       string
	Children []Node
	// Source mapping
	StartIndex   int `json:"startIndex,omitempty"`
	EndIndex     int `json:"endIndex,omitempty"`
	IRStartIndex int `json:"irStartIndex,omitempty"`
	IREndIndex   int `json:"irEndIndex,omitempty"`
}

func (f *ForNode) node() {}

// ImgNode represents an image element <img src="..." alt="..." position="..."/>
type ImgNode struct {
	Src      string
	Alt      string
	Position string // top, bottom, or empty for default
	// Source mapping
	StartIndex   int `json:"startIndex,omitempty"`
	EndIndex     int `json:"endIndex,omitempty"`
	IRStartIndex int `json:"irStartIndex,omitempty"`
	IREndIndex   int `json:"irEndIndex,omitempty"`
}

func (i *ImgNode) node() {}

// TableNode represents a simple table; children will be Element rows/cells.
type TableNode struct {
	Children []Node
	// Source mapping
	StartIndex   int `json:"startIndex,omitempty"`
	EndIndex     int `json:"endIndex,omitempty"`
	IRStartIndex int `json:"irStartIndex,omitempty"`
	IREndIndex   int `json:"irEndIndex,omitempty"`
}

func (t *TableNode) node() {}

// TemplateVariable represents a variable reference like .Name
type TemplateVariable struct {
	Name       string `json:"name"`       // variable name without dot (e.g., "Name")
	FullName   string `json:"fullName"`   // full variable name with dot (e.g., ".Name")
	StartIndex int    `json:"startIndex"` // position in template expression
	EndIndex   int    `json:"endIndex"`
}

func (t *TemplateVariable) node() {}

// TemplateFunction represents a function call like upper or replace
type TemplateFunction struct {
	Name       string        `json:"name"`       // function name (e.g., "upper")
	Args       []TemplateArg `json:"args"`       // function arguments
	StartIndex int           `json:"startIndex"` // position in template expression
	EndIndex   int           `json:"endIndex"`
}

func (t *TemplateFunction) node() {}

// TemplateArg represents an argument to a function (string literal, variable, or number)
type TemplateArg struct {
	Type       string `json:"type"`       // "string", "variable", "number"
	Value      string `json:"value"`      // the actual value
	StartIndex int    `json:"startIndex"` // position in template expression
	EndIndex   int    `json:"endIndex"`
}

func (t *TemplateArg) node() {}

// TemplatePipeline represents a pipeline operation: input | function
type TemplatePipeline struct {
	Input      TemplateComponent `json:"input"`      // input to the pipeline
	Function   *TemplateFunction `json:"function"`   // function being applied
	StartIndex int               `json:"startIndex"` // position in template expression
	EndIndex   int               `json:"endIndex"`
}

func (t *TemplatePipeline) node() {}

// TemplateComponent is a union type for template expression components
type TemplateComponent interface {
	Node
	templateComponent() // marker method
}

// Implement templateComponent for all component types
func (t *TemplateVariable) templateComponent() {}
func (t *TemplateFunction) templateComponent() {}
func (t *TemplatePipeline) templateComponent() {}

// TemplateExpr represents a template expression: {{...}}
type TemplateExpr struct {
	Expression  string            `json:"expression"`  // the raw template expression content
	Parsed      TemplateComponent `json:"parsed"`      // parsed AST representation (optional)
	HasPipeline bool              `json:"hasPipeline"` // whether this expression contains pipeline operations
	// Compilation results
	IsCompiled   bool   `json:"isCompiled"`   // whether template was successfully compiled at parse time
	CompileError string `json:"compileError"` // compilation error message (if any)
	// Variable resolution results
	ResolvedVariables       []string `json:"resolvedVariables"`       // variables successfully resolved in context
	UnresolvedVariables     []string `json:"unresolvedVariables"`     // variables not found in context
	IsVariablesResolved     bool     `json:"isVariablesResolved"`     // whether all variables were successfully resolved
	VariableResolutionError string   `json:"variableResolutionError"` // error message for variable resolution failures
	// Function validation results
	ValidatedFunctions        []string          `json:"validatedFunctions"`        // functions successfully validated
	InvalidFunctions          []string          `json:"invalidFunctions"`          // functions with validation errors
	IsFunctionsValidated      bool              `json:"isFunctionsValidated"`      // whether all functions were successfully validated
	FunctionValidationError   string            `json:"functionValidationError"`   // error message for function validation failures
	FunctionValidationDetails map[string]string `json:"functionValidationDetails"` // detailed validation results per function
	// Enhanced source mapping with line/column information
	StartIndex   int `json:"startIndex,omitempty"`
	EndIndex     int `json:"endIndex,omitempty"`
	IRStartIndex int `json:"irStartIndex,omitempty"`
	IREndIndex   int `json:"irEndIndex,omitempty"`
	// Line and column information for better error reporting
	StartLine   int    `json:"startLine,omitempty"`   // starting line number (1-based)
	StartColumn int    `json:"startColumn,omitempty"` // starting column number (1-based)
	EndLine     int    `json:"endLine,omitempty"`     // ending line number (1-based)
	EndColumn   int    `json:"endColumn,omitempty"`   // ending column number (1-based)
	SourceText  string `json:"sourceText,omitempty"`  // original source text for context
}

func (t *TemplateExpr) node() {}

// EnvNode represents an environment wrapper: <env presentation="...">...</env>
type EnvNode struct {
	Presentation string // multimedia, text, etc.
	Children     []Node
	// Source mapping
	StartIndex   int `json:"startIndex,omitempty"`
	EndIndex     int `json:"endIndex,omitempty"`
	IRStartIndex int `json:"irStartIndex,omitempty"`
	IREndIndex   int `json:"irEndIndex,omitempty"`
}

func (e *EnvNode) node() {}
