package ast

import "fmt"

type BindingPropertyNode struct {
	Parent   Node
	Children []Node

	// BindingIdentifier or PropertyName
	Target Node

	// For a single name property
	Initializer Node

	// For a pattern property
	BindingElement Node
}

func (n *BindingPropertyNode) GetNodeType() NodeType {
	return BindingProperty
}

func (n *BindingPropertyNode) GetParent() Node {
	return n.Parent
}

func (n *BindingPropertyNode) GetChildren() []Node {
	return n.Children
}

func (n *BindingPropertyNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BindingPropertyNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BindingPropertyNode) ToString() string {
	if n.Initializer == nil && n.BindingElement == nil {
		return fmt.Sprintf("BindingProperty(%s)", n.Target.ToString())
	}

	if n.Initializer != nil {
		return fmt.Sprintf("BindingProperty(%s = %s)", n.Target.ToString(), n.Initializer.ToString())
	}

	return fmt.Sprintf("BindingProperty(%s: %s)", n.Target.ToString(), n.BindingElement.ToString())
}
