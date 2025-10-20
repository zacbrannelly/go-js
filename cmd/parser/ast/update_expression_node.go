package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type UpdateExpressionNode struct {
	Operator lexer.Token
	IsPrefix bool

	parent Node
	value  Node
}

func NewUpdateExpressionNode() *UpdateExpressionNode {
	return &UpdateExpressionNode{
		Operator: lexer.Token{
			Type: -1,
		},
		IsPrefix: false,
	}
}

func (n *UpdateExpressionNode) GetNodeType() NodeType {
	return UpdateExpression
}

func (n *UpdateExpressionNode) GetParent() Node {
	return n.parent
}

func (n *UpdateExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *UpdateExpressionNode) GetChildren() []Node {
	return []Node{n.value}
}

func (n *UpdateExpressionNode) SetChildren(children []Node) {
	panic("UpdateExpressionNode does not support adding children")
}

func (n *UpdateExpressionNode) GetValue() Node {
	return n.value
}

func (n *UpdateExpressionNode) SetValue(value Node) {
	if value != nil {
		value.SetParent(n)
	}
	n.value = value
}

func (n *UpdateExpressionNode) IsComposable() bool {
	return false
}

func (n *UpdateExpressionNode) ToString() string {
	var leftVal string = n.Operator.Value
	var rightVal string = n.value.ToString()

	if !n.IsPrefix {
		leftVal = n.value.ToString()
		rightVal = n.Operator.Value
	}
	return fmt.Sprintf("UpdateExpression(%s %s)", leftVal, rightVal)
}
