package ast

import "fmt"

type BitwiseANDExpressionNode struct {
	Parent   Node
	Children []Node
	Left     Node
	Right    Node
}

func (n *BitwiseANDExpressionNode) GetNodeType() NodeType {
	return BitwiseANDExpression
}

func (n *BitwiseANDExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *BitwiseANDExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *BitwiseANDExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BitwiseANDExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BitwiseANDExpressionNode) ToString() string {
	return fmt.Sprintf("BitwiseANDExpression(%s & %s)", n.Left.ToString(), n.Right.ToString())
}

func (n *BitwiseANDExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *BitwiseANDExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *BitwiseANDExpressionNode) GetRight() Node {
	return n.Right
}

func (n *BitwiseANDExpressionNode) SetRight(right Node) {
	n.Right = right
}
