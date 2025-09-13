package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type BitwiseORExpressionNode struct {
	Parent   Node
	Children []Node
	Left     Node
	Right    Node
}

func (n *BitwiseORExpressionNode) GetNodeType() NodeType {
	return BitwiseORExpression
}

func (n *BitwiseORExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *BitwiseORExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *BitwiseORExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BitwiseORExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BitwiseORExpressionNode) ToString() string {
	return fmt.Sprintf("BitwiseORExpression(%s | %s)", n.Left.ToString(), n.Right.ToString())
}

func (n *BitwiseORExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *BitwiseORExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *BitwiseORExpressionNode) GetRight() Node {
	return n.Right
}

func (n *BitwiseORExpressionNode) SetRight(right Node) {
	n.Right = right
}

func (n *BitwiseORExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *BitwiseORExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.BitwiseOr, Value: "|"}
}
