package ast

import "fmt"

type ConditionalExpressionNode struct {
	Parent    Node
	Children  []Node
	Condition Node
	TrueExpr  Node
	FalseExpr Node
}

func (n *ConditionalExpressionNode) GetNodeType() NodeType {
	return ConditionalExpression
}

func (n *ConditionalExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *ConditionalExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *ConditionalExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ConditionalExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ConditionalExpressionNode) ToString() string {
	return fmt.Sprintf("ConditionalExpression(%s, %s, %s)", n.Condition.ToString(), n.TrueExpr.ToString(), n.FalseExpr.ToString())
}
