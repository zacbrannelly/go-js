package ast

import "fmt"

type StringLiteralNode struct {
	Parent   Node
	Children []Node
	Value    string
}

func (n *StringLiteralNode) GetNodeType() NodeType {
	return Expression
}

func (n *StringLiteralNode) GetParent() Node {
	return n.Parent
}

func (n *StringLiteralNode) GetChildren() []Node {
	return n.Children
}

func (n *StringLiteralNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *StringLiteralNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *StringLiteralNode) ToString() string {
	return fmt.Sprintf("StringLiteral(%s)", n.Value)
}
