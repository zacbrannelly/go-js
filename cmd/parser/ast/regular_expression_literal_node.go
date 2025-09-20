package ast

import "fmt"

type RegularExpressionLiteralNode struct {
	PatternAndFlags string

	parent Node
}

func NewRegularExpressionLiteralNode(patternAndFlags string) *RegularExpressionLiteralNode {
	return &RegularExpressionLiteralNode{
		PatternAndFlags: patternAndFlags,
	}
}

func (n *RegularExpressionLiteralNode) GetNodeType() NodeType {
	return RegularExpressionLiteral
}

func (n *RegularExpressionLiteralNode) GetParent() Node {
	return n.parent
}

func (n *RegularExpressionLiteralNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *RegularExpressionLiteralNode) GetChildren() []Node {
	return nil
}

func (n *RegularExpressionLiteralNode) SetChildren(children []Node) {
	panic("RegularExpressionLiteralNode does not support adding children")
}

func (n *RegularExpressionLiteralNode) ToString() string {
	return fmt.Sprintf("RegularExpressionLiteral(%s)", n.PatternAndFlags)
}
