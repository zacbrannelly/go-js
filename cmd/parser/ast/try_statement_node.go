package ast

import (
	"fmt"
	"slices"
)

type TryStatementNode struct {
	parent  Node
	block   Node
	catch   Node
	finally Node
}

func NewTryStatementNode(block Node, catch Node, finally Node) *TryStatementNode {
	newNode := &TryStatementNode{}
	newNode.SetBlock(block)
	newNode.SetCatch(catch)
	newNode.SetFinally(finally)
	return newNode
}

func (n *TryStatementNode) GetNodeType() NodeType {
	return TryStatement
}

func (n *TryStatementNode) GetParent() Node {
	return n.parent
}

func (n *TryStatementNode) GetChildren() []Node {
	return slices.DeleteFunc([]Node{n.block, n.catch, n.finally}, func(n Node) bool {
		return n == nil
	})
}

func (n *TryStatementNode) SetChildren(children []Node) {
	panic("TryStatementNode does not support adding children")
}

func (n *TryStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *TryStatementNode) GetBlock() Node {
	return n.block
}

func (n *TryStatementNode) SetBlock(block Node) {
	if block != nil {
		block.SetParent(n)
	}
	n.block = block
}

func (n *TryStatementNode) GetCatch() Node {
	return n.catch
}

func (n *TryStatementNode) SetCatch(catch Node) {
	if catch != nil {
		catch.SetParent(n)
	}
	n.catch = catch
}

func (n *TryStatementNode) GetFinally() Node {
	return n.finally
}

func (n *TryStatementNode) SetFinally(finally Node) {
	if finally != nil {
		finally.SetParent(n)
	}
	n.finally = finally
}

func (n *TryStatementNode) IsComposable() bool {
	return false
}

func (n *TryStatementNode) ToString() string {
	catchStr := ""
	if n.catch != nil {
		catchStr = " " + n.catch.ToString()
	}
	finallyStr := ""
	if n.finally != nil {
		finallyStr = fmt.Sprintf(" Finally(%s)", n.finally.ToString())
	}
	return fmt.Sprintf("TryStatement(%s%s%s)", n.block.ToString(), catchStr, finallyStr)
}
