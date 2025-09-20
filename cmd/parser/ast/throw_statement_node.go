package ast

import "fmt"

type ThrowStatementNode struct {
	parent     Node
	expression Node
}

func NewThrowStatementNode(expression Node) *ThrowStatementNode {
	newNode := &ThrowStatementNode{}
	newNode.SetExpression(expression)
	return newNode
}

func (n *ThrowStatementNode) GetNodeType() NodeType {
	return ThrowStatement
}

func (n *ThrowStatementNode) GetParent() Node {
	return n.parent
}

func (n *ThrowStatementNode) GetChildren() []Node {
	return nil
}

func (n *ThrowStatementNode) SetChildren(children []Node) {
	panic("ThrowStatementNode does not support adding children")
}

func (n *ThrowStatementNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ThrowStatementNode) GetExpression() Node {
	return n.expression
}

func (n *ThrowStatementNode) SetExpression(expression Node) {
	if expression != nil {
		expression.SetParent(n)
	}
	n.expression = expression
}

func (n *ThrowStatementNode) ToString() string {
	return fmt.Sprintf("ThrowStatement(%s)", n.expression.ToString())
}
