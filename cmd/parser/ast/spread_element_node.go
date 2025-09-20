package ast

import "fmt"

type SpreadElementNode struct {
	parent     Node
	expression Node
}

func NewSpreadElementNode(expression Node) *SpreadElementNode {
	newNode := &SpreadElementNode{}
	newNode.SetExpression(expression)
	return newNode
}

func (n *SpreadElementNode) GetNodeType() NodeType {
	return SpreadElement
}

func (n *SpreadElementNode) GetParent() Node {
	return n.parent
}

func (n *SpreadElementNode) GetChildren() []Node {
	return nil
}

func (n *SpreadElementNode) SetChildren(children []Node) {
	panic("SpreadElementNode does not support adding children")
}

func (n *SpreadElementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *SpreadElementNode) GetExpression() Node {
	return n.expression
}

func (n *SpreadElementNode) SetExpression(expression Node) {
	if expression != nil {
		expression.SetParent(n)
	}
	n.expression = expression
}

func (n *SpreadElementNode) ToString() string {
	return fmt.Sprintf("SpreadElement(%s)", n.expression.ToString())
}
