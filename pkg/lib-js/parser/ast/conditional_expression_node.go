package ast

import "fmt"

type ConditionalExpressionNode struct {
	parent    Node
	condition Node
	trueExpr  Node
	falseExpr Node
}

func NewConditionalExpressionNode() *ConditionalExpressionNode {
	return &ConditionalExpressionNode{}
}

func (n *ConditionalExpressionNode) GetNodeType() NodeType {
	return ConditionalExpression
}

func (n *ConditionalExpressionNode) GetParent() Node {
	return n.parent
}

func (n *ConditionalExpressionNode) GetChildren() []Node {
	return []Node{n.condition, n.trueExpr, n.falseExpr}
}

func (n *ConditionalExpressionNode) SetChildren(children []Node) {
	panic("ConditionalExpressionNode does not support adding children")
}

func (n *ConditionalExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ConditionalExpressionNode) GetCondition() Node {
	return n.condition
}

func (n *ConditionalExpressionNode) SetCondition(condition Node) {
	if condition != nil {
		condition.SetParent(n)
	}
	n.condition = condition
}

func (n *ConditionalExpressionNode) GetTrueExpr() Node {
	return n.trueExpr
}

func (n *ConditionalExpressionNode) SetTrueExpr(trueExpr Node) {
	if trueExpr != nil {
		trueExpr.SetParent(n)
	}
	n.trueExpr = trueExpr
}

func (n *ConditionalExpressionNode) GetFalseExpr() Node {
	return n.falseExpr
}

func (n *ConditionalExpressionNode) SetFalseExpr(falseExpr Node) {
	if falseExpr != nil {
		falseExpr.SetParent(n)
	}
	n.falseExpr = falseExpr
}

func (n *ConditionalExpressionNode) IsComposable() bool {
	return false
}

func (n *ConditionalExpressionNode) ToString() string {
	return fmt.Sprintf("ConditionalExpression(%s, %s, %s)", n.condition.ToString(), n.trueExpr.ToString(), n.falseExpr.ToString())
}
