package ast

import (
	"fmt"
)

type SwitchStatementNode struct {
	parent   Node
	children []Node
	target   Node
}

func NewSwitchStatementNode(target Node) *SwitchStatementNode {
	newNode := &SwitchStatementNode{
		children: make([]Node, 0),
	}
	newNode.SetTarget(target)
	return newNode
}

func (n *SwitchStatementNode) GetNodeType() NodeType {
	return SwitchStatement
}

func (n *SwitchStatementNode) GetParent() Node {
	return n.parent
}

func (n *SwitchStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *SwitchStatementNode) GetChildren() []Node {
	return n.children
}

func (n *SwitchStatementNode) SetChildren(children []Node) {
	n.children = children
}

func (n *SwitchStatementNode) GetTarget() Node {
	return n.target
}

func (n *SwitchStatementNode) SetTarget(target Node) {
	if target != nil {
		target.SetParent(n)
	}
	n.target = target
}

func (n *SwitchStatementNode) IsComposable() bool {
	return true
}

func (n *SwitchStatementNode) ToString() string {
	return fmt.Sprintf("SwitchStatement(%s)", n.target.ToString())
}
