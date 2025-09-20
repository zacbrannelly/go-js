package ast

import (
	"fmt"
)

type CatchNode struct {
	parent Node
	target Node
	block  Node
}

func NewCatchNode(target Node, block Node) *CatchNode {
	newNode := &CatchNode{}
	newNode.SetTarget(target)
	newNode.SetBlock(block)
	return newNode
}

func (n *CatchNode) GetNodeType() NodeType {
	return Catch
}

func (n *CatchNode) GetParent() Node {
	return n.parent
}

func (n *CatchNode) GetChildren() []Node {
	return nil
}

func (n *CatchNode) SetChildren(children []Node) {
	panic("CatchNode does not support adding children")
}

func (n *CatchNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *CatchNode) GetTarget() Node {
	return n.target
}

func (n *CatchNode) SetTarget(target Node) {
	if target != nil {
		target.SetParent(n)
	}
	n.target = target
}

func (n *CatchNode) GetBlock() Node {
	return n.block
}

func (n *CatchNode) SetBlock(block Node) {
	if block != nil {
		block.SetParent(n)
	}
	n.block = block
}

func (n *CatchNode) ToString() string {
	targetStr := ""
	if n.target != nil {
		targetStr = fmt.Sprintf("(%s)", n.target.ToString())
	}
	return fmt.Sprintf("Catch%s %s", targetStr, n.block.ToString())
}
