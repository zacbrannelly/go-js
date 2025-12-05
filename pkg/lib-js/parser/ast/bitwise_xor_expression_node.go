package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
)

type BitwiseXORExpressionNode struct {
	parent Node
	left   Node
	right  Node
}

func NewBitwiseXORExpressionNode() *BitwiseXORExpressionNode {
	return &BitwiseXORExpressionNode{}
}

func (n *BitwiseXORExpressionNode) GetNodeType() NodeType {
	return BitwiseXORExpression
}

func (n *BitwiseXORExpressionNode) GetParent() Node {
	return n.parent
}

func (n *BitwiseXORExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *BitwiseXORExpressionNode) SetChildren(children []Node) {
	panic("BitwiseXORExpressionNode does not support adding children")
}

func (n *BitwiseXORExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BitwiseXORExpressionNode) GetLeft() Node {
	return n.left
}

func (n *BitwiseXORExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *BitwiseXORExpressionNode) GetRight() Node {
	return n.right
}

func (n *BitwiseXORExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *BitwiseXORExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *BitwiseXORExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.BitwiseXor, Value: "^"}
}

func (n *BitwiseXORExpressionNode) IsComposable() bool {
	return false
}

func (n *BitwiseXORExpressionNode) ToString() string {
	return fmt.Sprintf("BitwiseXORExpression(%s ^ %s)", n.left.ToString(), n.right.ToString())
}
