package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type BitwiseXORExpressionNode struct {
	Parent   Node
	Children []Node
	Left     Node
	Right    Node
}

func (n *BitwiseXORExpressionNode) GetNodeType() NodeType {
	return BitwiseXORExpression
}

func (n *BitwiseXORExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *BitwiseXORExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *BitwiseXORExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BitwiseXORExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BitwiseXORExpressionNode) ToString() string {
	return fmt.Sprintf("BitwiseXORExpression(%s ^ %s)", n.Left.ToString(), n.Right.ToString())
}

func (n *BitwiseXORExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *BitwiseXORExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *BitwiseXORExpressionNode) GetRight() Node {
	return n.Right
}

func (n *BitwiseXORExpressionNode) SetRight(right Node) {
	n.Right = right
}

func (n *BitwiseXORExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *BitwiseXORExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.BitwiseXor, Value: "^"}
}
