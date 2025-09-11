package parser

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"zbrannelly.dev/go-js/cmd/lexer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

type TemplateMode int

const (
	TemplateModeNone TemplateMode = iota
	TemplateModeInSubstitution
	TemplateModeAfterSubstitution
)

type Parser struct {
	LexerState        *lexer.Lexer
	CurrentTokenIndex int
	RootNode          ast.Node

	// Lexer Goal State Flags
	ConsumedFirstSignificantToken bool
	ExpressionAllowed             bool
	TemplateMode                  TemplateMode

	// Flags (TODO: Do we need a stack for these?)
	AllowYield  bool
	AllowAwait  bool
	AllowReturn bool
}

func NewParser(input string, goalSymbol ast.NodeType) *Parser {
	var lexerGoalSymbol lexer.LexicalGoal
	switch goalSymbol {
	case ast.Script:
		lexerGoalSymbol = lexer.InputElementHashbangOrRegExp
	default:
		lexerGoalSymbol = lexer.InputElementDiv
	}

	lexerState := lexer.Lexer{
		Input:             input,
		Goal:              lexerGoalSymbol,
		Tokens:            []lexer.Token{},
		CurrentIndex:      0,
		CurrentTokenValue: "",
	}

	return &Parser{
		LexerState:                    &lexerState,
		CurrentTokenIndex:             0,
		RootNode:                      nil,
		ConsumedFirstSignificantToken: false,
		ExpressionAllowed:             false,
		TemplateMode:                  TemplateModeNone,
	}
}

func ParseText(input string, goalSymbol ast.NodeType) (ast.Node, error) {
	parser := NewParser(input, goalSymbol)

	switch goalSymbol {
	case ast.Script:
		return parseScriptNode(parser)
	default:
		return nil, errors.New("goal symbol not supported")
	}
}

func parseScriptNode(parser *Parser) (ast.Node, error) {
	scriptNode := &ast.ScriptNode{
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	statementList, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}
	ast.AddChild(scriptNode, statementList)
	return scriptNode, nil
}

func parseStatementList(parser *Parser) (ast.Node, error) {
	statementList := &ast.BasicNode{
		NodeType: ast.StatementList,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	for {
		if IsEOF(parser) {
			break
		}

		statementListItem, err := parseStatementListItem(parser)
		if err != nil {
			return nil, err
		}

		// Nil signals EOF.
		if statementListItem == nil {
			break
		}

		ast.AddChild(statementList, statementListItem)
	}

	if len(statementList.Children) == 0 {
		return nil, fmt.Errorf("expected at least one statement")
	}

	return statementList, nil
}

func parseStatementListItem(parser *Parser) (ast.Node, error) {
	statement, statementErr := parseStatement(parser)
	if statementErr != nil {
		return nil, statementErr
	}

	if statement != nil {
		return statement, nil
	}

	declaration, declarationErr := parseDeclaration(parser)
	if declarationErr != nil {
		return nil, declarationErr
	}

	if declaration != nil {
		return declaration, nil
	}

	return nil, nil
}

func parseStatement(parser *Parser) (ast.Node, error) {
	// EmptyStatement
	emptyStatement, emptyStatementErr := parseEmptyStatement(parser)
	if emptyStatementErr != nil {
		return nil, emptyStatementErr
	}

	if emptyStatement != nil {
		return emptyStatement, nil
	}

	// DebuggerStatement
	debuggerStatement, debuggerStatementErr := parseReservedWordStatement(parser, lexer.Debugger, ast.DebuggerStatement)
	if debuggerStatementErr != nil {
		return nil, debuggerStatementErr
	}

	if debuggerStatement != nil {
		return debuggerStatement, nil
	}

	// BlockStatement
	blockStatement, blockStatementErr := parseBlockStatement(parser)
	if blockStatementErr != nil {
		return nil, blockStatementErr
	}

	if blockStatement != nil {
		return blockStatement, nil
	}

	// ContinueStatement (TODO: Support the label extension.)
	continueStatement, continueStatementErr := parseReservedWordStatement(parser, lexer.Continue, ast.ContinueStatement)
	if continueStatementErr != nil {
		return nil, continueStatementErr
	}

	if continueStatement != nil {
		return continueStatement, nil
	}

	// BreakStatement (TODO: Support the label extension.)
	breakStatement, breakStatementErr := parseReservedWordStatement(parser, lexer.Break, ast.BreakStatement)
	if breakStatementErr != nil {
		return nil, breakStatementErr
	}

	if breakStatement != nil {
		return breakStatement, nil
	}

	// VariableStatement
	variableStatement, variableStatementErr := parseVariableStatement(parser)
	if variableStatementErr != nil {
		return nil, variableStatementErr
	}

	if variableStatement != nil {
		return variableStatement, nil
	}

	// TODO: ExpressionStatement
	// TODO: IfStatement
	// TODO: BreakableStatement
	// TODO: ContinueStatement
	// TODO: BreakStatement
	// TODO: WithStatement
	// TODO: LabelledStatement
	// TODO: ThrowStatement
	// TODO: TryStatement

	// TODO: Support the other extensions of this grammar ([Yield], [Await], [Return]).
	// TODO: [+Return]
	// TODO: ReturnStatement

	return nil, nil
}

func parseDeclaration(parser *Parser) (ast.Node, error) {
	return nil, errors.New("not implemented: parseDeclaration")
}

func parseEmptyStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.Semicolon {
		// Consume the semicolon token.
		ConsumeToken(parser)
		return &ast.BasicNode{
			NodeType: ast.EmptyStatement,
			Parent:   nil,
			Children: make([]ast.Node, 0),
		}, nil
	}

	return nil, nil
}

func parseReservedWordStatement(parser *Parser, tokenType lexer.TokenType, nodeType ast.NodeType) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != tokenType {
		return nil, nil
	}

	// Consume the reserved word token.
	ConsumeToken(parser)

	token = CurrentToken(parser)

	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("unexpected token: %v", token.Type)
	}

	// Consume the semicolon token.
	ConsumeToken(parser)

	return &ast.BasicNode{
		NodeType: nodeType,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}, nil
}

func parseBlockStatement(parser *Parser) (ast.Node, error) {
	block, err := parseBlock(parser)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, nil
	}

	return block, nil
}

func parseBlock(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.LeftBrace {
		return nil, nil
	}

	// Consume the left brace token.
	ConsumeToken(parser)

	statementList, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}

	// TODO: Create a SyntaxError type.
	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	// TODO: Create a SyntaxError type.
	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("unexpected token: %v", token.Type)
	}

	// Consume the right brace token.
	ConsumeToken(parser)

	block := &ast.BasicNode{
		NodeType: ast.Block,
		Parent:   nil,
		Children: []ast.Node{},
	}
	ast.AddChild(block, statementList)

	return block, nil
}

func parseVariableStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Var {
		return nil, nil
	}

	variableStatement := &ast.BasicNode{
		NodeType: ast.VariableStatement,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	// Consume the `var` token.
	ConsumeToken(parser)

	// No expression allowed after `var`.
	parser.ExpressionAllowed = false

	variableDeclarationList, err := parseVariableDeclarationList(parser)
	if err != nil {
		return nil, err
	}
	ast.AddChild(variableStatement, variableDeclarationList)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("unexpected token: %v", token.Type)
	}

	// Consume the semicolon token.
	ConsumeToken(parser)

	return variableStatement, nil
}

func parseVariableDeclarationList(parser *Parser) (ast.Node, error) {
	variableDeclarationList := &ast.BasicNode{
		NodeType: ast.VariableDeclarationList,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	for {
		variableDeclaration, err := parseVariableDeclaration(parser)
		if err != nil {
			return nil, err
		}

		if variableDeclaration == nil {
			break
		}
		ast.AddChild(variableDeclarationList, variableDeclaration)

		token := CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Comma {
			break
		}

		// Consume `,` token
		ConsumeToken(parser)
	}

	if len(variableDeclarationList.Children) == 0 {
		return nil, fmt.Errorf("expected at least one variable declaration")
	}

	return variableDeclarationList, nil
}

func parseVariableDeclaration(parser *Parser) (ast.Node, error) {
	variableDeclaration := &ast.BasicNode{
		NodeType: ast.VariableDeclaration,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	// No expression allowed after `var`.
	parser.ExpressionAllowed = false

	bindingIdentifier, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}

	if bindingIdentifier != nil {
		ast.AddChild(variableDeclaration, bindingIdentifier)
	} else {
		bindingPattern, err := parseBindingPattern(parser)
		if err != nil {
			return nil, err
		}

		if bindingPattern != nil {
			ast.AddChild(variableDeclaration, bindingPattern)
		}
	}

	if len(variableDeclaration.Children) == 0 {
		return nil, fmt.Errorf("expected at least one binding identifier or binding pattern")
	}

	initializer, err := parseInitializer(parser)
	if err != nil {
		return nil, err
	}

	if initializer != nil {
		ast.AddChild(variableDeclaration, initializer)
	}

	return variableDeclaration, nil
}

func parseBindingIdentifier(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Identifier {
		return nil, nil
	}

	// Consume the identifier token.
	ConsumeToken(parser)

	bindingIdentifier := &ast.BindingIdentifierNode{
		Parent:     nil,
		Children:   make([]ast.Node, 0),
		Identifier: token.Value,
	}

	// TODO: Support await and yield modifier here.

	return bindingIdentifier, nil
}

func parseInitializer(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Assignment {
		return nil, nil
	}

	// Consume the `=` token.
	ConsumeToken(parser)

	expression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the assignment token")
	}

	initializer := &ast.BasicNode{
		NodeType: ast.Initializer,
		Parent:   nil,
		Children: []ast.Node{expression},
	}
	return initializer, nil
}

func parseBindingPattern(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	objectBindingPattern, err := parseObjectBindingPattern(parser)
	if err != nil {
		return nil, err
	}

	if objectBindingPattern != nil {
		return objectBindingPattern, nil
	}

	arrayBindingPattern, err := parseArrayBindingPattern(parser)
	if err != nil {
		return nil, err
	}

	if arrayBindingPattern != nil {
		return arrayBindingPattern, nil
	}

	return nil, nil
}

func parseObjectBindingPattern(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.LeftBrace {
		return nil, nil
	}

	// Consume the left brace token.
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	// If we hit a right brace token, we're done.
	if token.Type == lexer.RightBrace {
		return &ast.ObjectBindingPatternNode{
			Properties: make([]ast.Node, 0),
		}, nil
	}

	bindingRestProperty, err := parseBindingPropertyRestNode(parser)
	if err != nil {
		return nil, err
	}

	if bindingRestProperty != nil {
		return &ast.ObjectBindingPatternNode{
			Properties: []ast.Node{bindingRestProperty},
		}, nil
	}

	propertyList := make([]ast.Node, 0)

	bindingPropertyList, err := parseBindingPropertyList(parser)
	if err != nil {
		return nil, err
	}

	if bindingPropertyList != nil {
		propertyList = bindingPropertyList
	}

	bindingRestProperty, err = parseBindingPropertyRestNode(parser)
	if err != nil {
		return nil, err
	}

	// Optional rest property on the end.
	if bindingRestProperty != nil {
		propertyList = append(propertyList, bindingRestProperty)
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the property definition list")
	}

	// Consume the right brace token.
	ConsumeToken(parser)

	return &ast.ObjectBindingPatternNode{
		Properties: propertyList,
	}, nil
}

func parseBindingPropertyRestNode(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Spread {
		return nil, nil
	}

	// Consume the spread token.
	ConsumeToken(parser)

	bindingIdentifier, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}

	if bindingIdentifier == nil {
		return nil, fmt.Errorf("expected a binding identifier after the spread token")
	}

	return &ast.BindingRestNode{
		Parent:     nil,
		Children:   make([]ast.Node, 0),
		Identifier: bindingIdentifier,
	}, nil
}

