package lexer

import (
	"fmt"
	"regexp"
	"unicode"
)

type LexicalGoal int

const (
	InputElementDiv LexicalGoal = iota
	InputElementRegExp
	InputElementRegExpOrTemplateTail
	InputElementTemplateTail
	InputElementHashbangOrRegExp
)

var ReservedWords = map[string]TokenType{
	"await":      Await,
	"break":      Break,
	"case":       Case,
	"catch":      Catch,
	"class":      Class,
	"const":      Const,
	"continue":   Continue,
	"debugger":   Debugger,
	"default":    Default,
	"delete":     Delete,
	"do":         Do,
	"else":       Else,
	"enum":       Enum,
	"export":     Export,
	"extends":    Extends,
	"false":      False,
	"finally":    Finally,
	"for":        For,
	"function":   Function,
	"if":         If,
	"import":     Import,
	"in":         In,
	"instanceof": InstanceOf,
	"new":        New,
	"null":       Null,
	"return":     Return,
	"super":      Super,
	"switch":     Switch,
	"this":       This,
	"throw":      Throw,
	"true":       True,
	"try":        Try,
	"typeof":     TypeOf,
	"var":        Var,
	"void":       Void,
	"while":      While,
	"with":       With,
	"yield":      Yield,
}

type TokenType int

const (
	WhiteSpace TokenType = iota
	LineTerminator
	Comment
	Identifier
	PrivateIdentifier
	// Reserved Words
	Await
	Break
	Case
	Catch
	Class
	Const
	Continue
	Debugger
	Default
	Delete
	Do
	Else
	Enum
	Export
	Extends
	False
	Finally
	For
	Function
	If
	Import
	In
	InstanceOf
	New
	Null
	Return
	Super
	Switch
	This
	Throw
	True
	Try
	TypeOf
	Var
	Void
	While
	With
	Yield
	// Punctuators
	LeftBrace                    // {
	RightBrace                   // }
	LeftBracket                  // [
	RightBracket                 // ]
	LeftParen                    // (
	RightParen                   // )
	Dot                          // .
	Spread                       // ...
	Semicolon                    // ;
	Comma                        // ,
	LessThan                     // <
	LessThanEqual                // <=
	GreaterThan                  // >
	GreaterThanEqual             // >=
	Equal                        // ==
	NotEqual                     // !=
	StrictEqual                  // ===
	StrictNotEqual               // !==
	Plus                         // +
	Minus                        // -
	Multiply                     // *
	Modulo                       // %
	Divide                       // /
	Exponentiation               // **
	Increment                    // ++
	Decrement                    // --
	LeftShift                    // <<
	RightShift                   // >>
	UnsignedRightShift           // >>>
	BitwiseAnd                   // &
	BitwiseOr                    // |
	BitwiseXor                   // ^
	Not                          // !
	BitwiseNot                   // ~
	And                          // &&
	Or                           // ||
	NullishCoalescing            // ??
	TernaryQuestionMark          // ?
	TernaryColon                 // :
	Assignment                   // =
	PlusAssignment               // +=
	MinusAssignment              // -=
	MultiplyAssignment           // *=
	ModuloAssignment             // %=
	DivideAssignment             // /=
	ExponentiationAssignment     // **=
	LeftShiftAssignment          // <<=
	RightShiftAssignment         // >>=
	UnsignedRightShiftAssignment // >>>=
	BitwiseAndAssignment         // &=
	BitwiseOrAssignment          // |=
	BitwiseXorAssignment         // ^=
	AndAssignment                // &&=
	OrAssignment                 // ||=
	NullishCoalescingAssignment  // ??=
	ArrowOperator                // =>
	OptionalChain                // ?.
	// End of Punctuators
	NumericLiteral                // 123, 123.456, 123.456e789, 0x123456789abcdef, 0b10101010, 0o12345670
	StringLiteral                 // "Hello, world!"
	TemplateNoSubstitutionLiteral // `No substitution`
	TemplateStartLiteral          // `Hello, ${`
	RegularExpressionLiteral      // /abc/g
	TemplateMiddle                // } maybes some other text ${
	TemplateTail                  // } maybes some other text`
	HashbangComment               // #!lol
)

var EqualityOperators = []TokenType{
	Equal,
	NotEqual,
	StrictEqual,
	StrictNotEqual,
}

var RelationalOperators = []TokenType{
	LessThan,
	LessThanEqual,
	GreaterThan,
	GreaterThanEqual,
	InstanceOf,
}

var ShiftOperators = []TokenType{
	LeftShift,
	RightShift,
	UnsignedRightShift,
}

var AdditiveOperators = []TokenType{
	Plus,
	Minus,
}

var MultiplicativeOperators = []TokenType{
	Multiply,
	Divide,
	Modulo,
}

var UnaryOperators = []TokenType{
	Delete,
	Void,
	TypeOf,
	Plus,
	Minus,
	Increment,
	Decrement,
	BitwiseNot,
}

var UpdateOperators = []TokenType{
	Increment,
	Decrement,
}

var AssignmentOperators = []TokenType{
	MultiplyAssignment,
	DivideAssignment,
	ModuloAssignment,
	PlusAssignment,
	MinusAssignment,
	LeftShiftAssignment,
	RightShiftAssignment,
	UnsignedRightShiftAssignment,
	BitwiseAndAssignment,
	BitwiseXorAssignment,
	BitwiseOrAssignment,
	ExponentiationAssignment,
}

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	Input             string
	Goal              LexicalGoal
	Tokens            []Token
	CurrentIndex      int
	CurrentTokenValue string
}

