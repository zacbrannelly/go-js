package ast

type ClassStaticBlockNode struct {
	Parent   Node
	Children []Node
	Body     Node
}

func (n *ClassStaticBlockNode) GetNodeType() NodeType {
	return ClassStaticBlock
}

func (n *ClassStaticBlockNode) GetParent() Node {
	return n.Parent
}

func (n *ClassStaticBlockNode) GetChildren() []Node {
	return n.Children
}

func (n *ClassStaticBlockNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ClassStaticBlockNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *ClassStaticBlockNode) ToString() string {
	body := ""
	if n.Body != nil {
		body = n.Body.ToString()
	}
	return "ClassStaticBlock { " + body + " }"
}
