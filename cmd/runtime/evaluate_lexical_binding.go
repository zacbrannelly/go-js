package runtime

import (
	"zbrannelly.dev/go-js/cmd/analyzer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateLexicalBinding(runtime *Runtime, lexicalBinding *ast.LexicalBindingNode) *Completion {
	maybeIdentifier := lexicalBinding.GetTarget()

	if maybeIdentifier.GetNodeType() == ast.BindingIdentifier {
		identifier := maybeIdentifier.(*ast.BindingIdentifierNode)
		isStrictMode := analyzer.IsStrictMode(lexicalBinding)
		lhs := ResolveBindingFromCurrentContext(identifier.Identifier, runtime, isStrictMode)
		if lhs.Type != Normal {
			return lhs
		}

		referenceValue := lhs.Value.(*JavaScriptValue)
		if referenceValue.Type != TypeReference {
			panic("Assert failed: Expected a Reference record when resolving a lexical binding.")
		}

		reference := referenceValue.Value.(*Reference)
		if lexicalBinding.GetInitializer() == nil {
			// BindingIdentifier
			reference.InitializeReferencedBinding(NewUndefinedValue())
		} else {
			// BindingIdentifier Initializer

			// TODO: Check if anon function. If so do something different according to the spec.
			initializer := lexicalBinding.GetInitializer()

			rhsCompletion := Evaluate(runtime, initializer)
			if rhsCompletion.Type != Normal {
				return rhsCompletion
			}

			// If rhs is a reference, we need to get the value of the reference.
			rhsValue := GetValue(rhsCompletion.Value.(*JavaScriptValue))
			if rhsValue.Type != Normal {
				return rhsValue
			}

			reference.InitializeReferencedBinding(rhsValue.Value.(*JavaScriptValue))
		}

		return NewUnusedCompletion()
	}

	// TODO: Support BindingPattern.
	panic("Unexpected lexical binding format.")
}
