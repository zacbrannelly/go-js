package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateLexicalDeclaration(runtime *Runtime, lexicalDeclaration *ast.BasicNode) *Completion {
	for _, child := range lexicalDeclaration.GetChildren() {
		completion := EvaluateLexicalBinding(runtime, child.(*ast.LexicalBindingNode))
		if completion.Type != Normal {
			return completion
		}
	}

	return NewUnusedCompletion()
}
