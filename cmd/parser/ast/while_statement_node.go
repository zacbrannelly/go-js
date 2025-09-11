package ast

import (
	"fmt"
)

type WhileStatementNode struct {
	Parent    Node
	Children  []Node
	Condition Node
	Statement Node
}

func (n *WhileStatementNode) GetNodeType() NodeType {
	return WhileStatement
}

func (n *WhileStatementNode) GetParent() Node {
	return n.Parent
}

func (n *WhileStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *WhileStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *WhileStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *WhileStatementNode) ToString() string {
	return fmt.Sprintf("WhileStatement(while (%s) %s)", n.Condition.ToString(), n.Statement.ToString())
}
