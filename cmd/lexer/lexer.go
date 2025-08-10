package lexer

import (
	"fmt"
	"regexp"
	"unicode"
)

type LexicalGoal int

const (
	InputElementDiv LexicalGoal = iota
)

type TokenType int

const (
	WhiteSpace TokenType = iota
	LineTerminator
	Comment
	Identifier
	PrivateIdentifier
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
	CurrentTokenValue string
}

func Lex(input string, goal LexicalGoal) []Token {
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

		if IsWhiteSpaceChar(char) {
			ConsumeWhiteSpace(lexer)
		} else if IsLineTerminator(char) {
			ConsumeLineTerminator(lexer)
		} else if IsCommentStart(lexer) {
			ConsumeComment(lexer)
		} else if IsIdentifierStartChar(char) {
			ConsumeIdentifier(lexer)
		} else if char == '#' {
			ConsumePrivateIdentifier(lexer)
		} else if IsOptionalChain(lexer) {
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
		} else if char == '.' {
			if CanLookahead(lexer) && LookaheadChar(lexer) == '.' && CanLookaheadN(lexer, 2) && LookaheadCharN(lexer, 2) == '.' {
				ConsumeChar(lexer)
				ConsumeChar(lexer)
				ConsumeChar(lexer)
				EmitToken(lexer, Spread)
			} else {
				ConsumeChar(lexer)
				EmitToken(lexer, Dot)
			}
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
		} else if char == '/' {
			if CanLookahead(lexer) && LookaheadChar(lexer) == '=' {
				// /=
				ConsumeChar(lexer)
				ConsumeChar(lexer)
				EmitToken(lexer, DivideAssignment)
			} else {
				// /
				ConsumeChar(lexer)
				EmitToken(lexer, Divide)
			}
		} else if char == '}' {
			ConsumeChar(lexer)
			EmitToken(lexer, RightBrace)
		} else {
			panic(fmt.Sprintf("Unexpected character: %c", char))
		}
	}

	return lexer.Tokens
}

func ConsumeWhiteSpace(lexer *Lexer) {
	for !IsEOF(lexer) && IsWhiteSpaceChar(CurrentChar(lexer)) {
		ConsumeChar(lexer)
	}
	EmitToken(lexer, WhiteSpace)
}

func ConsumeLineTerminator(lexer *Lexer) {
	if !IsEOF(lexer) && CurrentChar(lexer) == '\r' && CanLookahead(lexer) && LookaheadChar(lexer) == '\n' {
		ConsumeChar(lexer)
		ConsumeChar(lexer)
		EmitToken(lexer, LineTerminator)
	} else if !IsEOF(lexer) {
		ConsumeChar(lexer)
		EmitToken(lexer, LineTerminator)
	}
}

func ConsumeComment(lexer *Lexer) {
	if IsSingleLineCommentStart(lexer) {
		ConsumeSingleLineComment(lexer)
	} else if IsMultiLineCommentStart(lexer) {
		ConsumeMultiLineComment(lexer)
	} else {
		panic("Should not be reached")
	}
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

func ConsumeIdentifier(lexer *Lexer) {
	// Consume the first character (IdentifierStart)
	ConsumeChar(lexer)

	// Consume the remaining characters (IdentifierPart)
	for !IsEOF(lexer) && IsIdentifierPartChar(CurrentChar(lexer)) {
		ConsumeChar(lexer)
	}

	EmitToken(lexer, Identifier)
}

func ConsumePrivateIdentifier(lexer *Lexer) {
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
