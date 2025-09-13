package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type LogicalORExpressionNode struct {
	Parent   Node
	Children []Node
	Left     Node
	Right    Node
}

func (n *LogicalORExpressionNode) GetNodeType() NodeType {
	return LogicalORExpression
}

func (n *LogicalORExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *LogicalORExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *LogicalORExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *LogicalORExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *LogicalORExpressionNode) ToString() string {
	return fmt.Sprintf("LogicalORExpression(%s || %s)", n.Left.ToString(), n.Right.ToString())
}

func (n *LogicalORExpressionNode) GetLeft() Node {
	return n.Left
}

func (n *LogicalORExpressionNode) SetLeft(left Node) {
	n.Left = left
}

func (n *LogicalORExpressionNode) GetRight() Node {
	return n.Right
}

func (n *LogicalORExpressionNode) SetRight(right Node) {
	n.Right = right
}

func (n *LogicalORExpressionNode) SetOperator(operator lexer.Token) {
	// No-op
}

func (n *LogicalORExpressionNode) GetOperator() lexer.Token {
	return lexer.Token{Type: lexer.Or, Value: "||"}
}
