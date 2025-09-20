package ast

import (
	"fmt"
	"strings"
)

type CallExpressionNode struct {
	Super bool

	parent    Node
	callee    Node
	arguments []Node
}

func NewCallExpressionNode(callee Node, arguments []Node) *CallExpressionNode {
	newNode := &CallExpressionNode{}
	newNode.SetCallee(callee)
	newNode.SetArguments(arguments)
	return newNode
}

func NewCallExpressionNodeForSuper(arguments []Node) *CallExpressionNode {
	newNode := &CallExpressionNode{}
	newNode.SetArguments(arguments)
	newNode.Super = true
	return newNode
}

func (n *CallExpressionNode) GetNodeType() NodeType {
	return CallExpression
}

func (n *CallExpressionNode) GetParent() Node {
	return n.parent
}

func (n *CallExpressionNode) GetChildren() []Node {
	return nil
}

func (n *CallExpressionNode) SetChildren(children []Node) {
	panic("CallExpressionNode does not support adding children")
}

func (n *CallExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *CallExpressionNode) GetCallee() Node {
	return n.callee
}

func (n *CallExpressionNode) SetCallee(callee Node) {
	if callee != nil {
		callee.SetParent(n)
	}
	n.callee = callee
}

func (n *CallExpressionNode) GetArguments() []Node {
	return n.arguments
}

func (n *CallExpressionNode) SetArguments(arguments []Node) {
	for _, argument := range arguments {
		argument.SetParent(n)
	}
	n.arguments = arguments
}

func (n *CallExpressionNode) ToString() string {
	arguments := []string{}
	for _, argument := range n.arguments {
		arguments = append(arguments, argument.ToString())
	}

	callee := ""
	if n.Super {
		callee = "super "
	} else {
		callee = n.callee.ToString()
	}

	return fmt.Sprintf("CallExpression(%s(%s))", callee, strings.Join(arguments, ", "))
}
