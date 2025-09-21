package ast

import (
	"fmt"
	"slices"
)

type BindingRestNode struct {
	parent         Node
	identifier     Node
	bindingPattern Node
}

func NewBindingRestNodeForIdentifier(identifier Node) *BindingRestNode {
	newNode := &BindingRestNode{}
	newNode.SetIdentifier(identifier)
	return newNode
}

func NewBindingRestNodeForPattern(bindingPattern Node) *BindingRestNode {
	newNode := &BindingRestNode{}
	newNode.SetBindingPattern(bindingPattern)
	return newNode
}

func (n *BindingRestNode) GetNodeType() NodeType {
	return BindingRestProperty
}

func (n *BindingRestNode) GetParent() Node {
	return n.parent
}

func (n *BindingRestNode) GetChildren() []Node {
	return slices.DeleteFunc([]Node{n.identifier, n.bindingPattern}, func(n Node) bool {
		return n == nil
	})
}

func (n *BindingRestNode) SetChildren(children []Node) {
	panic("BindingRestNode does not support adding children")
}

func (n *BindingRestNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BindingRestNode) GetIdentifier() Node {
	return n.identifier
}

func (n *BindingRestNode) SetIdentifier(identifier Node) {
	if identifier != nil {
		identifier.SetParent(n)
	}
	n.identifier = identifier
}

func (n *BindingRestNode) GetBindingPattern() Node {
	return n.bindingPattern
}

func (n *BindingRestNode) SetBindingPattern(bindingPattern Node) {
	if bindingPattern != nil {
		bindingPattern.SetParent(n)
	}
	n.bindingPattern = bindingPattern
}

func (n *BindingRestNode) IsComposable() bool {
	return false
}

func (n *BindingRestNode) ToString() string {
	if n.bindingPattern != nil {
		return fmt.Sprintf("BindingRest(%s)", n.bindingPattern.ToString())
	}

	return fmt.Sprintf("BindingRest(%s)", n.identifier.ToString())
}
