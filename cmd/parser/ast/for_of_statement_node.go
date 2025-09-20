package ast

import (
	"fmt"
)

type ForOfStatementNode struct {
	Await bool

	parent   Node
	target   Node
	iterable Node
	body     Node
}

func NewForOfStatementNode(target Node, iterable Node, body Node) *ForOfStatementNode {
	newNode := &ForOfStatementNode{}
	newNode.SetTarget(target)
	newNode.SetIterable(iterable)
	newNode.SetBody(body)
	return newNode
}

func (n *ForOfStatementNode) GetNodeType() NodeType {
	return ForOfStatement
}

func (n *ForOfStatementNode) GetParent() Node {
	return n.parent
}

func (n *ForOfStatementNode) GetChildren() []Node {
	return nil
}

func (n *ForOfStatementNode) SetChildren(children []Node) {
	panic("ForOfStatementNode does not support adding children")
}

func (n *ForOfStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ForOfStatementNode) GetTarget() Node {
	return n.target
}

func (n *ForOfStatementNode) SetTarget(target Node) {
	if target != nil {
		target.SetParent(n)
	}
	n.target = target
}

func (n *ForOfStatementNode) GetIterable() Node {
	return n.iterable
}

func (n *ForOfStatementNode) SetIterable(iterable Node) {
	if iterable != nil {
		iterable.SetParent(n)
	}
	n.iterable = iterable
}

func (n *ForOfStatementNode) GetBody() Node {
	return n.body
}

func (n *ForOfStatementNode) SetBody(body Node) {
	if body != nil {
		body.SetParent(n)
	}
	n.body = body
}

func (n *ForOfStatementNode) ToString() string {
	var await string

	if n.Await {
		await = "await "
	}

	return fmt.Sprintf("ForOfStatement(%s%s of %s) %s",
		await,
		n.target.ToString(),
		n.iterable.ToString(),
		n.body.ToString())
}
