package ast

import "fmt"

type ArgumentListItemNode struct {
	Parent     Node
	Children   []Node
	Spread     bool
	Expression Node
}

func (n *ArgumentListItemNode) GetNodeType() NodeType {
	return ArgumentListItem
}

func (n *ArgumentListItemNode) GetParent() Node {
	return n.Parent
}

func (n *ArgumentListItemNode) GetChildren() []Node {
	return n.Children
}

func (n *ArgumentListItemNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ArgumentListItemNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ArgumentListItemNode) ToString() string {
	return fmt.Sprintf("ArgumentListItem(%s, spread: %t)", n.Expression.ToString(), n.Spread)
}
