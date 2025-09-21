package ast

import "fmt"

type IdentifierReferenceNode struct {
	parent     Node
	Identifier string
}

func NewIdentifierReferenceNode(identifier string) *IdentifierReferenceNode {
	return &IdentifierReferenceNode{
		Identifier: identifier,
	}
}

func (n *IdentifierReferenceNode) GetNodeType() NodeType {
	return IdentifierReference
}

func (n *IdentifierReferenceNode) GetParent() Node {
	return n.parent
}

func (n *IdentifierReferenceNode) GetChildren() []Node {
	return nil
}

func (n *IdentifierReferenceNode) SetChildren(children []Node) {
	panic("IdentifierReferenceNode does not support adding children")
}

func (n *IdentifierReferenceNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *IdentifierReferenceNode) IsComposable() bool {
	return false
}

func (n *IdentifierReferenceNode) ToString() string {
	return fmt.Sprintf("IdentifierReference(%s)", n.Identifier)
}
