package lexer

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"compiler/config"
	"compiler/pointer"
	"compiler/token"
)

const MAX_IDENTIFIER_LENGTH = 16

// Lexer represents a lexical analyzer
type Lexer struct {
	line   int
	cursor *pointer.Cursor[rune]
}

// New creates a new Lexer instance
func New() *Lexer {
	return &Lexer{
		line:   1,
		cursor: pointer.NewCursor([]rune(readSource())),
	}
}

// Tokenize processes the source file and generates tokens
func (l *Lexer) Tokenize() bool {
	tokens := []token.Token{}
	errors := []string{}

	for l.cursor.IsOpen() {
		tok, err := l.getNextToken()
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		tokens = append(tokens, tok)
	}

	tokens = append(tokens, token.Token{
		Type:  token.END_OF_FILE,
		Value: "EOF",
	})

	writeTokens(tokens)
	writeErrors(errors)

	return len(errors) == 0
}

func (l *Lexer) getNextToken() (token.Token, error) {
	// Skip whitespace
	for l.cursor.IsOpen() && l.cursor.Current() == ' ' {
		l.cursor.Consume()
	}

	if !l.cursor.IsOpen() {
		return token.Token{}, fmt.Errorf("unexpected end of input")
	}

	initial := l.cursor.Consume()

	if isLetter(initial) {
		value := string(initial)
		for l.cursor.IsOpen() && (isLetter(l.cursor.Current()) || isDigit(l.cursor.Current())) {
			value += string(l.cursor.Consume())
		}

		if keywordType := getKeywordType(value); keywordType != 0 {
			return token.Token{Type: keywordType, Value: value}, nil
		}

		if len(value) <= MAX_IDENTIFIER_LENGTH {
			return token.Token{Type: token.IDENTIFIER, Value: value}, nil
		}

		return token.Token{}, fmt.Errorf("line %d: Identifier name '%s' exceeds %d characters",
			l.line, value, MAX_IDENTIFIER_LENGTH)
	}

	if isDigit(initial) {
		value := string(initial)
		for l.cursor.IsOpen() && isDigit(l.cursor.Current()) {
			value += string(l.cursor.Consume())
		}
		return token.Token{Type: token.CONSTANT, Value: value}, nil
	}

	// Handle special characters
	switch initial {
	case '=':
		return token.Token{Type: token.EQUAL, Value: "="}, nil
	case '-':
		return token.Token{Type: token.SUBTRACT, Value: "-"}, nil
	case '*':
		return token.Token{Type: token.MULTIPLY, Value: "*"}, nil
	case '(':
		return token.Token{Type: token.LEFT_PARENTHESES, Value: "("}, nil
	case ')':
		return token.Token{Type: token.RIGHT_PARENTHESES, Value: ")"}, nil
	case '<':
		if l.cursor.IsOpen() {
			switch l.cursor.Current() {
			case '=':
				l.cursor.Consume()
				return token.Token{Type: token.LESS_THAN_OR_EQUAL, Value: "<="}, nil
			case '>':
				l.cursor.Consume()
				return token.Token{Type: token.NOT_EQUAL, Value: "<>"}, nil
			}
		}
		return token.Token{Type: token.LESS_THAN, Value: "<"}, nil
	case '>':
		if l.cursor.IsOpen() && l.cursor.Current() == '=' {
			l.cursor.Consume()
			return token.Token{Type: token.GREATER_THAN_OR_EQUAL, Value: ">="}, nil
		}
		return token.Token{Type: token.GREATER_THAN, Value: ">"}, nil
	case ':':
		if l.cursor.IsOpen() && l.cursor.Current() == '=' {
			l.cursor.Consume()
			return token.Token{Type: token.ASSIGN, Value: ":="}, nil
		}
		return token.Token{}, fmt.Errorf("line %d: Misused colon", l.line)
	case ';':
		return token.Token{Type: token.SEMICOLON, Value: ";"}, nil
	case '\n':
		l.line++
		return token.Token{Type: token.END_OF_LINE, Value: "EOLN"}, nil
	}

	return token.Token{}, fmt.Errorf("line %d: Invalid character '%c'", l.line, initial)
}

// Helper functions
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func getKeywordType(value string) token.TokenType {
	keywordMap := map[string]token.TokenType{
		"begin":    token.BEGIN,
		"end":      token.END,
		"integer":  token.INTEGER,
		"if":       token.IF,
		"then":     token.THEN,
		"else":     token.ELSE,
		"function": token.FUNCTION,
		"read":     token.READ,
		"write":    token.WRITE,
	}

	if tokType, ok := keywordMap[strings.ToLower(value)]; ok {
		return tokType
	}
	return 0
}

// File operations
func readSource() string {
	data, err := os.ReadFile(config.SOURCE_PATH)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(data))
}

func writeTokens(tokens []token.Token) error {
	var sb strings.Builder
	for _, tok := range tokens {
		padding := strings.Repeat(" ", 16-len(tok.Value))
		sb.WriteString(fmt.Sprintf("%s%s %02d\n", tok.Value, padding, tok.Type))
	}
	return os.WriteFile(config.DYD_PATH, []byte(sb.String()), 0644)
}

func writeErrors(errors []string) error {
	if len(errors) == 0 {
		return nil
	}
	return os.WriteFile(config.ERR_PATH, []byte(strings.Join(errors, "\n")), 0644)
}
