package ast

import (
	"fmt"
)

type AwaitExpressionNode struct {
	parent     Node
	expression Node
}

func NewAwaitExpressionNode(expression Node) *AwaitExpressionNode {
	newNode := &AwaitExpressionNode{}
	newNode.SetExpression(expression)
	return newNode
}

func (n *AwaitExpressionNode) GetNodeType() NodeType {
	return AwaitExpression
}

func (n *AwaitExpressionNode) GetParent() Node {
	return n.parent
}

func (n *AwaitExpressionNode) GetChildren() []Node {
	return []Node{n.expression}
}

func (n *AwaitExpressionNode) SetChildren(children []Node) {
	panic("AwaitExpressionNode does not support adding children")
}

func (n *AwaitExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *AwaitExpressionNode) GetExpression() Node {
	return n.expression
}

func (n *AwaitExpressionNode) SetExpression(expression Node) {
	if expression != nil {
		expression.SetParent(n)
	}
	n.expression = expression
}

func (n *AwaitExpressionNode) IsComposable() bool {
	return false
}

func (n *AwaitExpressionNode) ToString() string {
	return fmt.Sprintf("AwaitExpression(%s)", n.expression.ToString())
}
