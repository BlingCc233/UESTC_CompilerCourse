package parser

import (
	"fmt"
	"os"
	"strings"

	"compiler/config"
	"compiler/pointer"
	"compiler/token"
)

// Variable represents a variable in the program
type Variable struct {
	Name       string
	Procedure  string
	Kind       int // 0 or 1
	Type       string
	Level      int
	Address    int
	IsDeclared bool
}

// Procedure represents a procedure in the program
type Procedure struct {
	Name                 string
	Type                 string
	Level                int
	FirstVariableAddress int
	LastVariableAddress  int
}

// Parser represents the syntax analyzer
type Parser struct {
	line                   int
	callStack              []string
	currentVariableAddress int
	shouldAddError         bool

	correctTokens []token.Token
	variables     []Variable
	procedures    []Procedure
	errors        []string

	cursor *pointer.Cursor[token.Token]
}

// New creates a new Parser instance
func New() *Parser {
	return &Parser{
		line:                   1,
		callStack:              make([]string, 0),
		currentVariableAddress: -1,
		shouldAddError:         true,
		correctTokens:          make([]token.Token, 0),
		variables:              make([]Variable, 0),
		procedures:             make([]Procedure, 0),
		errors:                 make([]string, 0),
		cursor:                 pointer.NewCursor(readTokens()),
	}
}

// Parse starts the parsing process
func (p *Parser) Parse() bool {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				p.errors = append(p.errors, fmt.Sprintf("%v [FATAL]", err))
			}
		}
		writeCorrectTokens(p.correctTokens)
		writeVariables(p.variables)
		writeProcedures(p.procedures)
		writeErrors(p.errors)
	}()

	p.parseProgram()
	return len(p.errors) == 0
}

// Main parsing methods
func (p *Parser) parseProgram() {
	p.parseSubprogram()
	p.match(token.END_OF_FILE)
}

func (p *Parser) parseSubprogram() {
	p.callStack = append([]string{"main"}, p.callStack...)

	p.match(token.BEGIN)
	p.parseDeclarations()
	p.parseExecutions()
	p.match(token.END)

	p.callStack = p.callStack[1:]
}

func (p *Parser) parseDeclarations() {
	p.parseDeclaration()
	p.parseDeclarations_()
}

func (p *Parser) parseDeclarations_() {
	if p.hasType(token.INTEGER) {
		p.parseDeclaration()
		p.parseDeclarations_()
	}
}

func (p *Parser) parseDeclaration() {
	p.match(token.INTEGER, "Every program or procedure should have at least one declaration")
	p.parseDeclaration_()
	p.match(token.SEMICOLON)
}

func (p *Parser) parseDeclaration_() {
	if p.hasType(token.IDENTIFIER) {
		p.parseVariableDeclaration()
		return
	}

	if p.hasType(token.FUNCTION) {
		p.parseProcedureDeclaration()
		return
	}

	tok := p.consumeToken()
	p.throwError(fmt.Sprintf("'%s' is not a valid variable name", tok.Value))
}

func (p *Parser) parseVariableDeclaration() {
	tok := p.match(token.IDENTIFIER)
	p.registerVariable(tok.Value)
}

func (p *Parser) parseVariable() {
	tok := p.match(token.IDENTIFIER)
	if !p.findVariable(tok.Value) {
		p.addError(fmt.Sprintf("Undefined variable '%s'", tok.Value))
	}
}

func (p *Parser) parseProcedureDeclaration() {
	p.match(token.FUNCTION)
	p.parseProcedureNameDeclaration()
	p.match(token.LEFT_PARENTHESES)
	p.parseParameterDeclaration()
	p.match(token.RIGHT_PARENTHESES, "Unmatched '('")
	p.match(token.SEMICOLON)
	p.parseProcedureBody()
}

func (p *Parser) parseProcedureNameDeclaration() {
	tok := p.match(token.IDENTIFIER)
	p.registerProcedure(tok.Value)
}

func (p *Parser) parseProcedureName() {
	tok := p.match(token.IDENTIFIER)
	if !p.findProcedure(tok.Value) {
		p.addError(fmt.Sprintf("Undefined procedure '%s'", tok.Value))
	}
}

func (p *Parser) parseParameterDeclaration() {
	tok := p.match(token.IDENTIFIER)
	p.registerParameter(tok.Value)
}

func (p *Parser) parseProcedureBody() {
	p.match(token.BEGIN)
	p.parseDeclarations()
	p.parseExecutions()
	p.match(token.END)
	p.callStack = p.callStack[1:]
}

