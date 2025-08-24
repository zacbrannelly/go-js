package ast

type ExpressionNode struct {
	Parent   Node
	Children []Node
}

func (n *ExpressionNode) GetNodeType() NodeType {
	return Expression
}

func (n *ExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *ExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *ExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ExpressionNode) ToString() string {
	return "Expression"
}
