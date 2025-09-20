package ast

import "fmt"

type BindingPropertyNode struct {
	parent Node

	// BindingIdentifier or PropertyName
	target Node

	// For a single name property
	initializer Node

	// For a pattern property
	bindingElement Node
}

func NewBindingPropertyNodeForProperty(target Node, initializer Node) *BindingPropertyNode {
	newNode := &BindingPropertyNode{}
	newNode.SetTarget(target)
	newNode.SetInitializer(initializer)
	return newNode
}

func NewBindingPropertyNodeForPattern(target Node, bindingElement Node) *BindingPropertyNode {
	newNode := &BindingPropertyNode{}
	newNode.SetTarget(target)
	newNode.SetBindingElement(bindingElement)
	return newNode
}

func (n *BindingPropertyNode) GetNodeType() NodeType {
	return BindingProperty
}

func (n *BindingPropertyNode) GetParent() Node {
	return n.parent
}

func (n *BindingPropertyNode) GetChildren() []Node {
	return nil
}

func (n *BindingPropertyNode) SetChildren(children []Node) {
	panic("BindingPropertyNode does not support adding children")
}

func (n *BindingPropertyNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BindingPropertyNode) GetTarget() Node {
	return n.target
}

func (n *BindingPropertyNode) SetTarget(target Node) {
	if target != nil {
		target.SetParent(n)
	}
	n.target = target
}

func (n *BindingPropertyNode) GetInitializer() Node {
	return n.initializer
}

func (n *BindingPropertyNode) SetInitializer(initializer Node) {
	if initializer != nil {
		initializer.SetParent(n)
	}
	n.initializer = initializer
}

func (n *BindingPropertyNode) GetBindingElement() Node {
	return n.bindingElement
}

func (n *BindingPropertyNode) SetBindingElement(bindingElement Node) {
	if bindingElement != nil {
		bindingElement.SetParent(n)
	}
	n.bindingElement = bindingElement
}

func (n *BindingPropertyNode) ToString() string {
	if n.initializer == nil && n.bindingElement == nil {
		return fmt.Sprintf("BindingProperty(%s)", n.target.ToString())
	}

	if n.initializer != nil {
		return fmt.Sprintf("BindingProperty(%s = %s)", n.target.ToString(), n.initializer.ToString())
	}

	return fmt.Sprintf("BindingProperty(%s: %s)", n.target.ToString(), n.bindingElement.ToString())
}
