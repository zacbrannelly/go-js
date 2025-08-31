package ast

import "fmt"

type BindingElementNode struct {
	Parent      Node
	Children    []Node
	Target      Node
	Initializer Node
}

func (n *BindingElementNode) GetNodeType() NodeType {
	return BindingElement
}

func (n *BindingElementNode) GetParent() Node {
	return n.Parent
}

func (n *BindingElementNode) GetChildren() []Node {
	return n.Children
}

func (n *BindingElementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BindingElementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BindingElementNode) ToString() string {
	if n.Initializer == nil {
		return fmt.Sprintf("BindingElement(%s)", n.Target.ToString())
	}

	return fmt.Sprintf("BindingElement(%s = %s)", n.Target.ToString(), n.Initializer.ToString())
}
