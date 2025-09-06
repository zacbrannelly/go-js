package ast

import "fmt"

type PropertyDefinitionNode struct {
	Parent   Node
	Children []Node
	Key      Node
	Value    Node
	Computed bool
	Static   bool
}

func (n *PropertyDefinitionNode) GetNodeType() NodeType {
	return PropertyDefinition
}

func (n *PropertyDefinitionNode) GetParent() Node {
	return n.Parent
}

func (n *PropertyDefinitionNode) GetChildren() []Node {
	return n.Children
}

func (n *PropertyDefinitionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *PropertyDefinitionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (node *PropertyDefinitionNode) ToString() string {
	static := ""
	if node.Static {
		static = "static "
	}

	if node.Value == nil {
		return fmt.Sprintf("PropertyDefinition(%s%s)", static, node.Key.ToString())
	}

	return fmt.Sprintf("PropertyDefinition(%s%s: %s)", static, node.Key.ToString(), node.Value.ToString())
}
