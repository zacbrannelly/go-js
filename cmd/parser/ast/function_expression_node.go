package ast

import (
	"fmt"
	"strings"
)

type FunctionExpressionNode struct {
	Parent     Node
	Children   []Node
	Name       Node
	Parameters []Node
	Body       Node
	Generator  bool
	Async      bool
	Arrow      bool
}

func (n *FunctionExpressionNode) GetNodeType() NodeType {
	return FunctionExpression
}

func (n *FunctionExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *FunctionExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *FunctionExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *FunctionExpressionNode) SetParent(parent Node) {
	n.Parent = parent
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
	if n.Name != nil {
		name = n.Name.ToString()
	}

	parameters := []string{}
	for _, parameter := range n.Parameters {
		parameters = append(parameters, parameter.ToString())
	}

	body := ""
	if n.Body != nil {
		body = n.Body.ToString()
	}

	arrow := " "
	if n.Arrow {
		arrow = " => "
	}

	return fmt.Sprintf("FunctionExpression(%s%s%s(%s)%s{ %s })", async, generator, name, strings.Join(parameters, ", "), arrow, body)
}
