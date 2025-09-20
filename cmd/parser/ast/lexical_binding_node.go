package ast

import "fmt"

type LexicalBindingNode struct {
	Const bool

	parent      Node
	target      Node
	initializer Node
}

func NewLexicalBindingNode(target Node, initializer Node, isConst bool) *LexicalBindingNode {
	newNode := &LexicalBindingNode{}
	newNode.SetTarget(target)
	newNode.SetInitializer(initializer)
	newNode.Const = isConst
	return newNode
}

func (n *LexicalBindingNode) GetNodeType() NodeType {
	return LexicalBinding
}

func (n *LexicalBindingNode) GetParent() Node {
	return n.parent
}

func (n *LexicalBindingNode) GetChildren() []Node {
	return nil
}

func (n *LexicalBindingNode) SetChildren(children []Node) {
	panic("LexicalBindingNode does not support adding children")
}

func (n *LexicalBindingNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *LexicalBindingNode) GetTarget() Node {
	return n.target
}

func (n *LexicalBindingNode) SetTarget(target Node) {
	if target != nil {
		target.SetParent(n)
	}
	n.target = target
}

func (n *LexicalBindingNode) GetInitializer() Node {
	return n.initializer
}

func (n *LexicalBindingNode) SetInitializer(initializer Node) {
	if initializer != nil {
		initializer.SetParent(n)
	}
	n.initializer = initializer
}

func (n *LexicalBindingNode) ToString() string {
	keyword := "let"
	if n.Const {
		keyword = "const"
	}

	if n.initializer == nil {
		return fmt.Sprintf("LexicalBinding(%s %s)", keyword, n.target.ToString())
	}

	return fmt.Sprintf("LexicalBinding(%s %s = %s)", keyword, n.target.ToString(), n.initializer.ToString())
}
