package ast

import "fmt"

type BindingIdentifierNode struct {
	Parent     Node
	Children   []Node
	Identifier string
}

func (n *BindingIdentifierNode) GetNodeType() NodeType {
	return BindingIdentifier
}

func (n *BindingIdentifierNode) GetParent() Node {
	return n.Parent
}

func (n *BindingIdentifierNode) GetChildren() []Node {
	return n.Children
}

func (n *BindingIdentifierNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BindingIdentifierNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BindingIdentifierNode) ToString() string {
	return fmt.Sprintf("BindingIdentifier(%s)", n.Identifier)
}
