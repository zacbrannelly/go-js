package lexer

import "fmt"

type LexicalGoal int

const (
	InputElementDiv LexicalGoal = iota
)

type TokenType int

const (
	WhiteSpace TokenType = iota
	LineTerminator
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	Input             string
	Goal              LexicalGoal
	Tokens            []Token
	CurrentIndex      int
	CurrentTokenType  TokenType
	CurrentTokenValue string
}

func Lex(input string, goal LexicalGoal) []Token {
	lexer := Lexer{
		Input:             input,
		Goal:              goal,
		Tokens:            []Token{},
		CurrentIndex:      0,
		CurrentTokenType:  -1,
		CurrentTokenValue: "",
	}

	switch goal {
	case InputElementDiv:
		return LexInputElementDiv(&lexer)
	default:
		panic(fmt.Sprintf("Unsupported lexical goal: %d", goal))
	}
}

func EmitToken(lexer *Lexer) {
	lexer.Tokens = append(lexer.Tokens, Token{Type: lexer.CurrentTokenType, Value: lexer.CurrentTokenValue})
	lexer.CurrentTokenValue = ""
	lexer.CurrentTokenType = -1
}

func ConsumeChar(lexer *Lexer, tokenType TokenType) {
	currentChar := CurrentChar(lexer)
	lexer.CurrentTokenType = tokenType
	lexer.CurrentTokenValue += string(currentChar)
	lexer.CurrentIndex++
}

func CurrentChar(lexer *Lexer) rune {
	return rune(lexer.Input[lexer.CurrentIndex])
}

func CanLookahead(lexer *Lexer) bool {
	return lexer.CurrentIndex+1 < len(lexer.Input)
}

func LookaheadChar(lexer *Lexer) rune {
	return rune(lexer.Input[lexer.CurrentIndex+1])
}

func IsEOF(lexer *Lexer) bool {
	return lexer.CurrentIndex >= len(lexer.Input)
}

func LexInputElementDiv(lexer *Lexer) []Token {
	for !IsEOF(lexer) {
		char := CurrentChar(lexer)

		if IsWhiteSpaceChar(char) {
			ConsumeWhiteSpace(lexer)
		} else if IsLineTerminator(char) {
			ConsumeLineTerminator(lexer)
		} else {
			panic(fmt.Sprintf("Unexpected character: %c", char))
		}
	}

	return lexer.Tokens
}

func ConsumeWhiteSpace(lexer *Lexer) {
	for !IsEOF(lexer) && IsWhiteSpaceChar(CurrentChar(lexer)) {
		ConsumeChar(lexer, WhiteSpace)
	}
	EmitToken(lexer)
}

func ConsumeLineTerminator(lexer *Lexer) {
	if !IsEOF(lexer) && CurrentChar(lexer) == '\r' && CanLookahead(lexer) && LookaheadChar(lexer) == '\n' {
		ConsumeChar(lexer, LineTerminator)
		ConsumeChar(lexer, LineTerminator)
		EmitToken(lexer)
	} else if !IsEOF(lexer) {
		ConsumeChar(lexer, LineTerminator)
		EmitToken(lexer)
	}
}

func IsWhiteSpaceChar(char rune) bool {
	// TODO: Support Space_Separator character group
	return (char == ' ' ||
		char == '\t' ||
		char == '\v' ||
		char == '\f' ||
		char == '\uFEFF' ||
		char == '\u00BF')
}

func IsLineTerminator(char rune) bool {
	return (char == '\n' || char == '\r' || char == '\u2028' || char == '\u2029')
}
