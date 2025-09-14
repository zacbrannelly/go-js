package ast

import (
	"fmt"
)

type WithStatementNode struct {
	Parent     Node
	Children   []Node
	Expression Node
	Body       Node
}

func (n *WithStatementNode) GetNodeType() NodeType {
	return WithStatement
}

func (n *WithStatementNode) GetParent() Node {
	return n.Parent
}

func (n *WithStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *WithStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *WithStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *WithStatementNode) ToString() string {
	return fmt.Sprintf("WithStatement(%s) %s", n.Expression.ToString(), n.Body.ToString())
}