func LexNextToken(lexer *Lexer) bool {
	switch lexer.Goal {
	case InputElementDiv:
		return LexNextInputElementDiv(lexer)
	case InputElementRegExp:
		return LexNextInputElementRegExp(lexer)
	case InputElementRegExpOrTemplateTail:
		return LexNextInputElementRegExpOrTemplateTail(lexer)
	case InputElementTemplateTail:
		return LexNextInputElementTemplateTail(lexer)
	case InputElementHashbangOrRegExp:
		return LexNextInputElementHashbangOrRegExp(lexer)
	default:
		panic(fmt.Sprintf("Unsupported lexical goal: %d", lexer.Goal))
	}
}

func LexAll(input string, goal LexicalGoal) []Token {
	lexer := Lexer{
		Input:             input,
		Goal:              goal,
		Tokens:            []Token{},
		CurrentIndex:      0,
		CurrentTokenValue: "",
	}

	switch goal {
	case InputElementDiv:
		return LexInputElementDiv(&lexer)
	case InputElementRegExp:
		return LexInputElementRegExp(&lexer)
	case InputElementRegExpOrTemplateTail:
		return LexInputElementRegExpOrTemplateTail(&lexer)
	case InputElementTemplateTail:
		return LexInputElementTemplateTail(&lexer)
	case InputElementHashbangOrRegExp:
		return LexInputElementHashbangOrRegExp(&lexer)
	default:
		panic(fmt.Sprintf("Unsupported lexical goal: %d", goal))
	}
}

func EmitToken(lexer *Lexer, tokenType TokenType) {
	lexer.Tokens = append(lexer.Tokens, Token{Type: tokenType, Value: lexer.CurrentTokenValue})
	lexer.CurrentTokenValue = ""
}

func ConsumeChar(lexer *Lexer) {
	currentChar := CurrentChar(lexer)
	lexer.CurrentTokenValue += string(currentChar)
	lexer.CurrentIndex++
}

func CurrentChar(lexer *Lexer) rune {
	return rune(lexer.Input[lexer.CurrentIndex])
}

func CanLookahead(lexer *Lexer) bool {
	return lexer.CurrentIndex+1 < len(lexer.Input)
}

func CanLookaheadN(lexer *Lexer, n int) bool {
	return lexer.CurrentIndex+n < len(lexer.Input)
}

func LookaheadChar(lexer *Lexer) rune {
	return rune(lexer.Input[lexer.CurrentIndex+1])
}

func LookaheadCharN(lexer *Lexer, n int) rune {
	return rune(lexer.Input[lexer.CurrentIndex+n])
}

func IsEOF(lexer *Lexer) bool {
	return lexer.CurrentIndex >= len(lexer.Input)
}

func LexInputElementDiv(lexer *Lexer) []Token {
	for !IsEOF(lexer) {
		char := CurrentChar(lexer)

		if LexNextInputElementDiv(lexer) {
			continue
		}

		panic(fmt.Sprintf("Unexpected character: %c", char))
	}

	return lexer.Tokens
}

func LexNextInputElementDiv(lexer *Lexer) bool {
	if IsEOF(lexer) {
		return false
	}

	char := CurrentChar(lexer)

	if ConsumeWhiteSpace(lexer) {
		return true
	}

	if ConsumeLineTerminator(lexer) {
		return true
	}

	if ConsumeComment(lexer) {
		return true
	}

	if ConsumeCommonToken(lexer) {
		return true
	}

	// DivPunctuator
	if char == '/' {
		// /
		ConsumeChar(lexer)
		EmitToken(lexer, Divide)
		return true
	}

	// RightBracePunctuator
	if char == '}' {
		ConsumeChar(lexer)
		EmitToken(lexer, RightBrace)
		return true
	}

	return false
}

func LexInputElementRegExp(lexer *Lexer) []Token {
	for !IsEOF(lexer) {
		char := CurrentChar(lexer)

		if LexNextInputElementRegExp(lexer) {
			continue
		}

		panic(fmt.Sprintf("Unexpected character: %c", char))
	}

	return lexer.Tokens
}

func LexNextInputElementRegExp(lexer *Lexer) bool {
	if IsEOF(lexer) {
		return false
	}

	char := CurrentChar(lexer)

	if ConsumeWhiteSpace(lexer) {
		return true
	}

	if ConsumeLineTerminator(lexer) {
		return true
	}

	if ConsumeComment(lexer) {
		return true
	}

	if ConsumeCommonToken(lexer) {
		return true
	}

	// RightBracePunctuator
	if char == '}' {
		ConsumeChar(lexer)
		EmitToken(lexer, RightBrace)
		return true
	}

	if ConsumeRegularExpressionLiteral(lexer) {
		return true
	}

	return false
}

func LexInputElementRegExpOrTemplateTail(lexer *Lexer) []Token {
	for !IsEOF(lexer) {
		char := CurrentChar(lexer)

		if LexNextInputElementRegExpOrTemplateTail(lexer) {
			continue
		}

		panic(fmt.Sprintf("Unexpected character: %c", char))
	}

	return lexer.Tokens
}

func LexNextInputElementRegExpOrTemplateTail(lexer *Lexer) bool {
	if IsEOF(lexer) {
		return false
	}

	if ConsumeWhiteSpace(lexer) {
		return true
	}

	if ConsumeLineTerminator(lexer) {
		return true
	}

	if ConsumeComment(lexer) {
		return true
	}

	if ConsumeCommonToken(lexer) {
		return true
	}

	if ConsumeRegularExpressionLiteral(lexer) {
		return true
	}

	if ConsumeTemplateSubstitutionTail(lexer) {
		return true
	}

	return false
}

