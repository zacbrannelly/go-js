package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type CoalesceExpressionNode struct {
	parent Node
	left   Node
	right  Node
}

func NewCoalesceExpressionNode() *CoalesceExpressionNode {
	return &CoalesceExpressionNode{}
}

func (n *CoalesceExpressionNode) GetNodeType() NodeType {
	return CoalesceExpression
}

func (n *CoalesceExpressionNode) GetParent() Node {
	return n.parent
}

func (n *CoalesceExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *CoalesceExpressionNode) SetChildren(children []Node) {
	panic("CoalesceExpressionNode does not support adding children")
}

func (n *CoalesceExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *CoalesceExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *CoalesceExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *CoalesceExpressionNode) GetLeft() Node {
	return n.left
}

func (n *CoalesceExpressionNode) GetRight() Node {
	return n.right
}

func (n *CoalesceExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *CoalesceExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.NullishCoalescing, Value: "??"}
}

func (n *CoalesceExpressionNode) IsComposable() bool {
	return false
}

func (n *CoalesceExpressionNode) ToString() string {
	return fmt.Sprintf("CoalesceExpression(%s ?? %s)", n.left.ToString(), n.right.ToString())
}
