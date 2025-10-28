package runtime

import (
	"strconv"

	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateArrayLiteral(runtime *Runtime, arrayLiteral *ast.BasicNode) *Completion {
	array := NewArrayObject(0)
	arrayObj := NewJavaScriptValue(TypeObject, array)

	length := 0
	lengthStr := NewStringValue("length")

	for _, element := range arrayLiteral.GetChildren() {
		if element.GetNodeType() == ast.Elision {
			length++

			// NOTE: This should have the same semantics as the "Set(O, P, V, Throw)" operation in the spec.
			success := array.Set(lengthStr, NewNumberValue(float64(length), false), arrayObj)
			if success.Type != Normal {
				return success
			}

			if !success.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				return NewThrowCompletion(NewTypeError("Cannot build array because the length property is not writable."))
			}

			continue
		}

		maybeRefCompletion := Evaluate(runtime, element)
		if maybeRefCompletion.Type != Normal {
			return maybeRefCompletion
		}

		maybeRef := maybeRefCompletion.Value.(*JavaScriptValue)
		valCompletion := GetValue(maybeRef)
		if valCompletion.Type != Normal {
			return valCompletion
		}

		// NOTE: This should have the same semantics as "CreateDataPropertyOrThrow" in the spec.
		success := CreateDataProperty(array, NewStringValue(strconv.FormatInt(int64(length), 10)), valCompletion.Value.(*JavaScriptValue))
		if success.Type != Normal {
			return success
		}

		if !success.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError("Cannot build array because the length property is not writable."))
		}

		length++
	}

	return NewNormalCompletion(arrayObj)
}
