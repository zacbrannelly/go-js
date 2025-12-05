package ast

type ScriptNode struct {
	Parent   Node
	Children []Node
}

func (n *ScriptNode) GetNodeType() NodeType {
	return Script
}

func (n *ScriptNode) GetParent() Node {
	return n.Parent
}

func (n *ScriptNode) GetChildren() []Node {
	return n.Children
}

func (n *ScriptNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ScriptNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ScriptNode) IsComposable() bool {
	return true
}

func (n *ScriptNode) ToString() string {
	return "Script"
}