func parseBindingElementRestNode(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Spread {
		return nil, nil
	}

	// Consume the spread token.
	ConsumeToken(parser)

	bindingIdentifier, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}

	if bindingIdentifier == nil {
		bindingPattern, err := parseBindingPattern(parser)
		if err != nil {
			return nil, err
		}

		if bindingPattern == nil {
			return nil, fmt.Errorf("expected an identifier or binding pattern after the spread token")
		}

		return &ast.BindingRestNode{
			Parent:         nil,
			Children:       make([]ast.Node, 0),
			BindingPattern: bindingPattern,
		}, nil
	}

	return &ast.BindingRestNode{
		Parent:     nil,
		Children:   make([]ast.Node, 0),
		Identifier: bindingIdentifier,
	}, nil
}

func parseBindingPropertyList(parser *Parser) ([]ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.RightBrace {
		return nil, nil
	}

	bindingPropertyList := make([]ast.Node, 0)

	bindingProperty, err := parseBindingProperty(parser)
	if err != nil {
		return nil, err
	}

	if bindingProperty == nil {
		return nil, fmt.Errorf("expected a binding property after the `{` token")
	}

	bindingPropertyList = append(bindingPropertyList, bindingProperty)

	for {
		token = CurrentToken(parser)
		if token == nil {
			break
		}

		if token.Type != lexer.Comma {
			break
		}

		// Consume `,` token
		ConsumeToken(parser)

		bindingProperty, err = parseBindingProperty(parser)
		if err != nil {
			return nil, err
		}

		if bindingProperty == nil {
			return nil, fmt.Errorf("expected a binding property after the `,` token")
		}

		bindingPropertyList = append(bindingPropertyList, bindingProperty)
	}

	return bindingPropertyList, nil
}

func parseBindingProperty(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	// NOTE: The below will consume an identifier token OR yield & await.
	// Where the identifier part may match the PropertyName production.
	bindingIdentifier, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}

	if bindingIdentifier != nil {
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}

		token = CurrentToken(parser)
		if token == nil || token.Type != lexer.TernaryColon {
			return &ast.BindingPropertyNode{
				Target:      bindingIdentifier,
				Initializer: initializer,
			}, nil
		}
	}

	if bindingIdentifier != nil &&
		(bindingIdentifier.(*ast.BindingIdentifierNode).Identifier == "yield" || bindingIdentifier.(*ast.BindingIdentifierNode).Identifier == "await") {
		return nil, fmt.Errorf("invalid property name: %s", bindingIdentifier.(*ast.BindingIdentifierNode).Identifier)
	}

	token = CurrentToken(parser)
	if bindingIdentifier != nil && token != nil && token.Type == lexer.TernaryColon {
		// Consume `:` token
		ConsumeToken(parser)

		bindingElement, err := parseBindingElement(parser)
		if err != nil {
			return nil, err
		}

		if bindingElement == nil {
			return nil, fmt.Errorf("expected a binding element after the `:` token")
		}

		return &ast.BindingPropertyNode{
			Target: &ast.StringLiteralNode{
				Value: bindingIdentifier.(*ast.BindingIdentifierNode).Identifier,
			},
			BindingElement: bindingElement,
		}, nil
	}

	propertyName, err := parsePropertyName(parser)
	if err != nil {
		return nil, err
	}

	if propertyName == nil {
		return nil, nil
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.TernaryColon {
		return nil, fmt.Errorf("expected a ':' token after the property name")
	}

	// Consume `:` token
	ConsumeToken(parser)

	bindingElement, err := parseBindingElement(parser)
	if err != nil {
		return nil, err
	}

	if bindingElement == nil {
		return nil, fmt.Errorf("expected a binding element after the `:` token")
	}

	return &ast.BindingPropertyNode{
		Target:         propertyName,
		BindingElement: bindingElement,
	}, nil
}

func parseBindingElement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	bindingIdentifier, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}

	if bindingIdentifier != nil {
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}

		return &ast.BindingElementNode{
			Target:      bindingIdentifier,
			Initializer: initializer,
		}, nil
	}

	bindingPattern, err := parseBindingPattern(parser)
	if err != nil {
		return nil, err
	}

	if bindingPattern != nil {
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}

		return &ast.BindingElementNode{
			Target:      bindingPattern,
			Initializer: initializer,
		}, nil
	}

	return nil, nil
}

func parseArrayBindingPattern(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.LeftBracket {
		return nil, nil
	}

	// Consume the left bracket token.
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightBracket {
		return &ast.ArrayBindingPatternNode{
			Elements: make([]ast.Node, 0),
		}, nil
	}

	elementList := make([]ast.Node, 0)

	elisionCount, err := parseElisionSequence(parser)
	if err != nil {
		return nil, err
	}

	for range elisionCount {
		elementList = append(elementList, &ast.BindingElementNode{
			Parent:   nil,
			Children: make([]ast.Node, 0),
			Target: &ast.BasicNode{
				NodeType: ast.UndefinedLiteral,
			},
			Initializer: nil,
		})
	}

	bindingRestNode, err := parseBindingElementRestNode(parser)
	if err != nil {
		return nil, err
	}

	if bindingRestNode != nil {
		elementList = append(elementList, bindingRestNode)
		return &ast.ArrayBindingPatternNode{
			Elements: elementList,
		}, nil
	}

	bindingElementList, err := parseBindingElementList(parser)
	if err != nil {
		return nil, err
	}

	if bindingElementList != nil {
		elementList = append(elementList, bindingElementList...)
	}

	bindingRestNode, err = parseBindingElementRestNode(parser)
	if err != nil {
		return nil, err
	}

	if bindingRestNode != nil {
		elementList = append(elementList, bindingRestNode)
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightBracket {
		return nil, fmt.Errorf("expected a ']' token")
	}

	// Consume the right bracket token.
	ConsumeToken(parser)

	return &ast.ArrayBindingPatternNode{
		Elements: elementList,
	}, nil
}

func parseBindingElementList(parser *Parser) ([]ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.RightBracket {
		return nil, nil
	}

	bindingElementList := make([]ast.Node, 0)

	bindingElement, err := parseBindingElement(parser)
	if err != nil {
		return nil, err
	}

	if bindingElement != nil {
		bindingElementList = append(bindingElementList, bindingElement)
	}

	for {
		token = CurrentToken(parser)
		if token == nil {
			break
		}

		if token.Type != lexer.Comma {
			break
		}

		// Consume `,` token
		ConsumeToken(parser)

		elisionCount, err := parseElisionSequence(parser)
		if err != nil {
			return nil, err
		}

		for range elisionCount {
			bindingElementList = append(bindingElementList,
				&ast.BindingElementNode{
					Target: &ast.BasicNode{
						NodeType: ast.UndefinedLiteral,
					},
				},
			)
		}

		bindingElement, err = parseBindingElement(parser)
		if err != nil {
			return nil, err
		}

		if bindingElement != nil {
			bindingElementList = append(bindingElementList, bindingElement)
			continue
		}

		// No matches found, so we break out of the loop.
		break
	}

	return bindingElementList, nil
}

func parseAssignmentExpression(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	conditionalExpression, err := parseConditionalExpression(parser)
	if err != nil {
		return nil, err
	}

	if conditionalExpression != nil {
		token := CurrentToken(parser)

		if conditionalExpression.GetNodeType() == ast.CoverParenthesizedExpressionAndArrowParameterList {
			// ArrowParameters : CoverParenthesizedExpressionAndArrowParameterList[?Yield, ?Await]
			// ArrowFunction : ArrowParameters => ConciseBody[?Yield, ?Await]
			if token != nil && token.Type == lexer.ArrowOperator && !HasLineTerminatorBeforeCurrentToken(parser) {
				// Consume `=>` token
				ConsumeToken(parser)

				body, err := parseArrowFunctionConciseBody(parser)
				if err != nil {
					return nil, err
				}

				if body == nil {
					return nil, fmt.Errorf("expected a concise body after the arrow operator")
				}

				parameters := make([]ast.Node, 0)
				for _, child := range conditionalExpression.GetChildren() {
					if child.GetNodeType() == ast.Expression {
						// Destructure the expression.
						expression := child.(*ast.ExpressionNode)
						parameters = append(parameters, expression.Left)
						parameters = append(parameters, expression.Right)
					} else {
						parameters = append(parameters, child)
					}
				}

				parser.ExpressionAllowed = false
				return &ast.FunctionExpressionNode{
					Parameters: parameters,
					Body:       body,
					Arrow:      true,
				}, nil
			}

			if token != nil && token.Type == lexer.ArrowOperator {
				return nil, fmt.Errorf("expected a concise body after the arrow operator")
			}

			// ParenthesizedExpression : ( Expression )
			if len(conditionalExpression.GetChildren()) == 1 {
				parser.ExpressionAllowed = false
				return conditionalExpression.GetChildren()[0], nil
			}

			return nil, fmt.Errorf("this should not happen")
		}

		// ArrowFunction : BindingIdentifier => ConciseBody[?Yield, ?Await]
		if token != nil && token.Type == lexer.ArrowOperator && conditionalExpression.GetNodeType() == ast.IdentifierReference {
			// Consume `=>` token
			ConsumeToken(parser)

			body, err := parseArrowFunctionConciseBody(parser)
			if err != nil {
				return nil, err
			}

			if body == nil {
				return nil, fmt.Errorf("expected a concise body after the arrow operator")
			}

			parser.ExpressionAllowed = false
			return &ast.FunctionExpressionNode{
				Parameters: []ast.Node{conditionalExpression},
				Body:       body,
				Arrow:      true,
			}, nil
		}

		// AsyncArrowFunction : async BindingIdentifier => ConciseBody[?Yield, ?Await]
		if conditionalExpression.GetNodeType() == ast.IdentifierReference {
			keyword := conditionalExpression.(*ast.IdentifierReferenceNode).Identifier
			if keyword == "async" && !HasLineTerminatorBeforeCurrentToken(parser) {
				bindingIdentifier, err := parseBindingIdentifier(parser)
				if err != nil {
					return nil, err
				}

				if bindingIdentifier != nil {
					token = CurrentToken(parser)
					if token == nil {
						return nil, fmt.Errorf("unexpected EOF")
					}

					if token.Type != lexer.ArrowOperator {
						return nil, fmt.Errorf("expected an arrow operator after the binding identifier")
					}

					// Consume `=>` token
					ConsumeToken(parser)

					body, err := parseArrowFunctionConciseBody(parser)
					if err != nil {
						return nil, err
					}

					if body == nil {
						return nil, fmt.Errorf("expected a concise body after the arrow operator")
					}

					parser.ExpressionAllowed = false
					return &ast.FunctionExpressionNode{
						Parameters: []ast.Node{bindingIdentifier},
						Body:       body,
						Arrow:      true,
						Async:      true,
					}, nil
				}
			}
		}

		// AsyncArrowFunction : async ArrowParameters[?Yield, ?Await] => ConciseBody[?Yield, ?Await]
		if token != nil && token.Type == lexer.ArrowOperator && conditionalExpression.GetNodeType() == ast.CallExpression {
			callExpression := conditionalExpression.(*ast.CallExpressionNode)
			if callExpression.Callee.GetNodeType() == ast.IdentifierReference {
				keyword := callExpression.Callee.(*ast.IdentifierReferenceNode).Identifier

				if keyword == "async" && !HasLineTerminatorBeforeCurrentToken(parser) {
					// Consume `=>` token
					ConsumeToken(parser)

					body, err := parseArrowFunctionConciseBody(parser)
					if err != nil {
						return nil, err
					}

					if body == nil {
						return nil, fmt.Errorf("expected a concise body after the arrow operator")
					}

					return &ast.FunctionExpressionNode{
						Parameters: callExpression.Arguments,
						Body:       body,
						Arrow:      true,
						Async:      true,
					}, nil
				}
			}

			// TODO: Improve error message.
			return nil, fmt.Errorf("expected a valid async arrow function")
		}

		if token != nil && token.Type == lexer.Assignment {
			// Consume the assignment operator
			ConsumeToken(parser)

			expression, err := parseAssignmentExpression(parser)
			if err != nil {
				return nil, err
			}

			if expression == nil {
				return nil, fmt.Errorf("expected an expression after the assignment operator")
			}

			parser.ExpressionAllowed = false
			return &ast.AssignmentExpressionNode{
				Target:   conditionalExpression,
				Operator: *token,
				Value:    expression,
			}, nil
		}

		if token != nil && slices.Contains(lexer.AssignmentOperators, token.Type) {
			// Consume the assignment operator
			ConsumeToken(parser)

			expression, err := parseAssignmentExpression(parser)
			if err != nil {
				return nil, err
			}

			if expression == nil {
				return nil, fmt.Errorf("expected an expression after the assignment operator")
			}

			parser.ExpressionAllowed = false
			return &ast.AssignmentExpressionNode{
				Target:   conditionalExpression,
				Operator: *token,
				Value:    expression,
			}, nil
		}

		if token != nil && token.Type == lexer.AndAssignment {
			// Consume the assignment operator
			ConsumeToken(parser)

			expression, err := parseAssignmentExpression(parser)
			if err != nil {
				return nil, err
			}

			if expression == nil {
				return nil, fmt.Errorf("expected an expression after the assignment operator")
			}

			parser.ExpressionAllowed = false
			return &ast.AssignmentExpressionNode{
				Target:   conditionalExpression,
				Operator: *token,
				Value:    expression,
			}, nil
		}

		if token != nil && token.Type == lexer.OrAssignment {
			// Consume the assignment operator
			ConsumeToken(parser)

			expression, err := parseAssignmentExpression(parser)
			if err != nil {
				return nil, err
			}

			if expression == nil {
				return nil, fmt.Errorf("expected an expression after the assignment operator")
			}

			parser.ExpressionAllowed = false
			return &ast.AssignmentExpressionNode{
				Target:   conditionalExpression,
				Operator: *token,
				Value:    expression,
			}, nil
		}

		if token != nil && token.Type == lexer.NullishCoalescingAssignment {
			// Consume the assignment operator
			ConsumeToken(parser)

			expression, err := parseAssignmentExpression(parser)
			if err != nil {
				return nil, err
			}

			if expression == nil {
				return nil, fmt.Errorf("expected an expression after the assignment operator")
			}

			parser.ExpressionAllowed = false
			return &ast.AssignmentExpressionNode{
				Target:   conditionalExpression,
				Operator: *token,
				Value:    expression,
			}, nil
		}

		// Expression complete.
		parser.ExpressionAllowed = false
		return conditionalExpression, nil
	}

	// [+Yield] YieldExpression[?In, ?Await]
	if parser.AllowYield {
		token := CurrentToken(parser)

		if token != nil && token.Type == lexer.Yield && !HasLineTerminatorBeforeCurrentToken(parser) {
			// Consume `yield` keyword
			ConsumeToken(parser)

			token = CurrentToken(parser)
			generator := false
			if token != nil && token.Type == lexer.Multiply {
				// Consume `*` token
				ConsumeToken(parser)
				generator = true
			}

			// TODO: Figure out how to handle this note:
			// [Note 1] The syntactic context immediately following yield requires use of the InputElementRegExpOrTemplateTail lexical goal.

			expression, err := parseAssignmentExpression(parser)
			if err != nil {
				return nil, err
			}

			parser.ExpressionAllowed = false
			return &ast.YieldExpressionNode{
				Expression: expression,
				Generator:  generator,
			}, nil
		}
	}

	parser.ExpressionAllowed = false

	return nil, nil
}

func parseArrowFunctionConciseBody(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.LeftBrace {
		// Consume the left brace token.
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type == lexer.RightBrace {
			// Consume the right brace token.
			ConsumeToken(parser)
			return &ast.StatementListNode{
				Children: []ast.Node{},
			}, nil
		}

		// Consume the body.
		// TODO: Set [+Return = true, Await = false, Yield = false]
		body, err := parseStatementList(parser)
		if err != nil {
			return nil, err
		}

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.RightBrace {
			return nil, fmt.Errorf("expected a '}' token after the concise body")
		}

		// Consume the right brace token.
		ConsumeToken(parser)

		return body, nil
	}

	assignmentExpression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}

	if assignmentExpression != nil {
		return assignmentExpression, nil
	}

	return nil, fmt.Errorf("expected a function body after the arrow operator")
}

