package ast

import "fmt"

type SpreadElementNode struct {
	Parent     Node
	Children   []Node
	Expression Node
}

func (n *SpreadElementNode) GetNodeType() NodeType {
	return SpreadElement
}

func (n *SpreadElementNode) GetParent() Node {
	return n.Parent
}

func (n *SpreadElementNode) GetChildren() []Node {
	return n.Children
}

func (n *SpreadElementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *SpreadElementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *SpreadElementNode) ToString() string {
	return fmt.Sprintf("SpreadElement(%s)", n.Expression.ToString())
}
