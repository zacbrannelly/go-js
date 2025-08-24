package ast

import "fmt"

type BooleanLiteralNode struct {
	Parent   Node
	Children []Node
	Value    bool
}

func (n *BooleanLiteralNode) GetNodeType() NodeType {
	return Expression
}

func (n *BooleanLiteralNode) GetParent() Node {
	return n.Parent
}

func (n *BooleanLiteralNode) GetChildren() []Node {
	return n.Children
}

func (n *BooleanLiteralNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BooleanLiteralNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BooleanLiteralNode) ToString() string {
	return fmt.Sprintf("BooleanLiteral(%t)", n.Value)
}
