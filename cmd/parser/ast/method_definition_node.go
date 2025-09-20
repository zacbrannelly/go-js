package ast

import (
	"fmt"
	"strings"
)

type MethodDefinitionNode struct {
	Generator bool
	Async     bool
	Getter    bool
	Setter    bool
	Static    bool

	parent     Node
	name       Node
	parameters []Node
	body       Node
}

func NewMethodDefinitionNode(name Node, parameters []Node, body Node) *MethodDefinitionNode {
	newNode := &MethodDefinitionNode{}
	newNode.SetName(name)
	newNode.SetParameters(parameters)
	newNode.SetBody(body)
	return newNode
}

func NewMethodDefinitionNodeForGetter(name Node, parameters []Node, body Node) *MethodDefinitionNode {
	newNode := NewMethodDefinitionNode(name, parameters, body)
	newNode.Getter = true
	return newNode
}

func NewMethodDefinitionNodeForSetter(name Node, parameters []Node, body Node) *MethodDefinitionNode {
	newNode := NewMethodDefinitionNode(name, parameters, body)
	newNode.Setter = true
	return newNode
}

func (n *MethodDefinitionNode) GetNodeType() NodeType {
	return MethodDefinition
}

func (n *MethodDefinitionNode) GetParent() Node {
	return n.parent
}

func (n *MethodDefinitionNode) GetChildren() []Node {
	return nil
}

func (n *MethodDefinitionNode) SetChildren(children []Node) {
	panic("MethodDefinitionNode does not support adding children")
}

func (n *MethodDefinitionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *MethodDefinitionNode) GetName() Node {
	return n.name
}

func (n *MethodDefinitionNode) SetName(name Node) {
	if name != nil {
		name.SetParent(n)
	}
	n.name = name
}

func (n *MethodDefinitionNode) GetParameters() []Node {
	return n.parameters
}

func (n *MethodDefinitionNode) SetParameters(parameters []Node) {
	for _, param := range parameters {
		if param != nil {
			param.SetParent(n)
		}
	}
	n.parameters = parameters
}

func (n *MethodDefinitionNode) GetBody() Node {
	return n.body
}

func (n *MethodDefinitionNode) SetBody(body Node) {
	if body != nil {
		body.SetParent(n)
	}
	n.body = body
}

func (n *MethodDefinitionNode) ToString() string {
	parameters := []string{}
	for _, parameter := range n.parameters {
		parameters = append(parameters, parameter.ToString())
	}

	body := ""
	if n.body != nil {
		body = n.body.ToString()
	}

	static := ""
	if n.Static {
		static = "static "
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
		"MethodDefinition(%s%s%s(%s) { %s })",
		static,
		modifier,
		n.name.ToString(),
		strings.Join(parameters, ", "),
		body,
	)
}
