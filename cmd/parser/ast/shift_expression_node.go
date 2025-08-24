package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type ShiftExpressionNode struct {
	Parent   Node
	Children []Node
	Operator lexer.Token
	Left     Node
	Right    Node
}

func (n *ShiftExpressionNode) GetNodeType() NodeType {
	return ShiftExpression
}

func (n *ShiftExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *ShiftExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *ShiftExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ShiftExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ShiftExpressionNode) ToString() string {
	return fmt.Sprintf("ShiftExpression(%s %s %s)", n.Left.ToString(), n.Operator.Value, n.Right.ToString())
}

func (n *ShiftExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *ShiftExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *ShiftExpressionNode) GetRight() Node {
	return n.Right
}

func (n *ShiftExpressionNode) SetRight(right Node) {
	n.Right = right
}
