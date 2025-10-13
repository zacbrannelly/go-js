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
		return EvaluateStringOrNumericBinaryExpression(runtime, node.(*ast.AdditiveExpressionNode))
	case ast.MultiplicativeExpression:
		return EvaluateStringOrNumericBinaryExpression(runtime, node.(*ast.MultiplicativeExpressionNode))
	case ast.ExponentiationExpression:
		return EvaluateStringOrNumericBinaryExpression(runtime, node.(*ast.ExponentiationExpressionNode))
	case ast.ShiftExpression:
		return EvaluateStringOrNumericBinaryExpression(runtime, node.(*ast.ShiftExpressionNode))
	case ast.BitwiseANDExpression:
		return EvaluateStringOrNumericBinaryExpression(runtime, node.(*ast.BitwiseANDExpressionNode))
	case ast.BitwiseXORExpression:
		return EvaluateStringOrNumericBinaryExpression(runtime, node.(*ast.BitwiseXORExpressionNode))
	case ast.BitwiseORExpression:
		return EvaluateStringOrNumericBinaryExpression(runtime, node.(*ast.BitwiseORExpressionNode))
	case ast.NumericLiteral:
		return EvaluateNumericLiteral(runtime, node.(*ast.NumericLiteralNode))
	case ast.StringLiteral:
		return EvaluateStringLiteral(runtime, node.(*ast.StringLiteralNode))
	case ast.BooleanLiteral:
		return EvaluateBooleanLiteral(runtime, node.(*ast.BooleanLiteralNode))
	case ast.NullLiteral:
		return EvaluateNullLiteral(runtime, node.(*ast.BasicNode))
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
	case ast.AssignmentExpression:
		return EvaluateAssignmentExpression(runtime, node.(*ast.AssignmentExpressionNode))
	case ast.ConditionalExpression:
		return EvaluateConditionalExpression(runtime, node.(*ast.ConditionalExpressionNode))
	case ast.RelationalExpression:
		return EvaluateRelationalExpression(runtime, node.(*ast.RelationalExpressionNode))
	case ast.EqualityExpression:
		return EvaluateEqualityExpression(runtime, node.(*ast.EqualityExpressionNode))
	}

	panic(fmt.Sprintf("Assert failed: Evaluation of %s node not implemented.", ast.NodeTypeToString[node.GetNodeType()]))
}
