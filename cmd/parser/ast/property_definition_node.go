package ast

import "fmt"

type PropertyDefinitionNode struct {
	Static bool

	parent Node
	key    Node
	value  Node
}

func NewPropertyDefinitionNode(key Node, value Node) *PropertyDefinitionNode {
	newNode := &PropertyDefinitionNode{}
	newNode.SetKey(key)
	newNode.SetValue(value)
	return newNode
}

func (n *PropertyDefinitionNode) GetNodeType() NodeType {
	return PropertyDefinition
}

func (n *PropertyDefinitionNode) GetParent() Node {
	return n.parent
}

func (n *PropertyDefinitionNode) GetChildren() []Node {
	return nil
}

func (n *PropertyDefinitionNode) SetChildren(children []Node) {
	panic("PropertyDefinitionNode does not support adding children")
}

func (n *PropertyDefinitionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *PropertyDefinitionNode) GetKey() Node {
	return n.key
}

func (n *PropertyDefinitionNode) SetKey(key Node) {
	if key != nil {
		key.SetParent(n)
	}
	n.key = key
}

func (n *PropertyDefinitionNode) GetValue() Node {
	return n.value
}

func (n *PropertyDefinitionNode) SetValue(value Node) {
	if value != nil {
		value.SetParent(n)
	}
	n.value = value
}

func (node *PropertyDefinitionNode) ToString() string {
	static := ""
	if node.Static {
		static = "static "
	}

	if node.value == nil {
		return fmt.Sprintf("PropertyDefinition(%s%s)", static, node.key.ToString())
	}

	return fmt.Sprintf("PropertyDefinition(%s%s: %s)", static, node.key.ToString(), node.value.ToString())
}
