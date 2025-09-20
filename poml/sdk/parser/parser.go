package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/ZanzyTHEbar/poml/sdk/ast"
)

// VariableContext represents the context in which template variables are resolved
type VariableContext struct {
	Variables map[string]interface{} `json:"variables"` // available variables and their values
	Types     map[string]string      `json:"types"`     // variable type information
	Parent    *VariableContext       `json:"-"`         // parent context for scoping
}

// NewVariableContext creates a new empty variable context
func NewVariableContext() *VariableContext {
	return &VariableContext{
		Variables: make(map[string]interface{}),
		Types:     make(map[string]string),
		Parent:    nil,
	}
}

// NewVariableContextWithParent creates a new context with a parent
func NewVariableContextWithParent(parent *VariableContext) *VariableContext {
	return &VariableContext{
		Variables: make(map[string]interface{}),
		Types:     make(map[string]string),
		Parent:    parent,
	}
}

// AddVariable adds a variable to the context
func (vc *VariableContext) AddVariable(name string, value interface{}, varType string) {
	vc.Variables[name] = value
	if varType != "" {
		vc.Types[name] = varType
	}
}

// ResolveVariable resolves a variable by name, checking current and parent contexts
func (vc *VariableContext) ResolveVariable(name string) (interface{}, bool) {
	if value, exists := vc.Variables[name]; exists {
		return value, true
	}
	if vc.Parent != nil {
		return vc.Parent.ResolveVariable(name)
	}
	return nil, false
}

// GetVariableType gets the type of a variable
func (vc *VariableContext) GetVariableType(name string) (string, bool) {
	if varType, exists := vc.Types[name]; exists {
		return varType, true
	}
	if vc.Parent != nil {
		return vc.Parent.GetVariableType(name)
	}
	return "", false
}

// HasVariable checks if a variable exists in the context
func (vc *VariableContext) HasVariable(name string) bool {
	_, exists := vc.ResolveVariable(name)
	return exists
}

// GetAllVariables returns all variables available in this context (including parent contexts)
func (vc *VariableContext) GetAllVariables() map[string]interface{} {
	allVars := make(map[string]interface{})

	// Add from parent first (so child can override)
	if vc.Parent != nil {
		for k, v := range vc.Parent.GetAllVariables() {
			allVars[k] = v
		}
	}

	// Add from current context (will override parent if same key)
	for k, v := range vc.Variables {
		allVars[k] = v
	}

	return allVars
}

type Parser struct {
	lx      *Lexer
	cur     Token
	ctx     *VariableContext  // current variable context for template resolution
	funcReg *FunctionRegistry // function registry for validation
}

func NewParser(input string) *Parser {
	lx := NewLexer(input)
	ctx := NewVariableContext()
	funcReg := NewFunctionRegistry()
	return &Parser{lx: lx, cur: lx.Next(), ctx: ctx, funcReg: funcReg}
}

// NewParserWithContext creates a parser with a specific variable context
func NewParserWithContext(input string, ctx *VariableContext) *Parser {
	if ctx == nil {
		ctx = NewVariableContext()
	}
	lx := NewLexer(input)
	funcReg := NewFunctionRegistry()
	return &Parser{lx: lx, cur: lx.Next(), ctx: ctx, funcReg: funcReg}
}

// NewParserWithRegistry creates a parser with a specific function registry
func NewParserWithRegistry(input string, funcReg *FunctionRegistry) *Parser {
	lx := NewLexer(input)
	ctx := NewVariableContext()
	if funcReg == nil {
		funcReg = NewFunctionRegistry()
	}
	return &Parser{lx: lx, cur: lx.Next(), ctx: ctx, funcReg: funcReg}
}

// SetContext sets the variable context for the parser
func (p *Parser) SetContext(ctx *VariableContext) {
	if ctx == nil {
		ctx = NewVariableContext()
	}
	p.ctx = ctx
}

// GetContext returns the current variable context
func (p *Parser) GetContext() *VariableContext {
	return p.ctx
}

// SetFunctionRegistry sets the function registry for the parser
func (p *Parser) SetFunctionRegistry(funcReg *FunctionRegistry) {
	if funcReg == nil {
		funcReg = NewFunctionRegistry()
	}
	p.funcReg = funcReg
}

// GetFunctionRegistry returns the current function registry
func (p *Parser) GetFunctionRegistry() *FunctionRegistry {
	return p.funcReg
}

func (p *Parser) next() {
	p.cur = p.lx.Next()
}

func parseAttrsFromLit(lit string) []ast.Attr {
	res := []ast.Attr{}
	idx := strings.Index(lit, " ")
	if idx == -1 {
		return res
	}
	attrText := strings.TrimSpace(lit[idx+1:])
	parts := strings.Split(attrText, " ")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if eq := strings.Index(part, "="); eq != -1 {
			key := part[:eq]
			val := part[eq+1:]
			if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			}
			res = append(res, ast.Attr{Key: key, Val: val})
		} else {
			res = append(res, ast.Attr{Key: part, Val: ""})
		}
	}
	return res
}

