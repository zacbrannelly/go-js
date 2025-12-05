package ast

import (
	"fmt"
)

type WhileStatementNode struct {
	parent    Node
	condition Node
	statement Node
}

func NewWhileStatementNode(condition Node, statement Node) *WhileStatementNode {
	newNode := &WhileStatementNode{}
	newNode.SetCondition(condition)
	newNode.SetStatement(statement)
	return newNode
}

func (n *WhileStatementNode) GetNodeType() NodeType {
	return WhileStatement
}

func (n *WhileStatementNode) GetParent() Node {
	return n.parent
}

func (n *WhileStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *WhileStatementNode) GetChildren() []Node {
	return []Node{n.condition, n.statement}
}

func (n *WhileStatementNode) SetChildren(children []Node) {
	panic("WhileStatementNode does not support adding children")
}

func (n *WhileStatementNode) GetCondition() Node {
	return n.condition
}

func (n *WhileStatementNode) SetCondition(condition Node) {
	if condition != nil {
		condition.SetParent(n)
	}
	n.condition = condition
}

func (n *WhileStatementNode) GetStatement() Node {
	return n.statement
}

func (n *WhileStatementNode) SetStatement(statement Node) {
	if statement != nil {
		statement.SetParent(n)
	}
	n.statement = statement
}

func (n *WhileStatementNode) IsComposable() bool {
	return false
}

func (n *WhileStatementNode) ToString() string {
	return fmt.Sprintf("WhileStatement(while (%s) %s)", n.condition.ToString(), n.statement.ToString())
}
