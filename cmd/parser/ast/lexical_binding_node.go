package ast

import "fmt"

type LexicalBindingNode struct {
	Parent      Node
	Children    []Node
	Target      Node
	Initializer Node
	Const       bool
}

func (n *LexicalBindingNode) GetNodeType() NodeType {
	return LexicalBinding
}

func (n *LexicalBindingNode) GetParent() Node {
	return n.Parent
}

func (n *LexicalBindingNode) GetChildren() []Node {
	return n.Children
}

func (n *LexicalBindingNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *LexicalBindingNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *LexicalBindingNode) ToString() string {
	keyword := "let"
	if n.Const {
		keyword = "const"
	}

	if n.Initializer == nil {
		return fmt.Sprintf("LexicalBinding(%s %s)", keyword, n.Target.ToString())
	}

	return fmt.Sprintf("LexicalBinding(%s %s = %s)", keyword, n.Target.ToString(), n.Initializer.ToString())
}