func parseConditionalExpression(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	conditionalExpression := &ast.ConditionalExpressionNode{
		Parent:    nil,
		Children:  make([]ast.Node, 0),
		Condition: nil,
		TrueExpr:  nil,
		FalseExpr: nil,
	}

	// ShortCircuitExpression[?In, ?Yield, ?Await]
	shortCircuitExpression, err := parseShortCircuitExpression(parser)

	if err != nil {
		return nil, err
	}
	if shortCircuitExpression == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	// ShortCircuitExpression[?In, ?Yield, ?Await]
	//   ? AssignmentExpression[+In, ?Yield, ?Await]
	//   : AssignmentExpression[?In, ?Yield, ?Await]
	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		// Not actually a conditional expression, so just return the short circuit expression.
		return shortCircuitExpression, nil
	}

	if token.Type != lexer.TernaryQuestionMark {
		// Expression complete.
		parser.ExpressionAllowed = false

		// Not actually a ternary expression, so just return the short circuit expression.
		return shortCircuitExpression, nil
	}

	// Assign the short circuit expression to the condition.
	conditionalExpression.Condition = shortCircuitExpression

	// Consume the `?` token.
	ConsumeToken(parser)

	assignmentExpression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}

	if assignmentExpression == nil {
		return nil, fmt.Errorf("expected an assignment expression after the `?` token")
	}

	conditionalExpression.TrueExpr = assignmentExpression

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.TernaryColon {
		return nil, fmt.Errorf("unexpected token: %v", token.Type)
	}

	// Consume the `:` token.
	ConsumeToken(parser)

	assignmentExpression, err = parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}

	if assignmentExpression == nil {
		return nil, fmt.Errorf("expected an assignment expression after the `:` token")
	}

	conditionalExpression.FalseExpr = assignmentExpression

	// Expression complete.
	parser.ExpressionAllowed = false

	return conditionalExpression, nil
}

func parseShortCircuitExpression(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	logicalORExpression, err := parseLogicalORExpression(parser)
	if err != nil {
		return nil, err
	}

	if logicalORExpression != nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return logicalORExpression, nil
	}

	coalesceExpression, err := parseCoalesceExpression(parser)
	if err != nil {
		return nil, err
	}

	if coalesceExpression != nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return coalesceExpression, nil
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	return nil, nil
}

func parseLogicalORExpression(parser *Parser) (ast.Node, error) {
	return parseSingleOperatorExpression(
		parser,
		lexer.Or,
		func(*Parser) ast.OperatorNode {
			return &ast.LogicalORExpressionNode{}
		},
		parseLogicalANDExpression,
		parseLogicalANDExpression,
	)
}

func parseCoalesceExpression(parser *Parser) (ast.Node, error) {
	return parseSingleOperatorExpression(
		parser,
		lexer.NullishCoalescing,
		func(*Parser) ast.OperatorNode {
			return &ast.CoalesceExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
			}
		},
		parseBitwiseORExpression,
		parseBitwiseORExpression,
	)
}

func parseLogicalANDExpression(parser *Parser) (ast.Node, error) {
	return parseSingleOperatorExpression(
		parser,
		lexer.And,
		func(*Parser) ast.OperatorNode {
			return &ast.LogicalANDExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
			}
		},
		parseBitwiseORExpression,
		parseBitwiseORExpression,
	)
}

func parseBitwiseORExpression(parser *Parser) (ast.Node, error) {
	return parseSingleOperatorExpression(
		parser,
		lexer.BitwiseOr,
		func(*Parser) ast.OperatorNode {
			return &ast.BitwiseORExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
			}
		},
		parseBitwiseXORExpression,
		parseBitwiseXORExpression,
	)
}

func parseBitwiseXORExpression(parser *Parser) (ast.Node, error) {
	return parseSingleOperatorExpression(
		parser,
		lexer.BitwiseXor,
		func(*Parser) ast.OperatorNode {
			return &ast.BitwiseXORExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
			}
		},
		parseBitwiseANDExpression,
		parseBitwiseANDExpression,
	)
}

func parseBitwiseANDExpression(parser *Parser) (ast.Node, error) {
	return parseSingleOperatorExpression(
		parser,
		lexer.BitwiseAnd,
		func(*Parser) ast.OperatorNode {
			return &ast.BitwiseANDExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
			}
		},
		parseEqualityExpression,
		parseEqualityExpression,
	)
}

func parseEqualityExpression(parser *Parser) (ast.Node, error) {
	return parseOperatorExpression(
		parser,
		lexer.EqualityOperators,
		func(*Parser) ast.OperatorNode {
			return &ast.EqualityExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
				Operator: lexer.Token{
					Type: -1,
				},
			}
		},
		parseRelationalExpression,
		parseRelationalExpression,
	)
}

func parseRelationalExpression(parser *Parser) (ast.Node, error) {
	// TODO: [+In] PrivateIdentifier in ShiftExpression[?Yield, ?Await]
	return parseOperatorExpression(
		parser,
		lexer.RelationalOperators,
		func(*Parser) ast.OperatorNode {
			return &ast.RelationalExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
				Operator: lexer.Token{
					Type: -1,
				},
			}
		},
		parseShiftExpression,
		parseShiftExpression,
	)
}

func parseShiftExpression(parser *Parser) (ast.Node, error) {
	return parseOperatorExpression(
		parser,
		lexer.ShiftOperators,
		func(*Parser) ast.OperatorNode {
			return &ast.ShiftExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
				Operator: lexer.Token{
					Type: -1,
				},
			}
		},
		parseAdditiveExpression,
		parseAdditiveExpression,
	)
}

func parseAdditiveExpression(parser *Parser) (ast.Node, error) {
	return parseOperatorExpression(
		parser,
		lexer.AdditiveOperators,
		func(*Parser) ast.OperatorNode {
			return &ast.AdditiveExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
				Operator: lexer.Token{
					Type: -1,
				},
			}
		},
		parseMultiplicativeExpression,
		parseMultiplicativeExpression,
	)
}

func parseMultiplicativeExpression(parser *Parser) (ast.Node, error) {
	return parseOperatorExpression(
		parser,
		lexer.MultiplicativeOperators,
		func(*Parser) ast.OperatorNode {
			return &ast.MultiplicativeExpressionNode{
				Parent:   nil,
				Children: make([]ast.Node, 0),
				Left:     nil,
				Right:    nil,
				Operator: lexer.Token{
					Type: -1,
				},
			}
		},
		parseExponentiationExpression,
		parseExponentiationExpression,
	)
}

