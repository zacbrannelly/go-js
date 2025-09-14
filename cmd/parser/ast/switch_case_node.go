package ast

import (
	"fmt"
)

type SwitchCaseNode struct {
	Parent     Node
	Children   []Node
	Expression Node
}

func (n *SwitchCaseNode) GetNodeType() NodeType {
	return SwitchCase
}

func (n *SwitchCaseNode) GetParent() Node {
	return n.Parent
}

func (n *SwitchCaseNode) GetChildren() []Node {
	return n.Children
}

func (n *SwitchCaseNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *SwitchCaseNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *SwitchCaseNode) ToString() string {
	return fmt.Sprintf("SwitchCase(%s)", n.Expression.ToString())
}
