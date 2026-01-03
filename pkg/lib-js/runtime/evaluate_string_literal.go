package runtime

import (
	"regexp"
	"strconv"
	"strings"

	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateStringLiteral(runtime *Runtime, stringLiteral *ast.StringLiteralNode) *Completion {
	if strings.Contains(stringLiteral.Value, "\\u") {
		stringLiteral.Value = regexp.MustCompile(`\\u[0-9a-fA-F]{4}`).ReplaceAllStringFunc(stringLiteral.Value, func(match string) string {
			hexCode := match[2:]
			unicodeCodePoint, err := strconv.ParseInt(hexCode, 16, 32)
			if err != nil {
				panic(err)
			}
			return string(rune(unicodeCodePoint))
		})
	}
	return NewNormalCompletion(NewStringValue(stringLiteral.Value))
}
