package ast

import (
	"fmt"
	"strings"
)

type ClassExpressionNode struct {
	Parent   Node
	Children []Node
	Name     Node
	Heritage Node
	Elements []Node
}

func (n *ClassExpressionNode) GetNodeType() NodeType {
	return ClassExpression
}

func (n *ClassExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *ClassExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *ClassExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ClassExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ClassExpressionNode) ToString() string {
	name := ""
	if n.Name != nil {
		name = n.Name.ToString()
	}

	heritage := ""
	if n.Heritage != nil {
		heritage = " extends " + n.Heritage.ToString()
	}

	elements := []string{}
	for _, element := range n.Elements {
		elements = append(elements, element.ToString())
	}

	return fmt.Sprintf("ClassExpression(class %s%s { %s })", name, heritage, strings.Join(elements, ", "))
}