func (p *Parser) parseExecutions() {
	p.parseExecution()
	p.parseExecutions_()
}

func (p *Parser) parseExecutions_() {
	if p.hasType(token.SEMICOLON) {
		p.match(token.SEMICOLON)
		p.parseExecution()
		p.parseExecutions_()
	}
}

func (p *Parser) parseExecution() {
	if p.hasType(token.READ) {
		p.parseRead()
		return
	}

	if p.hasType(token.WRITE) {
		p.parseWrite()
		return
	}

	if p.hasType(token.IDENTIFIER) {
		p.parseAssignment()
		return
	}

	if p.hasType(token.IF) {
		p.parseCondition()
		return
	}

	if p.hasType(token.INTEGER) {
		p.consumeToken()
		p.throwError("Please move all declarations to the beginning of the procedure")
		return
	}

	tok := p.consumeToken()
	p.throwError(fmt.Sprintf("Execution cannot begin with '%s'", tok.Value))
}

func (p *Parser) parseRead() {
	p.match(token.READ)
	p.match(token.LEFT_PARENTHESES)
	p.parseVariable()
	p.match(token.RIGHT_PARENTHESES, "Unmatched '('")
}

func (p *Parser) parseWrite() {
	p.match(token.WRITE)
	p.match(token.LEFT_PARENTHESES)
	p.parseVariable()
	p.match(token.RIGHT_PARENTHESES, "Unmatched '('")
}

func (p *Parser) parseAssignment() {
	current := p.cursor.Current()
	if p.findVariable(current.Value) {
		p.parseVariable()
	} else if p.findProcedure(current.Value) {
		p.parseProcedureName()
	} else {
		tok := p.consumeToken()
		p.addError(fmt.Sprintf("Undefined variable or procedure '%s'", tok.Value))
	}

	p.match(token.ASSIGN)
	p.parseArithmeticExpression()
}

func (p *Parser) parseArithmeticExpression() {
	p.parseTerm()
	p.parseArithmeticExpression_()
}

func (p *Parser) parseArithmeticExpression_() {
	if p.hasType(token.SUBTRACT) {
		p.match(token.SUBTRACT)
		p.parseTerm()
		p.parseArithmeticExpression_()
	}
}

func (p *Parser) parseTerm() {
	p.parseFactor()
	p.parseTerm_()
}

func (p *Parser) parseTerm_() {
	if p.hasType(token.MULTIPLY) {
		p.match(token.MULTIPLY)
		p.parseFactor()
		p.parseTerm_()
	}
}

func (p *Parser) parseFactor() {
	if p.hasType(token.CONSTANT) {
		p.match(token.CONSTANT)
		return
	}

	if p.hasType(token.IDENTIFIER) {
		if p.findVariable(p.cursor.Current().Value) {
			p.parseVariable()
			return
		}
		if p.findProcedure(p.cursor.Current().Value) {
			p.parseProcedureCall()
			return
		}
		tok := p.consumeToken()
		p.throwError(fmt.Sprintf("Undefined variable or procedure '%s'", tok.Value))
	}

	tok := p.consumeToken()
	p.throwError(fmt.Sprintf("Expect variable, procedure or constant, but got '%s'", tok.Value))
}

func (p *Parser) parseProcedureCall() {
	p.parseProcedureName()
	p.match(token.LEFT_PARENTHESES)
	p.parseArithmeticExpression()
	p.match(token.RIGHT_PARENTHESES, "Unmatched '('")
}

func (p *Parser) parseCondition() {
	p.match(token.IF)
	p.parseConditionExpression()
	p.match(token.THEN)
	p.parseExecution()
	p.match(token.ELSE)
	p.parseExecution()
}

func (p *Parser) parseConditionExpression() {
	p.parseArithmeticExpression()
	p.parseOperator()
	p.parseArithmeticExpression()
}

func (p *Parser) parseOperator() {
	if p.hasType(token.EQUAL) {
		p.match(token.EQUAL)
		return
	}
	if p.hasType(token.NOT_EQUAL) {
		p.match(token.NOT_EQUAL)
		return
	}
	if p.hasType(token.LESS_THAN) {
		p.match(token.LESS_THAN)
		return
	}
	if p.hasType(token.LESS_THAN_OR_EQUAL) {
		p.match(token.LESS_THAN_OR_EQUAL)
		return
	}
	if p.hasType(token.GREATER_THAN) {
		p.match(token.GREATER_THAN)
		return
	}
	if p.hasType(token.GREATER_THAN_OR_EQUAL) {
		p.match(token.GREATER_THAN_OR_EQUAL)
		return
	}
	tok := p.consumeToken()
	p.addError(fmt.Sprintf("%s is not a valid operator", tok.Value))
}