func (p *Parser) Parse() (*ast.Document, error) {
	doc := &ast.Document{}
	for p.cur.Type != TokenEOF {
		if p.cur.Type == TokenText {
			if p.cur.Lit != "" {
				textNode := &ast.Text{Content: p.cur.Lit}
				// Source mapping for text nodes (using token positions)
				// Note: We can't set these directly as Text struct doesn't have source mapping fields
				doc.Children = append(doc.Children, textNode)
			}
			p.next()
			continue
		}
		if p.cur.Type == TokenError {
			return nil, fmt.Errorf("template syntax error: %s", p.cur.Lit)
		}
		if p.cur.Type == TokenTemplateContent {
			templateNode := &ast.TemplateExpr{
				Expression: p.cur.Lit,
				StartIndex: p.cur.StartPos,
				EndIndex:   p.cur.EndPos,
			}

			// Parse the template expression into AST components
			if parsed, hasPipeline := p.parseTemplateExpression(p.cur.Lit); parsed != nil {
				templateNode.Parsed = parsed
				templateNode.HasPipeline = hasPipeline
			}

			// Attempt to compile the template for validation with enhanced error reporting
			_, compileErr := p.compileTemplateExpression(p.cur.Lit)
			if compileErr != nil {
				templateNode.IsCompiled = false
				// Use contextual error analysis for better error reporting
				contextualError := p.analyzeTemplateError(p.cur.Lit, compileErr)

				// Additional validation using dedicated validation methods
				if varErr := p.validateTemplateVariables(p.cur.Lit); varErr != nil {
					contextualError.Suggestions = append(contextualError.Suggestions, fmt.Sprintf("Variable validation failed: %s", varErr.Error()))
				}
				if funcErr := p.validateTemplateFunctionsSimple(p.cur.Lit); funcErr != nil {
					contextualError.Suggestions = append(contextualError.Suggestions, fmt.Sprintf("Function validation failed: %s", funcErr.Error()))
				}
				templateNode.CompileError = p.formatTemplateError(contextualError)
				p.logTemplateError(p.cur.Lit, compileErr)
			} else {
				templateNode.IsCompiled = true
				templateNode.CompileError = ""
			}

			// Perform variable resolution validation with enhanced error reporting
			resolvedVars, unresolvedVars, varErr := p.resolveTemplateVariables(p.cur.Lit)
			templateNode.ResolvedVariables = resolvedVars
			templateNode.UnresolvedVariables = unresolvedVars
			if varErr != nil {
				templateNode.IsVariablesResolved = false
				// Use contextual error analysis for variable resolution errors
				contextualError := p.createContextualTemplateError(
					p.cur.Lit, p.cur.StartPos, p.cur.EndPos,
					TemplateErrorVariable, SeverityError, "Variable resolution failed")
				templateNode.VariableResolutionError = p.formatTemplateError(contextualError)
			} else {
				templateNode.IsVariablesResolved = true
				templateNode.VariableResolutionError = ""
			}

			// Perform function validation with enhanced error reporting
			validatedFuncs, invalidFuncs, funcDetails, funcErr := p.resolveTemplateFunctions(p.cur.Lit)
			templateNode.ValidatedFunctions = validatedFuncs
			templateNode.InvalidFunctions = invalidFuncs
			templateNode.FunctionValidationDetails = funcDetails

			// Use detailed validation for enhanced error reporting
			if detailedFuncValidation, detailedErr := p.validateTemplateFunctions(p.cur.Lit); detailedErr == nil {
				// Merge detailed validation results with existing function details
				if templateNode.FunctionValidationDetails == nil {
					templateNode.FunctionValidationDetails = make(map[string]string)
				}
				// Add detailed validation results, preserving existing details
				for funcName, validationResult := range detailedFuncValidation {
					if _, exists := templateNode.FunctionValidationDetails[funcName]; !exists {
						templateNode.FunctionValidationDetails[funcName] = validationResult
					}
				}
			}

			if funcErr != nil {
				templateNode.IsFunctionsValidated = false
				// Use contextual error analysis for function validation errors
				contextualError := p.createContextualTemplateError(
					p.cur.Lit, p.cur.StartPos, p.cur.EndPos,
					TemplateErrorFunction, SeverityError, "Function validation failed")
				templateNode.FunctionValidationError = p.formatTemplateError(contextualError)
			} else {
				templateNode.IsFunctionsValidated = true
				templateNode.FunctionValidationError = ""
			}

			// Calculate enhanced source mapping with line/column information
			startPos, endPos := p.calculateSourceRange(p.cur.StartPos, p.cur.EndPos)
			templateNode.StartLine = startPos.Line
			templateNode.StartColumn = startPos.Column
			templateNode.EndLine = endPos.Line
			templateNode.EndColumn = endPos.Column
			templateNode.SourceText = p.getSourceContext(p.cur.StartPos, p.cur.EndPos, 1)

			doc.Children = append(doc.Children, templateNode)
			p.next()
			continue
		}

		if p.cur.Type == TokenIdentifier {
			// opening tag (may include attrs in Lit)
			name := p.cur.Lit
			if len(name) > 0 && name[0] == '/' {
				return nil, errors.New("unexpected closing tag: " + name)
			}
			p.next()
			elem := &ast.Element{Name: name}
			// Set source mapping start position
			elem.StartIndex = p.cur.StartPos
			// parse attributes if present
			if i := strings.Index(name, " "); i != -1 {
				elem.Name = name[:i]
				elem.Attrs = parseAttrsFromLit(name)
			}
			p.next()
			for p.cur.Type != TokenEOF {

				// closing tag
				if p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+elem.Name {
					fmt.Printf("DEBUG PARSER: Found closing tag for element %s\n", elem.Name)
					// Set source mapping end position before consuming the closing tag
					elem.EndIndex = p.cur.StartPos
					p.next()
					break
				}
				// text child
				if p.cur.Type == TokenText {
					fmt.Printf("DEBUG PARSER: Adding text child: %q\n", p.cur.Lit)
					if p.cur.Lit != "" {
						elem.Children = append(elem.Children, &ast.Text{Content: p.cur.Lit})
					}
					p.next()
					continue
				}
				// template expression child
				if p.cur.Type == TokenError {
					return nil, fmt.Errorf("template syntax error: %s", p.cur.Lit)
				}
				fmt.Printf("DEBUG PARSER: Checking template condition: type=%d, TokenTemplateContent=%d\n", p.cur.Type, TokenTemplateContent)
				if p.cur.Type == TokenTemplateContent {
					fmt.Printf("DEBUG PARSER: Adding template child: %q\n", p.cur.Lit)
					templateNode := &ast.TemplateExpr{
						Expression: p.cur.Lit,
						StartIndex: p.cur.StartPos,
						EndIndex:   p.cur.EndPos,
					}

					// Parse the template expression into AST components
					if parsed, hasPipeline := p.parseTemplateExpression(p.cur.Lit); parsed != nil {
						templateNode.Parsed = parsed
						templateNode.HasPipeline = hasPipeline
					}

					// Attempt to compile the template for validation
					_, compileErr := p.compileTemplateExpression(p.cur.Lit)
					if compileErr != nil {
						templateNode.IsCompiled = false

						// Additional validation using dedicated validation methods
						contextualError := p.analyzeTemplateError(p.cur.Lit, compileErr)
						if varErr := p.validateTemplateVariables(p.cur.Lit); varErr != nil {
							contextualError.Suggestions = append(contextualError.Suggestions, fmt.Sprintf("Variable validation failed: %s", varErr.Error()))
						}
						if funcErr := p.validateTemplateFunctionsSimple(p.cur.Lit); funcErr != nil {
							contextualError.Suggestions = append(contextualError.Suggestions, fmt.Sprintf("Function validation failed: %s", funcErr.Error()))
						}
						templateNode.CompileError = p.formatTemplateError(contextualError)
						p.logTemplateError(p.cur.Lit, compileErr)
					} else {
						templateNode.IsCompiled = true
						templateNode.CompileError = ""
					}

					// Perform variable resolution validation
					resolvedVars, unresolvedVars, varErr := p.resolveTemplateVariables(p.cur.Lit)
					templateNode.ResolvedVariables = resolvedVars
					templateNode.UnresolvedVariables = unresolvedVars
					if varErr != nil {
						templateNode.IsVariablesResolved = false
						templateNode.VariableResolutionError = varErr.Error()
					} else {
						templateNode.IsVariablesResolved = true
						templateNode.VariableResolutionError = ""
					}

					// Perform function validation
					validatedFuncs, invalidFuncs, funcDetails, funcErr := p.resolveTemplateFunctions(p.cur.Lit)
					templateNode.ValidatedFunctions = validatedFuncs
					templateNode.InvalidFunctions = invalidFuncs
					templateNode.FunctionValidationDetails = funcDetails

					// Use detailed validation for enhanced error reporting
					if detailedFuncValidation, detailedErr := p.validateTemplateFunctions(p.cur.Lit); detailedErr == nil {
						// Merge detailed validation results with existing function details
						if templateNode.FunctionValidationDetails == nil {
							templateNode.FunctionValidationDetails = make(map[string]string)
						}
						for funcName, validationResult := range detailedFuncValidation {
							if _, exists := templateNode.FunctionValidationDetails[funcName]; !exists {
								templateNode.FunctionValidationDetails[funcName] = validationResult
							}
						}
					}

					if funcErr != nil {
						templateNode.IsFunctionsValidated = false
						templateNode.FunctionValidationError = p.createTemplateError(
							p.cur.Lit, p.cur.StartPos, p.cur.EndPos, "Function Validation Error", funcErr.Error())
					} else {
						templateNode.IsFunctionsValidated = true
						templateNode.FunctionValidationError = ""
					}

					// Calculate enhanced source mapping with line/column information
					startPos, endPos := p.calculateSourceRange(p.cur.StartPos, p.cur.EndPos)
					templateNode.StartLine = startPos.Line
					templateNode.StartColumn = startPos.Column
					templateNode.EndLine = endPos.Line
					templateNode.EndColumn = endPos.Column
					templateNode.SourceText = p.getSourceContext(p.cur.StartPos, p.cur.EndPos, 1)

					elem.Children = append(elem.Children, templateNode)
					fmt.Printf("DEBUG PARSER: Added template node to children, total children now: %d\n", len(elem.Children))
					p.next()
					continue
				}
				// nested constructs
				if p.cur.Type == TokenIdentifier {
					tok := p.cur.Lit
					fmt.Printf("DEBUG PARSER: Handling nested element: %q\n", tok)
					// handle if
					if strings.HasPrefix(tok, "if") {
						// parse cond attr
						cond := ""
						if j := strings.Index(tok, " "); j != -1 {
							attrText := strings.TrimSpace(tok[j+1:])
							if idx := strings.Index(attrText, "cond="); idx != -1 {
								val := attrText[idx+5:]
								if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
									val = val[1 : len(val)-1]
								}
								cond = val
							}
						}
						p.next()
						ifNode := &ast.IfNode{Cond: cond}
						// Set source mapping start position
						ifNode.StartIndex = p.cur.StartPos
						for p.cur.Type != TokenEOF && !(p.cur.Type == TokenIdentifier && p.cur.Lit == "/if") {
							if p.cur.Type == TokenText {
								if p.cur.Lit != "" {
									ifNode.Children = append(ifNode.Children, &ast.Text{Content: p.cur.Lit})
								}
								p.next()
								continue
							}
							if p.cur.Type == TokenError {
								return nil, fmt.Errorf("template syntax error: %s", p.cur.Lit)
							}
							if p.cur.Type == TokenTemplateContent {
								templateNode := &ast.TemplateExpr{
									Expression: p.cur.Lit,
									StartIndex: p.cur.StartPos,
									EndIndex:   p.cur.EndPos,
								}

								// Parse the template expression into AST components
								if parsed, hasPipeline := p.parseTemplateExpression(p.cur.Lit); parsed != nil {
									templateNode.Parsed = parsed
									templateNode.HasPipeline = hasPipeline
								}

								// Attempt to compile the template for validation
								_, compileErr := p.compileTemplateExpression(p.cur.Lit)
								if compileErr != nil {
									templateNode.IsCompiled = false

									// Additional validation using dedicated validation methods
									contextualError := p.analyzeTemplateError(p.cur.Lit, compileErr)
									if varErr := p.validateTemplateVariables(p.cur.Lit); varErr != nil {
										contextualError.Suggestions = append(contextualError.Suggestions, fmt.Sprintf("Variable validation failed: %s", varErr.Error()))
									}
									if funcErr := p.validateTemplateFunctionsSimple(p.cur.Lit); funcErr != nil {
										contextualError.Suggestions = append(contextualError.Suggestions, fmt.Sprintf("Function validation failed: %s", funcErr.Error()))
									}
									templateNode.CompileError = p.formatTemplateError(contextualError)
									p.logTemplateError(p.cur.Lit, compileErr)
								} else {
									templateNode.IsCompiled = true
									templateNode.CompileError = ""
								}

								// Perform variable resolution validation
								resolvedVars, unresolvedVars, varErr := p.resolveTemplateVariables(p.cur.Lit)
								templateNode.ResolvedVariables = resolvedVars
								templateNode.UnresolvedVariables = unresolvedVars
								if varErr != nil {
									templateNode.IsVariablesResolved = false
									templateNode.VariableResolutionError = varErr.Error()
								} else {
									templateNode.IsVariablesResolved = true
									templateNode.VariableResolutionError = ""
								}

								// Use detailed validation for enhanced error reporting
								if detailedFuncValidation, detailedErr := p.validateTemplateFunctions(p.cur.Lit); detailedErr == nil {
									// Add detailed validation results to function details
									if templateNode.FunctionValidationDetails == nil {
										templateNode.FunctionValidationDetails = make(map[string]string)
									}
									for funcName, validationResult := range detailedFuncValidation {
										if _, exists := templateNode.FunctionValidationDetails[funcName]; !exists {
											templateNode.FunctionValidationDetails[funcName] = validationResult
										}
									}
								}

								// Perform function validation
								validatedFuncs, invalidFuncs, funcDetails, funcErr := p.resolveTemplateFunctions(p.cur.Lit)
								templateNode.ValidatedFunctions = validatedFuncs
								templateNode.InvalidFunctions = invalidFuncs
								templateNode.FunctionValidationDetails = funcDetails
								if funcErr != nil {
									templateNode.IsFunctionsValidated = false
									templateNode.FunctionValidationError = p.createTemplateError(
										p.cur.Lit, p.cur.StartPos, p.cur.EndPos, "Function Validation Error", funcErr.Error())
								} else {
									templateNode.IsFunctionsValidated = true
									templateNode.FunctionValidationError = ""
								}

								// Calculate enhanced source mapping with line/column information
								startPos, endPos := p.calculateSourceRange(p.cur.StartPos, p.cur.EndPos)
								templateNode.StartLine = startPos.Line
								templateNode.StartColumn = startPos.Column
								templateNode.EndLine = endPos.Line
								templateNode.EndColumn = endPos.Column
								templateNode.SourceText = p.getSourceContext(p.cur.StartPos, p.cur.EndPos, 1)

								ifNode.Children = append(ifNode.Children, templateNode)
								p.next()
								continue
							}
							if p.cur.Type == TokenIdentifier {
								childName := p.cur.Lit
								p.next()
								child := &ast.Element{Name: childName}
								if k := strings.Index(childName, " "); k != -1 {
									child.Name = childName[:k]
									child.Attrs = parseAttrsFromLit(childName)
								}
								for p.cur.Type != TokenEOF && !(p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+child.Name) {
									if p.cur.Type == TokenText && p.cur.Lit != "" {
										child.Children = append(child.Children, &ast.Text{Content: p.cur.Lit})
									}
									p.next()
								}
								if p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+child.Name {
									p.next()
								}
								ifNode.Children = append(ifNode.Children, child)
								continue
							}
						}
						if p.cur.Type == TokenIdentifier && p.cur.Lit == "/if" {
							// Set source mapping end position
							ifNode.EndIndex = p.cur.StartPos
							p.next()
						}
						elem.Children = append(elem.Children, ifNode)
						continue
					}
					// handle img
					if strings.HasPrefix(tok, "img") {
						src := ""
						alt := ""
						position := ""
						if j := strings.Index(tok, " "); j != -1 {
							attrs := parseAttrsFromLit(tok)
							for _, attr := range attrs {
								switch attr.Key {
								case "src":
									src = attr.Val
								case "alt":
									alt = attr.Val
								case "position":
									position = attr.Val
								}
							}
						}
						imgNode := &ast.ImgNode{Src: src, Alt: alt, Position: position}
						// Set source mapping start position
						imgNode.StartIndex = p.cur.StartPos
						p.next()
						// expect closing tag
						if p.cur.Type == TokenIdentifier && p.cur.Lit == "/img" {
							// Set source mapping end position
							imgNode.EndIndex = p.cur.StartPos
							p.next()
							elem.Children = append(elem.Children, imgNode)
							continue
						}
					}
					// handle env
					if strings.HasPrefix(tok, "env") {
						// parse presentation attr
						presentation := ""
						if j := strings.Index(tok, " "); j != -1 {
							attrText := strings.TrimSpace(tok[j+1:])
							if idx := strings.Index(attrText, "presentation="); idx != -1 {
								val := attrText[idx+13:]
								if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
									val = val[1 : len(val)-1]
								}
								presentation = val
							}
						}
						p.next()
						envNode := &ast.EnvNode{Presentation: presentation}
						// Set source mapping start position
						envNode.StartIndex = p.cur.StartPos
						for p.cur.Type != TokenEOF && !(p.cur.Type == TokenIdentifier && p.cur.Lit == "/env") {
							if p.cur.Type == TokenText {
								if p.cur.Lit != "" {
									envNode.Children = append(envNode.Children, &ast.Text{Content: p.cur.Lit})
								}
								p.next()
								continue
							}
							if p.cur.Type == TokenIdentifier {
								childName := p.cur.Lit
								p.next()
								child := &ast.Element{Name: childName}
								if k := strings.Index(childName, " "); k != -1 {
									child.Name = childName[:k]
									child.Attrs = parseAttrsFromLit(childName)
								}
								for p.cur.Type != TokenEOF && !(p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+child.Name) {
									if p.cur.Type == TokenText && p.cur.Lit != "" {
										child.Children = append(child.Children, &ast.Text{Content: p.cur.Lit})
									}
									p.next()
								}
								if p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+child.Name {
									p.next()
								}
								envNode.Children = append(envNode.Children, child)
								continue
							}
						}
						if p.cur.Type == TokenIdentifier && p.cur.Lit == "/env" {
							// Set source mapping end position
							envNode.EndIndex = p.cur.StartPos
							p.next()
						}
						elem.Children = append(elem.Children, envNode)
						continue
					}
					// handle for
					if strings.HasPrefix(tok, "for") {
						varName := ""
						inExpr := ""
						if j := strings.Index(tok, " "); j != -1 {
							attrText := strings.TrimSpace(tok[j+1:])
							parts := strings.Split(attrText, " ")
							for _, part := range parts {
								if strings.HasPrefix(part, "var=") {
									v := part[4:]
									if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
										v = v[1 : len(v)-1]
									}
									varName = v
								}
								if strings.HasPrefix(part, "in=") {
									i := part[3:]
									if len(i) >= 2 && i[0] == '"' && i[len(i)-1] == '"' {
										i = i[1 : len(i)-1]
									}
									inExpr = i
								}
							}
						}
						p.next()
						forNode := &ast.ForNode{Var: varName, In: inExpr}
						// Set source mapping start position
						forNode.StartIndex = p.cur.StartPos
						for p.cur.Type != TokenEOF && !(p.cur.Type == TokenIdentifier && p.cur.Lit == "/for") {
							if p.cur.Type == TokenText {
								if p.cur.Lit != "" {
									forNode.Children = append(forNode.Children, &ast.Text{Content: p.cur.Lit})
								}
								p.next()
								continue
							}
							if p.cur.Type == TokenIdentifier {
								childName := p.cur.Lit
								p.next()
								child := &ast.Element{Name: childName}
								if i := strings.Index(childName, " "); i != -1 {
									child.Name = childName[:i]
									child.Attrs = parseAttrsFromLit(childName)
								}
								for p.cur.Type != TokenEOF && !(p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+child.Name) {
									if p.cur.Type == TokenText && p.cur.Lit != "" {
										child.Children = append(child.Children, &ast.Text{Content: p.cur.Lit})
									}
									p.next()
								}
								if p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+child.Name {
									p.next()
								}
								forNode.Children = append(forNode.Children, child)
								continue
							}
						}
						if p.cur.Type == TokenIdentifier && p.cur.Lit == "/for" {
							// Set source mapping end position
							forNode.EndIndex = p.cur.StartPos
							p.next()
						}
						elem.Children = append(elem.Children, forNode)
						continue
					}
					// generic nested element
					fmt.Printf("DEBUG PARSER: Handling generic nested element: %q\n", p.cur.Lit)
					childName := p.cur.Lit
					p.next()
					fmt.Printf("DEBUG PARSER: After next() in generic handler, current token: type=%d, lit=%q\n", p.cur.Type, p.cur.Lit)
					child := &ast.Element{Name: childName}
					if i := strings.Index(childName, " "); i != -1 {
						child.Name = childName[:i]
						child.Attrs = parseAttrsFromLit(childName)
					}
					fmt.Printf("DEBUG PARSER: Starting child processing loop for element %s, looking for closing tag %s\n", child.Name, "/"+child.Name)
					for p.cur.Type != TokenEOF && !(p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+child.Name) {
						fmt.Printf("DEBUG PARSER: In child loop, current token: type=%d, lit=%q\n", p.cur.Type, p.cur.Lit)
						if p.cur.Type == TokenText && p.cur.Lit != "" {
							child.Children = append(child.Children, &ast.Text{Content: p.cur.Lit})
							p.next()
							continue
						}
						if p.cur.Type == TokenTemplateContent {
							fmt.Printf("DEBUG PARSER: Processing template token in child loop: %q\n", p.cur.Lit)
							// Reconstruct full template expression with braces
							fullExpression := "{{" + p.cur.Lit + "}}"
							templateNode := &ast.TemplateExpr{
								Expression: fullExpression,
								StartIndex: p.cur.StartPos,
								EndIndex:   p.cur.EndPos,
							}

							// Calculate source mapping with line/column information
							startPos, endPos := p.calculateSourceRange(p.cur.StartPos, p.cur.EndPos)
							templateNode.StartLine = startPos.Line
							templateNode.StartColumn = startPos.Column
							templateNode.EndLine = endPos.Line
							templateNode.EndColumn = endPos.Column
							templateNode.SourceText = p.getSourceContext(p.cur.StartPos, p.cur.EndPos, 1)

							// Parse the template expression into AST components
							if parsed, hasPipeline := p.parseTemplateExpression(p.cur.Lit); parsed != nil {
								templateNode.Parsed = parsed
								templateNode.HasPipeline = hasPipeline
							}

							// Attempt to compile the template for validation
							_, compileErr := p.compileTemplateExpression(p.cur.Lit)
							if compileErr != nil {
								templateNode.IsCompiled = false
								templateNode.CompileError = p.createTemplateError(
									p.cur.Lit, p.cur.StartPos, p.cur.EndPos, "Template Compilation Error", compileErr.Error())
							} else {
								templateNode.IsCompiled = true
								templateNode.CompileError = ""
							}

							child.Children = append(child.Children, templateNode)
							p.next()
							continue
						}
						// For other token types, just advance
						p.next()
					}
					if p.cur.Type == TokenIdentifier && p.cur.Lit == "/"+child.Name {
						p.next()
					}
					elem.Children = append(elem.Children, child)
					continue
				}
			}
			fmt.Printf("DEBUG PARSER: Finished processing element %s, children count: %d\n", elem.Name, len(elem.Children))
			doc.Children = append(doc.Children, elem)
			continue
		}
		p.next()
	}
	return doc, nil
}