func LexInputElementTemplateTail(lexer *Lexer) []Token {
	for !IsEOF(lexer) {
		char := CurrentChar(lexer)

		if LexNextInputElementTemplateTail(lexer) {
			continue
		}

		panic(fmt.Sprintf("Unexpected character: %c", char))
	}

	return lexer.Tokens
}

func LexNextInputElementTemplateTail(lexer *Lexer) bool {
	if IsEOF(lexer) {
		return false
	}

	char := CurrentChar(lexer)

	if ConsumeWhiteSpace(lexer) {
		return true
	}

	if ConsumeLineTerminator(lexer) {
		return true
	}

	if ConsumeComment(lexer) {
		return true
	}

	if ConsumeCommonToken(lexer) {
		return true
	}

	// DivPunctuator
	if char == '/' {
		ConsumeChar(lexer)
		EmitToken(lexer, Divide)
		return true
	}

	if ConsumeTemplateSubstitutionTail(lexer) {
		return true
	}

	return false
}

func LexInputElementHashbangOrRegExp(lexer *Lexer) []Token {
	for !IsEOF(lexer) {
		char := CurrentChar(lexer)

		if LexNextInputElementHashbangOrRegExp(lexer) {
			continue
		}

		panic(fmt.Sprintf("Unexpected character: %c", char))
	}

	return lexer.Tokens
}

func LexNextInputElementHashbangOrRegExp(lexer *Lexer) bool {
	if IsEOF(lexer) {
		return false
	}

	if ConsumeWhiteSpace(lexer) {
		return true
	}

	if ConsumeLineTerminator(lexer) {
		return true
	}

	if ConsumeComment(lexer) {
		return true
	}

	if ConsumeCommonToken(lexer) {
		return true
	}

	if ConsumeHashbangComment(lexer) {
		return true
	}

	if ConsumeRegularExpressionLiteral(lexer) {
		return true
	}

	return false
}

func ConsumeWhiteSpace(lexer *Lexer) bool {
	if IsEOF(lexer) || !IsWhiteSpaceChar(CurrentChar(lexer)) {
		return false
	}

	for !IsEOF(lexer) && IsWhiteSpaceChar(CurrentChar(lexer)) {
		ConsumeChar(lexer)
	}
	EmitToken(lexer, WhiteSpace)
	return true
}

func ConsumeLineTerminator(lexer *Lexer) bool {
	if IsEOF(lexer) || !IsLineTerminator(CurrentChar(lexer)) {
		return false
	}

	if !IsEOF(lexer) && CurrentChar(lexer) == '\r' && CanLookahead(lexer) && LookaheadChar(lexer) == '\n' {
		// Consume "\r\n"
		ConsumeChar(lexer)
		ConsumeChar(lexer)
		EmitToken(lexer, LineTerminator)
	} else if !IsEOF(lexer) {
		ConsumeChar(lexer)
		EmitToken(lexer, LineTerminator)
	}

	return true
}

func ConsumeComment(lexer *Lexer) bool {
	if IsEOF(lexer) || !IsCommentStart(lexer) {
		return false
	}

	if IsSingleLineCommentStart(lexer) {
		ConsumeSingleLineComment(lexer)
		return true
	} else if IsMultiLineCommentStart(lexer) {
		ConsumeMultiLineComment(lexer)
		return true
	} else {
		panic("Should not be reached")
	}
}

func ConsumeHashbangComment(lexer *Lexer) bool {
	if IsEOF(lexer) || (CurrentChar(lexer) != '#' && (!CanLookahead(lexer) || LookaheadChar(lexer) != '!')) {
		return false
	}

	// Consume expected '#!' characters.
	ConsumeChar(lexer)
	ConsumeChar(lexer)

	// Consume the rest of the comment.
	for !IsEOF(lexer) && !IsLineTerminator(CurrentChar(lexer)) {
		ConsumeChar(lexer)
	}
	EmitToken(lexer, HashbangComment)
	return true
}

func ConsumeSingleLineComment(lexer *Lexer) {
	for !IsEOF(lexer) && !IsLineTerminator(CurrentChar(lexer)) {
		ConsumeChar(lexer)
	}
	EmitToken(lexer, Comment)
}

func ConsumeMultiLineComment(lexer *Lexer) {
	for !IsEOF(lexer) && !IsMultiLineCommentEnd(lexer) {
		ConsumeChar(lexer)
	}

	if !IsEOF(lexer) && IsMultiLineCommentEnd(lexer) {
		ConsumeChar(lexer)
		ConsumeChar(lexer)
		EmitToken(lexer, Comment)
	} else {
		panic("Expected */ to end multi-line comment")
	}
}

func ConsumeCommonToken(lexer *Lexer) bool {
	if IsEOF(lexer) {
		return false
	}

	char := CurrentChar(lexer)

	if ConsumeIdentifier(lexer) {
		return true
	}

	if ConsumePrivateIdentifier(lexer) {
		return true
	}

	if ConsumePunctuator(lexer) {
		return true
	}

	if IsDecimalStart(char, lexer) {
		ConsumeNumericLiteral(lexer)
		return true
	} else if char == '.' {
		// NOTE: This must be after numeric literals.
		if CanLookahead(lexer) && LookaheadChar(lexer) == '.' && CanLookaheadN(lexer, 2) && LookaheadCharN(lexer, 2) == '.' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, Spread)
		} else {
			ConsumeChar(lexer)
			EmitToken(lexer, Dot)
		}
		return true
	} else if char == '"' {
		ConsumeStringLiteral(lexer, '"')
		return true
	} else if char == '\'' {
		ConsumeStringLiteral(lexer, '\'')
		return true
	} else if char == '`' {
		ConsumeTemplateLiteralStart(lexer)
		return true
	}

	return false
}

