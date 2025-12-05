package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
)

type BitwiseANDExpressionNode struct {
	parent Node
	left   Node
	right  Node
}

func NewBitwiseANDExpressionNode() *BitwiseANDExpressionNode {
	return &BitwiseANDExpressionNode{}
}

func (n *BitwiseANDExpressionNode) GetNodeType() NodeType {
	return BitwiseANDExpression
}

func (n *BitwiseANDExpressionNode) GetParent() Node {
	return n.parent
}

func (n *BitwiseANDExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *BitwiseANDExpressionNode) SetChildren(children []Node) {
	panic("BitwiseANDExpressionNode does not support adding children")
}

func (n *BitwiseANDExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BitwiseANDExpressionNode) GetLeft() Node {
	return n.left
}

func (n *BitwiseANDExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *BitwiseANDExpressionNode) GetRight() Node {
	return n.right
}

func (n *BitwiseANDExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *BitwiseANDExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *BitwiseANDExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.BitwiseAnd, Value: "&"}
}

func (n *BitwiseANDExpressionNode) IsComposable() bool {
	return false
}

func (n *BitwiseANDExpressionNode) ToString() string {
	return fmt.Sprintf("BitwiseANDExpression(%s & %s)", n.left.ToString(), n.right.ToString())
}