// parseTemplatePipeline parses expressions with pipeline operations like "input | function"
func (p *Parser) parseTemplatePipeline(expr string) (ast.TemplateComponent, bool) {
	parts := strings.Split(expr, "|")
	if len(parts) < 2 {
		return nil, false
	}

	// Parse the input (left side of first pipe)
	inputExpr := strings.TrimSpace(parts[0])
	var input ast.TemplateComponent

	if strings.HasPrefix(inputExpr, ".") {
		input = &ast.TemplateVariable{
			Name:     inputExpr[1:],
			FullName: inputExpr,
		}
	} else {
		// Try parsing as function call
		if funcComp := p.parseTemplateFunction(inputExpr); funcComp != nil {
			input = funcComp
		} else {
			// Fallback to variable
			input = &ast.TemplateVariable{
				Name:     inputExpr,
				FullName: "." + inputExpr,
			}
		}
	}

	// Parse function calls (right side of pipes)
	for i := 1; i < len(parts); i++ {
		funcExpr := strings.TrimSpace(parts[i])
		function := p.parseTemplateFunction(funcExpr)
		if function == nil {
			// Invalid function, return nil
			return nil, false
		}

		// Create pipeline with current input and function
		input = &ast.TemplatePipeline{
			Input:    input,
			Function: function,
		}
	}

	return input, true
}