func (p *Parser) registerVariable(name string) {
	if param := p.findParameter(name); param != nil {
		param.IsDeclared = true
		return
	}

	if dup := p.findDuplicateVariable(name); dup != nil {
		p.addError(fmt.Sprintf("Variable '%s' has already been declared", name))
		return
	}

	p.variables = append(p.variables, Variable{
		Name:       name,
		Procedure:  p.callStack[0],
		Kind:       0,
		Type:       "integer",
		Level:      len(p.callStack),
		Address:    p.currentVariableAddress + 1,
		IsDeclared: true,
	})
	p.currentVariableAddress++

	p.updateProcedureVariableAddresses()
}

func (p *Parser) findDuplicateVariable(name string) *Variable {
	for _, v := range p.variables {
		if v.Name == name && v.Procedure == p.callStack[0] {
			return &v
		}
	}
	return nil
}

func (p *Parser) findVariable(name string) bool {
	for _, v := range p.variables {
		if v.Name == name && v.Level <= len(p.callStack) {
			if !v.IsDeclared {
				p.addError(fmt.Sprintf("Variable '%s' has not been declared", name))
			}
			return true
		}
	}
	return false
}

func (p *Parser) registerParameter(name string) {
	if dup := p.findDuplicateParameter(name); dup != nil {
		p.addError(fmt.Sprintf("Parameter '%s' has already been declared", name))
		return
	}

	p.variables = append(p.variables, Variable{
		Name:       "_" + name,
		Procedure:  p.callStack[0],
		Kind:       1,
		Type:       "integer",
		Level:      len(p.callStack),
		Address:    p.currentVariableAddress + 1,
		IsDeclared: false,
	})
	p.currentVariableAddress++

	p.updateProcedureVariableAddresses()
}

func (p *Parser) findDuplicateParameter(name string) *Variable {
	for _, v := range p.variables {
		if v.Name == name && v.Kind == 1 && v.Procedure == p.callStack[0] {
			return &v
		}
	}
	return nil
}

func (p *Parser) findParameter(name string) *Variable {
	for _, v := range p.variables {
		if v.Name == name && v.Kind == 1 && v.Level <= len(p.callStack) {
			return &v
		}
	}
	return nil
}

func (p *Parser) registerProcedure(name string) {
	if dup := p.findDuplicateProcedure(name); dup != nil {
		p.addError(fmt.Sprintf("Procedure '%s' has already been declared", name))
		return
	}

	p.procedures = append(p.procedures, Procedure{
		Name:                 name,
		Type:                 "integer",
		Level:                len(p.callStack) + 1,
		FirstVariableAddress: -1,
		LastVariableAddress:  -1,
	})
	p.callStack = append([]string{name}, p.callStack...)
}

func (p *Parser) findDuplicateProcedure(name string) *Procedure {
	for _, proc := range p.procedures {
		if proc.Name == name && proc.Level == len(p.callStack)+1 {
			return &proc
		}
	}
	return nil
}

func (p *Parser) findProcedure(name string) bool {
	for _, proc := range p.procedures {
		if proc.Name == name && proc.Level <= len(p.callStack)+1 {
			return true
		}
	}
	return false
}

func (p *Parser) updateProcedureVariableAddresses() {
	for i, proc := range p.procedures {
		if proc.Name == p.callStack[0] {
			if proc.FirstVariableAddress == -1 {
				p.procedures[i].FirstVariableAddress = p.currentVariableAddress
			}
			p.procedures[i].LastVariableAddress = p.currentVariableAddress
			break
		}
	}
}

// Helper methods
func (p *Parser) hasType(expectation token.TokenType) bool {
	return expectation == p.cursor.Current().Type
}

func (p *Parser) match(expectation token.TokenType, message ...string) token.Token {
	if !p.hasType(expectation) {
		msg := fmt.Sprintf("Expect %s, but got '%s'",
			translateToken(expectation),
			p.cursor.Current().Value)
		if len(message) > 0 {
			msg = message[0]
		}
		p.addError(msg)
	}
	return p.consumeToken()
}

