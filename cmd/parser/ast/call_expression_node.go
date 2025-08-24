package ast

type CallExpressionNode struct {
	Parent    Node
	Children  []Node
	Callee    Node
	Arguments Node
}

func (n *CallExpressionNode) GetNodeType() NodeType {
	return CallExpression
}

func (n *CallExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *CallExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *CallExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *CallExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *CallExpressionNode) ToString() string {
	// TODO
	return "CallExpression"
}
