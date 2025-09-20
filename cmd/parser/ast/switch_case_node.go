package ast

import (
	"fmt"
)

type SwitchCaseNode struct {
	parent     Node
	expression Node
}

func NewSwitchCaseNode(expression Node) *SwitchCaseNode {
	newNode := &SwitchCaseNode{}
	newNode.SetExpression(expression)
	return newNode
}

func (n *SwitchCaseNode) GetNodeType() NodeType {
	return SwitchCase
}

func (n *SwitchCaseNode) GetParent() Node {
	return n.parent
}

func (n *SwitchCaseNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *SwitchCaseNode) GetChildren() []Node {
	return nil
}

func (n *SwitchCaseNode) SetChildren(children []Node) {
	panic("SwitchCaseNode does not support adding children")
}

func (n *SwitchCaseNode) GetExpression() Node {
	return n.expression
}

func (n *SwitchCaseNode) SetExpression(expression Node) {
	if expression != nil {
		expression.SetParent(n)
	}
	n.expression = expression
}

func (n *SwitchCaseNode) ToString() string {
	return fmt.Sprintf("SwitchCase(%s)", n.expression.ToString())
}
