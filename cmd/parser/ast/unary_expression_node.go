package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type UnaryExpressionNode struct {
	Operator lexer.Token

	parent Node
	value  Node
}

func NewUnaryExpressionNode() *UnaryExpressionNode {
	return &UnaryExpressionNode{
		Operator: lexer.Token{
			Type: -1,
		},
	}
}

func (n *UnaryExpressionNode) GetNodeType() NodeType {
	return UnaryExpression
}

func (n *UnaryExpressionNode) GetParent() Node {
	return n.parent
}

func (n *UnaryExpressionNode) GetChildren() []Node {
	return []Node{n.value}
}

func (n *UnaryExpressionNode) SetChildren(children []Node) {
	panic("UnaryExpressionNode does not support adding children")
}

func (n *UnaryExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *UnaryExpressionNode) GetValue() Node {
	return n.value
}

func (n *UnaryExpressionNode) SetValue(value Node) {
	if value != nil {
		value.SetParent(n)
	}
	n.value = value
}

func (n *UnaryExpressionNode) IsComposable() bool {
	return false
}

func (n *UnaryExpressionNode) ToString() string {
	return fmt.Sprintf("UnaryExpression(%s %s)", n.Operator.Value, n.value.ToString())
}
