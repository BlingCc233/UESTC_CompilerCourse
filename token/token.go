package token

// TokenType represents the type of token
type TokenType int

const (
	BEGIN TokenType = iota + 1
	END
	INTEGER
	IF
	THEN
	ELSE
	FUNCTION
	READ
	WRITE
	IDENTIFIER
	CONSTANT
	EQUAL
	NOT_EQUAL
	LESS_THAN_OR_EQUAL
	LESS_THAN
	GREATER_THAN_OR_EQUAL
	GREATER_THAN
	SUBTRACT
	MULTIPLY
	ASSIGN
	LEFT_PARENTHESES
	RIGHT_PARENTHESES
	SEMICOLON
	END_OF_LINE
	END_OF_FILE
)

// Token represents a token with its type and value
type Token struct {
	Type  TokenType
	Value string
}
