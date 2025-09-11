package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type RelationalExpressionNode struct {
	Parent   Node
	Children []Node
	Operator lexer.Token
	Left     Node
	Right    Node
}

func (n *RelationalExpressionNode) GetNodeType() NodeType {
	return RelationalExpression
}

func (n *RelationalExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *RelationalExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *RelationalExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *RelationalExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *RelationalExpressionNode) ToString() string {
	return fmt.Sprintf("RelationalExpression(%s %s %s)", n.Left.ToString(), n.Operator.Value, n.Right.ToString())
}

func (n *RelationalExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *RelationalExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *RelationalExpressionNode) GetRight() Node {
	return n.Right
}

func (n *RelationalExpressionNode) SetRight(right Node) {
	n.Right = right
}