func (p *Parser) consumeToken() token.Token {
	p.goToNextLine()
	tok := p.cursor.Consume()
	p.correctTokens = append(p.correctTokens, tok)
	p.goToNextLine()
	return tok
}

func (p *Parser) goToNextLine() {
	for p.cursor.IsOpen() && p.hasType(token.END_OF_LINE) {
		tok := p.cursor.Consume()
		p.correctTokens = append(p.correctTokens, tok)
		p.line++
		p.shouldAddError = true
	}
}

func (p *Parser) throwError(error string) {
	panic(fmt.Errorf("***LINE %d: %s", p.line, error))
}

func (p *Parser) addError(error string) {
	if !p.shouldAddError {
		return
	}
	p.shouldAddError = false
	p.errors = append(p.errors, fmt.Sprintf("***LINE %d: %s", p.line, error))
}

func translateToken(t token.TokenType) string {
	tokenTranslation := map[token.TokenType]string{
		token.BEGIN:                 "'begin'",
		token.END:                   "'end'",
		token.INTEGER:               "'integer'",
		token.IF:                    "'if'",
		token.THEN:                  "'then'",
		token.ELSE:                  "'else'",
		token.FUNCTION:              "'function'",
		token.READ:                  "'read'",
		token.WRITE:                 "'write'",
		token.IDENTIFIER:            "identifier",
		token.CONSTANT:              "constant",
		token.EQUAL:                 "'='",
		token.NOT_EQUAL:             "'<>'",
		token.LESS_THAN_OR_EQUAL:    "'<='",
		token.LESS_THAN:             "'<'",
		token.GREATER_THAN_OR_EQUAL: "'>='",
		token.GREATER_THAN:          "'>'",
		token.SUBTRACT:              "'-'",
		token.MULTIPLY:              "'*'",
		token.ASSIGN:                "':='",
		token.LEFT_PARENTHESES:      "'('",
		token.RIGHT_PARENTHESES:     "')'",
		token.SEMICOLON:             "';'",
		token.END_OF_LINE:           "EOLN",
		token.END_OF_FILE:           "EOF",
	}
	return tokenTranslation[t]
}

func readTokens() []token.Token {
	data, err := os.ReadFile(config.DYD_PATH)
	if err != nil {
		panic(err)
	}
	text := strings.TrimSpace(string(data))
	tokens := make([]token.Token, 0)

	for _, line := range strings.Split(text, "\n") {
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) < 2 {
			continue
		}
		value := parts[0]
		typeVal := parts[1]
		var t token.TokenType
		fmt.Sscanf(typeVal, "%d", &t)
		tokens = append(tokens, token.Token{Type: t, Value: value})
	}
	return tokens
}

func writeCorrectTokens(tokens []token.Token) {
	var lines []string
	for _, t := range tokens {
		line := fmt.Sprintf("%-16s %02d", t.Value, t.Type)
		lines = append(lines, line)
	}
	text := strings.Join(lines, "\n")
	os.WriteFile(config.DYS_PATH, []byte(text), 0644)
}

func writeVariables(variables []Variable) {
	var lines []string
	for _, v := range variables {
		line := fmt.Sprintf("Var\n    Name      = %s\n    Procedure = %s\n    Kind      = %%!s(main.VarKind=%d)\n    Type      = %s\n    Level     = %d\n    Offset    = %d",
			v.Name, v.Procedure, v.Kind, v.Type, v.Level, v.Address)
		lines = append(lines, line)
	}
	text := strings.Join(lines, "\n")
	os.WriteFile(config.VAR_PATH, []byte(text), 0644)
}

func writeProcedures(procedures []Procedure) {
	var lines []string
	for _, p := range procedures {
		line := fmt.Sprintf("Proc\n    Name      = %s\n    Type      = %s\n    Level     = %d\n    FirstVar  = %d\n    LastVar   = %d",
			p.Name, p.Type, p.Level, p.FirstVariableAddress, p.LastVariableAddress)
		lines = append(lines, line)
	}
	text := strings.Join(lines, "\n")
	os.WriteFile(config.PRO_PATH, []byte(text), 0644)
}

func writeErrors(errors []string) {
	text := strings.Join(errors, "\n")
	os.WriteFile(config.ERR_PATH, []byte(text), 0644)
}

func (p *Parser) ListErrors() {
	for i, err := range p.errors {
		fmt.Printf("Error %d: %s\n", i+1, err)
	}
}
