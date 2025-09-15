package ast

import "fmt"

type RegularExpressionLiteralNode struct {
	BasicNode
	PatternAndFlags string
}

func (n *RegularExpressionLiteralNode) GetNodeType() NodeType {
	return RegularExpressionLiteral
}

func (n *RegularExpressionLiteralNode) GetParent() Node {
	return n.Parent
}

func (n *RegularExpressionLiteralNode) GetChildren() []Node {
	return n.Children
}

func (n *RegularExpressionLiteralNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *RegularExpressionLiteralNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *RegularExpressionLiteralNode) ToString() string {
	return fmt.Sprintf("RegularExpressionLiteral(%s)", n.PatternAndFlags)
}
