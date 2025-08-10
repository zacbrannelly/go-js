package main

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

func main() {
	input := `
		/** Hello, Multi-line comment on one line! */
		// Hello, Single-line comment!
		/*
		  Hello Multi-line comment!
		*/
		helloIdentifier
		#helloPrivateIdentifier
		helloAssignment = value
		{ helloBraces }
		helloBrackets[helloThere]
		helloParentheses(helloThere)
		helloSemicolon;
		helloComma, helloThere
		helloColon: helloThere
		helloDot.helloThere
		helloArrow => helloThere
		helloQuestionMark? helloThere
		helloExclamationMark! helloThere
		helloLessThan < helloThere
		helloGreaterThan > helloThere
		helloLessThanEqual <= helloThere
		helloGreaterThanEqual >= helloThere
		helloEqual == helloThere
		helloNotEqual != helloThere
		helloStrictEqual === helloThere
		helloStrictNotEqual !== helloThere
		helloPlus + helloThere
		helloMinus - helloThere
		helloMultiply * helloThere
		helloModulo % helloThere
		helloExponentiation ** helloThere
		helloIncrement++
		helloDecrement--
		helloLeftShift << helloThere
		helloRightShift >> helloThere
		helloUnsignedRightShift >>> helloThere
		helloBitwiseAnd & helloThere
		helloBitwiseOr | helloThere
		helloBitwiseXor ^ helloThere
		helloNot!
		helloBitwiseNot~
		helloAnd && helloThere
		helloOr || helloThere
		helloNullishCoalescing ?? helloThere
		helloTernary ? helloThen : helloElse
		helloAssignment = helloThere
		helloPlusAssignment += helloThere
		helloMinusAssignment -= helloThere
		helloMultiplyAssignment *= helloThere
		helloModuloAssignment %= helloThere
		helloExponentiationAssignment **= helloThere
		helloLeftShiftAssignment <<= helloThere
		helloRightShiftAssignment >>= helloThere
		helloUnsignedRightShiftAssignment >>>= helloThere
		helloBitwiseAndAssignment &= helloThere
		helloBitwiseOrAssignment |= helloThere
		helloBitwiseXorAssignment ^= helloThere
		helloAndAssignment &&= helloThere
		helloOrAssignment ||= helloThere
		helloNullishCoalescingAssignment ??= helloThere
		helloSpread ...helloThere
		helloArrowFunction => helloThere
		0x100
		0o100
		0b100
		100n
		0n
		0x100n
		0x10_100n
		0o100n
		0b100n
		0.1234
		1234.5678
		1234.5678
		1234.5678e1234
		1234.5678e+1234
		1234.5678e-1234
	`
	tokens := lexer.Lex(input, lexer.InputElementDiv)
	for _, token := range tokens {
		fmt.Printf("%d: %s\n", token.Type, token.Value)
	}
}
