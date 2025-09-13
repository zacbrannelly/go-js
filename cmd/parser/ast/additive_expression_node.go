package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type AdditiveExpressionNode struct {
	Parent   Node
	Children []Node
	Operator lexer.Token
	Left     Node
	Right    Node
}

func (n *AdditiveExpressionNode) GetNodeType() NodeType {
	return AdditiveExpression
}

func (n *AdditiveExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *AdditiveExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *AdditiveExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *AdditiveExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *AdditiveExpressionNode) ToString() string {
	return fmt.Sprintf("AdditiveExpression(%s %s %s)", n.Left.ToString(), n.Operator.Value, n.Right.ToString())
}

func (n *AdditiveExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *AdditiveExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *AdditiveExpressionNode) GetRight() Node {
	return n.Right
}

func (n *AdditiveExpressionNode) SetRight(right Node) {
	n.Right = right
}

func (n *AdditiveExpressionNode) SetOperator(operator lexer.Token) {
	n.Operator = operator
}

func (n *AdditiveExpressionNode) GetOperator() lexer.Token {
	return n.Operator
}
