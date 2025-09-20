package ast

import (
	"fmt"
	"strings"
)

type ObjectBindingPatternNode struct {
	parent     Node
	properties []Node
}

func NewObjectBindingPatternNode(properties []Node) *ObjectBindingPatternNode {
	newNode := &ObjectBindingPatternNode{}
	newNode.SetProperties(properties)
	return newNode
}

func (n *ObjectBindingPatternNode) GetNodeType() NodeType {
	return ObjectBindingPattern
}

func (n *ObjectBindingPatternNode) GetParent() Node {
	return n.parent
}

func (n *ObjectBindingPatternNode) GetChildren() []Node {
	return nil
}

func (n *ObjectBindingPatternNode) SetChildren(children []Node) {
	panic("ObjectBindingPatternNode does not support adding children")
}

func (n *ObjectBindingPatternNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ObjectBindingPatternNode) GetProperties() []Node {
	return n.properties
}

func (n *ObjectBindingPatternNode) SetProperties(properties []Node) {
	for _, property := range properties {
		property.SetParent(n)
	}
	n.properties = properties
}

func (n *ObjectBindingPatternNode) ToString() string {
	var properties []string
	for _, property := range n.properties {
		properties = append(properties, property.ToString())
	}

	return fmt.Sprintf("ObjectBindingPattern(%s)", strings.Join(properties, ", "))
}

type ArrayBindingPatternNode struct {
	parent   Node
	elements []Node
}

func NewArrayBindingPatternNode(elements []Node) *ArrayBindingPatternNode {
	newNode := &ArrayBindingPatternNode{}
	newNode.SetElements(elements)
	return newNode
}

func (n *ArrayBindingPatternNode) GetNodeType() NodeType {
	return ArrayBindingPattern
}

func (n *ArrayBindingPatternNode) GetParent() Node {
	return n.parent
}

func (n *ArrayBindingPatternNode) GetChildren() []Node {
	return nil
}

func (n *ArrayBindingPatternNode) SetChildren(children []Node) {
	panic("ArrayBindingPatternNode does not support adding children")
}

func (n *ArrayBindingPatternNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ArrayBindingPatternNode) GetElements() []Node {
	return n.elements
}

func (n *ArrayBindingPatternNode) SetElements(elements []Node) {
	for _, element := range elements {
		element.SetParent(n)
	}
	n.elements = elements
}

func (n *ArrayBindingPatternNode) ToString() string {
	var elements []string
	for _, element := range n.elements {
		elements = append(elements, element.ToString())
	}

	return fmt.Sprintf("ArrayBindingPattern(%s)", strings.Join(elements, ", "))
}
