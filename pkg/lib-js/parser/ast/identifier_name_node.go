package ast

import "fmt"

type IdentifierNameNode struct {
	Identifier string

	parent Node
}

func NewIdentifierNameNode(identifier string) *IdentifierNameNode {
	return &IdentifierNameNode{
		Identifier: identifier,
	}
}

func (n *IdentifierNameNode) GetNodeType() NodeType {
	return IdentifierName
}

func (n *IdentifierNameNode) GetParent() Node {
	return n.parent
}

func (n *IdentifierNameNode) GetChildren() []Node {
	return nil
}

func (n *IdentifierNameNode) SetChildren(children []Node) {
	panic("IdentifierNameNode does not support adding children")
}

func (n *IdentifierNameNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *IdentifierNameNode) IsComposable() bool {
	return false
}

func (n *IdentifierNameNode) ToString() string {
	return fmt.Sprintf("IdentifierName(%s)", n.Identifier)
}