func ConsumeIdentifier(lexer *Lexer) bool {
	if IsEOF(lexer) || !IsIdentifierStartChar(CurrentChar(lexer)) {
		return false
	}

	// Consume the first character (IdentifierStart)
	ConsumeChar(lexer)

	// Consume the remaining characters (IdentifierPart)
	for !IsEOF(lexer) && IsIdentifierPartChar(CurrentChar(lexer)) {
		ConsumeChar(lexer)
	}

	// If the identifier is a reserved word, emit the corresponding token.
	if _, ok := ReservedWords[lexer.CurrentTokenValue]; ok {
		EmitToken(lexer, ReservedWords[lexer.CurrentTokenValue])
		return true
	}

	EmitToken(lexer, Identifier)
	return true
}

func ConsumePrivateIdentifier(lexer *Lexer) bool {
	if IsEOF(lexer) || CurrentChar(lexer) != '#' {
		return false
	}

	// Consume expected '#' character.
	ConsumeChar(lexer)

	if IsEOF(lexer) {
		panic("Expected identifier after #")
	}

	// Consume the first character (IdentifierStart)
	if !IsIdentifierStartChar(CurrentChar(lexer)) {
		panic("Expected identifier after #")
	}
	ConsumeChar(lexer)

	// Consume the remaining characters (IdentifierPart)
	for !IsEOF(lexer) && IsIdentifierPartChar(CurrentChar(lexer)) {
		ConsumeChar(lexer)
	}

	EmitToken(lexer, PrivateIdentifier)
	return true
}

func ConsumePunctuator(lexer *Lexer) bool {
	if IsEOF(lexer) {
		return false
	}

	char := CurrentChar(lexer)

	if IsOptionalChain(lexer) {
		ConsumeChar(lexer)
		ConsumeChar(lexer)
		EmitToken(lexer, OptionalChain)
	} else if char == '{' {
		ConsumeChar(lexer)
		EmitToken(lexer, LeftBrace)
	} else if char == '(' {
		ConsumeChar(lexer)
		EmitToken(lexer, LeftParen)
	} else if char == ')' {
		ConsumeChar(lexer)
		EmitToken(lexer, RightParen)
	} else if char == '[' {
		ConsumeChar(lexer)
		EmitToken(lexer, LeftBracket)
	} else if char == ']' {
		ConsumeChar(lexer)
		EmitToken(lexer, RightBracket)
	} else if char == ';' {
		ConsumeChar(lexer)
		EmitToken(lexer, Semicolon)
	} else if char == ',' {
		ConsumeChar(lexer)
		EmitToken(lexer, Comma)
	} else if char == '<' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, LessThanEqual)
		} else if CanLookahead(lexer) && LookaheadChar(lexer) == '<' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)

			if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
				// <<=
				ConsumeChar(lexer)
				EmitToken(lexer, LeftShiftAssignment)
			} else {
				// <<
				EmitToken(lexer, LeftShift)
			}
		} else {
			ConsumeChar(lexer)
			EmitToken(lexer, LessThan)
		}
	} else if char == '>' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, GreaterThanEqual)
		} else if CanLookahead(lexer) && LookaheadChar(lexer) == '>' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)

			if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
				// >>=
				ConsumeChar(lexer)
				EmitToken(lexer, RightShiftAssignment)
			} else if !IsEOF(lexer) && CurrentChar(lexer) == '>' {
				ConsumeChar(lexer)
				if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
					// >>>=
					ConsumeChar(lexer)
					EmitToken(lexer, UnsignedRightShiftAssignment)
				} else {
					// >>>
					EmitToken(lexer, UnsignedRightShift)
				}
			} else {
				// >>
				EmitToken(lexer, RightShift)
			}
		} else {
			ConsumeChar(lexer)
			EmitToken(lexer, GreaterThan)
		}
	} else if char == '=' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)

			if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
				// ===
				ConsumeChar(lexer)
				EmitToken(lexer, StrictEqual)
			} else {
				// ==
				EmitToken(lexer, Equal)
			}
		} else if CanLookahead(lexer) && LookaheadChar(lexer) == '>' {
			// =>
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, ArrowOperator)
		} else {
			// =
			ConsumeChar(lexer)
			EmitToken(lexer, Assignment)
		}
	} else if char == '!' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)

			if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
				// !==
				ConsumeChar(lexer)
				EmitToken(lexer, StrictNotEqual)
			} else {
				// !=
				EmitToken(lexer, NotEqual)
			}
		} else {
			ConsumeChar(lexer)
			EmitToken(lexer, Not)
		}
	} else if char == '+' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '+' {
			// ++
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, Increment)
		} else if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			// +=
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, PlusAssignment)
		} else {
			// +
			ConsumeChar(lexer)
			EmitToken(lexer, Plus)
		}
	} else if char == '-' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '-' {
			// --
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, Decrement)
		} else if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			// -=
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, MinusAssignment)
		} else {
			// -
			ConsumeChar(lexer)
			EmitToken(lexer, Minus)
		}
	} else if char == '*' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			// *=
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, MultiplyAssignment)
		} else if CanLookahead(lexer) && LookaheadChar(lexer) == '*' {
			// **
			ConsumeChar(lexer)
			ConsumeChar(lexer)

			if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
				// **=
				ConsumeChar(lexer)
				EmitToken(lexer, ExponentiationAssignment)
			} else {
				// **
				EmitToken(lexer, Exponentiation)
			}
		} else {
			// *
			ConsumeChar(lexer)
			EmitToken(lexer, Multiply)
		}
	} else if char == '%' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			// %=
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, ModuloAssignment)
		} else {
			// %
			ConsumeChar(lexer)
			EmitToken(lexer, Modulo)
		}
	} else if char == '&' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			// &=
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, BitwiseAndAssignment)
		} else if CanLookahead(lexer) && LookaheadChar(lexer) == '&' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)

			if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
				// &&=
				ConsumeChar(lexer)
				EmitToken(lexer, AndAssignment)
			} else {
				// &&
				EmitToken(lexer, And)
			}
		} else {
			// &
			ConsumeChar(lexer)
			EmitToken(lexer, BitwiseAnd)
		}
	} else if char == '|' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			// |=
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, BitwiseOrAssignment)
		} else if CanLookahead(lexer) && LookaheadChar(lexer) == '|' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)

			if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
				// ||=
				ConsumeChar(lexer)
				EmitToken(lexer, OrAssignment)
			} else {
				// ||
				EmitToken(lexer, Or)
			}
		} else {
			// |
			ConsumeChar(lexer)
			EmitToken(lexer, BitwiseOr)
		}
	} else if char == '^' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
			// ^=
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, BitwiseXorAssignment)
		} else {
			// ^
			ConsumeChar(lexer)
			EmitToken(lexer, BitwiseXor)
		}
	} else if char == '~' {
		ConsumeChar(lexer)
		EmitToken(lexer, BitwiseNot)
	} else if char == '?' {
		if CanLookahead(lexer) && LookaheadChar(lexer) == '?' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			if !IsEOF(lexer) && CurrentChar(lexer) == '=' {
				// ??=
				ConsumeChar(lexer)
				EmitToken(lexer, NullishCoalescingAssignment)
			} else {
				// ??
				EmitToken(lexer, NullishCoalescing)
			}
		} else {
			// ?
			ConsumeChar(lexer)
			EmitToken(lexer, TernaryQuestionMark)
		}
	} else if char == ':' {
		ConsumeChar(lexer)
		EmitToken(lexer, TernaryColon)
	} else if char == '/' && CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
		// /=
		ConsumeChar(lexer)
		ConsumeChar(lexer)
		EmitToken(lexer, DivideAssignment)
	} else {
		return false
	}

	return true
}

