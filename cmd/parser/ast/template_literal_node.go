package ast

import (
	"fmt"
)

type TemplateLiteralNode struct {
	Tagged   bool
	Children []Node

	parent         Node
	tagFunctionRef Node
}

func NewTemplateLiteralNode() *TemplateLiteralNode {
	return &TemplateLiteralNode{
		Children: make([]Node, 0),
	}
}

func (n *TemplateLiteralNode) GetNodeType() NodeType {
	return TemplateLiteral
}

func (n *TemplateLiteralNode) GetChildren() []Node {
	return n.Children
}

func (n *TemplateLiteralNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *TemplateLiteralNode) GetParent() Node {
	return n.parent
}

func (n *TemplateLiteralNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *TemplateLiteralNode) GetTagFunctionRef() Node {
	return n.tagFunctionRef
}

func (n *TemplateLiteralNode) SetTagFunctionRef(tagFunctionRef Node) {
	if tagFunctionRef != nil {
		tagFunctionRef.SetParent(n)
		n.Tagged = true
	}
	n.tagFunctionRef = tagFunctionRef
}

func (n *TemplateLiteralNode) IsComposable() bool {
	return true
}

func (n *TemplateLiteralNode) ToString() string {
	if n.tagFunctionRef != nil {
		return fmt.Sprintf("TemplateLiteral(%s)", n.tagFunctionRef.ToString())
	}

	return "TemplateLiteral"
}
