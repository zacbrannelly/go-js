package ast

import "fmt"

type ExponentiationExpressionNode struct {
	Parent   Node
	Children []Node
	Left     Node
	Right    Node
}

func (n *ExponentiationExpressionNode) GetNodeType() NodeType {
	return ExponentiationExpression
}

func (n *ExponentiationExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *ExponentiationExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *ExponentiationExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ExponentiationExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ExponentiationExpressionNode) ToString() string {
	return fmt.Sprintf("ExponentiationExpression(%s ** %s)", n.Left.ToString(), n.Right.ToString())
}

func (n *ExponentiationExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *ExponentiationExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *ExponentiationExpressionNode) GetRight() Node {
	return n.Right
}

func (n *ExponentiationExpressionNode) SetRight(right Node) {
	n.Right = right
}
