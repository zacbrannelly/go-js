package ast

import "fmt"

type ThrowStatementNode struct {
	Parent     Node
	Children   []Node
	Expression Node
}

func (n *ThrowStatementNode) GetNodeType() NodeType {
	return ThrowStatement
}

func (n *ThrowStatementNode) GetParent() Node {
	return n.Parent
}

func (n *ThrowStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *ThrowStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ThrowStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ThrowStatementNode) ToString() string {
	return fmt.Sprintf("ThrowStatement(%s)", n.Expression.ToString())
}
