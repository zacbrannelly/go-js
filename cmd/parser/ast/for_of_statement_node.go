package ast

import (
	"fmt"
)

type ForOfStatementNode struct {
	Parent   Node
	Children []Node
	Target   Node
	Iterable Node
	Body     Node
	Await    bool
}

func (n *ForOfStatementNode) GetNodeType() NodeType {
	return ForOfStatement
}

func (n *ForOfStatementNode) GetParent() Node {
	return n.Parent
}

func (n *ForOfStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *ForOfStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ForOfStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ForOfStatementNode) ToString() string {
	var await string

	if n.Await {
		await = "await "
	}

	return fmt.Sprintf("ForOfStatement(%s%s of %s) %s",
		await,
		n.Target.ToString(),
		n.Iterable.ToString(),
		n.Body.ToString())
}
