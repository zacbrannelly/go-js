package ast

import "fmt"

type BooleanLiteralNode struct {
	parent Node
	Value  bool
}

func NewBooleanLiteralNode(value bool) *BooleanLiteralNode {
	return &BooleanLiteralNode{
		Value: value,
	}
}

func (n *BooleanLiteralNode) GetNodeType() NodeType {
	return Expression
}

func (n *BooleanLiteralNode) GetParent() Node {
	return n.parent
}

func (n *BooleanLiteralNode) GetChildren() []Node {
	return nil
}

func (n *BooleanLiteralNode) SetChildren(children []Node) {
	panic("BooleanLiteralNode does not support adding children")
}

func (n *BooleanLiteralNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BooleanLiteralNode) IsComposable() bool {
	return false
}

func (n *BooleanLiteralNode) ToString() string {
	return fmt.Sprintf("BooleanLiteral(%t)", n.Value)
}
