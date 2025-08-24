package ast

import "fmt"

type SpreadElementNode struct {
	BasicNode
	Expression Node
}

func (n *SpreadElementNode) Type() NodeType {
	return SpreadElement
}

func (n *SpreadElementNode) String() string {
	return fmt.Sprintf("SpreadElement(%s)", n.Expression.ToString())
}
