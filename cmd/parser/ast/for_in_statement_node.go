package ast

import (
	"fmt"
)

type ForInStatementNode struct {
	Parent   Node
	Children []Node
	Target   Node
	Iterable Node
	Body     Node
}

func (n *ForInStatementNode) GetNodeType() NodeType {
	return ForInStatement
}

func (n *ForInStatementNode) GetParent() Node {
	return n.Parent
}

func (n *ForInStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *ForInStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ForInStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ForInStatementNode) ToString() string {
	return fmt.Sprintf("ForInStatement(%s in %s) %s",
		n.Target.ToString(),
		n.Iterable.ToString(),
		n.Body.ToString())
}
