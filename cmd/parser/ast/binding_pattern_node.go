package ast

import (
	"fmt"
	"strings"
)

type ObjectBindingPatternNode struct {
	Parent     Node
	Children   []Node
	Properties []Node
}

func (n *ObjectBindingPatternNode) GetNodeType() NodeType {
	return ObjectBindingPattern
}

func (n *ObjectBindingPatternNode) GetParent() Node {
	return n.Parent
}

func (n *ObjectBindingPatternNode) GetChildren() []Node {
	return n.Children
}

func (n *ObjectBindingPatternNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ObjectBindingPatternNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ObjectBindingPatternNode) ToString() string {
	var properties []string
	for _, property := range n.Properties {
		properties = append(properties, property.ToString())
	}

	return fmt.Sprintf("ObjectBindingPattern(%s)", strings.Join(properties, ", "))
}

type ArrayBindingPatternNode struct {
	Parent   Node
	Children []Node
	Elements []Node
}

func (n *ArrayBindingPatternNode) GetNodeType() NodeType {
	return ArrayBindingPattern
}

func (n *ArrayBindingPatternNode) GetParent() Node {
	return n.Parent
}

func (n *ArrayBindingPatternNode) GetChildren() []Node {
	return n.Children
}

func (n *ArrayBindingPatternNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ArrayBindingPatternNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ArrayBindingPatternNode) ToString() string {
	var elements []string
	for _, element := range n.Elements {
		elements = append(elements, element.ToString())
	}

	return fmt.Sprintf("ArrayBindingPattern(%s)", strings.Join(elements, ", "))
}
