package ast

import "zbrannelly.dev/go-js/cmd/lexer"

type NodeType int

const (
	Script NodeType = iota
	Expression
	StatementList
	StatementListItem
	Statement
	Declaration
	EmptyStatement
	DebuggerStatement
	BlockStatement
	Block
	ContinueStatement
	BreakStatement
	VariableStatement
	VariableDeclarationList
	VariableDeclaration
	BindingIdentifier
	Initializer
	BindingPattern
	Identifier
	AssignmentExpression
	ConditionalExpression
	ShortCircuitExpression
	LogicalORExpression
	LogicalANDExpression
	CoalesceExpression
	CoalesceExpressionHead
	BitwiseORExpression
	BitwiseXORExpression
	BitwiseANDExpression
	EqualityExpression
	RelationalExpression
	ShiftExpression
	AdditiveExpression
	MultiplicativeExpression
	ExponentiationExpression
	UnaryExpression
	UpdateExpression
	LeftHandSideExpression
	NewExpression
	CallExpression
	OptionalExpression
	MemberExpression
	CoverCallExpressionAndAsyncArrowHead
	SuperCall
	ImportCall
	ThisExpression
	IdentifierReference
	NullLiteral
	BooleanLiteral
	NumericLiteral
	StringLiteral
	SpreadElement
	ArrayLiteral
	ObjectLiteral
	UndefinedLiteral
	PropertyDefinition
	ObjectBindingPattern
	ArrayBindingPattern
	BindingProperty
	BindingRestProperty
	BindingElement
	MethodDefinition
	FunctionExpression
	ClassExpression
	ClassStaticBlock
	RegularExpressionLiteral
	TemplateLiteral
	YieldExpression
	ImportMeta
	NewTarget
	IfStatement
	DoWhileStatement
	WhileStatement
	LexicalBinding
	LexicalDeclaration
	ForStatement
	ForInStatement
	ForOfStatement
	SwitchStatement
	SwitchCase
	SwitchDefault
	LabelIdentifier
	ReturnStatement
	WithStatement
	LabelledStatement
	ThrowStatement
	TryStatement
	Catch
	AwaitExpression
	CoverParenthesizedExpressionAndArrowParameterList
)

