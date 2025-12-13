package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func CompileScript(runtime *Runtime, script *ast.ScriptNode) []Instruction {
	if len(script.GetChildren()) == 0 {
		return []Instruction{}
	}

	statementList := script.GetChildren()[0]
	if statementList.GetNodeType() != ast.StatementList {
		panic("Assert failed: Statement list expected in script body.")
	}
	return Compile(runtime, statementList)
}

func EvaluateScript(runtime *Runtime, script *ast.ScriptNode) *Completion {
	if len(script.GetChildren()) == 0 {
		return NewUnusedCompletion()
	}

	statementList := script.GetChildren()[0]
	if statementList.GetNodeType() != ast.StatementList {
		panic("Assert failed: Statement list expected in script body.")
	}

	return Evaluate(runtime, statementList)
}
