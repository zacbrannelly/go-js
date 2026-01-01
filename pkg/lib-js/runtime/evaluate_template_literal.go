package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateTemplateLiteral(runtime *Runtime, templateLiteral *ast.TemplateLiteralNode) *Completion {
	result := ""

	for _, child := range templateLiteral.GetChildren() {
		completion := Evaluate(runtime, child)
		if completion.Type != Normal {
			return completion
		}

		maybeRef := completion.Value.(*JavaScriptValue)
		completion = GetValue(runtime, maybeRef)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)

		completion = ToString(runtime, value)
		if completion.Type != Normal {
			return completion
		}

		stringValue := completion.Value.(*JavaScriptValue).Value.(*String).Value
		result += stringValue
	}

	return NewNormalCompletion(NewStringValue(result))
}
