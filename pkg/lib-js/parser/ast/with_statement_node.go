package ast

import (
	"fmt"
)

type WithStatementNode struct {
	parent     Node
	expression Node
	body       Node
}

func NewWithStatementNode(expression Node, body Node) *WithStatementNode {
	newNode := &WithStatementNode{}
	newNode.SetExpression(expression)
	newNode.SetBody(body)
	return newNode
}

func (n *WithStatementNode) GetNodeType() NodeType {
	return WithStatement
}

func (n *WithStatementNode) GetParent() Node {
	return n.parent
}

func (n *WithStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *WithStatementNode) GetChildren() []Node {
	return []Node{n.expression, n.body}
}

func (n *WithStatementNode) SetChildren(children []Node) {
	panic("WithStatementNode does not support adding children")
}

func (n *WithStatementNode) GetExpression() Node {
	return n.expression
}

func (n *WithStatementNode) SetExpression(expression Node) {
	if expression != nil {
		expression.SetParent(n)
	}
	n.expression = expression
}

func (n *WithStatementNode) GetBody() Node {
	return n.body
}

func (n *WithStatementNode) SetBody(body Node) {
	if body != nil {
		body.SetParent(n)
	}
	n.body = body
}

func (n *WithStatementNode) IsComposable() bool {
	return false
}

func (n *WithStatementNode) ToString() string {
	return fmt.Sprintf("WithStatement(%s) %s", n.expression.ToString(), n.body.ToString())
}
