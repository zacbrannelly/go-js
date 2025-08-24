package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type EqualityExpressionNode struct {
	Parent   Node
	Children []Node
	Operator lexer.Token
	Left     Node
	Right    Node
}

func (n *EqualityExpressionNode) GetNodeType() NodeType {
	return EqualityExpression
}

func (n *EqualityExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *EqualityExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *EqualityExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *EqualityExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *EqualityExpressionNode) ToString() string {
	return fmt.Sprintf("EqualityExpression(%s %s %s)", n.Left.ToString(), n.Operator.Value, n.Right.ToString())
}

func (n *EqualityExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *EqualityExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *EqualityExpressionNode) GetRight() Node {
	return n.Right
}

func (n *EqualityExpressionNode) SetRight(right Node) {
	n.Right = right
}
