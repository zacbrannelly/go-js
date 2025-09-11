package ast

import "fmt"

type YieldExpressionNode struct {
	Parent     Node
	Children   []Node
	Expression Node
	Generator  bool
}

func (n *YieldExpressionNode) GetNodeType() NodeType {
	return YieldExpression
}

func (n *YieldExpressionNode) GetParent() Node {
	return n.Parent
}

func (n *YieldExpressionNode) GetChildren() []Node {
	return n.Children
}

func (n *YieldExpressionNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *YieldExpressionNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *YieldExpressionNode) ToString() string {
	if n.Expression == nil {
		return "YieldExpression()"
	}

	generator := ""
	if n.Generator {
		generator = "*"
	}

	return fmt.Sprintf("YieldExpression(%s%s)", generator, n.Expression.ToString())
}
