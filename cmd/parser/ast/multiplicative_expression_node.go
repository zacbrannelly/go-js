package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type MultiplicativeExpressionNode struct {
	Parent   Node
	Children []Node
	Operator lexer.Token
	Left     Node
	Right    Node
}

func (n *MultiplicativeExpressionNode) GetNodeType() NodeType {
	return MultiplicativeExpression
}

func (n *MultiplicativeExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *MultiplicativeExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *MultiplicativeExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *MultiplicativeExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *MultiplicativeExpressionNode) ToString() string {
	return fmt.Sprintf("MultiplicativeExpression(%s %s %s)", n.Left.ToString(), n.Operator.Value, n.Right.ToString())
}

func (n *MultiplicativeExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *MultiplicativeExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *MultiplicativeExpressionNode) GetRight() Node {
	return n.Right
}

func (n *MultiplicativeExpressionNode) SetRight(right Node) {
	n.Right = right
}
