package ast

import "fmt"

type MemberExpressionNode struct {
	Parent             Node
	Children           []Node
	Object             Node
	Property           Node
	PropertyIdentifier string
}

func (n *MemberExpressionNode) GetNodeType() NodeType {
	return MemberExpression
}

func (n *MemberExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *MemberExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *MemberExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *MemberExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *MemberExpressionNode) ToString() string {
	var identifier string
	if n.PropertyIdentifier != "" {
		identifier = n.PropertyIdentifier
	} else if n.Property != nil {
		identifier = n.Property.ToString()
	} else {
		identifier = "?"
	}
	return fmt.Sprintf("MemberExpression(%s[%s])", n.Object.ToString(), identifier)
}