// parseTemplateExpression parses any template expression and returns the AST component
// Returns the parsed component and whether it contains pipeline operations
func (p *Parser) parseTemplateExpression(expr string) (ast.TemplateComponent, bool) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, false
	}

	// Check if expression contains pipeline operations
	if strings.Contains(expr, "|") {
		return p.parseTemplatePipeline(expr)
	}

	// Parse simple expressions (no pipelines)
	if strings.HasPrefix(expr, ".") {
		// Variable reference like .Name
		return &ast.TemplateVariable{
			Name:     expr[1:],
			FullName: expr,
		}, false
	} else {
		// Try parsing as function call
		if funcComp := p.parseTemplateFunction(expr); funcComp != nil {
			return funcComp, false
		} else {
			// Fallback to variable (without leading dot)
			return &ast.TemplateVariable{
				Name:     expr,
				FullName: "." + expr,
			}, false
		}
	}
}

// parseTemplateFunction parses a function call like "upper" or 'replace "old" "new"'
func (p *Parser) parseTemplateFunction(expr string) *ast.TemplateFunction {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// Simple function without arguments
	if !strings.Contains(expr, "(") && !strings.Contains(expr, `"`) && !strings.Contains(expr, `'`) {
		return &ast.TemplateFunction{
			Name: expr,
			Args: []ast.TemplateArg{},
		}
	}

	// Parse function with arguments
	// For now, handle simple cases like: replace "old" "new"
	parts := strings.Fields(expr)
	if len(parts) == 0 {
		return nil
	}

	function := &ast.TemplateFunction{
		Name: parts[0],
		Args: []ast.TemplateArg{},
	}

	// Parse arguments (simplified - handles quoted strings)
	for i := 1; i < len(parts); i++ {
		arg := parts[i]
		argType := "string"

		// Remove quotes if present
		if (strings.HasPrefix(arg, `"`) && strings.HasSuffix(arg, `"`)) ||
			(strings.HasPrefix(arg, `'`) && strings.HasSuffix(arg, `'`)) {
			arg = arg[1 : len(arg)-1]
		} else if strings.HasPrefix(arg, ".") {
			argType = "variable"
		} else if _, err := parseNumber(arg); err == nil {
			argType = "number"
		}

		function.Args = append(function.Args, ast.TemplateArg{
			Type:  argType,
			Value: arg,
		})
	}

	return function
}

