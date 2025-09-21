package ast

import "fmt"

type ContinueStatementNode struct {
	parent Node
	label  Node
}

func NewContinueStatementNode(label Node) *ContinueStatementNode {
	newNode := &ContinueStatementNode{}
	newNode.SetLabel(label)
	return newNode
}

func (n *ContinueStatementNode) GetNodeType() NodeType {
	return ContinueStatement
}

func (n *ContinueStatementNode) GetParent() Node {
	return n.parent
}

func (n *ContinueStatementNode) GetChildren() []Node {
	if n.label != nil {
		return []Node{n.label}
	}
	return nil
}

func (n *ContinueStatementNode) SetChildren(children []Node) {
	panic("ContinueStatementNode does not support adding children")
}

func (n *ContinueStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ContinueStatementNode) GetLabel() Node {
	return n.label
}

func (n *ContinueStatementNode) SetLabel(label Node) {
	if label != nil {
		label.SetParent(n)
	}
	n.label = label
}

func (n *ContinueStatementNode) IsComposable() bool {
	return false
}

func (n *ContinueStatementNode) ToString() string {
	if n.label != nil {
		return fmt.Sprintf("ContinueStatement(%s)", n.label.ToString())
	}
	return "ContinueStatement"
}
