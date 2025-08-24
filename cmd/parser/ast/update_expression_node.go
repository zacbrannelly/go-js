package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type UpdateExpressionNode struct {
	Parent   Node
	Children []Node
	Operator lexer.Token
	Value    Node
}

func (n *UpdateExpressionNode) GetNodeType() NodeType {
	return UpdateExpression
}

func (n *UpdateExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *UpdateExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *UpdateExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *UpdateExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *UpdateExpressionNode) ToString() string {
	return fmt.Sprintf("UpdateExpression(%s %s)", n.Operator.Value, n.Value.ToString())
}
