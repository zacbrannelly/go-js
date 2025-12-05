package ast

import (
	"fmt"
)

type NewExpressionNode struct {
	parent      Node
	constructor Node
}

func NewNewExpressionNode(constructor Node) *NewExpressionNode {
	newNode := &NewExpressionNode{}
	newNode.SetConstructor(constructor)
	return newNode
}

func (n *NewExpressionNode) GetNodeType() NodeType {
	return NewExpression
}

func (n *NewExpressionNode) GetParent() Node {
	return n.parent
}

func (n *NewExpressionNode) GetChildren() []Node {
	return []Node{n.constructor}
}

func (n *NewExpressionNode) SetChildren(children []Node) {
	panic("NewExpressionNode does not support adding children")
}

func (n *NewExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *NewExpressionNode) GetConstructor() Node {
	return n.constructor
}

func (n *NewExpressionNode) SetConstructor(constructor Node) {
	if constructor != nil {
		constructor.SetParent(n)
	}
	n.constructor = constructor
}

func (n *NewExpressionNode) IsComposable() bool {
	return false
}

func (n *NewExpressionNode) ToString() string {
	return fmt.Sprintf("NewExpression(%s)", n.constructor.ToString())
}
