package ast

import "fmt"

type BreakStatementNode struct {
	Parent   Node
	Children []Node
	Label    Node
}

func (n *BreakStatementNode) GetNodeType() NodeType {
	return BreakStatement
}

func (n *BreakStatementNode) GetParent() Node {
	return n.Parent
}

func (n *BreakStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *BreakStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BreakStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BreakStatementNode) ToString() string {
	if n.Label != nil {
		return fmt.Sprintf("BreakStatement(%s)", n.Label.ToString())
	}
	return "BreakStatement"
}