func ConsumeNumericLiteral(lexer *Lexer) {
	if CurrentChar(lexer) == '0' && CanLookahead(lexer) && LookaheadChar(lexer) == 'x' {
		ConsumeHexIntegerLiteral(lexer)
	} else if CurrentChar(lexer) == '0' && CanLookahead(lexer) && unicode.ToLower(LookaheadChar(lexer)) == 'o' {
		ConsumeOctalIntegerLiteral(lexer)
	} else if CurrentChar(lexer) == '0' && CanLookahead(lexer) && unicode.ToLower(LookaheadChar(lexer)) == 'b' {
		ConsumeBinaryIntegerLiteral(lexer)
	} else if CurrentChar(lexer) == '.' {
		ConsumeDecimalDigitsAfterDot(lexer)
		EmitToken(lexer, NumericLiteral)
	} else if IsDecimalDigit(CurrentChar(lexer)) {
		ConsumeDecimalDigits(lexer)

		if !IsEOF(lexer) && CurrentChar(lexer) == '.' {
			ConsumeDecimalDigitsAfterDot(lexer)
		} else if !IsEOF(lexer) && CurrentChar(lexer) == 'n' {
			ConsumeBigIntLiteralSuffixIfPresent(lexer)
		} else if !IsEOF(lexer) && unicode.ToLower(CurrentChar(lexer)) == 'e' {
			ConsumeExponentPartIfPresent(lexer)
		}

		EmitToken(lexer, NumericLiteral)
	}
}

func ConsumeHexIntegerLiteral(lexer *Lexer) {
	// Consume '0x'
	ConsumeChar(lexer)
	ConsumeChar(lexer)

	if IsEOF(lexer) || !unicode.Is(unicode.Hex_Digit, CurrentChar(lexer)) {
		panic("Expected hex digit after 0x")
	}
	ConsumeChar(lexer)

	for IsHexIntegerPart(lexer) {
		ConsumeChar(lexer)
	}
	ConsumeBigIntLiteralSuffixIfPresent(lexer)

	EmitToken(lexer, NumericLiteral)
}

func ConsumeOctalIntegerLiteral(lexer *Lexer) {
	// Consume '0o'
	ConsumeChar(lexer)
	ConsumeChar(lexer)

	if IsEOF(lexer) || !IsOctalDigit(CurrentChar(lexer)) {
		panic("Expected octal digit after 0o")
	}
	ConsumeChar(lexer)

	for IsOctalIntegerPart(lexer) {
		ConsumeChar(lexer)
	}
	ConsumeBigIntLiteralSuffixIfPresent(lexer)

	EmitToken(lexer, NumericLiteral)
}

func ConsumeBinaryIntegerLiteral(lexer *Lexer) {
	// Consume '0b'
	ConsumeChar(lexer)
	ConsumeChar(lexer)

	if IsEOF(lexer) || !IsBinaryDigit(CurrentChar(lexer)) {
		panic("Expected binary digit after 0b")
	}
	ConsumeChar(lexer)

	for IsBinaryIntegerPart(lexer) {
		ConsumeChar(lexer)
	}
	ConsumeBigIntLiteralSuffixIfPresent(lexer)

	EmitToken(lexer, NumericLiteral)
}

