package ast

type ObjectLiteralNode struct {
	Parent     Node
	Children   []Node
	Properties []Node
}

func (n *ObjectLiteralNode) GetNodeType() NodeType {
	return ObjectLiteral
}

func (n *ObjectLiteralNode) GetChildren() []Node {
	return n.Children
}

func (n *ObjectLiteralNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ObjectLiteralNode) GetParent() Node {
	return n.Parent
}

func (n *ObjectLiteralNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ObjectLiteralNode) ToString() string {
	return "ObjectLiteral"
}

func (n *ObjectLiteralNode) String() string {
	// TODO: Implement this
	return "ObjectLiteral"
}
