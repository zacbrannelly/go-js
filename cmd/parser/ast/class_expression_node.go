package ast

import (
	"fmt"
	"strings"
)

type ClassExpressionNode struct {
	parent   Node
	name     Node
	heritage Node
	elements []Node
}

func NewClassExpressionNode(name Node, heritage Node, elements []Node) *ClassExpressionNode {
	newNode := &ClassExpressionNode{}
	newNode.SetName(name)
	newNode.SetHeritage(heritage)
	newNode.SetElements(elements)
	return newNode
}

func (n *ClassExpressionNode) GetNodeType() NodeType {
	return ClassExpression
}

func (n *ClassExpressionNode) GetParent() Node {
	return n.parent
}

func (n *ClassExpressionNode) GetChildren() []Node {
	return nil
}

func (n *ClassExpressionNode) SetChildren(children []Node) {
	panic("ClassExpressionNode does not support adding children")
}

func (n *ClassExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ClassExpressionNode) GetName() Node {
	return n.name
}

func (n *ClassExpressionNode) SetName(name Node) {
	if name != nil {
		name.SetParent(n)
	}
	n.name = name
}

func (n *ClassExpressionNode) GetHeritage() Node {
	return n.heritage
}

func (n *ClassExpressionNode) SetHeritage(heritage Node) {
	if heritage != nil {
		heritage.SetParent(n)
	}
	n.heritage = heritage
}

func (n *ClassExpressionNode) GetElements() []Node {
	return n.elements
}

func (n *ClassExpressionNode) SetElements(elements []Node) {
	for _, element := range elements {
		element.SetParent(n)
	}
	n.elements = elements
}

func (n *ClassExpressionNode) ToString() string {
	name := ""
	if n.name != nil {
		name = n.name.ToString()
	}

	heritage := ""
	if n.heritage != nil {
		heritage = " extends " + n.heritage.ToString()
	}

	elements := []string{}
	for _, element := range n.elements {
		elements = append(elements, element.ToString())
	}

	return fmt.Sprintf("ClassExpression(class %s%s { %s })", name, heritage, strings.Join(elements, ", "))
}
