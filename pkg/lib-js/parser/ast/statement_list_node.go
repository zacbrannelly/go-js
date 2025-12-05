package ast

type StatementListNode struct {
	Parent   Node
	Children []Node
}

func (n *StatementListNode) GetNodeType() NodeType {
	return StatementList
}

func (n *StatementListNode) GetParent() Node {
	return n.Parent
}

func (n *StatementListNode) GetChildren() []Node {
	return n.Children
}

func (n *StatementListNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *StatementListNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *StatementListNode) IsComposable() bool {
	return true
}

func (n *StatementListNode) ToString() string {
	return "StatementList"
}
