package ast

import (
	"fmt"
	"slices"
)

type BindingElementNode struct {
	parent      Node
	target      Node
	initializer Node
}

func NewBindingElementNode(target Node, initializer Node) *BindingElementNode {
	newNode := &BindingElementNode{}
	newNode.SetTarget(target)
	newNode.SetInitializer(initializer)
	return newNode
}

func (n *BindingElementNode) GetNodeType() NodeType {
	return BindingElement
}

func (n *BindingElementNode) GetParent() Node {
	return n.parent
}

func (n *BindingElementNode) GetChildren() []Node {
	return slices.DeleteFunc([]Node{n.target, n.initializer}, func(n Node) bool {
		return n == nil
	})
}

func (n *BindingElementNode) SetChildren(children []Node) {
	panic("BindingElementNode does not support adding children")
}

func (n *BindingElementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BindingElementNode) GetTarget() Node {
	return n.target
}

func (n *BindingElementNode) SetTarget(target Node) {
	if target != nil {
		target.SetParent(n)
	}
	n.target = target
}

func (n *BindingElementNode) GetInitializer() Node {
	return n.initializer
}

func (n *BindingElementNode) SetInitializer(initializer Node) {
	if initializer != nil {
		initializer.SetParent(n)
	}
	n.initializer = initializer
}

func (n *BindingElementNode) IsComposable() bool {
	return false
}

func (n *BindingElementNode) ToString() string {
	if n.initializer == nil {
		return fmt.Sprintf("BindingElement(%s)", n.target.ToString())
	}

	return fmt.Sprintf("BindingElement(%s = %s)", n.target.ToString(), n.initializer.ToString())
}