// parseNumber is a helper to check if a string represents a number
func parseNumber(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// extractVariablesFromTemplate extracts all variable references from a template expression
func (p *Parser) extractVariablesFromTemplate(expr string) []string {
	variables := make([]string, 0)
	expr = strings.TrimSpace(expr)

	// Simple regex-free approach to extract variable references
	parts := strings.Fields(expr)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, ".") {
			// Remove leading dot and any trailing operators
			varName := strings.TrimPrefix(part, ".")
			// Remove any pipeline operators or function calls
			if idx := strings.Index(varName, "|"); idx != -1 {
				varName = varName[:idx]
			}
			if idx := strings.Index(varName, "("); idx != -1 {
				varName = varName[:idx]
			}
			varName = strings.TrimSpace(varName)
			if varName != "" && !contains(variables, varName) {
				variables = append(variables, varName)
			}
		}
	}

	return variables
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// resolveTemplateVariables validates and resolves all variables in a template expression
func (p *Parser) resolveTemplateVariables(expr string) ([]string, []string, error) {
	variables := p.extractVariablesFromTemplate(expr)
	resolved := make([]string, 0)
	unresolved := make([]string, 0)

	for _, varName := range variables {
		if p.ctx.HasVariable(varName) {
			resolved = append(resolved, varName)
		} else {
			unresolved = append(unresolved, varName)
		}
	}

	var err error
	if len(unresolved) > 0 {
		err = fmt.Errorf("unresolved variables: %v", unresolved)
	}

	return resolved, unresolved, err
}

// validateTemplateVariables checks if all variables in a template are resolvable
// This is a convenience method that only returns validation errors without variable details.
// Use this when you only need to know if validation passes/fails, not the specific variables.
func (p *Parser) validateTemplateVariables(expr string) error {
	_, _, err := p.resolveTemplateVariables(expr)
	return err
}

// extractFunctionsFromTemplate extracts all potential function calls from a template expression
func (p *Parser) extractFunctionsFromTemplate(expr string) []string {
	functions := make([]string, 0)
	expr = strings.TrimSpace(expr)

	// Simple approach to extract function names from template expressions
	// This handles basic patterns like "functionName arg1 arg2 | otherFunc"
	parts := strings.Fields(expr)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Skip if it starts with dot (variable) or is a pipe operator
		if strings.HasPrefix(part, ".") || part == "|" {
			continue
		}

		// Extract potential function name by removing parentheses
		funcName := strings.TrimSuffix(strings.TrimSuffix(part, ")"), "(")
		if funcName != "" {
			// Check if this looks like a valid function name (contains letters)
			hasLetter := false
			for _, r := range funcName {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					hasLetter = true
					break
				}
			}

			if hasLetter {
				// Avoid duplicates
				found := false
				for _, existing := range functions {
					if existing == funcName {
						found = true
						break
					}
				}
				if !found {
					functions = append(functions, funcName)
				}
			}
		}
	}

	return functions
}

// validateTemplateFunctions validates all function calls in a template expression
// Returns detailed validation results for each function found in the expression.
// Use this when you need to know the status of individual functions, not just overall validity.
func (p *Parser) validateTemplateFunctions(expr string) (map[string]string, error) {
	functions := p.extractFunctionsFromTemplate(expr)
	validationResults := make(map[string]string)

	hasErrors := false
	for _, funcName := range functions {
		sig, exists := p.funcReg.GetFunction(funcName)
		if !exists {
			validationResults[funcName] = fmt.Sprintf("unknown function: %s", funcName)
			hasErrors = true
		} else {
			validationResults[funcName] = fmt.Sprintf("valid function: %s (%s)", sig.Description, sig.ReturnType)
		}
	}

	if hasErrors {
		return validationResults, fmt.Errorf("function validation failed for some functions")
	}

	return validationResults, nil
}

