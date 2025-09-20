package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type EqualityExpressionNode struct {
	// Public fields
	Operator lexer.Token

	// Private fields
	parent Node
	left   Node
	right  Node
}

func NewEqualityExpressionNode() *EqualityExpressionNode {
	return &EqualityExpressionNode{
		Operator: lexer.Token{
			Type: -1,
		},
	}
}

func (n *EqualityExpressionNode) GetNodeType() NodeType {
	return EqualityExpression
}

func (n *EqualityExpressionNode) GetParent() Node {
	return n.parent
}

func (n *EqualityExpressionNode) GetChildren() []Node {
	return nil
}

func (n *EqualityExpressionNode) SetChildren(children []Node) {
	panic("EqualityExpressionNode does not support adding children")
}

func (n *EqualityExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *EqualityExpressionNode) ToString() string {
	return fmt.Sprintf("EqualityExpression(%s %s %s)", n.left.ToString(), n.Operator.Value, n.right.ToString())
}

func (n *EqualityExpressionNode) GetLeft() Node {
	return n.left
}

func (n *EqualityExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *EqualityExpressionNode) GetRight() Node {
	return n.right
}

func (n *EqualityExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *EqualityExpressionNode) SetOperator(operator lexer.Token) {
	n.Operator = operator
}

func (n *EqualityExpressionNode) GetOperator() lexer.Token {
	return n.Operator
}