var NodeTypeToString = map[NodeType]string{
	Script:                               "Script",
	Expression:                           "Expression",
	StatementList:                        "StatementList",
	StatementListItem:                    "StatementListItem",
	Statement:                            "Statement",
	Declaration:                          "Declaration",
	EmptyStatement:                       "EmptyStatement",
	DebuggerStatement:                    "DebuggerStatement",
	BlockStatement:                       "BlockStatement",
	Block:                                "Block",
	ContinueStatement:                    "ContinueStatement",
	BreakStatement:                       "BreakStatement",
	VariableStatement:                    "VariableStatement",
	VariableDeclarationList:              "VariableDeclarationList",
	VariableDeclaration:                  "VariableDeclaration",
	BindingIdentifier:                    "BindingIdentifier",
	Initializer:                          "Initializer",
	BindingPattern:                       "BindingPattern",
	Identifier:                           "Identifier",
	AssignmentExpression:                 "AssignmentExpression",
	ConditionalExpression:                "ConditionalExpression",
	ShortCircuitExpression:               "ShortCircuitExpression",
	LogicalORExpression:                  "LogicalORExpression",
	LogicalANDExpression:                 "LogicalANDExpression",
	CoalesceExpression:                   "CoalesceExpression",
	CoalesceExpressionHead:               "CoalesceExpressionHead",
	BitwiseORExpression:                  "BitwiseORExpression",
	BitwiseXORExpression:                 "BitwiseXORExpression",
	BitwiseANDExpression:                 "BitwiseANDExpression",
	EqualityExpression:                   "EqualityExpression",
	RelationalExpression:                 "RelationalExpression",
	ShiftExpression:                      "ShiftExpression",
	AdditiveExpression:                   "AdditiveExpression",
	MultiplicativeExpression:             "MultiplicativeExpression",
	ExponentiationExpression:             "ExponentiationExpression",
	UnaryExpression:                      "UnaryExpression",
	UpdateExpression:                     "UpdateExpression",
	LeftHandSideExpression:               "LeftHandSideExpression",
	NewExpression:                        "NewExpression",
	CallExpression:                       "CallExpression",
	OptionalExpression:                   "OptionalExpression",
	MemberExpression:                     "MemberExpression",
	CoverCallExpressionAndAsyncArrowHead: "CoverCallExpressionAndAsyncArrowHead",
	SuperCall:                            "SuperCall",
	ImportCall:                           "ImportCall",
	ThisExpression:                       "ThisExpression",
	IdentifierReference:                  "IdentifierReference",
	NullLiteral:                          "NullLiteral",
	BooleanLiteral:                       "BooleanLiteral",
	NumericLiteral:                       "NumericLiteral",
	StringLiteral:                        "StringLiteral",
	SpreadElement:                        "SpreadElement",
	ArrayLiteral:                         "ArrayLiteral",
	ObjectLiteral:                        "ObjectLiteral",
	UndefinedLiteral:                     "UndefinedLiteral",
	PropertyDefinition:                   "PropertyDefinition",
	ObjectBindingPattern:                 "ObjectBindingPattern",
	ArrayBindingPattern:                  "ArrayBindingPattern",
	BindingProperty:                      "BindingProperty",
	BindingRestProperty:                  "BindingRestProperty",
	BindingElement:                       "BindingElement",
	MethodDefinition:                     "MethodDefinition",
	FunctionExpression:                   "FunctionExpression",
	ClassExpression:                      "ClassExpression",
	ClassStaticBlock:                     "ClassStaticBlock",
	RegularExpressionLiteral:             "RegularExpressionLiteral",
	TemplateLiteral:                      "TemplateLiteral",
	YieldExpression:                      "YieldExpression",
	ImportMeta:                           "ImportMeta",
	NewTarget:                            "NewTarget",
	IfStatement:                          "IfStatement",
	DoWhileStatement:                     "DoWhileStatement",
	WhileStatement:                       "WhileStatement",
	LexicalBinding:                       "LexicalBinding",
	LexicalDeclaration:                   "LexicalDeclaration",
	ForStatement:                         "ForStatement",
	ForInStatement:                       "ForInStatement",
	ForOfStatement:                       "ForOfStatement",
	SwitchStatement:                      "SwitchStatement",
	SwitchCase:                           "SwitchCase",
	SwitchDefault:                        "SwitchDefault",
	LabelIdentifier:                      "LabelIdentifier",
	ReturnStatement:                      "ReturnStatement",
	WithStatement:                        "WithStatement",
	LabelledStatement:                    "LabelledStatement",
	ThrowStatement:                       "ThrowStatement",
	TryStatement:                         "TryStatement",
	Catch:                                "Catch",
	AwaitExpression:                      "AwaitExpression",
	CoverParenthesizedExpressionAndArrowParameterList: "CoverParenthesizedExpressionAndArrowParameterList",
}

type Node interface {
	GetNodeType() NodeType
	GetParent() Node
	GetChildren() []Node
	SetChildren(children []Node)
	SetParent(parent Node)
	ToString() string
}

type OperatorNode interface {
	Node
	SetOperator(token lexer.Token)
	GetOperator() lexer.Token
	GetLeft() Node
	SetLeft(left Node)
	GetRight() Node
	SetRight(right Node)
}

func AddChild(parent Node, child Node) {
	parent.SetChildren(append(parent.GetChildren(), child))
	child.SetParent(parent)
}

type BasicNode struct {
	NodeType NodeType
	Parent   Node
	Children []Node
}

func (n *BasicNode) GetNodeType() NodeType {
	return n.NodeType
}

func (n *BasicNode) GetParent() Node {
	return n.Parent
}

func (n *BasicNode) GetChildren() []Node {
	return n.Children
}

func (n *BasicNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BasicNode) SetParent(parent Node) {
	n.Parent = parent
}

func (n *BasicNode) ToString() string {
	return NodeTypeToString[n.NodeType]
}
