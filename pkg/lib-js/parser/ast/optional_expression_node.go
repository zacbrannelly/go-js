package ast

import (
	"fmt"
)

type OptionalExpressionNode struct {
	parent     Node
	expression Node
}

func NewOptionalExpressionNode(expression Node) *OptionalExpressionNode {
	newNode := &OptionalExpressionNode{}
	newNode.SetExpression(expression)
	return newNode
}

func (n *OptionalExpressionNode) GetNodeType() NodeType {
	return OptionalExpression
}

func (n *OptionalExpressionNode) GetParent() Node {
	return n.parent
}

func (n *OptionalExpressionNode) GetChildren() []Node {
	return []Node{n.expression}
}

func (n *OptionalExpressionNode) SetChildren(children []Node) {
	panic("OptionalExpressionNode does not support adding children")
}

func (n *OptionalExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *OptionalExpressionNode) GetExpression() Node {
	return n.expression
}

func (n *OptionalExpressionNode) SetExpression(expression Node) {
	if expression != nil {
		expression.SetParent(n)
	}
	n.expression = expression
}

func (n *OptionalExpressionNode) IsComposable() bool {
	return false
}

func (n *OptionalExpressionNode) ToString() string {
	return fmt.Sprintf("OptionalExpression(%s)", n.expression.ToString())
}
