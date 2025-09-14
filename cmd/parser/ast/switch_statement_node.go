package ast

import (
	"fmt"
)

type SwitchStatementNode struct {
	Parent   Node
	Children []Node
	Target   Node
}

func (n *SwitchStatementNode) GetNodeType() NodeType {
	return SwitchStatement
}

func (n *SwitchStatementNode) GetParent() Node {
	return n.Parent
}

func (n *SwitchStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *SwitchStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *SwitchStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *SwitchStatementNode) ToString() string {
	return fmt.Sprintf("SwitchStatement(%s)", n.Target.ToString())
}