func ConsumeBigIntLiteralSuffixIfPresent(lexer *Lexer) {
	// BigIntLiteralSuffix
	if !IsEOF(lexer) && CurrentChar(lexer) == 'n' {
		// n
		ConsumeChar(lexer)
	}
}

func ConsumeDecimalDigits(lexer *Lexer) {
	for !IsEOF(lexer) && IsDecimalDigitPart(lexer) {
		ConsumeChar(lexer)
	}
}

func ConsumeExponentPartIfPresent(lexer *Lexer) {
	if !IsEOF(lexer) && (CurrentChar(lexer) == 'e' || CurrentChar(lexer) == 'E') {
		// Consume 'e' or 'E'
		ConsumeChar(lexer)
		ConsumeSignedInteger(lexer)
	}
}

func ConsumeSignedInteger(lexer *Lexer) {
	if IsEOF(lexer) {
		return
	}

	if CurrentChar(lexer) == '-' || CurrentChar(lexer) == '+' {
		ConsumeChar(lexer)
	}

	ConsumeDecimalDigits(lexer)
}

func ConsumeDecimalDigitsAfterDot(lexer *Lexer) {
	// Consume '.'
	ConsumeChar(lexer)

	if IsEOF(lexer) || (!IsDecimalDigit(CurrentChar(lexer)) && unicode.ToLower(CurrentChar(lexer)) != 'e') {
		panic("Expected decimal digits after .")
	}

	ConsumeDecimalDigits(lexer)
	ConsumeExponentPartIfPresent(lexer)
}

func ConsumeStringLiteral(lexer *Lexer, quote rune) {
	// Consume start quote
	ConsumeChar(lexer)

	for !IsEOF(lexer) && CurrentChar(lexer) != quote {
		if CurrentChar(lexer) == '\\' && CanLookahead(lexer) && IsLineTerminator(LookaheadChar(lexer)) {
			// LineContinuation
			ConsumeLineContinuation(lexer)
			continue
		} else if CurrentChar(lexer) == '\\' {
			ConsumeStringLiteralEscapeSequence(lexer)
			continue
		}

		if IsLineTerminator(CurrentChar(lexer)) {
			panic("Unexpected line terminator")
		}

		ConsumeChar(lexer)
	}

	if IsEOF(lexer) {
		panic("Unterminated string literal")
	}

	if CurrentChar(lexer) != quote {
		panic("Unterminated string literal")
	}

	// Consume end quote
	ConsumeChar(lexer)
	EmitToken(lexer, StringLiteral)
}

func ConsumeStringLiteralEscapeSequence(lexer *Lexer) {
	// Consume '\'
	ConsumeChar(lexer)

	if IsEOF(lexer) {
		panic("Expected escape sequence after \\")
	}

	// SingleEscapeCharacter
	switch CurrentChar(lexer) {
	case 'v':
	case 't':
	case 'r':
	case 'n':
	case 'f':
	case 'b':
	case '\\':
	case '"':
	case '\'':
		ConsumeChar(lexer)
		return
	}

	if !IsEOF(lexer) &&
		!IsDecimalDigit(CurrentChar(lexer)) &&
		CurrentChar(lexer) != 'x' &&
		CurrentChar(lexer) != 'u' &&
		!IsLineTerminator(CurrentChar(lexer)) {
		// SourceCharacter but not one of EscapeCharacter or LineTerminator
		ConsumeChar(lexer)
		return
	}

	if !IsEOF(lexer) && CurrentChar(lexer) == '0' && (!CanLookahead(lexer) || !IsDecimalDigit(LookaheadChar(lexer))) {
		// \0 [lookahead != DecimalDigit]
		ConsumeChar(lexer)
		return
	}

	if !IsEOF(lexer) && CurrentChar(lexer) == '0' && (LookaheadChar(lexer) == '8' || LookaheadChar(lexer) == '9') {
		// \0 [lookahead is 8 or 9]
		ConsumeChar(lexer)
		return
	}

	if !IsEOF(lexer) && IsOctalDigit(CurrentChar(lexer)) && CurrentChar(lexer) != '0' && CanLookahead(lexer) && !IsOctalDigit(LookaheadChar(lexer)) {
		// NonZeroOctalDigit [lookahead != OctalDigit]
		ConsumeChar(lexer)
		return
	}

	if !IsEOF(lexer) && IsZeroToThree(CurrentChar(lexer)) {
		// ZeroToThree
		ConsumeChar(lexer)

		currentIsOctalDigit := IsOctalDigit(CurrentChar(lexer))
		lookaheadIsOctalDigit := CanLookahead(lexer) && IsOctalDigit(LookaheadChar(lexer))
		if !IsEOF(lexer) && currentIsOctalDigit && !lookaheadIsOctalDigit {
			// ZeroToThree OctalDigit [lookahead != OctalDigit]
			ConsumeChar(lexer)
			return
		}

		if !IsEOF(lexer) && currentIsOctalDigit && lookaheadIsOctalDigit {
			// ZeroToThree OctalDigit OctalDigit
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			return
		}
	}

	if !IsEOF(lexer) && IsFourToSeven(CurrentChar(lexer)) && CanLookahead(lexer) && IsOctalDigit(LookaheadChar(lexer)) {
		// FourToSeven OctalDigit
		ConsumeChar(lexer)
		ConsumeChar(lexer)
		return
	}

	if !IsEOF(lexer) && (CurrentChar(lexer) == '8' || CurrentChar(lexer) == '9') {
		// NonOctalEscapeSequence = one of 8 or 9
		ConsumeChar(lexer)
		return
	}

	if !IsEOF(lexer) && CurrentChar(lexer) == 'x' && CanLookahead(lexer) && unicode.Is(unicode.Hex_Digit, LookaheadChar(lexer)) {
		// HexEscapeSequence = \x HexDigit HexDigit
		ConsumeChar(lexer)

		if IsEOF(lexer) || !unicode.Is(unicode.Hex_Digit, CurrentChar(lexer)) {
			panic("Invalid hex escape sequence")
		}

		ConsumeChar(lexer)

		if IsEOF(lexer) || !unicode.Is(unicode.Hex_Digit, CurrentChar(lexer)) {
			panic("Invalid hex escape sequence")
		}

		ConsumeChar(lexer)
		return
	}

	if !IsEOF(lexer) && CurrentChar(lexer) == 'u' {
		// UnicodeEscapeSequence = \u HexDigit HexDigit HexDigit HexDigit
		ConsumeChar(lexer)

		if IsEOF(lexer) {
			panic("Invalid unicode escape sequence")
		}

		if CurrentChar(lexer) == '{' {
			ConsumeChar(lexer)

			// Consume N HexDigits with optional separators.
			for IsHexIntegerPart(lexer) {
				ConsumeChar(lexer)
			}

			if IsEOF(lexer) || CurrentChar(lexer) != '}' {
				panic("Invalid unicode escape sequence")
			}

			ConsumeChar(lexer)
			return
		} else if unicode.Is(unicode.Hex_Digit, CurrentChar(lexer)) {
			// Consume 4 HexDigits (no separators)
			for range 4 {
				if IsEOF(lexer) || !unicode.Is(unicode.Hex_Digit, CurrentChar(lexer)) {
					panic("Invalid unicode escape sequence")
				}
				ConsumeChar(lexer)
			}
			return
		} else {
			panic("Invalid unicode escape sequence")
		}
	}

	panic("Invalid escape sequence")
}

