package ast

type ObjectLiteralNode struct {
	BasicNode
	Properties []Node
}

func (n *ObjectLiteralNode) Type() NodeType {
	return ObjectLiteral
}

func (n *ObjectLiteralNode) String() string {
	// TODO: Implement this
	return "ObjectLiteral"
}
