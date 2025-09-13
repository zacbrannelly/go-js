package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type CoalesceExpressionNode struct {
	Parent   Node
	Children []Node
	Left     Node
	Right    Node
}

func (n *CoalesceExpressionNode) GetNodeType() NodeType {
	return CoalesceExpression
}

func (n *CoalesceExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *CoalesceExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *CoalesceExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *CoalesceExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *CoalesceExpressionNode) ToString() string {
	return fmt.Sprintf("CoalesceExpression(%s ?? %s)", n.Left.ToString(), n.Right.ToString())
}

func (n *CoalesceExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *CoalesceExpressionNode) SetRight(right Node) {
	n.Right = right
}

func (n *CoalesceExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *CoalesceExpressionNode) GetRight() Node {
	return n.Right
}

func (n *CoalesceExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *CoalesceExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.NullishCoalescing, Value: "??"}
}
