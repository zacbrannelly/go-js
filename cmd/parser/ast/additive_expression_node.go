package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type AdditiveExpressionNode struct {
	// Public fields
	Operator lexer.Token

	// Private fields
	parent Node
	left   Node
	right  Node
}

func NewAdditiveExpressionNode() *AdditiveExpressionNode {
	return &AdditiveExpressionNode{
		Operator: lexer.Token{
			Type: -1,
		},
	}
}

func (n *AdditiveExpressionNode) GetNodeType() NodeType {
	return AdditiveExpression
}

func (n *AdditiveExpressionNode) GetParent() Node {
	return n.parent
}

func (n *AdditiveExpressionNode) GetChildren() []Node {
	return nil
}

func (n *AdditiveExpressionNode) SetChildren(children []Node) {
	panic("AdditiveExpressionNode does not support adding children")
}

func (n *AdditiveExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *AdditiveExpressionNode) ToString() string {
	return fmt.Sprintf("AdditiveExpression(%s %s %s)", n.left.ToString(), n.Operator.Value, n.right.ToString())
}

func (n *AdditiveExpressionNode) GetLeft() Node {
	return n.left
}

func (n *AdditiveExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *AdditiveExpressionNode) GetRight() Node {
	return n.right
}

func (n *AdditiveExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *AdditiveExpressionNode) SetOperator(operator lexer.Token) {
	n.Operator = operator
}

func (n *AdditiveExpressionNode) GetOperator() lexer.Token {
	return n.Operator
}
