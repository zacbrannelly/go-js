package runtime

import (
	"zbrannelly.dev/go-js/cmd/analyzer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateIdentifierReference(runtime *Runtime, identifierReference *ast.IdentifierReferenceNode) *Completion {
	isStrictMode := analyzer.IsStrictMode(identifierReference)
	return ResolveBindingFromCurrentContext(identifierReference.Identifier, runtime, isStrictMode)
}
