package ast

import "fmt"

type YieldExpressionNode struct {
	Generator bool

	parent     Node
	expression Node
}

func NewYieldExpressionNode(expression Node, generator bool) *YieldExpressionNode {
	newNode := &YieldExpressionNode{
		Generator: generator,
	}
	newNode.SetExpression(expression)
	return newNode
}

func (n *YieldExpressionNode) GetNodeType() NodeType {
	return YieldExpression
}

func (n *YieldExpressionNode) GetParent() Node {
	return n.parent
}

func (n *YieldExpressionNode) GetChildren() []Node {
	return nil
}

func (n *YieldExpressionNode) SetChildren(children []Node) {
	panic("YieldExpressionNode does not support adding children")
}

func (n *YieldExpressionNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *YieldExpressionNode) GetExpression() Node {
	return n.expression
}

func (n *YieldExpressionNode) SetExpression(expression Node) {
	if expression != nil {
		expression.SetParent(n)
	}
	n.expression = expression
}

func (n *YieldExpressionNode) ToString() string {
	if n.expression == nil {
		return "YieldExpression()"
	}

	generator := ""
	if n.Generator {
		generator = "*"
	}

	return fmt.Sprintf("YieldExpression(%s%s)", generator, n.expression.ToString())
}
