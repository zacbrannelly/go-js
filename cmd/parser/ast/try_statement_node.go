package ast

import (
	"fmt"
)

type TryStatementNode struct {
	Parent   Node
	Children []Node
	Block    Node
	Catch    Node
	Finally  Node
}

func (n *TryStatementNode) GetNodeType() NodeType {
	return TryStatement
}

func (n *TryStatementNode) GetParent() Node {
	return n.Parent
}

func (n *TryStatementNode) GetChildren() []Node {
	return n.Children
}

func (n *TryStatementNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *TryStatementNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *TryStatementNode) ToString() string {
	catchStr := ""
	if n.Catch != nil {
		catchStr = " " + n.Catch.ToString()
	}
	finallyStr := ""
	if n.Finally != nil {
		finallyStr = fmt.Sprintf(" Finally(%s)", n.Finally.ToString())
	}
	return fmt.Sprintf("TryStatement(%s%s%s)", n.Block.ToString(), catchStr, finallyStr)
}
