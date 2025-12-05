package ast

import (
	"fmt"
)

type ExpressionNode struct {
	parent Node
	left   Node
	right  Node
}

func NewExpressionNodeEmpty() *ExpressionNode {
	return &ExpressionNode{}
}

func (n *ExpressionNode) GetNodeType() NodeType {
	return Expression
}

func (n *ExpressionNode) GetParent() Node {
	return n.parent
}

func (n *ExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *ExpressionNode) SetChildren(children []Node) {
	panic("ExpressionNode does not support adding children")
}

func (n *ExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ExpressionNode) GetLeft() Node {
	return n.left
}

func (n *ExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *ExpressionNode) GetRight() Node {
	return n.right
}

func (n *ExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *ExpressionNode) IsComposable() bool {
	return false
}

func (n *ExpressionNode) ToString() string {
	return fmt.Sprintf("Expression(%s, %s)", n.left.ToString(), n.right.ToString())
}