func ConsumeLineContinuation(lexer *Lexer) {
	// Consume \
	ConsumeChar(lexer)

	// <CR> <LF>
	if CurrentChar(lexer) == '\r' && CanLookahead(lexer) && LookaheadChar(lexer) == '\n' {
		ConsumeChar(lexer)
		ConsumeChar(lexer)
	} else {
		ConsumeChar(lexer)
	}
}

func ConsumeTemplateLiteralStart(lexer *Lexer) {
	// Consume start `
	ConsumeChar(lexer)

	for !IsEOF(lexer) && CurrentChar(lexer) != '`' {
		if CurrentChar(lexer) == '$' && CanLookahead(lexer) && LookaheadChar(lexer) == '{' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, TemplateStartLiteral)
			return
		}

		if CurrentChar(lexer) == '\\' && CanLookahead(lexer) && IsLineTerminator(LookaheadChar(lexer)) {
			ConsumeLineContinuation(lexer)
			continue
		}

		if CurrentChar(lexer) == '\\' {
			ConsumeStringLiteralEscapeSequence(lexer)
			continue
		}

		if IsLineTerminator(CurrentChar(lexer)) {
			panic("Unexpected line terminator")
		}

		ConsumeChar(lexer)
	}

	if IsEOF(lexer) {
		panic("Unterminated template literal")
	}

	if CurrentChar(lexer) != '`' {
		panic("Unterminated template literal")
	}

	// Consume end `
	ConsumeChar(lexer)
	EmitToken(lexer, TemplateNoSubstitutionLiteral)
}

func ConsumeTemplateSubstitutionTail(lexer *Lexer) bool {
	if IsEOF(lexer) {
		return false
	}

	for !IsEOF(lexer) && CurrentChar(lexer) != '`' {
		if CurrentChar(lexer) == '$' && CanLookahead(lexer) && LookaheadChar(lexer) == '{' {
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			EmitToken(lexer, TemplateMiddle)
			return true
		}

		if CurrentChar(lexer) == '\\' && CanLookahead(lexer) && IsLineTerminator(LookaheadChar(lexer)) {
			ConsumeLineContinuation(lexer)
			continue
		}

		if CurrentChar(lexer) == '\\' {
			ConsumeStringLiteralEscapeSequence(lexer)
			continue
		}

		if IsLineTerminator(CurrentChar(lexer)) {
			panic("Unexpected line terminator")
		}

		ConsumeChar(lexer)
	}

	if IsEOF(lexer) {
		panic("Unterminated template substitution tail")
	}

	if CurrentChar(lexer) != '`' {
		panic("Unterminated template substitution tail")
	}

	// Consume end `
	ConsumeChar(lexer)
	EmitToken(lexer, TemplateTail)
	return true
}

