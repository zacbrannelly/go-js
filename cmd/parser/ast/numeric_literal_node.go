package ast

import "fmt"

type NumericLiteralNode struct {
	Parent   Node
	Children []Node
	Value    float64
}

func (n *NumericLiteralNode) GetNodeType() NodeType {
	return Expression
}

func (n *NumericLiteralNode) GetParent() Node {
	return n.Parent
}

func (n *NumericLiteralNode) GetChildren() []Node {
	return n.Children
}

func (n *NumericLiteralNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *NumericLiteralNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *NumericLiteralNode) ToString() string {
	return fmt.Sprintf("NumericLiteral(%f)", n.Value)
}
