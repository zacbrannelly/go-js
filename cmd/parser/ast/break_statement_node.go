package ast

import "fmt"

type BreakStatementNode struct {
	parent Node
	label  Node
}

func NewBreakStatementNode(label Node) *BreakStatementNode {
	newNode := &BreakStatementNode{}
	newNode.SetLabel(label)
	return newNode
}

func (n *BreakStatementNode) GetNodeType() NodeType {
	return BreakStatement
}

func (n *BreakStatementNode) GetParent() Node {
	return n.parent
}

func (n *BreakStatementNode) GetChildren() []Node {
	if n.label != nil {
		return []Node{n.label}
	}
	return nil
}

func (n *BreakStatementNode) SetChildren(children []Node) {
	panic("BreakStatementNode does not support adding children")
}

func (n *BreakStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BreakStatementNode) GetLabel() Node {
	return n.label
}

func (n *BreakStatementNode) SetLabel(label Node) {
	if label != nil {
		label.SetParent(n)
	}
	n.label = label
}

func (n *BreakStatementNode) IsComposable() bool {
	return false
}

func (n *BreakStatementNode) ToString() string {
	if n.label != nil {
		return fmt.Sprintf("BreakStatement(%s)", n.label.ToString())
	}
	return "BreakStatement"
}
