package ast

import (
	"fmt"
	"slices"
	"strings"
)

type FunctionExpressionNode struct {
	parent     Node
	name       Node
	parameters []Node
	body       Node
	Generator  bool
	Async      bool
	Arrow      bool
}

func NewFunctionExpressionNode(name Node, parameters []Node, body Node) *FunctionExpressionNode {
	newNode := &FunctionExpressionNode{}
	newNode.SetName(name)
	newNode.SetParameters(parameters)
	newNode.SetBody(body)
	return newNode
}

func NewFunctionExpressionNodeForArrowFunc(parameters []Node, body Node) *FunctionExpressionNode {
	newNode := &FunctionExpressionNode{}
	newNode.SetParameters(parameters)
	newNode.SetBody(body)
	newNode.Arrow = true
	return newNode
}

func (n *FunctionExpressionNode) GetNodeType() NodeType {
	return FunctionExpression
}

func (n *FunctionExpressionNode) GetParent() Node {
	return n.parent
}

func (n *FunctionExpressionNode) GetChildren() []Node {
	children := make([]Node, len(n.parameters))
	copy(children, n.parameters)
	children = append(children, n.name, n.body)

	return slices.DeleteFunc(children, func(n Node) bool {
		return n == nil
	})
}

func (n *FunctionExpressionNode) SetChildren(children []Node) {
	panic("FunctionExpressionNode does not support adding children")
}

func (n *FunctionExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *FunctionExpressionNode) GetName() Node {
	return n.name
}

func (n *FunctionExpressionNode) SetName(name Node) {
	if name != nil {
		name.SetParent(n)
	}
	n.name = name
}

func (n *FunctionExpressionNode) GetParameters() []Node {
	return n.parameters
}

func (n *FunctionExpressionNode) SetParameters(parameters []Node) {
	for _, param := range parameters {
		if param != nil {
			param.SetParent(n)
		}
	}
	n.parameters = parameters
}

func (n *FunctionExpressionNode) GetBody() Node {
	return n.body
}

func (n *FunctionExpressionNode) SetBody(body Node) {
	if body != nil {
		body.SetParent(n)
	}
	n.body = body
}

func (n *FunctionExpressionNode) IsComposable() bool {
	return false
}

func (n *FunctionExpressionNode) ToString() string {
	async := ""
	if n.Async {
		async = "async "
	}

	generator := ""
	if n.Generator {
		generator = "*"
	}

	name := ""
	if n.name != nil {
		name = n.name.ToString()
	}

	parameters := []string{}
	for _, parameter := range n.parameters {
		parameters = append(parameters, parameter.ToString())
	}

	body := ""
	if n.body != nil {
		body = n.body.ToString()
	}

	arrow := " "
	if n.Arrow {
		arrow = " => "
	}

	return fmt.Sprintf("FunctionExpression(%s%s%s(%s)%s{ %s })", async, generator, name, strings.Join(parameters, ", "), arrow, body)
}
