package ast

import "fmt"

type ReturnStatementNode struct {
	Parent   Node
	Children []Node
	Value    Node
}

func (n *ReturnStatementNode) GetNodeType() NodeType {
	return ReturnStatement
}

func (n *ReturnStatementNode) GetParent() Node {
	return n.Parent
}

func (n *ReturnStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *ReturnStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ReturnStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ReturnStatementNode) ToString() string {
	if n.Value != nil {
		return fmt.Sprintf("ReturnStatement(%s)", n.Value.ToString())
	}
	return "ReturnStatement(undefined)"
}
