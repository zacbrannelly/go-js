package ast

import "fmt"

type NumericLiteralNode struct {
	parent Node
	Value  float64
}

func NewNumericLiteralNode(value float64) *NumericLiteralNode {
	return &NumericLiteralNode{
		Value: value,
	}
}

func (n *NumericLiteralNode) GetNodeType() NodeType {
	return NumericLiteral
}

func (n *NumericLiteralNode) GetParent() Node {
	return n.parent
}

func (n *NumericLiteralNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *NumericLiteralNode) GetChildren() []Node {
	return nil
}

func (n *NumericLiteralNode) SetChildren(children []Node) {
	panic("NumericLiteralNode does not support adding children")
}

func (n *NumericLiteralNode) IsComposable() bool {
	return false
}

func (n *NumericLiteralNode) ToString() string {
	return fmt.Sprintf("NumericLiteral(%f)", n.Value)
}
