package ast

import (
	"fmt"
)

type ForInStatementNode struct {
	parent   Node
	target   Node
	iterable Node
	body     Node
}

func NewForInStatementNode(target Node, iterable Node, body Node) *ForInStatementNode {
	newNode := &ForInStatementNode{}
	newNode.SetTarget(target)
	newNode.SetIterable(iterable)
	newNode.SetBody(body)
	return newNode
}

func (n *ForInStatementNode) GetNodeType() NodeType {
	return ForInStatement
}

func (n *ForInStatementNode) GetParent() Node {
	return n.parent
}

func (n *ForInStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ForInStatementNode) GetChildren() []Node {
	return []Node{n.target, n.iterable, n.body}
}

func (n *ForInStatementNode) SetChildren(children []Node) {
	panic("ForInStatementNode does not support adding children")
}

func (n *ForInStatementNode) GetTarget() Node {
	return n.target
}

func (n *ForInStatementNode) SetTarget(target Node) {
	if target != nil {
		target.SetParent(n)
	}
	n.target = target
}

func (n *ForInStatementNode) GetIterable() Node {
	return n.iterable
}

func (n *ForInStatementNode) SetIterable(iterable Node) {
	if iterable != nil {
		iterable.SetParent(n)
	}
	n.iterable = iterable
}

func (n *ForInStatementNode) GetBody() Node {
	return n.body
}

func (n *ForInStatementNode) SetBody(body Node) {
	if body != nil {
		body.SetParent(n)
	}
	n.body = body
}

func (n *ForInStatementNode) IsComposable() bool {
	return false
}

func (n *ForInStatementNode) ToString() string {
	return fmt.Sprintf("ForInStatement(%s in %s) %s",
		n.target.ToString(),
		n.iterable.ToString(),
		n.body.ToString())
}