func parseExponentiationExpression(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	exponentiationExpression := &ast.ExponentiationExpressionNode{
		Parent:   nil,
		Children: make([]ast.Node, 0),
		Left:     nil,
		Right:    nil,
	}

	unaryExpression, err := parseUnaryExpression(parser)
	if err != nil {
		return nil, err
	}

	if unaryExpression == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	// The LHS of an exponentiation expression must be an UpdateExpression.
	if !slices.ContainsFunc(unaryExpression.GetChildren(), func(node ast.Node) bool {
		return node.GetNodeType() == ast.UpdateExpression
	}) {
		// Expression complete.
		parser.ExpressionAllowed = false

		return unaryExpression, nil
	}

	token = CurrentToken(parser)
	if token == nil || token.Type != lexer.Exponentiation {
		// Expression complete.
		parser.ExpressionAllowed = false

		return unaryExpression, nil
	}

	// Consume the exponentiation operator.
	ConsumeToken(parser)

	exponentiationExpression.Left = unaryExpression
	exponentiationExpression.Right, err = parseExponentiationExpression(parser)
	if err != nil {
		return nil, err
	}

	if exponentiationExpression.Right == nil {
		return nil, fmt.Errorf("expected a right-hand side expression after the exponentiation operator")
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	return exponentiationExpression, nil
}

func parseUnaryExpression(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	unaryExpression := &ast.UnaryExpressionNode{
		Parent:   nil,
		Children: make([]ast.Node, 0),
		Operator: lexer.Token{
			Type: -1,
		},
		Value: nil,
	}

	updateExpression, err := parseUpdateExpression(parser)
	if err != nil {
		return nil, err
	}

	if updateExpression != nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return updateExpression, nil
	}

	if !slices.Contains(lexer.UnaryOperators, token.Type) {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	// Consume the unary operator.
	unaryExpression.Operator = *token
	ConsumeToken(parser)

	unaryExpression.Value, err = parseUnaryExpression(parser)
	if err != nil {
		return nil, err
	}

	if unaryExpression.Value == nil {
		return nil, fmt.Errorf("expected a value expression after the %s operator", token.Value)
	}

	// TODO: [+Await] AwaitExpression[?Yield]

	// Expression complete.
	parser.ExpressionAllowed = false

	return unaryExpression, nil
}

func parseUpdateExpression(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	updateExpression := &ast.UpdateExpressionNode{
		Parent:   nil,
		Children: make([]ast.Node, 0),
		Operator: lexer.Token{
			Type: -1,
		},
		Value: nil,
	}

	// Prefix update expression.
	if slices.Contains(lexer.UpdateOperators, token.Type) {
		updateExpression.Operator = *token
		ConsumeToken(parser)

		unaryExpression, err := parseUnaryExpression(parser)
		if err != nil {
			return nil, err
		}

		if unaryExpression == nil {
			return nil, fmt.Errorf("expected a unary expression after the %s operator", token.Value)
		}

		updateExpression.Value = unaryExpression

		// Expression complete.
		parser.ExpressionAllowed = false

		return updateExpression, nil
	}

	leftHandSideExpression, err := parseLeftHandSideExpression(parser)
	if err != nil {
		return nil, err
	}

	if leftHandSideExpression == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	token = CurrentToken(parser)

	// Not an actual update expression, so just return the left-hand side expression.
	// Also support [No LineTerminator here] after the left-hand side expression.
	if token == nil || !slices.Contains(lexer.UpdateOperators, token.Type) || HasLineTerminatorBeforeCurrentToken(parser) {
		// Expression complete.
		parser.ExpressionAllowed = false

		return leftHandSideExpression, nil
	}

	// Consume the operator token.
	updateExpression.Operator = *token
	ConsumeToken(parser)

	updateExpression.Value = leftHandSideExpression

	// Expression complete.
	parser.ExpressionAllowed = false

	return updateExpression, nil
}

func parseLeftHandSideExpression(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	baseNode, err := parseMemberExpression(parser)
	if err != nil {
		return nil, err
	}

	if baseNode == nil {
		if token.Type == lexer.New {
			// Consume `new` token
			ConsumeToken(parser)

			memberExpression, err := parseMemberExpression(parser)
			if err != nil {
				return nil, err
			}

			if memberExpression == nil {
				return nil, fmt.Errorf("expected a member expression after the 'new' keyword")
			}

			baseNode = &ast.NewExpressionNode{
				Constructor: memberExpression,
			}
		}
	}

	if baseNode == nil {
		if token.Type == lexer.Super {
			// Consume `super` token
			ConsumeToken(parser)

			arguments, err := parseArguments(parser)
			if err != nil {
				return nil, err
			}

			if arguments == nil {
				return nil, fmt.Errorf("expected arguments after the 'super' keyword")
			}

			baseNode = &ast.CallExpressionNode{
				Parent:    nil,
				Children:  make([]ast.Node, 0),
				Arguments: arguments,
				Super:     true,
			}
		}
	}

	if baseNode == nil {
		baseNode, err = parseImportCall(parser)
		if err != nil {
			return nil, err
		}
	}

	if baseNode == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	for {
		token = CurrentToken(parser)
		if token == nil {
			break
		}

		// Property access via expression.
		if token.Type == lexer.LeftBracket {
			// Consume `[` token
			ConsumeToken(parser)

			expression, err := parseExpression(parser)
			if err != nil {
				return nil, err
			}

			if expression == nil {
				return nil, fmt.Errorf("expected an expression after the '[' token")
			}

			token = CurrentToken(parser)
			if token == nil || token.Type != lexer.RightBracket {
				return nil, fmt.Errorf("expected a ']' token after the expression")
			}

			// Consume `]` token
			ConsumeToken(parser)

			baseNode = &ast.MemberExpressionNode{
				Object:             baseNode,
				Property:           expression,
				PropertyIdentifier: "",
			}
			continue
		}

		// Property access via identifier.
		if token.Type == lexer.Dot {
			// Consume `.` token
			ConsumeToken(parser)

			token = CurrentToken(parser)
			if token == nil || (token.Type != lexer.Identifier && token.Type != lexer.PrivateIdentifier) {
				return nil, fmt.Errorf("expected an identifier after the '.' token")
			}

			// Consume the identifier token
			ConsumeToken(parser)

			baseNode = &ast.MemberExpressionNode{
				Object:             baseNode,
				Property:           nil,
				PropertyIdentifier: token.Value,
			}
			continue
		}

		// Call expression.
		arguments, err := parseArguments(parser)
		if err != nil {
			return nil, err
		}

		if arguments != nil {
			baseNode = &ast.CallExpressionNode{
				Callee:    baseNode,
				Arguments: arguments,
			}
			continue
		}

		// Optional chain.
		if token.Type == lexer.OptionalChain {
			// Consume .? token
			ConsumeToken(parser)

			// Optional CallExpression.
			arguments, err := parseArguments(parser)
			if err != nil {
				return nil, err
			}

			if arguments != nil {
				baseNode = &ast.OptionalExpressionNode{
					Expression: &ast.CallExpressionNode{
						Callee:    baseNode,
						Arguments: arguments,
					},
				}
				continue
			}

			token = CurrentToken(parser)
			if token == nil {
				break
			}

			// Optional property access via expression
			if token.Type == lexer.LeftBracket {
				// Consume `[` token
				ConsumeToken(parser)

				expression, err := parseExpression(parser)
				if err != nil {
					return nil, err
				}

				if expression == nil {
					return nil, fmt.Errorf("expected an expression after the '[' token")
				}

				token = CurrentToken(parser)
				if token == nil || token.Type != lexer.RightBracket {
					return nil, fmt.Errorf("expected a ']' token after the expression")
				}

				// Consume `]` token
				ConsumeToken(parser)

				baseNode = &ast.OptionalExpressionNode{
					Expression: &ast.MemberExpressionNode{
						Object:             baseNode,
						Property:           expression,
						PropertyIdentifier: "",
					},
				}
				continue
			}

			// Optional property access via identifier.
			if token.Type == lexer.Identifier || token.Type == lexer.PrivateIdentifier {
				// Consume the identifier token
				ConsumeToken(parser)

				baseNode = &ast.OptionalExpressionNode{
					Expression: &ast.MemberExpressionNode{
						Object:             baseNode,
						Property:           nil,
						PropertyIdentifier: token.Value,
					},
				}
				continue
			}

			// TODO: Tagged Template parsing.
		}

		// No continuation, so we're done.
		break
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	return baseNode, nil
}

func parseImportCall(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	if token.Type != lexer.Import {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	// Consume `import` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil || token.Type != lexer.LeftParen {
		// TODO: Should this be an error? Or should we just lookahead for the left paren?
		return nil, fmt.Errorf("expected a '(' token after the 'import' keyword")
	}

	// Consume `(` token
	ConsumeToken(parser)

	importCall := &ast.BasicNode{
		NodeType: ast.ImportCall,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	for {
		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}

		if assignmentExpression == nil {
			break
		}

		ast.AddChild(importCall, assignmentExpression)

		token = CurrentToken(parser)
		if token == nil || token.Type != lexer.Comma {
			break
		}

		// Consume `,` token
		ConsumeToken(parser)
	}

	if importCall.GetChildren() == nil || len(importCall.GetChildren()) == 0 {
		return nil, fmt.Errorf("expected at least one assignment expression after the 'import' keyword")
	}

	token = CurrentToken(parser)
	if token == nil || token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the assignment expressions")
	}

	// Consume `)` token
	ConsumeToken(parser)

	// Expression complete.
	parser.ExpressionAllowed = false

	return importCall, nil
}

func parseArguments(parser *Parser) ([]ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	if token.Type != lexer.LeftParen {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	// Consume `(` token
	ConsumeToken(parser)

	arguments := make([]ast.Node, 0)

	argumentList, err := parseArgumentList(parser)
	if err != nil {
		return nil, err
	}

	if argumentList != nil {
		arguments = append(arguments, argumentList...)
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("expected a ')' token after the argument list")
	}

	// Comma is allowed after the argument list.
	if token.Type == lexer.Comma {
		// Consume `,` token
		ConsumeToken(parser)
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	token = CurrentToken(parser)
	if token == nil || token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the argument list")
	}

	// Consume `)` token
	ConsumeToken(parser)

	return arguments, nil
}

func parseArgumentList(parser *Parser) ([]ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.RightParen {
		return nil, nil
	}

	isSpread := false
	if token.Type == lexer.Spread {
		// Consume `...` token
		ConsumeToken(parser)

		isSpread = true
	}

	assignmentExpression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}

	if assignmentExpression == nil && !isSpread {
		return nil, nil
	} else if assignmentExpression == nil {
		return nil, fmt.Errorf("expected an assignment expression after the '...' token")
	}

	argumentList := make([]ast.Node, 0)

	if isSpread {
		argumentList = append(argumentList, &ast.SpreadElementNode{
			Expression: assignmentExpression,
		})
	} else {
		argumentList = append(argumentList, assignmentExpression)
	}

	token = CurrentToken(parser)
	if token == nil {
		return argumentList, nil
	}

	// Recurse if there is a comma.
	if token.Type == lexer.Comma {
		// Consume `,` token
		ConsumeToken(parser)

		childList, err := parseArgumentList(parser)
		if err != nil {
			return nil, err
		}

		if len(childList) == 0 {
			// If the child list is nil or empty, we need to reverse consume the ',' token.
			ReverseConsumeToken(parser)
			return argumentList, nil
		}
		argumentList = append(argumentList, childList...)
	}

	return argumentList, nil
}

func parseExpression(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	expression := &ast.ExpressionNode{
		Left:  nil,
		Right: nil,
	}

	assignmentExpression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}

	if assignmentExpression == nil {
		return nil, nil
	}

	expression.Left = assignmentExpression

	for {
		token = CurrentToken(parser)
		if token == nil {
			break
		}

		if token.Type != lexer.Comma {
			break
		}

		// Stop looking for more assignment expressions if we hit a spread operator.
		// This comma needs to be handled outside of this function.
		lookahead := LookaheadToken(parser)
		if lookahead != nil && lookahead.Type == lexer.Spread {
			break
		}

		// Consume `,` token
		ConsumeToken(parser)

		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}

		if assignmentExpression == nil {
			return nil, fmt.Errorf("expected an assignment expression after the ',' token")
		}

		expression.Right = assignmentExpression
		expression = &ast.ExpressionNode{
			Left:  expression,
			Right: nil,
		}
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	if expression.Right == nil {
		return expression.Left, nil
	}

	return expression, nil
}

