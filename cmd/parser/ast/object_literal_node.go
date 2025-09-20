package ast

import (
	"fmt"
	"strings"
)

type ObjectLiteralNode struct {
	parent     Node
	properties []Node
}

func NewObjectLiteralNode(properties []Node) *ObjectLiteralNode {
	newNode := &ObjectLiteralNode{}
	newNode.SetProperties(properties)
	return newNode
}

func (n *ObjectLiteralNode) GetNodeType() NodeType {
	return ObjectLiteral
}

func (n *ObjectLiteralNode) GetChildren() []Node {
	return nil
}

func (n *ObjectLiteralNode) SetChildren(children []Node) {
	panic("ObjectLiteralNode does not support adding children")
}

func (n *ObjectLiteralNode) GetParent() Node {
	return n.parent
}

func (n *ObjectLiteralNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ObjectLiteralNode) GetProperties() []Node {
	return n.properties
}

func (n *ObjectLiteralNode) SetProperties(properties []Node) {
	for _, property := range properties {
		property.SetParent(n)
	}
	n.properties = properties
}

func (n *ObjectLiteralNode) ToString() string {
	properties := []string{}

	for _, property := range n.properties {
		properties = append(properties, property.ToString())
	}

	return fmt.Sprintf("ObjectLiteral(%s)", strings.Join(properties, ", "))
}
