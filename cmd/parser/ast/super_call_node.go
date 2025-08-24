package ast

type SuperCallNode struct {
	Parent    Node
	Children  []Node
	Arguments Node
}

func (n *SuperCallNode) GetNodeType() NodeType {
	return SuperCall
}

func (n *SuperCallNode) GetParent() Node {
	return n.Parent
}

func (n *SuperCallNode) GetChildren() []Node {
	return n.Children
}

func (n *SuperCallNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *SuperCallNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *SuperCallNode) ToString() string {
	// TODO
	return "SuperCall"
}
