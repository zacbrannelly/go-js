package ast

import (
	"fmt"
)

type ForStatementNode struct {
	Parent      Node
	Children    []Node
	Initializer Node
	Condition   Node
	Update      Node
	Body        Node
}

func (n *ForStatementNode) GetNodeType() NodeType {
	return ForStatement
}

func (n *ForStatementNode) GetParent() Node {
	return n.Parent
}

func (n *ForStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *ForStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ForStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ForStatementNode) ToString() string {
	var init, cond, update string

	if n.Initializer != nil {
		init = n.Initializer.ToString()
	}

	if n.Condition != nil {
		cond = n.Condition.ToString()
	}

	if n.Update != nil {
		update = n.Update.ToString()
	}

	return fmt.Sprintf("ForStatement(%s; %s; %s) %s",
		init,
		cond,
		update,
		n.Body.ToString())
}