func parseMemberExpression(parser *Parser) (ast.Node, error) {
	// MemberExpression[Yield, Await] :
	// PrimaryExpression[?Yield, ?Await]
	// SuperProperty[?Yield, ?Await]
	// MetaProperty
	// new MemberExpression[?Yield, ?Await] Arguments[?Yield, ?Await]
	// MemberExpression[?Yield, ?Await] [ Expression[+In, ?Yield, ?Await] ]
	// MemberExpression[?Yield, ?Await] . IdentifierName
	// MemberExpression[?Yield, ?Await] TemplateLiteral[?Yield, ?Await, +Tagged]
	// MemberExpression[?Yield, ?Await] . PrivateIdentifier

	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	baseNode, err := parsePrimaryExpression(parser)
	if err != nil {
		return nil, err
	}

	if baseNode == nil {
		baseNode, err = parseSuperProperty(parser)
		if err != nil {
			return nil, err
		}
	}

	if baseNode == nil {
		baseNode, err = parseMetaProperty(parser)
		if err != nil {
			return nil, err
		}
	}

	if baseNode == nil {
		if token.Type != lexer.New {
			// Expression complete.
			parser.ExpressionAllowed = false

			return nil, nil
		}

		// Consume `new` token
		ConsumeToken(parser)

		memberExpression, err := parseMemberExpression(parser)
		if err != nil {
			return nil, err
		}

		if memberExpression == nil {
			return nil, fmt.Errorf("expected a member expression after the 'new' keyword")
		}

		arguments, err := parseArguments(parser)
		if err != nil {
			return nil, err
		}

		if arguments == nil {
			return nil, fmt.Errorf("expected an arguments list")
		}

		baseNode = &ast.NewExpressionNode{
			Constructor: &ast.CallExpressionNode{
				Callee:    memberExpression,
				Arguments: arguments,
			},
		}
	}

	memberExpressionNode := &ast.MemberExpressionNode{
		Object:             baseNode,
		Property:           nil,
		PropertyIdentifier: "",
	}

	for {
		token = CurrentToken(parser)
		if token == nil {
			break
		}

		matchFound := false

		switch token.Type {
		case lexer.LeftBracket:
			// Consume `[` token
			ConsumeToken(parser)

			expression, err := parseExpression(parser)
			if err != nil {
				return nil, err
			}

			if expression == nil {
				return nil, fmt.Errorf("expected an expression after the '[' token")
			}

			token = CurrentToken(parser)
			if token == nil {
				return nil, fmt.Errorf("expected a ']' token after the expression")
			}

			if token.Type != lexer.RightBracket {
				return nil, fmt.Errorf("expected a ']' token after the expression")
			}

			// Consume `]` token
			ConsumeToken(parser)

			memberExpressionNode.Property = expression
			matchFound = true
		case lexer.Dot:
			// Consume `.` token
			ConsumeToken(parser)

			token = CurrentToken(parser)
			if token == nil || (token.Type != lexer.Identifier && token.Type != lexer.PrivateIdentifier) {
				return nil, fmt.Errorf("expected an identifier after the '.' token")
			}

			// Consume the identifier token.
			ConsumeToken(parser)

			memberExpressionNode.PropertyIdentifier = token.Value
			matchFound = true
		}

		// No match, break the loop.
		if !matchFound {
			break
		}

		memberExpressionNode = &ast.MemberExpressionNode{
			Object:             memberExpressionNode,
			Property:           nil,
			PropertyIdentifier: "",
		}
	}

	if memberExpressionNode.PropertyIdentifier == "" && memberExpressionNode.Property == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return memberExpressionNode.Object, nil
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	return memberExpressionNode, nil
}

func parsePrimaryExpression(parser *Parser) (ast.Node, error) {
	// PrimaryExpression[Yield, Await] :
	// this
	// IdentifierReference[?Yield, ?Await]
	// Literal
	// ArrayLiteral[?Yield, ?Await]
	// ObjectLiteral[?Yield, ?Await]
	// FunctionExpression
	// ClassExpression[?Yield, ?Await]
	// GeneratorExpression
	// AsyncFunctionExpression
	// AsyncGeneratorExpression
	// RegularExpressionLiteral
	// TemplateLiteral[?Yield, ?Await, ~Tagged]
	// CoverParenthesizedExpressionAndArrowParameterList[?Yield, ?Await]

	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	if token.Type == lexer.This {
		ConsumeToken(parser)

		// Expression complete.
		parser.ExpressionAllowed = false

		return &ast.BasicNode{
			NodeType: ast.ThisExpression,
		}, nil
	}

	asyncFunctionExpression, err := parseAsyncFunctionOrGeneratorExpression(parser)
	if err != nil {
		return nil, err
	}

	if asyncFunctionExpression != nil {
		return asyncFunctionExpression, nil
	}

	identifierReference, err := parseIdentifierReference(parser)
	if err != nil {
		return nil, err
	}

	if identifierReference != nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return identifierReference, nil
	}

	literal, err := parseLiteral(parser)
	if err != nil {
		return nil, err
	}

	if literal != nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return literal, nil
	}

	arrayLiteral, err := parseArrayLiteral(parser)
	if err != nil {
		return nil, err
	}

	if arrayLiteral != nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return arrayLiteral, nil
	}

	objectLiteral, err := parseObjectLiteral(parser)
	if err != nil {
		return nil, err
	}

	if objectLiteral != nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return objectLiteral, nil
	}

	classExpression, err := parseClassExpression(parser)
	if err != nil {
		return nil, err
	}

	if classExpression != nil {
		return classExpression, nil
	}

	functionExpression, err := parseFunctionOrGeneratorExpression(parser, false /* Async = false */)
	if err != nil {
		return nil, err
	}

	if functionExpression != nil {
		return functionExpression, nil
	}

	token = CurrentToken(parser)
	if token != nil && token.Type == lexer.RegularExpressionLiteral {
		// Consume `RegularExpressionLiteral` token
		ConsumeToken(parser)

		return &ast.RegularExpressionLiteralNode{
			PatternAndFlags: token.Value,
		}, nil
	}

	// TODO: Set [Tagged = false]
	templateLiteral, err := parseTemplateLiteral(parser)
	if err != nil {
		return nil, err
	}

	if templateLiteral != nil {
		return templateLiteral, nil
	}

	coverParenthesizedExpressionAndArrowParameterList, err := parseCoverParenthesizedExpressionAndArrowParameterList(parser)
	if err != nil {
		return nil, err
	}

	// NOTE: Callers of parsePrimaryExpression should further refine the cover node.
	if coverParenthesizedExpressionAndArrowParameterList != nil {
		return coverParenthesizedExpressionAndArrowParameterList, nil
	}

	return nil, nil
}

func parseIdentifierReference(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.Identifier {
		ConsumeToken(parser)
		return &ast.IdentifierReferenceNode{
			Identifier: token.Value,
		}, nil
	}

	// TODO [Await=false] allow "await" as an identifier reference.
	// TODO [Yield=false] allow "yield" as an identifier reference.

	return nil, nil
}

func parseLiteral(parser *Parser) (ast.Node, error) {
	// Literal :
	// NullLiteral
	// BooleanLiteral
	// NumericLiteral
	// StringLiteral

	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	if token.Type == lexer.Null {
		ConsumeToken(parser)

		// Expression complete.
		parser.ExpressionAllowed = false

		return &ast.BasicNode{
			NodeType: ast.NullLiteral,
		}, nil
	}

	if token.Type == lexer.True || token.Type == lexer.False {
		ConsumeToken(parser)

		// Expression complete.
		parser.ExpressionAllowed = false

		return &ast.BooleanLiteralNode{
			Value: token.Type == lexer.True,
		}, nil
	}

	numericLiteral, err := parseNumericLiteral(parser)
	if err != nil {
		return nil, err
	}

	if numericLiteral != nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return numericLiteral, nil
	}

	if token.Type == lexer.StringLiteral {
		ConsumeToken(parser)

		// Expression complete.
		parser.ExpressionAllowed = false

		// Remove the quotes from the string literal.
		value := token.Value[1 : len(token.Value)-1]

		return &ast.StringLiteralNode{
			Value: value,
		}, nil
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	return nil, nil
}

func parseNumericLiteral(parser *Parser) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	if token.Type == lexer.NumericLiteral {
		ConsumeToken(parser)

		// Remove underscore separators from the numeric literal.
		valueStr := strings.ReplaceAll(token.Value, "_", "")

		isBigInt := strings.Contains(valueStr, "n")
		valueStr = strings.TrimSuffix(valueStr, "n")

		if strings.HasPrefix(strings.ToLower(valueStr), "0x") {
			// TODO: Hex
			return nil, errors.New("not implemented: parseNumericLiteral - Hex")
		}

		if strings.HasPrefix(strings.ToLower(valueStr), "0b") {
			// TODO: Binary
			return nil, errors.New("not implemented: parseNumericLiteral - Binary")
		}

		if strings.HasPrefix(strings.ToLower(valueStr), "0o") {
			// TODO: Octal
			return nil, errors.New("not implemented: parseNumericLiteral - Octal")
		}

		// TODO: Handle legacy octal integer literals?

		// This parses decimals using Go's float parser, which should handle scientific notation.
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid numeric literal: %w", err)
		}

		if isBigInt {
			// TODO: Handle big ints.
			return nil, errors.New("not implemented: parseNumericLiteral - BigInt")
		}

		// Expression complete.
		parser.ExpressionAllowed = false

		return &ast.NumericLiteralNode{
			Value: value,
		}, nil
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	return nil, nil
}

func parseArrayLiteral(parser *Parser) (ast.Node, error) {
	// ArrayLiteral[Yield, Await] :
	// [ Elision[opt] ]
	// [ ElementList[?Yield, ?Await] ]
	// [ ElementList[?Yield, ?Await] , Elision[opt] ]

	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	if token.Type != lexer.LeftBracket {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	// Consume `[` token
	ConsumeToken(parser)

	elementList := make([]ast.Node, 0)

	elisionCount, err := parseElisionSequence(parser)
	if err != nil {
		return nil, err
	}

	for range elisionCount {
		elementList = append(elementList, &ast.BasicNode{
			NodeType: ast.UndefinedLiteral,
		})
	}

	elementListContinued, err := parseElementList(parser)
	if err != nil {
		return nil, err
	}

	if elementListContinued != nil {
		elementList = append(elementList, elementListContinued...)
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("expected a ']' token after the expression")
	}

	if token.Type != lexer.RightBracket {
		return nil, fmt.Errorf("expected a ']' token after the expression")
	}

	// Consume `]` token
	ConsumeToken(parser)

	// Expression complete.
	parser.ExpressionAllowed = false

	return &ast.BasicNode{
		NodeType: ast.ArrayLiteral,
		Children: elementList,
	}, nil
}

func parseElisionSequence(parser *Parser) (int, error) {
	count := 0

	for {
		token := CurrentToken(parser)
		if token == nil {
			break
		}

		if token.Type != lexer.Comma {
			break
		}

		// Consume `,` token
		ConsumeToken(parser)
		count++
	}

	return count, nil
}

func parseElementList(parser *Parser) ([]ast.Node, error) {
	// ElementList[Yield, Await] :
	// Elision[opt] AssignmentExpression[+In, ?Yield, ?Await]
	// Elision[opt] SpreadElement[?Yield, ?Await]
	// ElementList[?Yield, ?Await] , Elision[opt] AssignmentExpression[+In, ?Yield, ?Await]
	// ElementList[?Yield, ?Await] , Elision[opt] SpreadElement[?Yield, ?Await]

	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	elisionCount, err := parseElisionSequence(parser)
	if err != nil {
		return nil, err
	}

	elementListItems := make([]ast.Node, 0)

	for range elisionCount {
		elementListItems = append(elementListItems, &ast.BasicNode{
			NodeType: ast.UndefinedLiteral,
		})
	}

	// Avoid trying to parse an assignment expression if we're at the end of the element list.
	token = CurrentToken(parser)
	if token == nil || token.Type == lexer.RightBracket {
		// Expression complete.
		parser.ExpressionAllowed = false

		return elementListItems, nil
	}

	if token.Type == lexer.Spread {
		// Consume `...` token
		ConsumeToken(parser)

		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}

		if assignmentExpression == nil {
			return nil, fmt.Errorf("expected an assignment expression after the '...' token")
		}

		elementListItems = append(elementListItems, &ast.SpreadElementNode{
			Expression: assignmentExpression,
		})
	} else {
		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}

		if assignmentExpression != nil {
			elementListItems = append(elementListItems, assignmentExpression)
		}
	}

	token = CurrentToken(parser)

	if token == nil || token.Type != lexer.Comma {
		// Expression complete.
		parser.ExpressionAllowed = false

		return elementListItems, nil
	}

	// Consume `,` token
	ConsumeToken(parser)

	// Stop looking for elements if we hit a `]` token.
	token = CurrentToken(parser)
	if token.Type == lexer.RightBracket {
		// Expression complete.
		parser.ExpressionAllowed = false

		return elementListItems, nil
	}

	// Otherwise, we're looking for more elements.
	elementListItemsContinued, err := parseElementList(parser)
	if err != nil {
		return nil, err
	}

	if elementListItemsContinued != nil {
		elementListItems = append(elementListItems, elementListItemsContinued...)
	}

	// Expression complete.
	parser.ExpressionAllowed = false

	return elementListItems, nil
}