// resolveTemplateFunctions validates and resolves all function calls in a template expression
func (p *Parser) resolveTemplateFunctions(expr string) ([]string, []string, map[string]string, error) {
	functions := p.extractFunctionsFromTemplate(expr)
	validated := make([]string, 0)
	invalid := make([]string, 0)
	details := make(map[string]string)

	for _, funcName := range functions {
		sig, exists := p.funcReg.GetFunction(funcName)
		if exists {
			validated = append(validated, funcName)
			details[funcName] = fmt.Sprintf("valid: %s", sig.Description)
		} else {
			invalid = append(invalid, funcName)
			details[funcName] = fmt.Sprintf("unknown function: %s", funcName)
		}
	}

	var err error
	if len(invalid) > 0 {
		err = fmt.Errorf("invalid functions found: %v", invalid)
	}

	return validated, invalid, details, err
}

// validateTemplateFunctionsSimple checks if all functions in a template are valid
// This is a convenience method that only returns validation errors without function details.
// Use this when you only need to know if validation passes/fails, not the specific functions.
func (p *Parser) validateTemplateFunctionsSimple(expr string) error {
	_, _, _, err := p.resolveTemplateFunctions(expr)
	return err
}

// SourcePosition represents a position in the source code with line and column information
type SourcePosition struct {
	Line   int // 1-based line number
	Column int // 1-based column number
	Index  int // character index in source
}

// calculateLineColumn calculates line and column information from a character index
func (p *Parser) calculateLineColumn(charIndex int) SourcePosition {
	if charIndex < 0 || charIndex >= len(p.lx.input) {
		return SourcePosition{Line: -1, Column: -1, Index: charIndex}
	}

	line := 1
	column := 1
	currentIndex := 0

	for currentIndex < charIndex && currentIndex < len(p.lx.input) {
		if p.lx.input[currentIndex] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
		currentIndex++
	}

	return SourcePosition{
		Line:   line,
		Column: column,
		Index:  charIndex,
	}
}

// calculateSourceRange calculates source position information for a character range
func (p *Parser) calculateSourceRange(startIndex, endIndex int) (SourcePosition, SourcePosition) {
	startPos := p.calculateLineColumn(startIndex)
	endPos := p.calculateLineColumn(endIndex)
	return startPos, endPos
}

// getSourceContext extracts source text context around a position for error reporting
func (p *Parser) getSourceContext(startIndex, endIndex int, contextLines int) string {
	if startIndex < 0 || endIndex > len(p.lx.input) || startIndex > endIndex {
		return ""
	}

	// Find the line containing the error
	errorLineStart := startIndex
	for errorLineStart > 0 && p.lx.input[errorLineStart-1] != '\n' {
		errorLineStart--
	}
	errorLineEnd := endIndex
	for errorLineEnd < len(p.lx.input) && p.lx.input[errorLineEnd] != '\n' {
		errorLineEnd++
	}

	// Expand context to include surrounding lines
	contextStart := errorLineStart
	contextEnd := errorLineEnd

	// Add lines before the error
	for i := 0; i < contextLines && contextStart > 0; i++ {
		// Move to start of previous line
		for contextStart > 0 && p.lx.input[contextStart-1] != '\n' {
			contextStart--
		}
		// Include the newline character
		if contextStart > 0 {
			contextStart--
		}
	}

	// Add lines after the error
	for i := 0; i < contextLines && contextEnd < len(p.lx.input); i++ {
		// Move to end of next line (including newline)
		for contextEnd < len(p.lx.input) && p.lx.input[contextEnd] != '\n' {
			contextEnd++
		}
		if contextEnd < len(p.lx.input) {
			contextEnd++ // Include the newline
		}
	}

	if contextStart >= contextEnd {
		return ""
	}

	return p.lx.input[contextStart:contextEnd]
}

// TemplateError represents a categorized template error with context and suggestions
type TemplateError struct {
	Category    TemplateErrorCategory
	Severity    TemplateErrorSeverity
	Message     string
	Suggestions []string
	Location    SourcePosition
	Context     string
	Expression  string
}

// TemplateErrorCategory defines different types of template errors
type TemplateErrorCategory int

const (
	TemplateErrorSyntax TemplateErrorCategory = iota
	TemplateErrorSemantic
	TemplateErrorCompilation
	TemplateErrorVariable
	TemplateErrorFunction
	TemplateErrorPipeline
)

// TemplateErrorSeverity defines error severity levels
type TemplateErrorSeverity int

const (
	SeverityInfo TemplateErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityFatal
)

// String returns the string representation of the error category
func (c TemplateErrorCategory) String() string {
	switch c {
	case TemplateErrorSyntax:
		return "Syntax Error"
	case TemplateErrorSemantic:
		return "Semantic Error"
	case TemplateErrorCompilation:
		return "Compilation Error"
	case TemplateErrorVariable:
		return "Variable Error"
	case TemplateErrorFunction:
		return "Function Error"
	case TemplateErrorPipeline:
		return "Pipeline Error"
	default:
		return "Unknown Error"
	}
}

