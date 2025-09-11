package ast

import (
	"fmt"
	"strings"
)

type CallExpressionNode struct {
	Parent    Node
	Children  []Node
	Callee    Node
	Arguments []Node
	Super     bool
}

func (n *CallExpressionNode) GetNodeType() NodeType {
	return CallExpression
}

func (n *CallExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *CallExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *CallExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *CallExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *CallExpressionNode) ToString() string {
	arguments := []string{}
	for _, argument := range n.Arguments {
		arguments = append(arguments, argument.ToString())
	}

	callee := ""
	if n.Super {
		callee = "super "
	} else {
		callee = n.Callee.ToString()
	}

	return fmt.Sprintf("CallExpression(%s(%s))", callee, strings.Join(arguments, ", "))
}
