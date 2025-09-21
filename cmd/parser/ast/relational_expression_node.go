package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type RelationalExpressionNode struct {
	Operator lexer.Token

	parent Node
	left   Node
	right  Node
}

func NewRelationalExpressionNode() *RelationalExpressionNode {
	return &RelationalExpressionNode{
		Operator: lexer.Token{
			Type: -1,
		},
	}
}

func (n *RelationalExpressionNode) GetNodeType() NodeType {
	return RelationalExpression
}

func (n *RelationalExpressionNode) GetParent() Node {
	return n.parent
}

func (n *RelationalExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *RelationalExpressionNode) SetChildren(children []Node) {
	panic("RelationalExpressionNode does not support adding children")
}

func (n *RelationalExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *RelationalExpressionNode) GetLeft() Node {
	return n.left
}

func (n *RelationalExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *RelationalExpressionNode) GetRight() Node {
	return n.right
}

func (n *RelationalExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *RelationalExpressionNode) SetOperator(operator lexer.Token) {
	n.Operator = operator
}

func (n *RelationalExpressionNode) GetOperator() lexer.Token {
	return n.Operator
}

func (n *RelationalExpressionNode) IsComposable() bool {
	return false
}

func (n *RelationalExpressionNode) ToString() string {
	return fmt.Sprintf("RelationalExpression(%s %s %s)", n.left.ToString(), n.Operator.Value, n.right.ToString())
}
