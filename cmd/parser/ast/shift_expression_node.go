package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type ShiftExpressionNode struct {
	Operator lexer.Token

	parent Node
	left   Node
	right  Node
}

func NewShiftExpressionNode() *ShiftExpressionNode {
	return &ShiftExpressionNode{
		Operator: lexer.Token{
			Type: -1,
		},
	}
}

func (n *ShiftExpressionNode) GetNodeType() NodeType {
	return ShiftExpression
}

func (n *ShiftExpressionNode) GetParent() Node {
	return n.parent
}

func (n *ShiftExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *ShiftExpressionNode) SetChildren(children []Node) {
	panic("ShiftExpressionNode does not support adding children")
}

func (n *ShiftExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ShiftExpressionNode) GetLeft() Node {
	return n.left
}

func (n *ShiftExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *ShiftExpressionNode) GetRight() Node {
	return n.right
}

func (n *ShiftExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *ShiftExpressionNode) SetOperator(operator lexer.Token) {
	n.Operator = operator
}

func (n *ShiftExpressionNode) GetOperator() lexer.Token {
	return n.Operator
}

func (n *ShiftExpressionNode) IsComposable() bool {
	return false
}

func (n *ShiftExpressionNode) ToString() string {
	return fmt.Sprintf("ShiftExpression(%s %s %s)", n.left.ToString(), n.Operator.Value, n.right.ToString())
}
