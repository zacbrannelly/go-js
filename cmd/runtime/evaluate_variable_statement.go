package runtime

import (
	"zbrannelly.dev/go-js/cmd/analyzer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateVariableStatement(runtime *Runtime, variableDeclaration *ast.BasicNode) *Completion {
	return EvaluateVariableDeclarationList(runtime, variableDeclaration.GetChildren()[0].(*ast.BasicNode))
}

func EvaluateVariableDeclarationList(runtime *Runtime, variableDeclarationList *ast.BasicNode) *Completion {
	for _, variableDeclaration := range variableDeclarationList.GetChildren() {
		completion := EvaluateVariableDeclaration(runtime, variableDeclaration.(*ast.BasicNode))
		if completion.Type != Normal {
			return completion
		}
	}
	return NewUnusedCompletion()
}

func EvaluateVariableDeclaration(runtime *Runtime, variableDeclaration *ast.BasicNode) *Completion {
	if len(variableDeclaration.GetChildren()) == 0 {
		panic("Assert failed: Expected a variable declaration with at least one child.")
	}

	if len(variableDeclaration.GetChildren()) == 1 {
		// BindingIdentifier
		// NOTE: These are already initialized to undefined during GlobalDeclarationInstantiation.
		return NewUnusedCompletion()
	}

	target := variableDeclaration.GetChildren()[0]
	initializer := variableDeclaration.GetChildren()[1]

	if target.GetNodeType() == ast.BindingIdentifier {
		identifier := target.(*ast.BindingIdentifierNode)
		isStrictMode := analyzer.IsStrictMode(variableDeclaration)
		lhs := ResolveBindingFromCurrentContext(identifier.Identifier, runtime, isStrictMode)
		if lhs.Type != Normal {
			return lhs
		}

		reference := lhs.Value.(*JavaScriptValue)

		// TODO: Check if anon function. If so do something different according to the spec.

		rhsCompletion := Evaluate(runtime, initializer)
		if rhsCompletion.Type != Normal {
			return rhsCompletion
		}

		rhsValue := GetValue(rhsCompletion.Value.(*JavaScriptValue))
		if rhsValue.Type != Normal {
			return rhsValue
		}

		putCompletion := PutValue(runtime, reference, rhsValue.Value.(*JavaScriptValue))
		if putCompletion.Type != Normal {
			return putCompletion
		}

		// TODO: Return EMPTY in the spec, unsure if this matters.
		return NewUnusedCompletion()
	}

	// TODO: Support BindingPattern.
	panic("Unexpected variable declaration format.")
}
