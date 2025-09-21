package ast

import (
	"fmt"
)

type DoWhileStatementNode struct {
	parent    Node
	condition Node
	statement Node
}

func NewDoWhileStatementNode(condition Node, statement Node) *DoWhileStatementNode {
	newNode := &DoWhileStatementNode{}
	newNode.SetCondition(condition)
	newNode.SetStatement(statement)
	return newNode
}

func (n *DoWhileStatementNode) GetNodeType() NodeType {
	return DoWhileStatement
}

func (n *DoWhileStatementNode) GetParent() Node {
	return n.parent
}

func (n *DoWhileStatementNode) GetChildren() []Node {
	return []Node{n.condition, n.statement}
}

func (n *DoWhileStatementNode) SetChildren(children []Node) {
	panic("DoWhileStatementNode does not support adding children")
}

func (n *DoWhileStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *DoWhileStatementNode) GetCondition() Node {
	return n.condition
}

func (n *DoWhileStatementNode) SetCondition(condition Node) {
	if condition != nil {
		condition.SetParent(n)
	}
	n.condition = condition
}

func (n *DoWhileStatementNode) GetStatement() Node {
	return n.statement
}

func (n *DoWhileStatementNode) SetStatement(statement Node) {
	if statement != nil {
		statement.SetParent(n)
	}
	n.statement = statement
}

func (n *DoWhileStatementNode) IsComposable() bool {
	return false
}

func (n *DoWhileStatementNode) ToString() string {
	return fmt.Sprintf("DoWhileStatement(do %s while (%s))", n.statement.ToString(), n.condition.ToString())
}
