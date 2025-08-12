package lexer

import (
	"testing"
)

func compareTokens(t *testing.T, expected, actual []Token) {
	if len(expected) != len(actual) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(actual))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Actual: %v", actual)
		return
	}

	for i, expectedToken := range expected {
		actualToken := actual[i]
		if expectedToken.Type != actualToken.Type {
			t.Errorf("Token %d: Expected type %d, got %d", i, expectedToken.Type, actualToken.Type)
		}
		if expectedToken.Value != actualToken.Value {
			t.Errorf("Token %d: Expected value '%s', got '%s'", i, expectedToken.Value, actualToken.Value)
		}
	}
}

func executeTest(t *testing.T, input string, goal LexicalGoal, expected []Token) {
	func() {
		defer func() {
			// Catch panics and report them.
			if r := recover(); r != nil {
				t.Errorf("Lexing %q: %v", input, r)
			}
		}()
		tokens := Lex(input, goal)
		compareTokens(t, expected, tokens)
	}()
}

func executeTests(t *testing.T, tests []struct {
	input    string
	expected []Token
}, goal LexicalGoal) {
	for _, test := range tests {
		executeTest(t, test.input, goal, test.expected)
	}
}

// Test whitespace tokens
func TestWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: " ",
			expected: []Token{
				{Type: WhiteSpace, Value: " "},
			},
		},
		{
			input: "\t",
			expected: []Token{
				{Type: WhiteSpace, Value: "\t"},
			},
		},
		{
			input: "\v",
			expected: []Token{
				{Type: WhiteSpace, Value: "\v"},
			},
		},
		{
			input: "\f",
			expected: []Token{
				{Type: WhiteSpace, Value: "\f"},
			},
		},
		{
			input: "   \t\v",
			expected: []Token{
				{Type: WhiteSpace, Value: "   \t\v"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test line terminators
func TestLineTerminators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "\n",
			expected: []Token{
				{Type: LineTerminator, Value: "\n"},
			},
		},
		{
			input: "\r",
			expected: []Token{
				{Type: LineTerminator, Value: "\r"},
			},
		},
		{
			input: "\r\n",
			expected: []Token{
				{Type: LineTerminator, Value: "\r\n"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test comments
func TestComments(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "// single line comment",
			expected: []Token{
				{Type: Comment, Value: "// single line comment"},
			},
		},
		{
			input: "/* multi line comment */",
			expected: []Token{
				{Type: Comment, Value: "/* multi line comment */"},
			},
		},
		{
			input: "/* multi\nline\ncomment */",
			expected: []Token{
				{Type: Comment, Value: "/* multi\nline\ncomment */"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test identifiers
func TestIdentifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "variable",
			expected: []Token{
				{Type: Identifier, Value: "variable"},
			},
		},
		{
			input: "$dollar",
			expected: []Token{
				{Type: Identifier, Value: "$dollar"},
			},
		},
		{
			input: "_underscore",
			expected: []Token{
				{Type: Identifier, Value: "_underscore"},
			},
		},
		{
			input: "var123",
			expected: []Token{
				{Type: Identifier, Value: "var123"},
			},
		},
		{
			input: "camelCase",
			expected: []Token{
				{Type: Identifier, Value: "camelCase"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test private identifiers
func TestPrivateIdentifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "#private",
			expected: []Token{
				{Type: PrivateIdentifier, Value: "#private"},
			},
		},
		{
			input: "#_private",
			expected: []Token{
				{Type: PrivateIdentifier, Value: "#_private"},
			},
		},
		{
			input: "#private123",
			expected: []Token{
				{Type: PrivateIdentifier, Value: "#private123"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test basic punctuators
func TestBasicPunctuators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "{",
			expected: []Token{
				{Type: LeftBrace, Value: "{"},
			},
		},
		{
			input: "}",
			expected: []Token{
				{Type: RightBrace, Value: "}"},
			},
		},
		{
			input: "[",
			expected: []Token{
				{Type: LeftBracket, Value: "["},
			},
		},
		{
			input: "]",
			expected: []Token{
				{Type: RightBracket, Value: "]"},
			},
		},
		{
			input: "(",
			expected: []Token{
				{Type: LeftParen, Value: "("},
			},
		},
		{
			input: ")",
			expected: []Token{
				{Type: RightParen, Value: ")"},
			},
		},
		{
			input: ".",
			expected: []Token{
				{Type: Dot, Value: "."},
			},
		},
		{
			input: "...",
			expected: []Token{
				{Type: Spread, Value: "..."},
			},
		},
		{
			input: ";",
			expected: []Token{
				{Type: Semicolon, Value: ";"},
			},
		},
		{
			input: ",",
			expected: []Token{
				{Type: Comma, Value: ","},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test comparison operators
func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "<",
			expected: []Token{
				{Type: LessThan, Value: "<"},
			},
		},
		{
			input: "<=",
			expected: []Token{
				{Type: LessThanEqual, Value: "<="},
			},
		},
		{
			input: ">",
			expected: []Token{
				{Type: GreaterThan, Value: ">"},
			},
		},
		{
			input: ">=",
			expected: []Token{
				{Type: GreaterThanEqual, Value: ">="},
			},
		},
		{
			input: "==",
			expected: []Token{
				{Type: Equal, Value: "=="},
			},
		},
		{
			input: "===",
			expected: []Token{
				{Type: StrictEqual, Value: "==="},
			},
		},
		{
			input: "!=",
			expected: []Token{
				{Type: NotEqual, Value: "!="},
			},
		},
		{
			input: "!==",
			expected: []Token{
				{Type: StrictNotEqual, Value: "!=="},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test arithmetic operators
func TestArithmeticOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "+",
			expected: []Token{
				{Type: Plus, Value: "+"},
			},
		},
		{
			input: "++",
			expected: []Token{
				{Type: Increment, Value: "++"},
			},
		},
		{
			input: "-",
			expected: []Token{
				{Type: Minus, Value: "-"},
			},
		},
		{
			input: "--",
			expected: []Token{
				{Type: Decrement, Value: "--"},
			},
		},
		{
			input: "*",
			expected: []Token{
				{Type: Multiply, Value: "*"},
			},
		},
		{
			input: "**",
			expected: []Token{
				{Type: Exponentiation, Value: "**"},
			},
		},
		{
			input: "/",
			expected: []Token{
				{Type: Divide, Value: "/"},
			},
		},
		{
			input: "%",
			expected: []Token{
				{Type: Modulo, Value: "%"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test shift operators
func TestShiftOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "<<",
			expected: []Token{
				{Type: LeftShift, Value: "<<"},
			},
		},
		{
			input: ">>",
			expected: []Token{
				{Type: RightShift, Value: ">>"},
			},
		},
		{
			input: ">>>",
			expected: []Token{
				{Type: UnsignedRightShift, Value: ">>>"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test bitwise operators
func TestBitwiseOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "&",
			expected: []Token{
				{Type: BitwiseAnd, Value: "&"},
			},
		},
		{
			input: "|",
			expected: []Token{
				{Type: BitwiseOr, Value: "|"},
			},
		},
		{
			input: "^",
			expected: []Token{
				{Type: BitwiseXor, Value: "^"},
			},
		},
		{
			input: "~",
			expected: []Token{
				{Type: BitwiseNot, Value: "~"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test logical operators
func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "!",
			expected: []Token{
				{Type: Not, Value: "!"},
			},
		},
		{
			input: "&&",
			expected: []Token{
				{Type: And, Value: "&&"},
			},
		},
		{
			input: "||",
			expected: []Token{
				{Type: Or, Value: "||"},
			},
		},
		{
			input: "??",
			expected: []Token{
				{Type: NullishCoalescing, Value: "??"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test assignment operators
func TestAssignmentOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "=",
			expected: []Token{
				{Type: Assignment, Value: "="},
			},
		},
		{
			input: "+=",
			expected: []Token{
				{Type: PlusAssignment, Value: "+="},
			},
		},
		{
			input: "-=",
			expected: []Token{
				{Type: MinusAssignment, Value: "-="},
			},
		},
		{
			input: "*=",
			expected: []Token{
				{Type: MultiplyAssignment, Value: "*="},
			},
		},
		{
			input: "/=",
			expected: []Token{
				{Type: DivideAssignment, Value: "/="},
			},
		},
		{
			input: "%=",
			expected: []Token{
				{Type: ModuloAssignment, Value: "%="},
			},
		},
		{
			input: "**=",
			expected: []Token{
				{Type: ExponentiationAssignment, Value: "**="},
			},
		},
		{
			input: "<<=",
			expected: []Token{
				{Type: LeftShiftAssignment, Value: "<<="},
			},
		},
		{
			input: ">>=",
			expected: []Token{
				{Type: RightShiftAssignment, Value: ">>="},
			},
		},
		{
			input: ">>>=",
			expected: []Token{
				{Type: UnsignedRightShiftAssignment, Value: ">>>="},
			},
		},
		{
			input: "&=",
			expected: []Token{
				{Type: BitwiseAndAssignment, Value: "&="},
			},
		},
		{
			input: "|=",
			expected: []Token{
				{Type: BitwiseOrAssignment, Value: "|="},
			},
		},
		{
			input: "^=",
			expected: []Token{
				{Type: BitwiseXorAssignment, Value: "^="},
			},
		},
		{
			input: "&&=",
			expected: []Token{
				{Type: AndAssignment, Value: "&&="},
			},
		},
		{
			input: "||=",
			expected: []Token{
				{Type: OrAssignment, Value: "||="},
			},
		},
		{
			input: "??=",
			expected: []Token{
				{Type: NullishCoalescingAssignment, Value: "??="},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test special operators
func TestSpecialOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "=>",
			expected: []Token{
				{Type: ArrowOperator, Value: "=>"},
			},
		},
		{
			input: "?.",
			expected: []Token{
				{Type: OptionalChain, Value: "?."},
			},
		},
		{
			input: "?",
			expected: []Token{
				{Type: TernaryQuestionMark, Value: "?"},
			},
		},
		{
			input: ":",
			expected: []Token{
				{Type: TernaryColon, Value: ":"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test numeric literals
func TestNumericLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		// Basic decimal integers
		{
			input: "0",
			expected: []Token{
				{Type: NumericLiteral, Value: "0"},
			},
		},
		{
			input: "123",
			expected: []Token{
				{Type: NumericLiteral, Value: "123"},
			},
		},
		{
			input: "999",
			expected: []Token{
				{Type: NumericLiteral, Value: "999"},
			},
		},
		// Decimal integers with separators
		{
			input: "1_000",
			expected: []Token{
				{Type: NumericLiteral, Value: "1_000"},
			},
		},
		{
			input: "1_000_000",
			expected: []Token{
				{Type: NumericLiteral, Value: "1_000_000"},
			},
		},
		// Decimal floats
		{
			input: "0.5",
			expected: []Token{
				{Type: NumericLiteral, Value: "0.5"},
			},
		},
		{
			input: "123.456",
			expected: []Token{
				{Type: NumericLiteral, Value: "123.456"},
			},
		},
		{
			input: ".5",
			expected: []Token{
				{Type: NumericLiteral, Value: ".5"},
			},
		},
		{
			input: ".123",
			expected: []Token{
				{Type: NumericLiteral, Value: ".123"},
			},
		},
		// Decimal floats with separators
		{
			input: "1_000.5",
			expected: []Token{
				{Type: NumericLiteral, Value: "1_000.5"},
			},
		},
		{
			input: "123.456_789",
			expected: []Token{
				{Type: NumericLiteral, Value: "123.456_789"},
			},
		},
		// Scientific notation
		{
			input: "1e5",
			expected: []Token{
				{Type: NumericLiteral, Value: "1e5"},
			},
		},
		{
			input: "1E5",
			expected: []Token{
				{Type: NumericLiteral, Value: "1E5"},
			},
		},
		{
			input: "1e+5",
			expected: []Token{
				{Type: NumericLiteral, Value: "1e+5"},
			},
		},
		{
			input: "1e-5",
			expected: []Token{
				{Type: NumericLiteral, Value: "1e-5"},
			},
		},
		{
			input: "123.456e7",
			expected: []Token{
				{Type: NumericLiteral, Value: "123.456e7"},
			},
		},
		{
			input: "123.456E-7",
			expected: []Token{
				{Type: NumericLiteral, Value: "123.456E-7"},
			},
		},
		{
			input: ".5e10",
			expected: []Token{
				{Type: NumericLiteral, Value: ".5e10"},
			},
		},
		// Scientific notation with separators
		{
			input: "1_000e5",
			expected: []Token{
				{Type: NumericLiteral, Value: "1_000e5"},
			},
		},
		{
			input: "1.5_00e-1_0",
			expected: []Token{
				{Type: NumericLiteral, Value: "1.5_00e-1_0"},
			},
		},
		// Hexadecimal integers
		{
			input: "0x0",
			expected: []Token{
				{Type: NumericLiteral, Value: "0x0"},
			},
		},
		{
			input: "0x123",
			expected: []Token{
				{Type: NumericLiteral, Value: "0x123"},
			},
		},
		{
			input: "0xabc",
			expected: []Token{
				{Type: NumericLiteral, Value: "0xabc"},
			},
		},
		{
			input: "0xABC",
			expected: []Token{
				{Type: NumericLiteral, Value: "0xABC"},
			},
		},
		{
			input: "0x123abc",
			expected: []Token{
				{Type: NumericLiteral, Value: "0x123abc"},
			},
		},
		{
			input: "0xDEADBEEF",
			expected: []Token{
				{Type: NumericLiteral, Value: "0xDEADBEEF"},
			},
		},
		// Hexadecimal with separators
		{
			input: "0x1_23_abc",
			expected: []Token{
				{Type: NumericLiteral, Value: "0x1_23_abc"},
			},
		},
		{
			input: "0xFF_FF_FF",
			expected: []Token{
				{Type: NumericLiteral, Value: "0xFF_FF_FF"},
			},
		},
		// Octal integers
		{
			input: "0o0",
			expected: []Token{
				{Type: NumericLiteral, Value: "0o0"},
			},
		},
		{
			input: "0o123",
			expected: []Token{
				{Type: NumericLiteral, Value: "0o123"},
			},
		},
		{
			input: "0o7654321",
			expected: []Token{
				{Type: NumericLiteral, Value: "0o7654321"},
			},
		},
		{
			input: "0O777", // uppercase O
			expected: []Token{
				{Type: NumericLiteral, Value: "0O777"},
			},
		},
		// Octal with separators
		{
			input: "0o1_23_456",
			expected: []Token{
				{Type: NumericLiteral, Value: "0o1_23_456"},
			},
		},
		// Binary integers
		{
			input: "0b0",
			expected: []Token{
				{Type: NumericLiteral, Value: "0b0"},
			},
		},
		{
			input: "0b1",
			expected: []Token{
				{Type: NumericLiteral, Value: "0b1"},
			},
		},
		{
			input: "0b101010",
			expected: []Token{
				{Type: NumericLiteral, Value: "0b101010"},
			},
		},
		{
			input: "0B11111111", // uppercase B
			expected: []Token{
				{Type: NumericLiteral, Value: "0B11111111"},
			},
		},
		// Binary with separators
		{
			input: "0b1010_1010",
			expected: []Token{
				{Type: NumericLiteral, Value: "0b1010_1010"},
			},
		},
		{
			input: "0b1111_0000_1111_0000",
			expected: []Token{
				{Type: NumericLiteral, Value: "0b1111_0000_1111_0000"},
			},
		},
		// BigInt literals
		{
			input: "123n",
			expected: []Token{
				{Type: NumericLiteral, Value: "123n"},
			},
		},
		{
			input: "0x123n",
			expected: []Token{
				{Type: NumericLiteral, Value: "0x123n"},
			},
		},
		{
			input: "0o123n",
			expected: []Token{
				{Type: NumericLiteral, Value: "0o123n"},
			},
		},
		{
			input: "0b101n",
			expected: []Token{
				{Type: NumericLiteral, Value: "0b101n"},
			},
		},
		// BigInt with separators
		{
			input: "1_000n",
			expected: []Token{
				{Type: NumericLiteral, Value: "1_000n"},
			},
		},
		{
			input: "0x1_23_ABCn",
			expected: []Token{
				{Type: NumericLiteral, Value: "0x1_23_ABCn"},
			},
		},

		// Edge cases
		{
			input: "0.0",
			expected: []Token{
				{Type: NumericLiteral, Value: "0.0"},
			},
		},
		{
			input: "0e0",
			expected: []Token{
				{Type: NumericLiteral, Value: "0e0"},
			},
		},
		{
			input: "1.23e+45",
			expected: []Token{
				{Type: NumericLiteral, Value: "1.23e+45"},
			},
		},
		{
			input: "9.876e-54",
			expected: []Token{
				{Type: NumericLiteral, Value: "9.876e-54"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		// Basic double-quoted strings
		{
			input: `""`,
			expected: []Token{
				{Type: StringLiteral, Value: `""`},
			},
		},
		{
			input: `"hello"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"hello"`},
			},
		},
		{
			input: `"Hello, world!"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"Hello, world!"`},
			},
		},
		// Basic single-quoted strings
		{
			input: `''`,
			expected: []Token{
				{Type: StringLiteral, Value: `''`},
			},
		},
		{
			input: `'hello'`,
			expected: []Token{
				{Type: StringLiteral, Value: `'hello'`},
			},
		},
		{
			input: `'Hello, world!'`,
			expected: []Token{
				{Type: StringLiteral, Value: `'Hello, world!'`},
			},
		},
		// Strings with spaces and special characters
		{
			input: `"   spaces   "`,
			expected: []Token{
				{Type: StringLiteral, Value: `"   spaces   "`},
			},
		},
		{
			input: `"!@#$%^&*()_+-=[]{}|;:,.<>?"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"!@#$%^&*()_+-=[]{}|;:,.<>?"`},
			},
		},
		// Basic escape sequences
		{
			input: `"\"quoted\""`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\"quoted\""`},
			},
		},
		{
			input: `'\'quoted\''`,
			expected: []Token{
				{Type: StringLiteral, Value: `'\'quoted\''`},
			},
		},
		{
			input: `"\\backslash"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\\backslash"`},
			},
		},
		{
			input: `"\n\t\r"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\n\t\r"`},
			},
		},
		{
			input: `"\b\f\v"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\b\f\v"`},
			},
		},
		// Null escape sequence
		{
			input: `"\0"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\0"`},
			},
		},
		// Hex escape sequences
		{
			input: `"\x41"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\x41"`},
			},
		},
		{
			input: `"\x20\x48\x65\x6C\x6C\x6F"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\x20\x48\x65\x6C\x6C\x6F"`},
			},
		},
		{
			input: `"\xff\xFF"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\xff\xFF"`},
			},
		},
		// Unicode escape sequences (4-digit)
		{
			input: `"\u0041"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\u0041"`},
			},
		},
		{
			input: `"\u2603"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\u2603"`},
			},
		},
		{
			input: `"\uD83D\uDE00"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\uD83D\uDE00"`},
			},
		},
		// Unicode escape sequences (variable length)
		{
			input: `"\u{41}"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\u{41}"`},
			},
		},
		{
			input: `"\u{1F600}"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\u{1F600}"`},
			},
		},
		{
			input: `"\u{10FFFF}"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\u{10FFFF}"`},
			},
		},
		// Octal escape sequences
		{
			input: `"\101"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\101"`},
			},
		},
		{
			input: `"\377"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\377"`},
			},
		},
		{
			input: `"\1\12\123"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\1\12\123"`},
			},
		},
		{
			input: `"\4\56"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\4\56"`},
			},
		},
		// Non-octal escape sequences (8 and 9)
		{
			input: `"\8\9"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\8\9"`},
			},
		},
		// Mixed escape sequences
		{
			input: `"Hello\nWorld\t\u0021"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"Hello\nWorld\t\u0021"`},
			},
		},
		{
			input: `"Quote: \"Hello\", Backslash: \\"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"Quote: \"Hello\", Backslash: \\"`},
			},
		},
		// Non-escape sequences (characters after backslash that don't form valid escapes)
		{
			input: `"\a\c\d\e\g\h\i\j\k\l\m\o\p\q\s\w\y\z"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\a\c\d\e\g\h\i\j\k\l\m\o\p\q\s\w\y\z"`},
			},
		},
		// Strings with numbers and identifiers-like content
		{
			input: `"123"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"123"`},
			},
		},
		{
			input: `"identifier"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"identifier"`},
			},
		},
		{
			input: `"var x = 42;"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"var x = 42;"`},
			},
		},
		// Mixed quotes (single quotes inside double quotes and vice versa)
		{
			input: `"It's a beautiful day"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"It's a beautiful day"`},
			},
		},
		{
			input: `'He said "Hello"'`,
			expected: []Token{
				{Type: StringLiteral, Value: `'He said "Hello"'`},
			},
		},
		// Empty and whitespace-only strings
		{
			input: `" "`,
			expected: []Token{
				{Type: StringLiteral, Value: `" "`},
			},
		},
		{
			input: `"\t\n\r"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"\t\n\r"`},
			},
		},
		// Long strings
		{
			input: `"This is a very long string that contains multiple words and should be tokenized as a single string literal token."`,
			expected: []Token{
				{Type: StringLiteral, Value: `"This is a very long string that contains multiple words and should be tokenized as a single string literal token."`},
			},
		},
		// Strings with all printable ASCII characters
		{
			input: `"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"`,
			expected: []Token{
				{Type: StringLiteral, Value: `"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"`},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test complex expressions
func TestComplexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "obj?.prop",
			expected: []Token{
				{Type: Identifier, Value: "obj"},
				{Type: OptionalChain, Value: "?."},
				{Type: Identifier, Value: "prop"},
			},
		},
		{
			input: "a + b * c",
			expected: []Token{
				{Type: Identifier, Value: "a"},
				{Type: WhiteSpace, Value: " "},
				{Type: Plus, Value: "+"},
				{Type: WhiteSpace, Value: " "},
				{Type: Identifier, Value: "b"},
				{Type: WhiteSpace, Value: " "},
				{Type: Multiply, Value: "*"},
				{Type: WhiteSpace, Value: " "},
				{Type: Identifier, Value: "c"},
			},
		},
		{
			input: "arr[identifier]",
			expected: []Token{
				{Type: Identifier, Value: "arr"},
				{Type: LeftBracket, Value: "["},
				{Type: Identifier, Value: "identifier"},
				{Type: RightBracket, Value: "]"},
			},
		},
	}
	executeTests(t, tests, InputElementDiv)
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	// Test empty input
	tokens := Lex("", InputElementDiv)
	if len(tokens) != 0 {
		t.Errorf("Expected 0 tokens for empty input, got %d", len(tokens))
	}

	// Test optional chain vs ternary with decimal
	expected := []Token{
		{Type: TernaryQuestionMark, Value: "?"},
		{Type: NumericLiteral, Value: ".5"},
	}
	executeTest(t, "?.5", InputElementDiv, expected)
}
