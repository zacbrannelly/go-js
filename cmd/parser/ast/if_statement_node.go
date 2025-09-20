package ast

import (
	"fmt"
)

type IfStatementNode struct {
	parent        Node
	condition     Node
	trueStatement Node
	elseStatement Node
}

func NewIfStatementNode(condition Node, trueStatement Node, elseStatement Node) *IfStatementNode {
	newNode := &IfStatementNode{}
	newNode.SetCondition(condition)
	newNode.SetTrueStatement(trueStatement)
	newNode.SetElseStatement(elseStatement)
	return newNode
}

func (n *IfStatementNode) GetNodeType() NodeType {
	return IfStatement
}

func (n *IfStatementNode) GetParent() Node {
	return n.parent
}

func (n *IfStatementNode) GetChildren() []Node {
	return nil
}

func (n *IfStatementNode) SetChildren(children []Node) {
	panic("IfStatementNode does not support adding children")
}

func (n *IfStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *IfStatementNode) GetCondition() Node {
	return n.condition
}

func (n *IfStatementNode) SetCondition(condition Node) {
	if condition != nil {
		condition.SetParent(n)
	}
	n.condition = condition
}

func (n *IfStatementNode) GetTrueStatement() Node {
	return n.trueStatement
}

func (n *IfStatementNode) SetTrueStatement(statement Node) {
	if statement != nil {
		statement.SetParent(n)
	}
	n.trueStatement = statement
}

func (n *IfStatementNode) GetElseStatement() Node {
	return n.elseStatement
}

func (n *IfStatementNode) SetElseStatement(statement Node) {
	if statement != nil {
		statement.SetParent(n)
	}
	n.elseStatement = statement
}

func (n *IfStatementNode) ToString() string {
	elseStr := ""
	if n.elseStatement != nil {
		elseStr = fmt.Sprintf(" else %s", n.elseStatement.ToString())
	}
	return fmt.Sprintf("IfStatement(%s) %s%s", n.condition.ToString(), n.trueStatement.ToString(), elseStr)
}