func parseObjectLiteral(parser *Parser) (ast.Node, error) {
	// ObjectLiteral[Yield, Await] :
	// { }
	// { PropertyDefinitionList[?Yield, ?Await] }
	// { PropertyDefinitionList[?Yield, ?Await] , }

	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	if token.Type != lexer.LeftBrace {
		// Expression complete.
		parser.ExpressionAllowed = false

		return nil, nil
	}

	// Consume `{` token
	ConsumeToken(parser)

	// Expressions aren't allowed straight away.
	parser.ExpressionAllowed = false

	objectLiteral := &ast.ObjectLiteralNode{
		Properties: make([]ast.Node, 0),
	}

	propertyDefinitionList, err := parsePropertyDefinitionList(parser)
	if err != nil {
		return nil, err
	}

	if propertyDefinitionList != nil {
		objectLiteral.Properties = propertyDefinitionList
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("expected a '}' token after the expression")
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the expression")
	}

	// Consume `}` token
	ConsumeToken(parser)

	// Expression complete.
	parser.ExpressionAllowed = false

	return objectLiteral, nil
}

func parsePropertyDefinitionList(parser *Parser) ([]ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	propertyDefinition, err := parsePropertyDefinition(parser)
	if err != nil {
		return nil, err
	}

	propertyDefinitionList := make([]ast.Node, 0)
	if propertyDefinition != nil {
		propertyDefinitionList = append(propertyDefinitionList, propertyDefinition)
	}

	token = CurrentToken(parser)
	if token == nil {
		return propertyDefinitionList, nil
	}

	if token.Type != lexer.Comma {
		return propertyDefinitionList, nil
	}

	// Consume `,` token
	ConsumeToken(parser)

	// Stop looking for properties if we hit a `}` token.
	token = CurrentToken(parser)
	if token == nil || token.Type == lexer.RightBrace {
		return propertyDefinitionList, nil
	}

	// Otherwise, we're looking for more properties.
	propertyDefinitionListContinued, err := parsePropertyDefinitionList(parser)
	if err != nil {
		return nil, err
	}

	if propertyDefinitionListContinued != nil {
		propertyDefinitionList = append(propertyDefinitionList, propertyDefinitionListContinued...)
	}

	return propertyDefinitionList, nil
}

func parsePropertyDefinition(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.RightBrace {
		return nil, nil
	}

	propertyName, err := parsePropertyName(parser)
	if err != nil {
		return nil, err
	}

	if propertyName != nil && propertyName.GetNodeType() == ast.IdentifierReference {
		// Identifier Initializer
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}

		if initializer != nil {
			return &ast.PropertyDefinitionNode{
				Key:   propertyName,
				Value: initializer,
			}, nil
		}

		// MethodDefinition : async GeneratorMethod
		identifier := propertyName.(*ast.IdentifierReferenceNode).Identifier
		if identifier == "async" && token.Type == lexer.Multiply && !HasLineTerminatorBeforeCurrentToken(parser) {
			// Consume `*` token
			ConsumeToken(parser)

			methodDefinition, err := parseBaseMethod(parser, true, true)
			if err != nil {
				return nil, err
			}

			if methodDefinition == nil {
				return nil, fmt.Errorf("expected a method definition after the 'async' keyword")
			}

			methodDefinition.(*ast.MethodDefinitionNode).Async = true
			methodDefinition.(*ast.MethodDefinitionNode).Generator = true
			return methodDefinition, nil
		}

		// MethodDefinition : async MethodDefinition
		if identifier == "async" && !HasLineTerminatorBeforeCurrentToken(parser) {
			methodDefinition, err := parseBaseMethod(parser, true, false)
			if err != nil {
				return nil, err
			}

			if methodDefinition == nil {
				return nil, fmt.Errorf("expected a method definition after the 'async' keyword")
			}

			methodDefinition.(*ast.MethodDefinitionNode).Async = true
			return methodDefinition, nil
		}

		// MethodDefinition : get ClassElementName[?Yield, ?Await] ( ) { FunctionBody[~Yield, ~Await] }
		if identifier == "get" {
			return parseGetterMethodAfterGetKeyword(parser)
		}

		// MethodDefinition : set ClassElementName[?Yield, ?Await] ( UniqueFormalParameters ) { FunctionBody[~Yield, ~Await] }
		if identifier == "set" {
			return parseSetterMethodAfterSetKeyword(parser)
		}
	}

	// PropertyName : AssignmentExpression
	token = CurrentToken(parser)
	if propertyName != nil && token != nil && token.Type == lexer.TernaryColon {
		// Consume `:` token
		ConsumeToken(parser)

		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}

		if assignmentExpression == nil {
			return nil, fmt.Errorf("expected an assignment expression after the ':' token")
		}

		return &ast.PropertyDefinitionNode{
			Key:   propertyName,
			Value: assignmentExpression,
		}, nil
	}

	// MethodDefinition : PropertyName ( UniqueFormalParameters ) { FunctionBody }
	if propertyName != nil && token != nil && token.Type == lexer.LeftParen {
		return parseMethodBodyAfterClassName(parser, propertyName)
	}

	// MethodDefinition : PrivateIdentifier ( UniqueFormalParameters ) { FunctionBody }
	if propertyName == nil && token != nil && token.Type == lexer.PrivateIdentifier {
		// Consume the private identifier token
		ConsumeToken(parser)

		return parseMethodBodyAfterClassName(parser, &ast.StringLiteralNode{
			Value: token.Value,
		})
	}

	generatorMethod, err := parseGeneratorMethod(parser)
	if err != nil {
		return nil, err
	}

	if generatorMethod != nil {
		return generatorMethod, nil
	}

	asyncMethod, err := parseAsyncMethodOrAsyncGeneratorMethod(parser)
	if err != nil {
		return nil, err
	}

	if asyncMethod != nil {
		return asyncMethod, nil
	}

	// IdentifierReference
	if propertyName != nil && propertyName.GetNodeType() == ast.IdentifierReference {
		return propertyName, nil
	}

	// Property name is not an identifier, but we didn't parse a value after it.
	if propertyName != nil {
		return nil, fmt.Errorf("expected a value after the property name")
	}

	return nil, nil
}

func parsePropertyName(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.Identifier {
		// Consume the identifier token
		ConsumeToken(parser)

		return &ast.IdentifierReferenceNode{
			Identifier: token.Value,
		}, nil
	}

	if token.Type == lexer.StringLiteral {
		// Consume the string literal token
		ConsumeToken(parser)

		// Remove the quotes from the string literal.
		value := token.Value[1 : len(token.Value)-1]

		return &ast.StringLiteralNode{
			Value: value,
		}, nil
	}

	numericLiteral, err := parseNumericLiteral(parser)
	if err != nil {
		return nil, err
	}

	if numericLiteral != nil {
		return numericLiteral, nil
	}

	if token.Type == lexer.LeftBracket {
		// Consume `[` token
		ConsumeToken(parser)

		computedPropertyName, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}

		if computedPropertyName == nil {
			return nil, fmt.Errorf("expected an assignment expression after the '[' token")
		}

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.RightBracket {
			return nil, fmt.Errorf("expected a ']' token after the assignment expression")
		}

		// Consume `]` token
		ConsumeToken(parser)

		return computedPropertyName, nil
	}

	return nil, nil
}

func parseClassElementName(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	propertyName, err := parsePropertyName(parser)
	if err != nil {
		return nil, err
	}

	if propertyName == nil && token.Type == lexer.PrivateIdentifier {
		propertyName = &ast.IdentifierReferenceNode{
			Identifier: token.Value,
		}
	}

	return propertyName, nil
}

func parseFormalParameters(parser *Parser) ([]ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightParen {
		return nil, nil
	}

	formalParameters := make([]ast.Node, 0)

	functionRestParameter, err := parseBindingElementRestNode(parser)
	if err != nil {
		return nil, err
	}

	if functionRestParameter != nil {
		formalParameters = append(formalParameters, functionRestParameter)
		return formalParameters, nil
	}

	formalParameterList, err := parseFormalParameterList(parser)
	if err != nil {
		return nil, err
	}

	if formalParameterList != nil {
		formalParameters = append(formalParameters, formalParameterList...)
	}

	functionRestParameter, err = parseBindingElementRestNode(parser)
	if err != nil {
		return nil, err
	}

	if functionRestParameter != nil {
		formalParameters = append(formalParameters, functionRestParameter)
	}

	return formalParameters, nil
}

func parseFormalParameterList(parser *Parser) ([]ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightParen {
		return nil, nil
	}

	formalParameters := make([]ast.Node, 0)

	formalParameter, err := parseBindingElement(parser)
	if err != nil {
		return nil, err
	}

	if formalParameter != nil {
		formalParameters = append(formalParameters, formalParameter)
	}

	for {
		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Comma {
			break
		}

		// Consume `,` token
		ConsumeToken(parser)

		formalParameter, err = parseBindingElement(parser)
		if err != nil {
			return nil, err
		}

		if formalParameter != nil {
			formalParameters = append(formalParameters, formalParameter)
			continue
		}

		// No matches found, so we break out of the loop.
		break
	}

	return formalParameters, nil
}

func parseGeneratorMethod(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Multiply {
		return nil, nil
	}

	// Consume `*` token
	ConsumeToken(parser)

	methodDefinition, err := parseBaseMethod(parser, false, true /* Yield = true */)
	if err != nil {
		return nil, err
	}

	if methodDefinition == nil {
		return nil, fmt.Errorf("expected a method definition after the '*' token")
	}

	methodDefinition.(*ast.MethodDefinitionNode).Generator = true
	return methodDefinition, nil
}

