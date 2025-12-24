package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateBlockStatement(runtime *Runtime, blockStatement *ast.BasicNode) *Completion {
	if len(blockStatement.GetChildren()) == 0 {
		// TODO: In the spec this is EMPTY, unsure if this matters.
		return NewUnusedCompletion()
	}

	runningContext := runtime.GetRunningExecutionContext()
	oldEnv := runningContext.LexicalEnvironment

	blockEnv := NewDeclarativeEnvironment(oldEnv)
	BlockDeclarationInstantiation(runtime, blockStatement, blockEnv)

	runningContext.LexicalEnvironment = blockEnv
	completion := EvaluateStatementList(runtime, blockStatement.GetChildren()[0].(*ast.StatementListNode))
	runningContext.LexicalEnvironment = oldEnv

	return completion
}

func BlockDeclarationInstantiation(runtime *Runtime, blockStatement ast.Node, env *DeclarativeEnvironment) *Completion {
	declarations := LexicallyScopedDeclarations(blockStatement)

	for _, declaration := range declarations {
		boundNames := BoundNames(declaration)
		isConst := false

		if declaration.GetNodeType() == ast.LexicalDeclaration {
			isConst = declaration.GetChildren()[0].(*ast.LexicalBindingNode).Const
		}

		for _, name := range boundNames {
			if isConst {
				env.CreateImmutableBinding(runtime, name, true)
			} else {
				env.CreateMutableBinding(runtime, name, false)
			}
		}

		if declaration.GetNodeType() == ast.FunctionExpression {
			// TODO: Handle function declaration instantiation.
			panic("TODO: Handle function declaration instantiation.")
		}
	}

	return NewUnusedCompletion()
}
