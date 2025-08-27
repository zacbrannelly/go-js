package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type OptionalExpressionNode struct {
	Parent     Node
	Children   []Node
	Operator   lexer.Token
	Expression Node
}

func (n *OptionalExpressionNode) GetNodeType() NodeType {
	return OptionalExpression
}

func (n *OptionalExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *OptionalExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *OptionalExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *OptionalExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *OptionalExpressionNode) ToString() string {
	return fmt.Sprintf("OptionalExpression(%s)", n.Expression.ToString())
}
