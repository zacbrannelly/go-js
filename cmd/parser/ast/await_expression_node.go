package ast

import (
	"fmt"
)

type AwaitExpressionNode struct {
	Parent     Node
	Children   []Node
	Expression Node
}

func (n *AwaitExpressionNode) GetNodeType() NodeType {
	return AwaitExpression
}

func (n *AwaitExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *AwaitExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *AwaitExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *AwaitExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *AwaitExpressionNode) ToString() string {
	return fmt.Sprintf("AwaitExpression(%s)", n.Expression.ToString())
}
