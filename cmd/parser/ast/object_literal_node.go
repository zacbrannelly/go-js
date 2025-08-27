package ast

import (
	"fmt"
	"strings"
)

type ObjectLiteralNode struct {
	Parent     Node
	Children   []Node
	Properties []Node
}

func (n *ObjectLiteralNode) GetNodeType() NodeType {
	return ObjectLiteral
}

func (n *ObjectLiteralNode) GetChildren() []Node {
	return n.Children
}

func (n *ObjectLiteralNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ObjectLiteralNode) GetParent() Node {
	return n.Parent
}

func (n *ObjectLiteralNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ObjectLiteralNode) ToString() string {
	properties := []string{}

	for _, property := range n.Properties {
		properties = append(properties, property.ToString())
	}

	return fmt.Sprintf("ObjectLiteral(%s)", strings.Join(properties, ", "))
}
