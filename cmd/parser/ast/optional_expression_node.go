package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type OptionalExpressionNode struct {
	Parent   Node
	Children []Node
	Operator lexer.Token
	Left     Node
	Right    Node
}

func (n *OptionalExpressionNode) GetNodeType() NodeType {
	return OptionalExpression
}

func (n *OptionalExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *OptionalExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *OptionalExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *OptionalExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *OptionalExpressionNode) ToString() string {
	return fmt.Sprintf("OptionalExpression(%s %s %s)", n.Left.ToString(), n.Operator.Value, n.Right.ToString())
}

func (n *OptionalExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *OptionalExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *OptionalExpressionNode) GetRight() Node {
	return n.Right
}

func (n *OptionalExpressionNode) SetRight(right Node) {
	n.Right = right
}
