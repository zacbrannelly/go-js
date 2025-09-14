package ast

import "fmt"

type LabelIdentifierNode struct {
	Parent     Node
	Children   []Node
	Identifier string
}

func (n *LabelIdentifierNode) GetNodeType() NodeType {
	return LabelIdentifier
}

func (n *LabelIdentifierNode) GetParent() Node {
	return n.Parent
}

func (n *LabelIdentifierNode) GetChildren() []Node {
	return n.Children
}

func (n *LabelIdentifierNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *LabelIdentifierNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *LabelIdentifierNode) ToString() string {
	return fmt.Sprintf("LabelIdentifier(%s)", n.Identifier)
}
