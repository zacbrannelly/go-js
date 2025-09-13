package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type LogicalANDExpressionNode struct {
	Parent   Node
	Children []Node
	Left     Node
	Right    Node
}

func (n *LogicalANDExpressionNode) GetNodeType() NodeType {
	return LogicalANDExpression
}

func (n *LogicalANDExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *LogicalANDExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *LogicalANDExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *LogicalANDExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *LogicalANDExpressionNode) ToString() string {
	return fmt.Sprintf("LogicalANDExpression(%s || %s)", n.Left.ToString(), n.Right.ToString())
}

func (n *LogicalANDExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *LogicalANDExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *LogicalANDExpressionNode) GetRight() Node {
	return n.Right
}

func (n *LogicalANDExpressionNode) SetRight(right Node) {
	n.Right = right
}

func (n *LogicalANDExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *LogicalANDExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.And, Value: "&&"}
}
