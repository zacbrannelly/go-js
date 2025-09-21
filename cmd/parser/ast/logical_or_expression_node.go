package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type LogicalORExpressionNode struct {
	parent Node
	left   Node
	right  Node
}

func NewLogicalORExpressionNode() *LogicalORExpressionNode {
	return &LogicalORExpressionNode{}
}

func (n *LogicalORExpressionNode) GetNodeType() NodeType {
	return LogicalORExpression
}

func (n *LogicalORExpressionNode) GetParent() Node {
	return n.parent
}

func (n *LogicalORExpressionNode) GetChildren() []Node {
	return []Node{n.left, n.right}
}

func (n *LogicalORExpressionNode) SetChildren(children []Node) {
	panic("LogicalORExpressionNode does not support adding children")
}

func (n *LogicalORExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *LogicalORExpressionNode) GetLeft() Node {
	return n.left
}

func (n *LogicalORExpressionNode) SetLeft(left Node) {
	if left != nil {
		left.SetParent(n)
	}
	n.left = left
}

func (n *LogicalORExpressionNode) GetRight() Node {
	return n.right
}

func (n *LogicalORExpressionNode) SetRight(right Node) {
	if right != nil {
		right.SetParent(n)
	}
	n.right = right
}

func (n *LogicalORExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *LogicalORExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.Or, Value: "||"}
}

func (n *LogicalORExpressionNode) IsComposable() bool {
	return false
}

func (n *LogicalORExpressionNode) ToString() string {
	return fmt.Sprintf("LogicalORExpression(%s || %s)", n.left.ToString(), n.right.ToString())
}
