package runtime

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
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
	case ast.VariableDeclarationList:
		return EvaluateVariableDeclarationList(runtime, node.(*ast.BasicNode))
	case ast.AssignmentExpression:
		return EvaluateAssignmentExpression(runtime, node.(*ast.AssignmentExpressionNode))
	case ast.ConditionalExpression:
		return EvaluateConditionalExpression(runtime, node.(*ast.ConditionalExpressionNode))
	case ast.RelationalExpression:
		return EvaluateRelationalExpression(runtime, node.(*ast.RelationalExpressionNode))
	case ast.EqualityExpression:
		return EvaluateEqualityExpression(runtime, node.(*ast.EqualityExpressionNode))
	case ast.LogicalANDExpression:
		return EvaluateLogicalANDExpression(runtime, node.(*ast.LogicalANDExpressionNode))
	case ast.LogicalORExpression:
		return EvaluateLogicalORExpression(runtime, node.(*ast.LogicalORExpressionNode))
	case ast.CoalesceExpression:
		return EvaluateCoalesceExpression(runtime, node.(*ast.CoalesceExpressionNode))
	case ast.IfStatement:
		return EvaluateIfStatement(runtime, node.(*ast.IfStatementNode))
	case ast.DoWhileStatement:
		return EvaluateDoWhileStatement(runtime, node.(*ast.DoWhileStatementNode))
	case ast.WhileStatement:
		return EvaluateWhileStatement(runtime, node.(*ast.WhileStatementNode))
	case ast.ForStatement:
		return EvaluateForStatement(runtime, node.(*ast.ForStatementNode))
	case ast.Block:
		return EvaluateBlockStatement(runtime, node.(*ast.BasicNode))
	case ast.EmptyStatement:
		// TODO: In the spec this is EMPTY, unsure if this matters.
		return NewUnusedCompletion()
	case ast.UpdateExpression:
		return EvaluateUpdateExpression(runtime, node.(*ast.UpdateExpressionNode))
	case ast.UnaryExpression:
		return EvaluateUnaryExpression(runtime, node.(*ast.UnaryExpressionNode))
	case ast.Expression:
		return EvaluateExpression(runtime, node.(*ast.ExpressionNode))
	case ast.ContinueStatement:
		return EvaluateContinueStatement(runtime, node.(*ast.ContinueStatementNode))
	case ast.BreakStatement:
		return EvaluateBreakStatement(runtime, node.(*ast.BreakStatementNode))
	case ast.ArrayLiteral:
		return EvaluateArrayLiteral(runtime, node.(*ast.BasicNode))
	case ast.MemberExpression:
		return EvaluateMemberExpression(runtime, node.(*ast.MemberExpressionNode))
	case ast.FunctionExpression:
		return EvaluateFunctionExpression(runtime, node.(*ast.FunctionExpressionNode))
	case ast.CallExpression:
		return EvaluateCallExpression(runtime, node.(*ast.CallExpressionNode))
	case ast.ReturnStatement:
		return EvaluateReturnStatement(runtime, node.(*ast.ReturnStatementNode))
	case ast.ObjectLiteral:
		return EvaluateObjectLiteral(runtime, node.(*ast.ObjectLiteralNode))
	case ast.NewExpression:
		return EvaluateNewExpression(runtime, node.(*ast.NewExpressionNode))
	case ast.ThrowStatement:
		return EvaluateThrowStatement(runtime, node.(*ast.ThrowStatementNode))
	case ast.ThisExpression:
		return EvaluateThisExpression(runtime, node.(*ast.BasicNode))
	case ast.TryStatement:
		return EvaluateTryStatement(runtime, node.(*ast.TryStatementNode))
	case ast.SwitchStatement:
		return EvaluateSwitchStatement(runtime, node.(*ast.SwitchStatementNode))
	case ast.CoverParenthesizedExpressionAndArrowParameterList:
		return EvaluateCoverParenthesizedExpressionAndArrowParameterList(runtime, node.(*ast.BasicNode))
	}

	panic(fmt.Sprintf("Assert failed: Evaluation of %s node not implemented.", ast.NodeTypeToString[node.GetNodeType()]))
}
