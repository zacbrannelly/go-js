package ast

import "fmt"

type ReturnStatementNode struct {
	parent Node
	value  Node
}

func NewReturnStatementNode(value Node) *ReturnStatementNode {
	newNode := &ReturnStatementNode{}
	newNode.SetValue(value)
	return newNode
}

func (n *ReturnStatementNode) GetNodeType() NodeType {
	return ReturnStatement
}

func (n *ReturnStatementNode) GetParent() Node {
	return n.parent
}

func (n *ReturnStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ReturnStatementNode) GetChildren() []Node {
	if n.value != nil {
		return []Node{n.value}
	}
	return nil
}

func (n *ReturnStatementNode) SetChildren(children []Node) {
	panic("ReturnStatementNode does not support adding children")
}

func (n *ReturnStatementNode) GetValue() Node {
	return n.value
}

func (n *ReturnStatementNode) SetValue(value Node) {
	if value != nil {
		value.SetParent(n)
	}
	n.value = value
}

func (n *ReturnStatementNode) IsComposable() bool {
	return false
}

func (n *ReturnStatementNode) ToString() string {
	if n.value != nil {
		return fmt.Sprintf("ReturnStatement(%s)", n.value.ToString())
	}
	return "ReturnStatement(undefined)"
}