func ConsumeRegularExpressionLiteral(lexer *Lexer) bool {
	if IsEOF(lexer) || CurrentChar(lexer) != '/' {
		return false
	}

	// Consume start /
	ConsumeChar(lexer)

	if IsEOF(lexer) || CurrentChar(lexer) == '*' {
		panic("Invalid regular expression literal")
	}

	// RegularExpressionBody
	for !IsEOF(lexer) && CurrentChar(lexer) != '/' {
		if IsLineTerminator(CurrentChar(lexer)) {
			panic("Unexpected line terminator")
		}

		if CurrentChar(lexer) == '[' {
			ConsumeChar(lexer)

			for !IsEOF(lexer) && CurrentChar(lexer) != ']' && !IsLineTerminator(CurrentChar(lexer)) {
				ConsumeChar(lexer)
			}

			if IsEOF(lexer) || CurrentChar(lexer) != ']' {
				panic("Unterminated character class")
			}

			ConsumeChar(lexer)
			continue
		}

		// \ SourceCharacter but not one of LineTerminator
		if CurrentChar(lexer) == '\\' && CanLookahead(lexer) && !IsLineTerminator(LookaheadChar(lexer)) {
			ConsumeChar(lexer)
			ConsumeChar(lexer)
			continue
		}

		ConsumeChar(lexer)
	}

	if IsEOF(lexer) || CurrentChar(lexer) != '/' {
		panic("Unterminated regular expression literal")
	}

	// Consume end /
	ConsumeChar(lexer)

	// Consume RegularExpressionFlags
	for !IsEOF(lexer) && IsIdentifierPartChar(CurrentChar(lexer)) {
		ConsumeChar(lexer)
	}

	EmitToken(lexer, RegularExpressionLiteral)
	return true
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

func IsIDStartChar(char rune) bool {
	return ((unicode.Is(unicode.L, char) ||
		unicode.Is(unicode.Nl, char) ||
		unicode.Is(unicode.Other_ID_Start, char)) &&
		!unicode.Is(unicode.Pattern_Syntax, char) && !unicode.Is(unicode.Pattern_White_Space, char))
}

func IsIdentifierStartChar(char rune) bool {
	return (IsIDStartChar(char) ||
		char == '$' ||
		char == '_')
}

func IsIdentifierPartChar(char rune) bool {
	return ((unicode.Is(unicode.L, char) ||
		unicode.Is(unicode.Nl, char) ||
		unicode.Is(unicode.Other_ID_Start, char) ||
		unicode.Is(unicode.Other_ID_Continue, char) ||
		unicode.Is(unicode.Mn, char) ||
		unicode.Is(unicode.Mc, char) ||
		unicode.Is(unicode.Nd, char) ||
		unicode.Is(unicode.Pc, char) ||
		char == '$') &&
		!unicode.Is(unicode.Pattern_Syntax, char) && !unicode.Is(unicode.Pattern_White_Space, char))
}

func IsPrivateIdentifierStartChar(char rune) bool {
	return char == '#'
}

func IsCommentStart(lexer *Lexer) bool {
	return IsSingleLineCommentStart(lexer) || IsMultiLineCommentStart(lexer)
}

func IsSingleLineCommentStart(lexer *Lexer) bool {
	return CurrentChar(lexer) == '/' && CanLookahead(lexer) && LookaheadChar(lexer) == '/'
}

func IsMultiLineCommentStart(lexer *Lexer) bool {
	return CurrentChar(lexer) == '/' && CanLookahead(lexer) && LookaheadChar(lexer) == '*'
}

func IsMultiLineCommentEnd(lexer *Lexer) bool {
	return CurrentChar(lexer) == '*' && CanLookahead(lexer) && LookaheadChar(lexer) == '/'
}

func IsOptionalChain(lexer *Lexer) bool {
	// ?. [lookahead != DecimalDigit]
	return (CurrentChar(lexer) == '?' &&
		CanLookahead(lexer) &&
		LookaheadChar(lexer) == '.' &&
		(!CanLookaheadN(lexer, 2) ||
			!IsDecimalDigit(LookaheadCharN(lexer, 2))))
}

func IsDecimalDigit(char rune) bool {
	return regexp.MustCompile(`[0-9]`).MatchString(string(char))
}

func IsDecimalStart(currentChar rune, lexer *Lexer) bool {
	// DecimalDigit
	// . DecimalDigits
	return (IsDecimalDigit(currentChar) ||
		(currentChar == '.' && CanLookahead(lexer) && IsDecimalDigit(LookaheadChar(lexer))))
}

func IsHexIntegerPart(lexer *Lexer) bool {
	return !IsEOF(lexer) &&
		(unicode.Is(unicode.Hex_Digit, CurrentChar(lexer)) ||
			(CurrentChar(lexer) == '_' && CanLookahead(lexer) && unicode.Is(unicode.Hex_Digit, LookaheadChar(lexer))))
}

func IsOctalIntegerPart(lexer *Lexer) bool {
	return !IsEOF(lexer) &&
		(IsOctalDigit(CurrentChar(lexer)) ||
			(CurrentChar(lexer) == '_' && CanLookahead(lexer) && IsOctalDigit(LookaheadChar(lexer))))
}

func IsOctalDigit(char rune) bool {
	return regexp.MustCompile(`[0-7]`).MatchString(string(char))
}

func IsBinaryIntegerPart(lexer *Lexer) bool {
	return !IsEOF(lexer) &&
		(IsBinaryDigit(CurrentChar(lexer)) ||
			(CurrentChar(lexer) == '_' && CanLookahead(lexer) && IsBinaryDigit(LookaheadChar(lexer))))
}

func IsBinaryDigit(char rune) bool {
	return char == '0' || char == '1'
}

func IsDecimalDigitPart(lexer *Lexer) bool {
	return !IsEOF(lexer) &&
		(IsDecimalDigit(CurrentChar(lexer)) ||
			(CurrentChar(lexer) == '_' && CanLookahead(lexer) && IsDecimalDigit(LookaheadChar(lexer))))
}

func IsZeroToThree(char rune) bool {
	return char == '0' || char == '1' || char == '2' || char == '3'
}

func IsFourToSeven(char rune) bool {
	return char == '4' || char == '5' || char == '6' || char == '7'
}
