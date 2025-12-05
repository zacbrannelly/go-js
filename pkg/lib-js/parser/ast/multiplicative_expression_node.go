package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
)

type MultiplicativeExpressionNode struct {
	Operator lexer.Token

	parent Node
	left   Node
	right  Node
}

func NewMultiplicativeExpressionNode() *MultiplicativeExpressionNode {
	return &MultiplicativeExpressionNode{
		Operator: lexer.Token{
			Type: -1,
		},
	}
}

func (n *MultiplicativeExpressionNode) GetNodeType() NodeType {
	return MultiplicativeExpression
}

func (n *MultiplicativeExpressionNode) GetParent() Node {
	return n.parent
}

func (n *MultiplicativeExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *MultiplicativeExpressionNode) SetChildren(children []Node) {
	panic("MultiplicativeExpressionNode does not support adding children")
}

func (n *MultiplicativeExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *MultiplicativeExpressionNode) GetLeft() Node {
	return n.left
}

func (n *MultiplicativeExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *MultiplicativeExpressionNode) GetRight() Node {
	return n.right
}

func (n *MultiplicativeExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *MultiplicativeExpressionNode) SetOperator(operator lexer.Token) {
	n.Operator = operator
}

func (n *MultiplicativeExpressionNode) GetOperator() lexer.Token {
	return n.Operator
}

func (n *MultiplicativeExpressionNode) IsComposable() bool {
	return false
}

func (n *MultiplicativeExpressionNode) ToString() string {
	return fmt.Sprintf("MultiplicativeExpression(%s %s %s)", n.left.ToString(), n.Operator.Value, n.right.ToString())
}
