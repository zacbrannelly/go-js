package ast

import "fmt"

type ContinueStatementNode struct {
	Parent   Node
	Children []Node
	Label    Node
}

func (n *ContinueStatementNode) GetNodeType() NodeType {
	return ContinueStatement
}

func (n *ContinueStatementNode) GetParent() Node {
	return n.Parent
}

func (n *ContinueStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *ContinueStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ContinueStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ContinueStatementNode) ToString() string {
	if n.Label != nil {
		return fmt.Sprintf("ContinueStatement(%s)", n.Label.ToString())
	}
	return "ContinueStatement"
}
