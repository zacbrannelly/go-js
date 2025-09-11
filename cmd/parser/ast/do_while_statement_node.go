package ast

import (
	"fmt"
)

type DoWhileStatementNode struct {
	Parent    Node
	Children  []Node
	Condition Node
	Statement Node
}

func (n *DoWhileStatementNode) GetNodeType() NodeType {
	return DoWhileStatement
}

func (n *DoWhileStatementNode) GetParent() Node {
	return n.Parent
}

func (n *DoWhileStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *DoWhileStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *DoWhileStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *DoWhileStatementNode) ToString() string {
	return fmt.Sprintf("DoWhileStatement(do %s while (%s))", n.Statement.ToString(), n.Condition.ToString())
}
