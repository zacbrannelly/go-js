package ast

import (
	"fmt"
)

type ExponentiationExpressionNode struct {
	parent Node
	left   Node
	right  Node
}

func NewExponentiationExpressionNode() *ExponentiationExpressionNode {
	return &ExponentiationExpressionNode{}
}

func (n *ExponentiationExpressionNode) GetNodeType() NodeType {
	return ExponentiationExpression
}

func (n *ExponentiationExpressionNode) GetParent() Node {
	return n.parent
}

func (n *ExponentiationExpressionNode) GetChildren() []Node {
	return nil
}

func (n *ExponentiationExpressionNode) SetChildren(children []Node) {
	panic("ExponentiationExpressionNode does not support adding children")
}

func (n *ExponentiationExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ExponentiationExpressionNode) ToString() string {
	return fmt.Sprintf("ExponentiationExpression(%s ** %s)", n.left.ToString(), n.right.ToString())
}

func (n *ExponentiationExpressionNode) GetLeft() Node {
	return n.left
}

func (n *ExponentiationExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *ExponentiationExpressionNode) GetRight() Node {
	return n.right
}

func (n *ExponentiationExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}
