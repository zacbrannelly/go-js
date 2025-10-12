package runtime

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func Evaluate(runtime *Runtime, node ast.Node) *Completion {
	if node == nil {
		panic("Assert failed: Node is nil.")
	}

	switch node.GetNodeType() {
	case ast.Script:
		return EvaluateScript(runtime, node.(*ast.ScriptNode))
	case ast.StatementList:
		return EvaluateStatementList(runtime, node.(*ast.StatementListNode))
	case ast.AdditiveExpression:
		return EvaluateAdditiveExpression(runtime, node.(*ast.AdditiveExpressionNode))
	case ast.NumericLiteral:
		return EvaluateNumericLiteral(runtime, node.(*ast.NumericLiteralNode))
	case ast.StringLiteral:
		return EvaluateStringLiteral(runtime, node.(*ast.StringLiteralNode))
	case ast.IdentifierReference:
		return EvaluateIdentifierReference(runtime, node.(*ast.IdentifierReferenceNode))
	case ast.LexicalDeclaration:
		return EvaluateLexicalDeclaration(runtime, node.(*ast.BasicNode))
	case ast.LexicalBinding:
		return EvaluateLexicalBinding(runtime, node.(*ast.LexicalBindingNode))
	case ast.Initializer:
		return EvaluateInitializer(runtime, node.(*ast.BasicNode))
	case ast.VariableStatement:
		return EvaluateVariableStatement(runtime, node.(*ast.BasicNode))
	}

	panic(fmt.Sprintf("Assert failed: Evaluation of %s node not implemented.", ast.NodeTypeToString[node.GetNodeType()]))
}