func parseAsyncMethodOrAsyncGeneratorMethod(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Identifier || token.Value != "async" {
		return nil, nil
	}

	// Consume `async` keyword
	ConsumeToken(parser)

	if HasLineTerminatorBeforeCurrentToken(parser) {
		return nil, fmt.Errorf("unexpected line terminator after the 'async' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.Multiply {
		// Consume `*` token
		ConsumeToken(parser)

		methodDefinition, err := parseBaseMethod(
			parser,
			true, /* Await = true */
			true, /* Yield = true */
		)
		if err != nil {
			return nil, err
		}

		if methodDefinition == nil {
			return nil, fmt.Errorf("expected a method definition after the '*' token")
		}

		methodDefinition.(*ast.MethodDefinitionNode).Async = true
		methodDefinition.(*ast.MethodDefinitionNode).Generator = true
		return methodDefinition, nil
	}

	methodDefinition, err := parseBaseMethod(parser, true /* Await = true */, false /* Yield = false */)
	if err != nil {
		return nil, err
	}

	if methodDefinition == nil {
		return nil, fmt.Errorf("expected a method definition after the 'async' keyword")
	}

	methodDefinition.(*ast.MethodDefinitionNode).Async = true
	return methodDefinition, nil
}

func parseBaseMethod(parser *Parser, await bool, yield bool) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	classElementName, err := parseClassElementName(parser)
	if err != nil {
		return nil, err
	}

	if classElementName == nil {
		return nil, nil
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the class element name")
	}

	// Consume `(` token
	ConsumeToken(parser)

	// TODO: Set [Await = await, Yield = yield]
	formalParameters, err := parseFormalParameters(parser)
	if err != nil {
		return nil, err
	}

	if formalParameters == nil {
		return nil, fmt.Errorf("expected a formal parameters after the 'base' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the formal parameters")
	}

	// Consume `)` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftBrace {
		return nil, fmt.Errorf("expected a '{' token after the formal parameters")
	}

	// Consume `{` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	// Avoid trying to parse the body if we have an empty body.
	if token.Type == lexer.RightBrace {
		// Consume `}` token
		ConsumeToken(parser)
		return &ast.MethodDefinitionNode{
			Name:       classElementName,
			Parameters: formalParameters,
			Body: &ast.StatementListNode{
				Children: []ast.Node{},
			},
		}, nil
	}

	// Parse base body.
	// TODO: Set [Await = await, Yield = yield, +Return = true]
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the base body")
	}

	// Consume `}` token
	ConsumeToken(parser)

	return &ast.MethodDefinitionNode{
		Name:       classElementName,
		Parameters: formalParameters,
		Body:       functionBody,
	}, nil
}

func parseMethodBodyAfterClassName(parser *Parser, identifier ast.Node) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the private identifier")
	}

	// Consume `(` token
	ConsumeToken(parser)

	formalParameters, err := parseFormalParameters(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the formal parameters")
	}

	// Consume `)` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftBrace {
		return nil, fmt.Errorf("expected a '{' token after the formal parameters")
	}

	// Consume `{` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightBrace {
		// Consume `}` token
		ConsumeToken(parser)
		return &ast.MethodDefinitionNode{
			Name:       identifier,
			Parameters: formalParameters,
			Body: &ast.StatementListNode{
				Children: []ast.Node{},
			},
		}, nil
	}

	// TODO: Set [+Return = true, Await = false, Yield = false]
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the function body")
	}

	// Consume `}` token
	ConsumeToken(parser)

	return &ast.MethodDefinitionNode{
		Name:       identifier,
		Parameters: formalParameters,
		Body:       functionBody,
	}, nil
}

func parseGetterMethod(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Identifier || token.Value != "get" {
		return nil, nil
	}

	// Consume `get` keyword
	ConsumeToken(parser)

	return parseGetterMethodAfterGetKeyword(parser)
}

func parseGetterMethodAfterGetKeyword(parser *Parser) (ast.Node, error) {
	classElementName, err := parseClassElementName(parser)
	if err != nil {
		return nil, err
	}

	if classElementName == nil {
		return nil, fmt.Errorf("expected a class element name after the 'get' keyword")
	}

	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'get' keyword")
	}

	// Consume `(` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected no arguments after the '(' token for a getter method")
	}

	// Consume `)` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftBrace {
		return nil, fmt.Errorf("expected a '{' token after the arguments for a getter method")
	}

	// Consume `{` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightBrace {
		// Consume `}` token
		ConsumeToken(parser)
		return &ast.MethodDefinitionNode{
			Name:       classElementName,
			Parameters: nil,
			Body: &ast.StatementListNode{
				Children: []ast.Node{},
			},
			Getter: true,
		}, nil
	}

	// TODO: Set [+Return = true, Await = false, Yield = false]
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the function body for a getter method")
	}

	// Consume `}` token
	ConsumeToken(parser)

	return &ast.MethodDefinitionNode{
		Name:       classElementName,
		Parameters: nil,
		Body:       functionBody,
		Getter:     true,
	}, nil
}

func parseSetterMethod(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Identifier || token.Value != "set" {
		return nil, nil
	}

	// Consume `set` keyword
	ConsumeToken(parser)

	return parseSetterMethodAfterSetKeyword(parser)
}

func parseSetterMethodAfterSetKeyword(parser *Parser) (ast.Node, error) {
	classElementName, err := parseClassElementName(parser)
	if err != nil {
		return nil, err
	}

	if classElementName == nil {
		return nil, fmt.Errorf("expected a class element name after the 'set' keyword")
	}

	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'set' keyword")
	}

	// Consume `(` token
	ConsumeToken(parser)

	// TODO: Set [Await = false, Yield = false]
	formalParameter, err := parseBindingElement(parser)
	if err != nil {
		return nil, err
	}

	if formalParameter == nil {
		return nil, fmt.Errorf("expected a single parameter for a setter method")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the formal parameter for a setter method")
	}

	// Consume `)` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftBrace {
		return nil, fmt.Errorf("expected a '{' token after the formal parameter for a setter method")
	}

	// Consume `{` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightBrace {
		// Consume `}` token
		ConsumeToken(parser)
		return &ast.MethodDefinitionNode{
			Name:       classElementName,
			Parameters: []ast.Node{formalParameter},
			Body: &ast.StatementListNode{
				Children: []ast.Node{},
			},
			Setter: true,
		}, nil
	}

	// TODO: Set [Await = false, Yield = false, +Return = true]
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the function body for a setter method")
	}

	// Consume `}` token
	ConsumeToken(parser)

	return &ast.MethodDefinitionNode{
		Name:       classElementName,
		Parameters: []ast.Node{formalParameter},
		Body:       functionBody,
		Setter:     true,
	}, nil
}

func parseAsyncFunctionOrGeneratorExpression(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Identifier || token.Value != "async" {
		return nil, nil
	}

	lookahead := LookaheadToken(parser)

	if lookahead != nil && lookahead.Type != lexer.Function && lookahead.Type != lexer.Multiply {
		return nil, nil
	}

	// Consume `async` keyword
	ConsumeToken(parser)

	return parseFunctionOrGeneratorExpression(parser, true /* Async = true */)
}

func parseFunctionOrGeneratorExpression(parser *Parser, async bool) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Function {
		return nil, nil
	}

	// Consume `function` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	isGenerator := false
	if token.Type == lexer.Multiply {
		// Consume `*` token
		ConsumeToken(parser)
		isGenerator = true
	}

	// TODO: Set [Await = async, Yield = isGenerator]
	bindingIdentifier, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the function keyword")
	}

	// Consume `(` token
	ConsumeToken(parser)

	// TODO: Set [Await = async, Yield = isGenerator]
	formalParameters, err := parseFormalParameters(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the formal parameters")
	}

	// Consume `)` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftBrace {
		return nil, fmt.Errorf("expected a '{' token after the formal parameters")
	}

	// Consume `{` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightBrace {
		// Consume `}` token
		ConsumeToken(parser)
		return &ast.FunctionExpressionNode{
			Name:       bindingIdentifier,
			Parameters: formalParameters,
			Body: &ast.StatementListNode{
				Children: []ast.Node{},
			},
			Generator: isGenerator,
			Async:     async,
		}, nil
	}

	// TODO: Set [+Return = true, Await = async, Yield = isGenerator]
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the function body")
	}

	// Consume `}` token
	ConsumeToken(parser)

	return &ast.FunctionExpressionNode{
		Name:       bindingIdentifier,
		Parameters: formalParameters,
		Body:       functionBody,
		Generator:  isGenerator,
		Async:      async,
	}, nil
}

func parseClassExpression(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Class {
		return nil, nil
	}

	// Consume `class` keyword
	ConsumeToken(parser)

	bindingIdentifier, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}

	classHeritage, err := parseClassHeritage(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftBrace {
		return nil, fmt.Errorf("expected a '{' token after the class heritage")
	}

	// Consume `{` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightBrace {
		// Consume `}` token
		ConsumeToken(parser)
		return &ast.ClassExpressionNode{
			Name:     bindingIdentifier,
			Heritage: classHeritage,
			Elements: []ast.Node{},
		}, nil
	}

	classElements, err := parseClassElements(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the class elements")
	}

	// Consume `}` token
	ConsumeToken(parser)

	return &ast.ClassExpressionNode{
		Name:     bindingIdentifier,
		Heritage: classHeritage,
		Elements: classElements,
	}, nil
}

func parseClassHeritage(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Extends {
		return nil, nil
	}

	// Consume `extends` keyword
	ConsumeToken(parser)

	heritage, err := parseLeftHandSideExpression(parser)
	if err != nil {
		return nil, err
	}

	if heritage == nil {
		return nil, fmt.Errorf("expected a left-hand side expression after the 'extends' keyword")
	}

	return heritage, nil
}

func parseClassElements(parser *Parser) ([]ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	classElements := []ast.Node{}

	for {
		element, err := parseClassElement(parser)
		if err != nil {
			return nil, err
		}
		if element != nil {
			classElements = append(classElements, element)
		}

		token = CurrentToken(parser)
		if token == nil {
			break
		}

		if token.Type == lexer.RightBrace {
			break
		}
	}

	return classElements, nil
}

func parseClassElement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.Semicolon {
		// Consume `;` token
		ConsumeToken(parser)
		return nil, nil
	}

	if token.Type == lexer.Identifier && token.Value == "static" {
		// Consume `static` keyword
		ConsumeToken(parser)
		return parseStaticClassElement(parser)
	}

	asyncMethod, err := parseAsyncMethodOrAsyncGeneratorMethod(parser)
	if err != nil {
		return nil, err
	}

	if asyncMethod != nil {
		return asyncMethod, nil
	}

	generatorMethod, err := parseGeneratorMethod(parser)
	if err != nil {
		return nil, err
	}

	if generatorMethod != nil {
		return generatorMethod, nil
	}

	getterMethod, err := parseGetterMethod(parser)
	if err != nil {
		return nil, err
	}

	if getterMethod != nil {
		return getterMethod, nil
	}

	setterMethod, err := parseSetterMethod(parser)
	if err != nil {
		return nil, err
	}

	if setterMethod != nil {
		return setterMethod, nil
	}

	classElementName, err := parseClassElementName(parser)
	if err != nil {
		return nil, err
	}

	if classElementName != nil {
		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type == lexer.LeftParen {
			return parseMethodBodyAfterClassName(parser, classElementName)
		}

		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Semicolon {
			return nil, fmt.Errorf("expected a ';' token after the initializer")
		}

		// Consume `;` token
		ConsumeToken(parser)

		return &ast.PropertyDefinitionNode{
			Key:   classElementName,
			Value: initializer,
		}, nil
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	return nil, fmt.Errorf("unexpected token inside class body: %s", token.Value)
}

