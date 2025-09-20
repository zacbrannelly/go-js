package ast

import (
	"fmt"
)

type ForStatementNode struct {
	parent      Node
	initializer Node
	condition   Node
	update      Node
	body        Node
}

func NewForStatementNode(initializer Node, condition Node, update Node, body Node) *ForStatementNode {
	newNode := &ForStatementNode{}
	newNode.SetInitializer(initializer)
	newNode.SetCondition(condition)
	newNode.SetUpdate(update)
	newNode.SetBody(body)
	return newNode
}

func (n *ForStatementNode) GetNodeType() NodeType {
	return ForStatement
}

func (n *ForStatementNode) GetParent() Node {
	return n.parent
}

func (n *ForStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ForStatementNode) GetChildren() []Node {
	return nil
}

func (n *ForStatementNode) SetChildren(children []Node) {
	panic("ForStatementNode does not support adding children")
}

func (n *ForStatementNode) GetInitializer() Node {
	return n.initializer
}

func (n *ForStatementNode) SetInitializer(initializer Node) {
	if initializer != nil {
		initializer.SetParent(n)
	}
	n.initializer = initializer
}

func (n *ForStatementNode) GetCondition() Node {
	return n.condition
}

func (n *ForStatementNode) SetCondition(condition Node) {
	if condition != nil {
		condition.SetParent(n)
	}
	n.condition = condition
}

func (n *ForStatementNode) GetUpdate() Node {
	return n.update
}

func (n *ForStatementNode) SetUpdate(update Node) {
	if update != nil {
		update.SetParent(n)
	}
	n.update = update
}

func (n *ForStatementNode) GetBody() Node {
	return n.body
}

func (n *ForStatementNode) SetBody(body Node) {
	if body != nil {
		body.SetParent(n)
	}
	n.body = body
}

func (n *ForStatementNode) ToString() string {
	var init, cond, update string

	if n.initializer != nil {
		init = n.initializer.ToString()
	}

	if n.condition != nil {
		cond = n.condition.ToString()
	}

	if n.update != nil {
		update = n.update.ToString()
	}

	return fmt.Sprintf("ForStatement(%s; %s; %s) %s",
		init,
		cond,
		update,
		n.body.ToString())
}
