package parser

import (
	"fmt"
	"strings"
)

// Token types
type TokenType int

const (
	TokenText  TokenType = iota
	TokenLT              // '<'
	TokenGT              // '>'
	TokenSlash           // '/'
	TokenIdentifier
	TokenEq     // '='
	TokenString // quoted string
	TokenEOF

	// Template-related tokens
	TokenTemplateStart   // '{{'
	TokenTemplateEnd     // '}}'
	TokenTemplateContent // template expression content
	TokenError           // lexer error (e.g., invalid template syntax)
)

type Token struct {
	Type     TokenType
	Lit      string
	StartPos int // Start position in input
	EndPos   int // End position in input
}

// very small lexer that recognizes tags and text
type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, pos: 0}
}

func (l *Lexer) read() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r := rune(l.input[l.pos])
	l.pos++
	return r
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

func (l *Lexer) Next() Token {
	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, StartPos: l.pos, EndPos: l.pos}
	}
	// if starts with '<'
	if l.peek() == '<' {
		l.read()
		// check slash
		if l.peek() == '/' {
			startPos := l.pos - 1 // Account for '<' already read
			l.read()
			// read identifier
			ident := l.readIdentifier()
			// skip '>' if present
			if l.peek() == '>' {
				l.read()
			}
			token := Token{Type: TokenIdentifier, Lit: "/" + ident, StartPos: startPos, EndPos: l.pos}
			fmt.Printf("DEBUG LEXER: Produced token: %v\n", token)
			return token
		}
		// open tag: read tag name
		startPos := l.pos - 1 // Account for '<' already read
		ident := l.readIdentifier()
		// read attributes until '>'
		attrs := ""
		for {
			l.skipSpaces()
			if l.peek() == '>' || l.peek() == 0 {
				break
			}
			// read attribute key
			key := l.readIdentifier()
			if key == "" {
				// consume one char to avoid infinite loop
				if l.peek() == '>' || l.peek() == 0 {
					break
				}
				l.read()
				continue
			}
			l.skipSpaces()
			val := ""
			if l.peek() == '=' {
				l.read()
				l.skipSpaces()
				if l.peek() == '"' {
					val = l.readQuoted()
				} else {
					val = l.readIdentifier()
				}
			}
			if val != "" {
				attrs += " " + key + "=" + val
			} else {
				attrs += " " + key
			}
		}
		// skip '>' if present
		if l.peek() == '>' {
			l.read()
		}
		token := Token{Type: TokenIdentifier, Lit: ident + attrs, StartPos: startPos, EndPos: l.pos}
		fmt.Printf("DEBUG LEXER: Produced token: %v\n", token)
		return token
	}
	// check for template expressions '{{'
	if l.peek() == '{' && l.peekN(2) == '{' {
		token := l.readTemplateExpression()
		fmt.Printf("DEBUG LEXER: Produced template token: %v\n", token)
		return token
	}
	// otherwise read until next '<' or '{{'
	startPos := l.pos
	fmt.Printf("DEBUG LEXER: Starting text read at position %d\n", startPos)
	for {
		currentChar := l.peek()
		fmt.Printf("DEBUG LEXER: Reading char at pos %d: %q\n", l.pos, string(currentChar))
		if currentChar == '<' || currentChar == 0 {
			fmt.Printf("DEBUG LEXER: Breaking on '<' or EOF\n")
			break
		}
		// check for template start within text
		if currentChar == '{' && l.peekN(2) == '{' {
			fmt.Printf("DEBUG LEXER: Breaking on template start\n")
			break
		}
		l.read()
	}
	token := Token{Type: TokenText, Lit: l.input[startPos:l.pos], StartPos: startPos, EndPos: l.pos}
	fmt.Printf("DEBUG LEXER: Produced text token: %q (from %d to %d), next char: %q\n", token.Lit, startPos, l.pos, string(l.peek()))
	return token
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for {
		r := l.peek()
		if r == 0 || r == '>' || r == '<' || r == '/' || r == ' ' || r == '\n' || r == '\t' {
			break
		}
		l.read()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readQuoted() string {
	if l.peek() != '"' {
		return ""
	}
	// consume opening quote
	l.read()
	start := l.pos
	for {
		if l.peek() == '"' || l.peek() == 0 {
			break
		}
		l.read()
	}
	val := l.input[start:l.pos]
	// consume closing quote if present
	if l.peek() == '"' {
		l.read()
	}
	return "\"" + val + "\""
}

func (l *Lexer) skipSpaces() {
	for {
		r := l.peek()
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			l.read()
			continue
		}
		break
	}
}

// peekN looks ahead n characters without consuming them
func (l *Lexer) peekN(n int) rune {
	if l.pos+n-1 >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos+n-1])
}

// readTemplateExpression reads a complete template expression {{...}}
func (l *Lexer) readTemplateExpression() Token {
	startPos := l.pos

	// consume opening {{
	l.read() // consume first {
	l.read() // consume second {

	// read template content until }}
	contentStart := l.pos
	braceCount := 0 // track nested braces
	for {
		if l.peek() == '}' && l.peekN(2) == '}' {
			if braceCount == 0 {
				// found closing }}, and no nested braces
				break
			}
			braceCount-- // consume one level of nesting
		}
		if l.peek() == '{' && l.peekN(2) == '{' {
			braceCount++ // found nested template start
		}
		if l.peek() == 0 {
			// EOF - unclosed template expression
			// Return error token for unclosed template
			return Token{
				Type:     TokenError,
				Lit:      "ERROR: Unclosed template expression",
				StartPos: startPos,
				EndPos:   l.pos,
			}
		}
		l.read()
	}

	content := l.input[contentStart:l.pos]

	// consume closing }}
	l.read() // consume first }
	l.read() // consume second }

	// Basic validation of template content
	if err := l.validateTemplateContent(content); err != nil {
		return Token{
			Type:     TokenError,
			Lit:      "ERROR: " + err.Error(),
			StartPos: startPos,
			EndPos:   l.pos,
		}
	}

	return Token{
		Type:     TokenTemplateContent,
		Lit:      content,
		StartPos: startPos,
		EndPos:   l.pos,
	}
}

// validateTemplateContent performs basic syntax validation on template content
func (l *Lexer) validateTemplateContent(content string) error {
	if content == "" {
		return nil // empty templates are allowed
	}

	// Check for obviously invalid syntax
	content = strings.TrimSpace(content)

	// Check for lone operators
	if strings.HasPrefix(content, "|") {
		return fmt.Errorf("template cannot start with pipe operator")
	}
	// Check for consecutive pipes
	if strings.Contains(content, "||") {
		return fmt.Errorf("consecutive pipe operators")
	}

	// Check for unmatched quotes (basic check)
	singleQuotes := strings.Count(content, "'")
	if singleQuotes%2 != 0 {
		return fmt.Errorf("unmatched single quotes")
	}
	doubleQuotes := strings.Count(content, `"`)
	if doubleQuotes%2 != 0 {
		return fmt.Errorf("unmatched double quotes")
	}

	// Check for basic variable reference
	if !strings.HasPrefix(content, ".") && !strings.Contains(content, ".") {
		// Not a variable reference, could be function call or other valid syntax
		if !strings.Contains(content, "(") && !strings.Contains(content, ")") {
			// No parentheses, should be variable reference
			return fmt.Errorf("invalid template syntax: expected variable reference starting with '.'")
		}
	}

	return nil
}
