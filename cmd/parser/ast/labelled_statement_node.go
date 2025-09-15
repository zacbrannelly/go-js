package ast

import "fmt"

type LabelledStatementNode struct {
	Parent       Node
	Children     []Node
	Label        Node
	LabelledItem Node
}

func (n *LabelledStatementNode) GetNodeType() NodeType {
	return LabelledStatement
}

func (n *LabelledStatementNode) GetParent() Node {
	return n.Parent
}

func (n *LabelledStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *LabelledStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *LabelledStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *LabelledStatementNode) ToString() string {
	return fmt.Sprintf("LabelledStatement(%s) %s", n.Label.ToString(), n.LabelledItem.ToString())
}
