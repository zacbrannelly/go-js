package ast

import (
	"fmt"
)

type NewExpressionNode struct {
	Parent      Node
	Children    []Node
	Constructor Node
}

func (n *NewExpressionNode) GetNodeType() NodeType {
	return NewExpression
}

func (n *NewExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *NewExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *NewExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *NewExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *NewExpressionNode) ToString() string {
	return fmt.Sprintf("NewExpression(%s)", n.Constructor.ToString())
}
