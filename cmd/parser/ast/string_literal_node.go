package ast

import "fmt"

type StringLiteralNode struct {
	Value string

	parent Node
}

func NewStringLiteralNode(value string) *StringLiteralNode {
	return &StringLiteralNode{
		Value: value,
	}
}

func (n *StringLiteralNode) GetNodeType() NodeType {
	return Expression
}

func (n *StringLiteralNode) GetParent() Node {
	return n.parent
}

func (n *StringLiteralNode) GetChildren() []Node {
	return nil
}

func (n *StringLiteralNode) SetChildren(children []Node) {
	panic("StringLiteralNode does not support adding children")
}

func (n *StringLiteralNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *StringLiteralNode) ToString() string {
	return fmt.Sprintf("StringLiteral(%s)", n.Value)
}
