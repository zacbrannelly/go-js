package ast

import (
	"fmt"
	"slices"
)

type MemberExpressionNode struct {
	PropertyIdentifier string
	Super              bool

	parent   Node
	object   Node
	property Node
}

func NewMemberExpressionNode() *MemberExpressionNode {
	return &MemberExpressionNode{
		PropertyIdentifier: "",
	}
}

func (n *MemberExpressionNode) GetNodeType() NodeType {
	return MemberExpression
}

func (n *MemberExpressionNode) GetParent() Node {
	return n.parent
}

func (n *MemberExpressionNode) GetChildren() []Node {
	return slices.DeleteFunc([]Node{n.object, n.property}, func(n Node) bool {
		return n == nil
	})
}

func (n *MemberExpressionNode) SetChildren(children []Node) {
	panic("MemberExpressionNode does not support adding children")
}

func (n *MemberExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *MemberExpressionNode) GetObject() Node {
	return n.object
}

func (n *MemberExpressionNode) SetObject(object Node) {
	if object != nil {
		object.SetParent(n)
	}
	n.object = object
}

func (n *MemberExpressionNode) GetProperty() Node {
	return n.property
}

func (n *MemberExpressionNode) SetProperty(property Node) {
	if property != nil {
		property.SetParent(n)
	}
	n.property = property
}

func (n *MemberExpressionNode) IsComposable() bool {
	return false
}

func (n *MemberExpressionNode) ToString() string {
	var identifier string
	if n.PropertyIdentifier != "" {
		identifier = n.PropertyIdentifier
	} else if n.property != nil {
		identifier = n.property.ToString()
	} else {
		identifier = "?"
	}

	object := ""
	if n.Super {
		object = "super"
	} else {
		object = n.object.ToString()
	}

	return fmt.Sprintf("MemberExpression(%s[%s])", object, identifier)
}