func parseStaticClassElement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.LeftBrace {
		// Consume `{` token
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type == lexer.RightBrace {
			// Consume `}` token
			ConsumeToken(parser)
			return nil, nil
		}

		// TODO: Set [+Return = false, Await = true, Yield = false]
		body, err := parseStatementList(parser)
		if err != nil {
			return nil, err
		}

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.RightBrace {
			return nil, fmt.Errorf("expected a '}' token after the class static block body")
		}

		// Consume `}` token
		ConsumeToken(parser)

		return &ast.ClassStaticBlockNode{
			Body: body,
		}, nil
	}

	element, err := parseClassElement(parser)
	if err != nil {
		return nil, err
	}

	if element == nil {
		return nil, fmt.Errorf("expected a class element after the 'static' keyword")
	}

	if element.GetNodeType() == ast.PropertyDefinition {
		element.(*ast.PropertyDefinitionNode).Static = true
	} else if element.GetNodeType() == ast.MethodDefinition {
		element.(*ast.MethodDefinitionNode).Static = true
	} else {
		return nil, fmt.Errorf("unexpected class element after the 'static' keyword: %s", element.ToString())
	}

	return element, nil
}

func parseTemplateLiteral(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.TemplateNoSubstitutionLiteral {
		// Consume `TemplateNoSubstitutionLiteral` token
		ConsumeToken(parser)

		literalNode := &ast.BasicNode{
			NodeType: ast.TemplateLiteral,
		}

		// Remove the backticks from the template literal.
		value := token.Value[1 : len(token.Value)-1]

		ast.AddChild(literalNode, &ast.StringLiteralNode{
			Value: value,
		})
		return literalNode, nil
	}

	if token.Type != lexer.TemplateStartLiteral {
		return nil, nil
	}

	// Consume `TemplateStartLiteral` token
	ConsumeToken(parser)

	// Remove the start backtick and the start of the substitution.
	startValue := token.Value[1 : len(token.Value)-2]

	literalNode := &ast.BasicNode{
		NodeType: ast.TemplateLiteral,
	}

	if startValue != "" {
		ast.AddChild(literalNode, &ast.StringLiteralNode{
			Value: startValue,
		})
	}

	for {
		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		parser.TemplateMode = TemplateModeInSubstitution

		expression, err := parseExpression(parser)
		if err != nil {
			return nil, err
		}

		if expression == nil {
			return nil, fmt.Errorf("expected an expression after the template start literal")
		}

		ast.AddChild(literalNode, expression)

		parser.TemplateMode = TemplateModeAfterSubstitution

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type == lexer.TemplateMiddle {
			// Consume `TemplateMiddle` token
			ConsumeToken(parser)

			// Remove the `}` and `${` from the value.
			value := token.Value[1 : len(token.Value)-2]

			if value != "" {
				ast.AddChild(literalNode, &ast.StringLiteralNode{
					Value: value,
				})
			}
			continue
		}

		if token.Type == lexer.TemplateTail {
			// Consume `TemplateTail` token
			ConsumeToken(parser)

			// Remove the `}` from the start of the tail.
			value := token.Value[1 : len(token.Value)-1]

			if value != "" {
				ast.AddChild(literalNode, &ast.StringLiteralNode{
					Value: value,
				})
			}
			break
		}

		return nil, fmt.Errorf("unexpected token inside template literal: %s", token.Value)
	}

	parser.TemplateMode = TemplateModeNone

	return literalNode, nil
}

func parseCoverParenthesizedExpressionAndArrowParameterList(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.LeftParen {
		return nil, nil
	}

	node := &ast.BasicNode{
		NodeType: ast.CoverParenthesizedExpressionAndArrowParameterList,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	// Consume `(` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightParen {
		// Consume `)` token
		ConsumeToken(parser)
		return node, nil
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	bindingRestElement, err := parseBindingElementRestNode(parser)
	if err != nil {
		return nil, err
	}

	if bindingRestElement != nil {
		ast.AddChild(node, bindingRestElement)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.RightParen {
			return nil, fmt.Errorf("expected a ')' token after the binding rest element")
		}

		// Consume `)` token
		ConsumeToken(parser)

		return node, nil
	}

	// TODO: Set [+In = true]
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the '(' token")
	}

	ast.AddChild(node, expression)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Comma && token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ',' or ')' token after the expression")
	}

	if token.Type == lexer.RightParen {
		// Consume `)` token
		ConsumeToken(parser)
		return node, nil
	}

	// Consume `,` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightParen {
		// Consume `)` token
		ConsumeToken(parser)
		return node, nil
	}

	bindingRestElement, err = parseBindingElementRestNode(parser)
	if err != nil {
		return nil, err
	}

	if bindingRestElement != nil {
		ast.AddChild(node, bindingRestElement)
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the binding rest element")
	}

	// Consume `)` token
	ConsumeToken(parser)

	return node, nil
}

func parseSuperProperty(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	return nil, errors.New("not implemented: parseSuperProperty")
}

func parseMetaProperty(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	return nil, errors.New("not implemented: parseMetaProperty")
}

func parseSingleOperatorExpression(
	parser *Parser,
	operatorToken lexer.TokenType,
	newOperatorNode func(*Parser) ast.OperatorNode,
	valueParser func(*Parser) (ast.Node, error),
	rightParser func(*Parser) (ast.Node, error),
) (ast.Node, error) {
	return parseOperatorExpression(
		parser,
		[]lexer.TokenType{operatorToken},
		newOperatorNode,
		valueParser,
		rightParser,
	)
}

func parseOperatorExpression(
	parser *Parser,
	operatorTokens []lexer.TokenType,
	newOperatorNode func(*Parser) ast.OperatorNode,
	valueParser func(*Parser) (ast.Node, error),
	rightParser func(*Parser) (ast.Node, error),
) (ast.Node, error) {
	// Expressions are allowed.
	parser.ExpressionAllowed = true

	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	opNode := newOperatorNode(parser)

	left, err := valueParser(parser)
	if err != nil {
		return nil, err
	}

	if left == nil {
		return nil, nil
	}

	opNode.SetLeft(left)

	for {
		token = CurrentToken(parser)
		if token == nil {
			break
		}

		if !slices.Contains(operatorTokens, token.Type) {
			break
		}

		// Consume the operator token.
		ConsumeToken(parser)

		right, err := rightParser(parser)
		if err != nil {
			return nil, err
		}

		if right == nil {
			return nil, fmt.Errorf("expected a right-hand side expression after the operator token '%s'", token.Value)
		}

		opNode.SetRight(right)

		left = opNode
		opNode = newOperatorNode(parser)
		opNode.SetLeft(left)
	}

	// Return just the left-hand side if there's no right-hand side.
	if opNode.GetLeft() != nil && opNode.GetRight() == nil {
		return opNode.GetLeft(), nil
	}

	return opNode, nil
}

func UpdateLexerGoalState(parser *Parser) {
	var goalState lexer.LexicalGoal

	if !parser.ConsumedFirstSignificantToken {
		goalState = lexer.InputElementHashbangOrRegExp
	} else if parser.TemplateMode == TemplateModeAfterSubstitution {
		goalState = lexer.InputElementTemplateTail
	} else if parser.TemplateMode == TemplateModeInSubstitution {
		goalState = lexer.InputElementRegExpOrTemplateTail
	} else if parser.ExpressionAllowed {
		goalState = lexer.InputElementRegExp
	} else {
		goalState = lexer.InputElementDiv
	}

	parser.LexerState.Goal = goalState
}

func CurrentToken(parser *Parser) *lexer.Token {
	// Update the lexer goal state.
	UpdateLexerGoalState(parser)

	// No tokens in the buffer - we need to lex the next token.
	if parser.CurrentTokenIndex == len(parser.LexerState.Tokens) {
		if !lexer.LexNextToken(parser.LexerState) {
			return nil
		}
	}

	// Consume whitespace, line terminators, and comments.
	token := parser.LexerState.Tokens[parser.CurrentTokenIndex]
	for token.Type == lexer.WhiteSpace || token.Type == lexer.LineTerminator || token.Type == lexer.Comment {
		ConsumeToken(parser)
		if !lexer.LexNextToken(parser.LexerState) {
			return nil
		}
		token = parser.LexerState.Tokens[parser.CurrentTokenIndex]
	}

	return &token
}

func HasLineTerminatorBeforeCurrentToken(parser *Parser) bool {
	// Make sure lexer is at the current token.
	CurrentToken(parser)

	offset := 0
	for {
		if parser.CurrentTokenIndex-offset < 0 {
			return false
		}

		token := parser.LexerState.Tokens[parser.CurrentTokenIndex-offset]
		if token.Type == lexer.LineTerminator {
			return true
		}

		if token.Type == lexer.WhiteSpace {
			offset++
			continue
		}

		return false
	}
}

func ConsumeToken(parser *Parser) {
	if !parser.ConsumedFirstSignificantToken {
		if parser.LexerState.Tokens[parser.CurrentTokenIndex].Type != lexer.WhiteSpace &&
			parser.LexerState.Tokens[parser.CurrentTokenIndex].Type != lexer.LineTerminator &&
			parser.LexerState.Tokens[parser.CurrentTokenIndex].Type != lexer.Comment {
			// Track that we've consumed a significant token.
			parser.ConsumedFirstSignificantToken = true
			parser.ExpressionAllowed = true
		}
	}

	// Consume the token.
	parser.CurrentTokenIndex++
}

func ReverseConsumeToken(parser *Parser) {
	if parser.CurrentTokenIndex == 0 {
		return
	}

	// Consume whitespace, line terminators, and comments.
	for {
		token := parser.LexerState.Tokens[parser.CurrentTokenIndex-1]

		// No more whitespace, line terminators, or comments - we're done.
		if token.Type != lexer.WhiteSpace && token.Type != lexer.LineTerminator && token.Type != lexer.Comment {
			break
		}

		// Consume the token.
		parser.CurrentTokenIndex--

		if parser.CurrentTokenIndex == 0 {
			// Reset the lexer.
			parser.LexerState.CurrentIndex = 0
			parser.LexerState.CurrentTokenValue = ""
			parser.LexerState.Tokens = make([]lexer.Token, 0)
			return
		}
	}

	// Consume the token.
	parser.CurrentTokenIndex--

	// Reset the lexer to the previous token.
	parser.LexerState.CurrentIndex = parser.CurrentTokenIndex
	parser.LexerState.CurrentTokenValue = ""
	parser.LexerState.Tokens = parser.LexerState.Tokens[:parser.CurrentTokenIndex]
}

func CanLookaheadToken(parser *Parser) bool {
	if parser.CurrentTokenIndex+1 == len(parser.LexerState.Tokens) {
		return !lexer.IsEOF(parser.LexerState)
	}

	return true
}

func LookaheadToken(parser *Parser) *lexer.Token {
	// Backup the lexer state.
	tokens := make([]lexer.Token, len(parser.LexerState.Tokens))
	copy(tokens, parser.LexerState.Tokens)
	currentLexerGoal := parser.LexerState.Goal
	currentLexerIndex := parser.LexerState.CurrentIndex
	currentLexerTokenValue := parser.LexerState.CurrentTokenValue

	var token *lexer.Token = nil
	tokenIdx := parser.CurrentTokenIndex

	// Lex forward until we find a significant token.
	for lexer.LexNextToken(parser.LexerState) {
		tokenIdx++
		token = &parser.LexerState.Tokens[tokenIdx]
		if token.Type == lexer.WhiteSpace || token.Type == lexer.LineTerminator || token.Type == lexer.Comment {
			continue
		}
		break
	}

	// Restore the lexer state.
	parser.LexerState.Tokens = tokens
	parser.LexerState.Goal = currentLexerGoal
	parser.LexerState.CurrentIndex = currentLexerIndex
	parser.LexerState.CurrentTokenValue = currentLexerTokenValue

	return token
}

func IsEOF(parser *Parser) bool {
	return parser.CurrentTokenIndex == len(parser.LexerState.Tokens) && lexer.IsEOF(parser.LexerState)
}
