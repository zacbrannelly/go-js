package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type UnaryExpressionNode struct {
	Parent   Node
	Children []Node
	Operator lexer.Token
	Value    Node
}

func (n *UnaryExpressionNode) GetNodeType() NodeType {
	return UnaryExpression
}

func (n *UnaryExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *UnaryExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *UnaryExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *UnaryExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *UnaryExpressionNode) ToString() string {
	return fmt.Sprintf("UnaryExpression(%s %s)", n.Operator.Value, n.Value.ToString())
}
