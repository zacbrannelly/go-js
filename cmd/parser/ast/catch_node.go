package ast

import (
	"fmt"
)

type CatchNode struct {
	Parent   Node
	Children []Node
	Target   Node
	Block    Node
}

func (n *CatchNode) GetNodeType() NodeType {
	return Catch
}

func (n *CatchNode) GetParent() Node {
	return n.Parent
}

func (n *CatchNode) GetChildren() []Node {
	return n.Children
}

func (n *CatchNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *CatchNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *CatchNode) ToString() string {
	targetStr := ""
	if n.Target != nil {
		targetStr = fmt.Sprintf("(%s)", n.Target.ToString())
	}
	return fmt.Sprintf("Catch%s %s", targetStr, n.Block.ToString())
}
