package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
)

type LogicalANDExpressionNode struct {
	parent Node
	left   Node
	right  Node
}

func NewLogicalANDExpressionNode() *LogicalANDExpressionNode {
	return &LogicalANDExpressionNode{}
}

func (n *LogicalANDExpressionNode) GetNodeType() NodeType {
	return LogicalANDExpression
}

func (n *LogicalANDExpressionNode) GetParent() Node {
	return n.parent
}

func (n *LogicalANDExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *LogicalANDExpressionNode) SetChildren(children []Node) {
	panic("LogicalANDExpressionNode does not support adding children")
}

func (n *LogicalANDExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *LogicalANDExpressionNode) GetLeft() Node {
	return n.left
}

func (n *LogicalANDExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *LogicalANDExpressionNode) GetRight() Node {
	return n.right
}

func (n *LogicalANDExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *LogicalANDExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *LogicalANDExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.And, Value: "&&"}
}

func (n *LogicalANDExpressionNode) IsComposable() bool {
	return false
}

func (n *LogicalANDExpressionNode) ToString() string {
	return fmt.Sprintf("LogicalANDExpression(%s && %s)", n.left.ToString(), n.right.ToString())
}
