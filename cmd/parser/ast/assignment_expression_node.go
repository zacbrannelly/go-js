package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type AssignmentExpressionNode struct {
	// Public fields
	Operator lexer.Token

	// Private fields
	parent Node
	target Node
	value  Node
}

func NewAssignmentExpressionNode(target Node, operator lexer.Token, value Node) *AssignmentExpressionNode {
	newNode := &AssignmentExpressionNode{
		Operator: operator,
	}
	newNode.SetTarget(target)
	newNode.SetValue(value)
	return newNode
}

func (n *AssignmentExpressionNode) GetNodeType() NodeType {
	return AssignmentExpression
}

func (n *AssignmentExpressionNode) GetParent() Node {
	return n.parent
}

func (n *AssignmentExpressionNode) GetChildren() []Node {
	return []Node{n.target, n.value}
}

func (n *AssignmentExpressionNode) SetChildren(children []Node) {
	panic("AssignmentExpressionNode does not support adding children")
}

func (n *AssignmentExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *AssignmentExpressionNode) GetTarget() Node {
	return n.target
}

func (n *AssignmentExpressionNode) SetTarget(target Node) {
	if target != nil {
		target.SetParent(n)
	}
	n.target = target
}

func (n *AssignmentExpressionNode) GetValue() Node {
	return n.value
}

func (n *AssignmentExpressionNode) SetValue(value Node) {
	if value != nil {
		value.SetParent(n)
	}
	n.value = value
}

func (n *AssignmentExpressionNode) IsComposable() bool {
	return false
}

func (n *AssignmentExpressionNode) ToString() string {
	return fmt.Sprintf("AssignmentExpression(%s %s %s)", n.target.ToString(), n.Operator.Value, n.value.ToString())
}
