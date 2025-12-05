package ast

import "fmt"

type LabelledStatementNode struct {
	parent       Node
	label        Node
	labelledItem Node
}

func NewLabelledStatementNode(label Node, labelledItem Node) *LabelledStatementNode {
	newNode := &LabelledStatementNode{}
	newNode.SetLabel(label)
	newNode.SetLabelledItem(labelledItem)
	return newNode
}

func (n *LabelledStatementNode) GetNodeType() NodeType {
	return LabelledStatement
}

func (n *LabelledStatementNode) GetParent() Node {
	return n.parent
}

func (n *LabelledStatementNode) GetChildren() []Node {
	return []Node{n.label, n.labelledItem}
}

func (n *LabelledStatementNode) SetChildren(children []Node) {
	panic("LabelledStatementNode does not support adding children")
}

func (n *LabelledStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *LabelledStatementNode) GetLabel() Node {
	return n.label
}

func (n *LabelledStatementNode) SetLabel(label Node) {
	if label != nil {
		label.SetParent(n)
	}
	n.label = label
}

func (n *LabelledStatementNode) GetLabelledItem() Node {
	return n.labelledItem
}

func (n *LabelledStatementNode) SetLabelledItem(labelledItem Node) {
	if labelledItem != nil {
		labelledItem.SetParent(n)
	}
	n.labelledItem = labelledItem
}

func (n *LabelledStatementNode) IsComposable() bool {
	return false
}

func (n *LabelledStatementNode) ToString() string {
	return fmt.Sprintf("LabelledStatement(%s) %s", n.label.ToString(), n.labelledItem.ToString())
}
