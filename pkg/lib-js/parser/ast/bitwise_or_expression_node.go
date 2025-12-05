package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
)

type BitwiseORExpressionNode struct {
	parent Node
	left   Node
	right  Node
}

func NewBitwiseORExpressionNode() *BitwiseORExpressionNode {
	return &BitwiseORExpressionNode{}
}

func (n *BitwiseORExpressionNode) GetNodeType() NodeType {
	return BitwiseORExpression
}

func (n *BitwiseORExpressionNode) GetParent() Node {
	return n.parent
}

func (n *BitwiseORExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *BitwiseORExpressionNode) SetChildren(children []Node) {
	panic("BitwiseORExpressionNode does not support adding children")
}

func (n *BitwiseORExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BitwiseORExpressionNode) GetLeft() Node {
	return n.left
}

func (n *BitwiseORExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *BitwiseORExpressionNode) GetRight() Node {
	return n.right
}

func (n *BitwiseORExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *BitwiseORExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *BitwiseORExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.BitwiseOr, Value: "|"}
}

func (n *BitwiseORExpressionNode) IsComposable() bool {
	return false
}

func (n *BitwiseORExpressionNode) ToString() string {
	return fmt.Sprintf("BitwiseORExpression(%s | %s)", n.left.ToString(), n.right.ToString())
}
