package ast

import (
	"fmt"
	"strings"
)

type MethodDefinitionNode struct {
	Parent     Node
	Children   []Node
	Name       Node
	Parameters []Node
	Body       Node
	Generator  bool
	Async      bool
	Getter     bool
	Setter     bool
}

func (n *MethodDefinitionNode) GetNodeType() NodeType {
	return MethodDefinition
}

func (n *MethodDefinitionNode) GetParent() Node {
	return n.Parent
}

func (n *MethodDefinitionNode) GetChildren() []Node {
	return n.Children
}

func (n *MethodDefinitionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *MethodDefinitionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *MethodDefinitionNode) ToString() string {
	parameters := []string{}
	for _, parameter := range n.Parameters {
		parameters = append(parameters, parameter.ToString())
	}

	body := ""
	if n.Body != nil {
		body = n.Body.ToString()
	}

	modifier := ""
	if n.Generator {
		modifier = "*"
	} else if n.Async {
		modifier = "async "
	} else if n.Getter {
		modifier = "get "
	} else if n.Setter {
		modifier = "set "
	}

	return fmt.Sprintf(
		"MethodDefinition(%s%s(%s) { %s })",
		modifier,
		n.Name.ToString(),
		strings.Join(parameters, ", "),
		body,
	)
}
