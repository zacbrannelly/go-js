package ast

import "fmt"

type BindingRestNode struct {
	Parent         Node
	Children       []Node
	Identifier     Node
	BindingPattern Node
}

func (n *BindingRestNode) GetNodeType() NodeType {
	return BindingRestProperty
}

func (n *BindingRestNode) GetParent() Node {
	return n.Parent
}

func (n *BindingRestNode) GetChildren() []Node {
	return n.Children
}

func (n *BindingRestNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BindingRestNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BindingRestNode) ToString() string {
	if n.BindingPattern != nil {
		return fmt.Sprintf("BindingRest(%s)", n.BindingPattern.ToString())
	}

	return fmt.Sprintf("BindingRest(%s)", n.Identifier.ToString())
}