// String returns the string representation of the error severity
func (s TemplateErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// createTemplateError creates a detailed error message with source location information
func (p *Parser) createTemplateError(expr string, startIndex, endIndex int, errorType, message string) string {
	startPos, endPos := p.calculateSourceRange(startIndex, endIndex)
	sourceContext := p.getSourceContext(startIndex, endIndex, 1)

	errorMsg := fmt.Sprintf("%s at line %d, column %d: %s (expression: %s)", errorType, startPos.Line, startPos.Column, message, expr)
	if sourceContext != "" {
		errorMsg += fmt.Sprintf("\nContext: %s", sourceContext)
		if startPos.Line == endPos.Line {
			// Add pointer to the error location on the same line
			pointer := strings.Repeat(" ", startPos.Column-1) + strings.Repeat("^", endPos.Column-startPos.Column)
			errorMsg += fmt.Sprintf("\n%s", pointer)
		}
	}

	return errorMsg
}

// createContextualTemplateError creates a sophisticated template error with categorization and suggestions
func (p *Parser) createContextualTemplateError(expr string, startIndex, endIndex int, category TemplateErrorCategory, severity TemplateErrorSeverity, baseMessage string) *TemplateError {
	startPos, _ := p.calculateSourceRange(startIndex, endIndex)
	sourceContext := p.getSourceContext(startIndex, endIndex, 1)

	templateError := &TemplateError{
		Category:    category,
		Severity:    severity,
		Message:     baseMessage,
		Location:    startPos,
		Context:     sourceContext,
		Expression:  expr,
		Suggestions: p.generateTemplateSuggestions(expr, category),
	}

	return templateError
}

// generateTemplateSuggestions provides context-aware suggestions for common template errors
func (p *Parser) generateTemplateSuggestions(expr string, category TemplateErrorCategory) []string {
	var suggestions []string

	switch category {
	case TemplateErrorSyntax:
		if strings.Contains(expr, "{{") && !strings.Contains(expr, "}}") {
			suggestions = append(suggestions, "Missing closing '}}' - add the closing braces")
		} else if strings.Contains(expr, "}}") && !strings.Contains(expr, "{{") {
			suggestions = append(suggestions, "Missing opening '{{' - add the opening braces")
		} else if strings.Contains(expr, "|") {
			if strings.HasPrefix(strings.TrimSpace(expr), "|") {
				suggestions = append(suggestions, "Pipeline cannot start with '|' - remove the leading pipe or add a variable before it")
			} else if strings.HasSuffix(strings.TrimSpace(expr), "|") {
				suggestions = append(suggestions, "Pipeline cannot end with '|' - remove the trailing pipe or add a function after it")
			} else if strings.Contains(expr, "||") {
				suggestions = append(suggestions, "Consecutive pipes '||' are not allowed - separate functions with spaces")
			}
		}

	case TemplateErrorVariable:
		// First validate basic variable syntax
		if varErr := p.validateTemplateVariable(expr); varErr != nil {
			suggestions = append(suggestions, fmt.Sprintf("Syntax error: %s", varErr.Error()))
			suggestions = append(suggestions, "Variables should start with a dot (.) for template expressions")
			suggestions = append(suggestions, "Example: '.Name', '.Count', '.User.Email'")
		} else if strings.Contains(expr, ".") {
			if strings.HasPrefix(expr, ".") {
				suggestions = append(suggestions, "Variable syntax is valid, but check if the variable exists in the current context")
				suggestions = append(suggestions, "Ensure the variable is defined before this template expression")
				suggestions = append(suggestions, "Consider providing a default value if the variable might be undefined")
			} else {
				suggestions = append(suggestions, "Check if the variable is defined in the current context")
				suggestions = append(suggestions, "Consider using a different variable name if this one is not available")
			}
		} else {
			suggestions = append(suggestions, "Variables should be referenced with dot notation (e.g., '.Name')")
			suggestions = append(suggestions, "Check if you meant to use a string literal instead of a variable")
		}

	case TemplateErrorFunction:
		// Check for syntax issues first
		if strings.Contains(expr, "(") && !strings.Contains(expr, ")") {
			suggestions = append(suggestions, "Missing closing ')' for function call")
		} else if strings.Contains(expr, ")") && !strings.Contains(expr, "(") {
			suggestions = append(suggestions, "Missing opening '(' for function call")
		} else {
			// Try to extract function names and validate them
			functions := p.extractFunctionsFromTemplate(expr)
			if len(functions) > 0 {
				validFunctions := []string{}
				invalidFunctions := []string{}
				for _, funcName := range functions {
					if p.validateTemplateFunction(funcName) {
						validFunctions = append(validFunctions, funcName)
					} else {
						invalidFunctions = append(invalidFunctions, funcName)
					}
				}
				if len(invalidFunctions) > 0 {
					suggestions = append(suggestions, fmt.Sprintf("Invalid functions: %s", strings.Join(invalidFunctions, ", ")))
					suggestions = append(suggestions, "Available functions: upper, lower, trimSpace, trim, replace, len, contains, hasPrefix, hasSuffix, title, trimStart, trimEnd")
				}
				if len(validFunctions) > 0 {
					suggestions = append(suggestions, fmt.Sprintf("Valid functions in expression: %s", strings.Join(validFunctions, ", ")))
				}
			} else {
				suggestions = append(suggestions, "No recognizable function calls found in expression")
			}
		}

		// General function error suggestions
		suggestions = append(suggestions, "Check if the function name is spelled correctly")
		suggestions = append(suggestions, "Verify that all required function arguments are provided")
		suggestions = append(suggestions, "Ensure function arguments are properly separated by commas")

	case TemplateErrorPipeline:
		suggestions = append(suggestions, "Ensure each pipeline stage produces a value that the next stage can consume")
		suggestions = append(suggestions, "Check that function signatures are compatible between pipeline stages")
		suggestions = append(suggestions, "Consider using intermediate variables for complex pipelines")

	case TemplateErrorCompilation:
		suggestions = append(suggestions, "Review Go template syntax documentation")
		suggestions = append(suggestions, "Check for unsupported template constructs")
		suggestions = append(suggestions, "Consider simplifying the template expression")

	case TemplateErrorSemantic:
		suggestions = append(suggestions, "Review the template context and available variables")
		suggestions = append(suggestions, "Check function signatures and return types")
		suggestions = append(suggestions, "Verify data types are compatible with operations")
	}

	// Add general suggestions for all error types
	suggestions = append(suggestions, "Refer to template documentation for syntax examples")
	suggestions = append(suggestions, "Test template expressions individually before combining them")

	return suggestions
}

// analyzeTemplateError analyzes a template error and categorizes it appropriately
func (p *Parser) analyzeTemplateError(expr string, err error) *TemplateError {
	errorMsg := err.Error()

	// Determine error category based on error message patterns
	var category TemplateErrorCategory
	var severity TemplateErrorSeverity
	var baseMessage string

	if strings.Contains(errorMsg, "function") && strings.Contains(errorMsg, "not defined") {
		category = TemplateErrorFunction
		severity = SeverityError
		baseMessage = "Undefined function in template expression"
	} else if strings.Contains(errorMsg, "unexpected") || strings.Contains(errorMsg, "invalid") {
		category = TemplateErrorSyntax
		severity = SeverityError
		baseMessage = "Syntax error in template expression"
	} else if strings.Contains(errorMsg, "variable") || strings.Contains(errorMsg, "field") {
		category = TemplateErrorVariable
		severity = SeverityError
		baseMessage = "Variable resolution error"
	} else if strings.Contains(errorMsg, "template") && strings.Contains(errorMsg, "compilation") {
		category = TemplateErrorCompilation
		severity = SeverityError
		baseMessage = "Template compilation failed"
	} else if strings.Contains(errorMsg, "pipeline") {
		category = TemplateErrorPipeline
		severity = SeverityError
		baseMessage = "Pipeline execution error"
	} else {
		category = TemplateErrorSemantic
		severity = SeverityError
		baseMessage = "Template processing error"
	}

	// Find the error location in the expression (simplified approach)
	errorIndex := 0
	if strings.Contains(errorMsg, "at line") && strings.Contains(errorMsg, "column") {
		// Error message already contains location info from template engine
		// We'll use the start of the expression as the location
		errorIndex = 0
	}

	return p.createContextualTemplateError(expr, errorIndex, errorIndex+len(expr), category, severity, baseMessage)
}

// formatTemplateError formats a TemplateError into a human-readable string
func (p *Parser) formatTemplateError(tmplErr *TemplateError) string {
	var builder strings.Builder

	// Header with category and severity
	builder.WriteString(fmt.Sprintf("[%s %s] ", tmplErr.Severity.String(), tmplErr.Category.String()))

	// Location information
	builder.WriteString(fmt.Sprintf("at line %d, column %d: ", tmplErr.Location.Line, tmplErr.Location.Column))

	// Main message
	builder.WriteString(tmplErr.Message)
	builder.WriteString("\n")

	// Expression
	builder.WriteString(fmt.Sprintf("Expression: %s\n", tmplErr.Expression))

	// Context (if available)
	if tmplErr.Context != "" {
		builder.WriteString(fmt.Sprintf("Context: %s\n", tmplErr.Context))
	}

	// Suggestions
	if len(tmplErr.Suggestions) > 0 {
		builder.WriteString("Suggestions:\n")
		for i, suggestion := range tmplErr.Suggestions {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
		}
	}

	return builder.String()
}

// FunctionSignature represents a function signature with validation rules
type FunctionSignature struct {
	Name        string
	MinArgs     int
	MaxArgs     int
	ArgTypes    []string // expected argument types
	ReturnType  string
	Description string
}

// FunctionRegistry holds all available function signatures
type FunctionRegistry struct {
	Functions map[string]*FunctionSignature
}

// NewFunctionRegistry creates a new function registry with built-in functions
func NewFunctionRegistry() *FunctionRegistry {
	reg := &FunctionRegistry{
		Functions: make(map[string]*FunctionSignature),
	}

	// Register built-in string functions
	reg.RegisterFunction(&FunctionSignature{
		Name:        "upper",
		MinArgs:     1,
		MaxArgs:     1,
		ArgTypes:    []string{"string"},
		ReturnType:  "string",
		Description: "Convert string to uppercase",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "lower",
		MinArgs:     1,
		MaxArgs:     1,
		ArgTypes:    []string{"string"},
		ReturnType:  "string",
		Description: "Convert string to lowercase",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "trimSpace",
		MinArgs:     1,
		MaxArgs:     1,
		ArgTypes:    []string{"string"},
		ReturnType:  "string",
		Description: "Remove leading and trailing whitespace",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "trim",
		MinArgs:     1,
		MaxArgs:     2,
		ArgTypes:    []string{"string", "string"},
		ReturnType:  "string",
		Description: "Remove specified characters from string",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "replace",
		MinArgs:     3,
		MaxArgs:     3,
		ArgTypes:    []string{"string", "string", "string"},
		ReturnType:  "string",
		Description: "Replace occurrences of old string with new string",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "len",
		MinArgs:     1,
		MaxArgs:     1,
		ArgTypes:    []string{"any"},
		ReturnType:  "int",
		Description: "Return length of string, slice, or map",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "contains",
		MinArgs:     2,
		MaxArgs:     2,
		ArgTypes:    []string{"string", "string"},
		ReturnType:  "bool",
		Description: "Check if string contains substring",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "hasPrefix",
		MinArgs:     2,
		MaxArgs:     2,
		ArgTypes:    []string{"string", "string"},
		ReturnType:  "bool",
		Description: "Check if string has specified prefix",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "hasSuffix",
		MinArgs:     2,
		MaxArgs:     2,
		ArgTypes:    []string{"string", "string"},
		ReturnType:  "bool",
		Description: "Check if string has specified suffix",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "title",
		MinArgs:     1,
		MaxArgs:     1,
		ArgTypes:    []string{"string"},
		ReturnType:  "string",
		Description: "Convert string to title case",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "trimStart",
		MinArgs:     1,
		MaxArgs:     2,
		ArgTypes:    []string{"string", "string"},
		ReturnType:  "string",
		Description: "Remove leading characters from string",
	})

	reg.RegisterFunction(&FunctionSignature{
		Name:        "trimEnd",
		MinArgs:     1,
		MaxArgs:     2,
		ArgTypes:    []string{"string", "string"},
		ReturnType:  "string",
		Description: "Remove trailing characters from string",
	})

	return reg
}

// RegisterFunction adds a function signature to the registry
func (fr *FunctionRegistry) RegisterFunction(sig *FunctionSignature) {
	fr.Functions[sig.Name] = sig
}

// GetFunction returns a function signature by name
func (fr *FunctionRegistry) GetFunction(name string) (*FunctionSignature, bool) {
	sig, exists := fr.Functions[name]
	return sig, exists
}

// ListFunctions returns all registered function names
func (fr *FunctionRegistry) ListFunctions() []string {
	names := make([]string, 0, len(fr.Functions))
	for name := range fr.Functions {
		names = append(names, name)
	}
	return names
}

// compileTemplateExpression compiles and validates a template expression during parsing
// Returns the compiled template and any validation errors
func (p *Parser) compileTemplateExpression(expr string) (*template.Template, error) {
	if expr == "" {
		return nil, nil
	}

	// Perform lexer-style validation before compilation
	if err := p.validateTemplateExpressionSyntax(expr); err != nil {
		return nil, fmt.Errorf("template syntax error: %w", err)
	}

	// Get the function library from the runtime (we'll need to access this)
	// For now, create a basic template with common functions
	funcMap := template.FuncMap{
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"trimSpace": strings.TrimSpace,
		"trim":      strings.TrimSpace,
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"len": func(v interface{}) int {
			// Basic length function for common types
			switch val := v.(type) {
			case string:
				return len(val)
			case []interface{}:
				return len(val)
			case map[string]interface{}:
				return len(val)
			default:
				return 0
			}
		},
	}

	// Create a new template with the function map
	tmpl, err := template.New("poml-template").Funcs(funcMap).Parse("{{" + expr + "}}")
	if err != nil {
		return nil, fmt.Errorf("template compilation failed: %w", err)
	}

	// Validate that all variables referenced in the template exist in the context
	// This requires extracting variable names from the template and checking them
	if varErr := p.validateTemplateVariables(expr); varErr != nil {
		return nil, fmt.Errorf("variable validation failed: %w", varErr)
	}

	return tmpl, nil
}

// validateTemplateFunction checks if a function name is supported
// This provides a quick lookup for built-in template functions without registry overhead.
// Use this for fast validation of individual function names.
func (p *Parser) validateTemplateFunction(name string) bool {
	supportedFunctions := map[string]bool{
		"upper":     true,
		"lower":     true,
		"trimSpace": true,
		"trim":      true,
		"replace":   true,
		"len":       true,
		"contains":  true,
		"hasPrefix": true,
		"hasSuffix": true,
		"title":     true,
		"trimStart": true,
		"trimEnd":   true,
	}

	return supportedFunctions[name]
}

// validateTemplateExpressionSyntax performs basic syntax validation on template expressions
// This replicates the lexer validation for expressions that bypass normal parsing
func (p *Parser) validateTemplateExpressionSyntax(expr string) error {
	if expr == "" {
		return nil // empty expressions are allowed
	}

	content := strings.TrimSpace(expr)

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

	// Check for unmatched quotes
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

// validateTemplateVariable checks if a variable reference is syntactically valid
// This provides basic validation for variable names without checking context resolution.
// Use this for quick syntax validation of individual variable references.
func (p *Parser) validateTemplateVariable(variable string) error {
	if variable == "" {
		return fmt.Errorf("empty variable reference")
	}

	// Basic validation - variable should start with a letter or dot
	if len(variable) > 0 && !strings.HasPrefix(variable, ".") &&
		(variable[0] < 'a' || variable[0] > 'z') &&
		(variable[0] < 'A' || variable[0] > 'Z') {
		return fmt.Errorf("invalid variable name: %s", variable)
	}

	return nil
}

// logTemplateError logs template compilation errors (could be extended for better error reporting)
func (p *Parser) logTemplateError(expr string, err error) {
	// For now, we'll just log to console. This could be enhanced to collect errors
	// and provide better error reporting in the future.
	fmt.Printf("Template compilation warning for '{{%s}}': %v\n", expr, err)
}
