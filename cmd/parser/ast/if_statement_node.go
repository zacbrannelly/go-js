package ast

import (
	"fmt"
)

type IfStatementNode struct {
	Parent        Node
	Children      []Node
	Condition     Node
	TrueStatement Node
	ElseStatement Node
}

func (n *IfStatementNode) GetNodeType() NodeType {
	return IfStatement
}

func (n *IfStatementNode) GetParent() Node {
	return n.Parent
}

func (n *IfStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *IfStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *IfStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *IfStatementNode) ToString() string {
	elseStr := ""
	if n.ElseStatement != nil {
		elseStr = fmt.Sprintf(" else %s", n.ElseStatement.ToString())
	}
	return fmt.Sprintf("IfStatement(%s) %s%s", n.Condition.ToString(), n.TrueStatement.ToString(), elseStr)
}
