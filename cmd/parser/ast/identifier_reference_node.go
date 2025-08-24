package ast

import "fmt"

type IdentifierReferenceNode struct {
	Parent     Node
	Children   []Node
	Identifier string
}

func (n *IdentifierReferenceNode) GetNodeType() NodeType {
	return IdentifierReference
}

func (n *IdentifierReferenceNode) GetParent() Node {
	return n.Parent
}

func (n *IdentifierReferenceNode) GetChildren() []Node {
	return n.Children
}

func (n *IdentifierReferenceNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *IdentifierReferenceNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *IdentifierReferenceNode) ToString() string {
	return fmt.Sprintf("IdentifierReference(%s)", n.Identifier)
}
