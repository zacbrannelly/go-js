package ast

import "fmt"

type LabelIdentifierNode struct {
	parent     Node
	Identifier string
}

func NewLabelIdentifierNode(identifier string) *LabelIdentifierNode {
	return &LabelIdentifierNode{
		Identifier: identifier,
	}
}

func (n *LabelIdentifierNode) GetNodeType() NodeType {
	return LabelIdentifier
}

func (n *LabelIdentifierNode) GetParent() Node {
	return n.parent
}

func (n *LabelIdentifierNode) GetChildren() []Node {
	return nil
}

func (n *LabelIdentifierNode) SetChildren(children []Node) {
	panic("LabelIdentifierNode does not support adding children")
}

func (n *LabelIdentifierNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *LabelIdentifierNode) IsComposable() bool {
	return false
}

func (n *LabelIdentifierNode) ToString() string {
	return fmt.Sprintf("LabelIdentifier(%s)", n.Identifier)
}
