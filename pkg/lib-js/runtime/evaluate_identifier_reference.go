package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/analyzer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateIdentifierReference(runtime *Runtime, identifierReference *ast.IdentifierReferenceNode) *Completion {
	isStrictMode := analyzer.IsStrictMode(identifierReference)
	return ResolveBindingFromCurrentContext(identifierReference.Identifier, runtime, isStrictMode)
}
