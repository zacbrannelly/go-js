package ast

import "fmt"

type BindingIdentifierNode struct {
	// Public fields
	Identifier string

	// Private fields
	parent Node
}

func NewBindingIdentifierNode(identifier string) *BindingIdentifierNode {
	return &BindingIdentifierNode{
		Identifier: identifier,
	}
}

func (n *BindingIdentifierNode) GetNodeType() NodeType {
	return BindingIdentifier
}

func (n *BindingIdentifierNode) GetParent() Node {
	return n.parent
}

func (n *BindingIdentifierNode) GetChildren() []Node {
	return nil
}

func (n *BindingIdentifierNode) SetChildren(children []Node) {
	panic("BindingIdentifierNode does not support adding children")
}

func (n *BindingIdentifierNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BindingIdentifierNode) IsComposable() bool {
	return false
}

func (n *BindingIdentifierNode) ToString() string {
	return fmt.Sprintf("BindingIdentifier(%s)", n.Identifier)
}
