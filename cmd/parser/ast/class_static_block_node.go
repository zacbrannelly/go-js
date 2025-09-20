package ast

type ClassStaticBlockNode struct {
	parent Node
	body   Node
}

func NewClassStaticBlockNode(body Node) *ClassStaticBlockNode {
	newNode := &ClassStaticBlockNode{}
	newNode.SetBody(body)
	return newNode
}

func (n *ClassStaticBlockNode) GetNodeType() NodeType {
	return ClassStaticBlock
}

func (n *ClassStaticBlockNode) GetParent() Node {
	return n.parent
}

func (n *ClassStaticBlockNode) GetChildren() []Node {
	return nil
}

func (n *ClassStaticBlockNode) SetChildren(children []Node) {
	panic("ClassStaticBlockNode does not support adding children")
}

func (n *ClassStaticBlockNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *ClassStaticBlockNode) GetBody() Node {
	return n.body
}

func (n *ClassStaticBlockNode) SetBody(body Node) {
	if body != nil {
		body.SetParent(n)
	}
	n.body = body
}

func (n *ClassStaticBlockNode) ToString() string {
	body := ""
	if n.body != nil {
		body = n.body.ToString()
	}
	return "ClassStaticBlock { " + body + " }"
}
