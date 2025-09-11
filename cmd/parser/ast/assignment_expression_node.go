package ast

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

type AssignmentExpressionNode struct {
	Parent   Node
	Children []Node
	Target   Node
	Value    Node
	Operator lexer.Token
}

func (n *AssignmentExpressionNode) GetNodeType() NodeType {
	return AssignmentExpression
}

func (n *AssignmentExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *AssignmentExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *AssignmentExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *AssignmentExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *AssignmentExpressionNode) ToString() string {
	return fmt.Sprintf("AssignmentExpression(%s %s %s)", n.Target.ToString(), n.Operator.Value, n.Value.ToString())
}
