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

type Parser struct {
	LexerState              *lexer.Lexer
	CurrentTokenIndex       int
	CurrentTemplateSubDepth int
	CanStartRegExp          bool
	RootNode                ast.Node

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
		LexerState:              &lexerState,
		CurrentTokenIndex:       0,
		CurrentTemplateSubDepth: 0,
		CanStartRegExp:          true,
		RootNode:                nil,
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
		return nil, fmt.Errorf("unexpected token: %v", token)
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
		return nil, fmt.Errorf("unexpected token: %v", token)
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
		return nil, fmt.Errorf("unexpected token: %v", token)
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

	return nil, errors.New("not implemented: parseBindingPattern")
}

func parseAssignmentExpression(parser *Parser) (ast.Node, error) {
	conditionalExpression, err := parseConditionalExpression(parser)
	if err != nil {
		return nil, err
	}

	if conditionalExpression != nil {
		return conditionalExpression, nil
	}

	// TODO: [+Yield] YieldExpression[?In, ?Await]
	// TODO: ArrowFunction[?In, ?Yield, ?Await]
	// TODO: AsyncArrowFunction[?In, ?Yield, ?Await]
	// TODO: LeftHandSideExpression[?Yield, ?Await] = AssignmentExpression[?In, ?Yield, ?Await]
	// TODO: LeftHandSideExpression[?Yield, ?Await] AssignmentOperator AssignmentExpression[?In, ?Yield, ?Await]
	// TODO: LeftHandSideExpression[?Yield, ?Await] &&= AssignmentExpression[?In, ?Yield, ?Await]
	// TODO: LeftHandSideExpression[?Yield, ?Await] ||= AssignmentExpression[?In, ?Yield, ?Await]
	// TODO: LeftHandSideExpression[?Yield, ?Await] ??= AssignmentExpression[?In, ?Yield, ?Await]

	return nil, nil
}

func parseConditionalExpression(parser *Parser) (ast.Node, error) {
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
		return nil, nil
	}

	// ShortCircuitExpression[?In, ?Yield, ?Await]
	//   ? AssignmentExpression[+In, ?Yield, ?Await]
	//   : AssignmentExpression[?In, ?Yield, ?Await]
	token := CurrentToken(parser)
	if token == nil {
		// Not actually a conditional expression, so just return the short circuit expression.
		return shortCircuitExpression, nil
	}

	if token.Type != lexer.TernaryQuestionMark {
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
		return nil, fmt.Errorf("unexpected token: %v", token)
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

	return conditionalExpression, nil
}

func parseShortCircuitExpression(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	logicalORExpression, err := parseLogicalORExpression(parser)
	if err != nil {
		return nil, err
	}

	if logicalORExpression != nil {
		return logicalORExpression, nil
	}

	coalesceExpression, err := parseCoalesceExpression(parser)
	if err != nil {
		return nil, err
	}

	if coalesceExpression != nil {
		return coalesceExpression, nil
	}

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
	token := CurrentToken(parser)
	if token == nil {
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
		return nil, nil
	}

	// The LHS of an exponentiation expression must be an UpdateExpression.
	if !slices.ContainsFunc(unaryExpression.GetChildren(), func(node ast.Node) bool {
		return node.GetNodeType() == ast.UpdateExpression
	}) {
		return unaryExpression, nil
	}

	token = CurrentToken(parser)
	if token == nil || token.Type != lexer.Exponentiation {
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

	return exponentiationExpression, nil
}

func parseUnaryExpression(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
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
		return updateExpression, nil
	}

	if !slices.Contains(lexer.UnaryOperators, token.Type) {
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

	return unaryExpression, nil
}

func parseUpdateExpression(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
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
		return updateExpression, nil
	}

	leftHandSideExpression, err := parseLeftHandSideExpression(parser)
	if err != nil {
		return nil, err
	}

	if leftHandSideExpression == nil {
		return nil, nil
	}

	token = CurrentToken(parser)

	// Not an actual update expression, so just return the left-hand side expression.
	// Also support [No LineTerminator here] after the left-hand side expression.
	if token == nil || !slices.Contains(lexer.UpdateOperators, token.Type) || HasLineTerminatorBeforeCurrentToken(parser) {
		return leftHandSideExpression, nil
	}

	// Consume the operator token.
	updateExpression.Operator = *token
	ConsumeToken(parser)

	updateExpression.Value = leftHandSideExpression
	return updateExpression, nil
}

func parseLeftHandSideExpression(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	newExpression, err := parseNewExpression(parser)
	if err != nil {
		return nil, err
	}

	if newExpression != nil {
		return newExpression, nil
	}

	callExpression, err := parseCallExpression(parser)
	if err != nil {
		return nil, err
	}

	if callExpression != nil {
		return callExpression, nil
	}

	optionalExpression, err := parseOptionalExpression(parser)
	if err != nil {
		return nil, err
	}

	if optionalExpression != nil {
		return optionalExpression, nil
	}

	return nil, nil
}

func parseNewExpression(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	memberExpression, err := parseMemberExpression(parser)
	if err != nil {
		return nil, err
	}

	if memberExpression != nil {
		return memberExpression, nil
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.New {
		return nil, nil
	}

	// Consume the "new" keyword.
	ConsumeToken(parser)

	newExpression := &ast.NewExpressionNode{
		Parent:      nil,
		Children:    make([]ast.Node, 0),
		Constructor: nil,
	}

	newExpression.Constructor, err = parseMemberExpression(parser)
	if err != nil {
		return nil, err
	}

	if newExpression.Constructor == nil {
		return nil, fmt.Errorf("expected a member expression after the 'new' keyword")
	}

	return newExpression, nil
}

func parseCallExpression(parser *Parser) (ast.Node, error) {
	// CallExpression[Yield, Await] :
	// CoverCallExpressionAndAsyncArrowHead[?Yield, ?Await]
	// SuperCall[?Yield, ?Await]
	// ImportCall[?Yield, ?Await]
	// CallExpression[?Yield, ?Await] Arguments[?Yield, ?Await]
	// CallExpression[?Yield, ?Await] [ Expression[+In, ?Yield, ?Await] ]
	// CallExpression[?Yield, ?Await] . IdentifierName
	// CallExpression[?Yield, ?Await] TemplateLiteral[?Yield, ?Await, +Tagged]
	// CallExpression[?Yield, ?Await] . PrivateIdentifier

	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	baseNode, err := parseCoverCallExpressionAndAsyncArrowHead(parser)
	if err != nil {
		return nil, err
	}

	if baseNode == nil {
		baseNode, err = parseSuperCall(parser)
		if err != nil {
			return nil, err
		}
	}

	if baseNode == nil {
		baseNode, err = parseImportCall(parser)
		if err != nil {
			return nil, err
		}
	}

	if baseNode == nil {
		return nil, nil
	}

	var accumulator ast.Node = &ast.CallExpressionNode{
		Callee:    baseNode,
		Arguments: nil,
	}

	for {
		token = CurrentToken(parser)
		if token == nil {
			break
		}

		hasMatch := false

		arguments, err := parseArguments(parser)
		if err != nil {
			return nil, err
		}

		if arguments != nil {
			accumulator.(*ast.CallExpressionNode).Arguments = arguments
			hasMatch = true
		}

		if !hasMatch && token.Type == lexer.LeftBracket {
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

			accumulator = &ast.MemberExpressionNode{
				Object:             accumulator,
				Property:           expression,
				PropertyIdentifier: "",
			}

			hasMatch = true
		}

		if !hasMatch && token.Type == lexer.Dot {
			// Consume `.` token
			ConsumeToken(parser)

			token = CurrentToken(parser)
			if token == nil || (token.Type != lexer.Identifier && token.Type != lexer.PrivateIdentifier) {
				return nil, fmt.Errorf("expected an identifier after the '.' token")
			}

			accumulator = &ast.MemberExpressionNode{
				Object:             accumulator,
				Property:           nil,
				PropertyIdentifier: token.Value,
			}

			// Consume the identifier token
			ConsumeToken(parser)

			hasMatch = true
		}

		// TODO: Tagged Template parsing.
		// TODO: CallExpression[?Yield, ?Await] TemplateLiteral[?Yield, ?Await, +Tagged]

		// No continuation, so we're done.
		if !hasMatch {
			break
		}

		accumulator = &ast.CallExpressionNode{
			Callee:    accumulator,
			Arguments: nil,
		}
	}

	if accumulator.(*ast.CallExpressionNode).Arguments == nil {
		return accumulator.(*ast.CallExpressionNode).Callee, nil
	}

	return accumulator, nil
}

func parseCoverCallExpressionAndAsyncArrowHead(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	memberExpression, err := parseMemberExpression(parser)
	if err != nil {
		return nil, err
	}

	if memberExpression == nil {
		return nil, nil
	}

	arguments, err := parseArguments(parser)
	if err != nil {
		return nil, err
	}

	if arguments == nil {
		return nil, fmt.Errorf("expected arguments after the member expression")
	}

	return &ast.CallExpressionNode{
		Callee:    memberExpression,
		Arguments: arguments,
	}, nil
}

func parseSuperCall(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Super {
		return nil, nil
	}

	superCall := &ast.SuperCallNode{
		Parent:    nil,
		Children:  make([]ast.Node, 0),
		Arguments: nil,
	}

	// Consume `super` keyword
	ConsumeToken(parser)

	arguments, err := parseArguments(parser)
	if err != nil {
		return nil, err
	}

	if arguments == nil {
		return nil, fmt.Errorf("expected arguments after the 'super' keyword")
	}

	superCall.Arguments = arguments
	return superCall, nil
}

func parseImportCall(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Import {
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

	return importCall, nil
}

func parseArguments(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.LeftParen {
		return nil, nil
	}

	arguments := &ast.BasicNode{
		NodeType: ast.Arguments,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	// Consume `(` token
	ConsumeToken(parser)

	argumentList, err := parseArgumentList(parser)
	if err != nil {
		return nil, err
	}

	if argumentList != nil {
		ast.AddChild(arguments, argumentList)
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

	token = CurrentToken(parser)
	if token == nil || token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the argument list")
	}

	// Consume `)` token
	ConsumeToken(parser)

	return arguments, nil
}

func parseArgumentList(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
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
	} else if assignmentExpression == nil && isSpread {
		return nil, fmt.Errorf("expected an assignment expression after the '...' token")
	}

	argumentList := &ast.BasicNode{
		NodeType: ast.ArgumentList,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	listItem := &ast.ArgumentListItemNode{
		Parent:     nil,
		Children:   make([]ast.Node, 0),
		Spread:     isSpread,
		Expression: assignmentExpression,
	}

	ast.AddChild(argumentList, listItem)

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

		if childList == nil || len(childList.GetChildren()) == 0 {
			// If the child list is nil or empty, we need to reverse consume the ',' token.
			ReverseConsumeToken(parser)
			return argumentList, nil
		}

		for _, child := range childList.GetChildren() {
			ast.AddChild(argumentList, child)
		}
	}

	return argumentList, nil
}

func parseExpression(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	return nil, errors.New("not implemented: parseExpression")
}

func parseOptionalExpression(parser *Parser) (ast.Node, error) {
	// OptionalExpression[Yield, Await] :
	// MemberExpression[?Yield, ?Await] OptionalChain[?Yield, ?Await]
	// CallExpression[?Yield, ?Await] OptionalChain[?Yield, ?Await]
	// OptionalExpression[?Yield, ?Await] OptionalChain[?Yield, ?Await]

	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	targetNode, err := parseMemberExpression(parser)
	if err != nil {
		return nil, err
	}

	if targetNode == nil {
		targetNode, err = parseCallExpression(parser)
		if err != nil {
			return nil, err
		}
	}

	if targetNode == nil {
		return nil, nil
	}

	optionalExpression := &ast.OptionalExpressionNode{
		Parent:   nil,
		Children: make([]ast.Node, 0),
		Left:     targetNode,
		Right:    nil,
	}

	for {
		optionalExpression.Right, err = parseOptionalChain(parser)
		if err != nil {
			return nil, err
		}

		if optionalExpression.Right == nil {
			break
		}

		optionalExpression = &ast.OptionalExpressionNode{
			Parent:   nil,
			Children: make([]ast.Node, 0),
			Left:     optionalExpression,
			Right:    nil,
		}
	}

	// No right-hand side, so we return the left-hand side.
	if optionalExpression.Right == nil {
		return optionalExpression.Left, nil
	}

	return optionalExpression, nil
}

func parseOptionalChain(parser *Parser) (ast.Node, error) {
	// OptionalChain[Yield, Await] :
	// ?. Arguments[?Yield, ?Await]
	// ?. [ Expression[+In, ?Yield, ?Await] ]
	// ?. IdentifierName
	// ?. TemplateLiteral[?Yield, ?Await, +Tagged]
	// ?. PrivateIdentifier
	// OptionalChain[?Yield, ?Await] Arguments[?Yield, ?Await]
	// OptionalChain[?Yield, ?Await] [ Expression[+In, ?Yield, ?Await] ]
	// OptionalChain[?Yield, ?Await] . IdentifierName
	// OptionalChain[?Yield, ?Await] TemplateLiteral[?Yield, ?Await, +Tagged]
	// OptionalChain[?Yield, ?Await] . PrivateIdentifier

	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.OptionalChain {
		return nil, nil
	}

	// Consume `?.` token
	ConsumeToken(parser)

	accessorNode, err := parseArguments(parser)
	if err != nil {
		return nil, err
	}

	if accessorNode == nil {
		accessorNode, err = parseMemberExpression(parser)
		if err != nil {
			return nil, err
		}
	}

	if accessorNode == nil {
		// TODO: Should we have a parse function for IdentifierName?
	}

	return nil, errors.New("not implemented: parseOptionalChain")
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

	token := CurrentToken(parser)
	if token == nil {
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

			if token.Type != lexer.Identifier && token.Type != lexer.PrivateIdentifier {
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
		return memberExpressionNode.Object, nil
	}

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

	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.This {
		ConsumeToken(parser)
		return &ast.BasicNode{
			NodeType: ast.ThisExpression,
		}, nil
	}

	identifierReference, err := parseIdentifierReference(parser)
	if err != nil {
		return nil, err
	}

	if identifierReference != nil {
		return identifierReference, nil
	}

	literal, err := parseLiteral(parser)
	if err != nil {
		return nil, err
	}

	if literal != nil {
		return literal, nil
	}

	arrayLiteral, err := parseArrayLiteral(parser)
	if err != nil {
		return nil, err
	}

	if arrayLiteral != nil {
		return arrayLiteral, nil
	}

	objectLiteral, err := parseObjectLiteral(parser)
	if err != nil {
		return nil, err
	}

	if objectLiteral != nil {
		return objectLiteral, nil
	}

	return nil, errors.New("not implemented: parsePrimaryExpression")
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

	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.Null {
		ConsumeToken(parser)
		return &ast.BasicNode{
			NodeType: ast.NullLiteral,
		}, nil
	}

	if token.Type == lexer.True || token.Type == lexer.False {
		ConsumeToken(parser)
		return &ast.BooleanLiteralNode{
			Value: token.Type == lexer.True,
		}, nil
	}

	numericLiteral, err := parseNumericLiteral(parser)
	if err != nil {
		return nil, err
	}

	if numericLiteral != nil {
		return numericLiteral, nil
	}

	if token.Type == lexer.StringLiteral {
		ConsumeToken(parser)
		return &ast.StringLiteralNode{
			Value: token.Value,
		}, nil
	}

	return nil, nil
}

func parseNumericLiteral(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.NumericLiteral {
		ConsumeToken(parser)

		// Remove underscore separators from the numeric literal.
		valueStr := strings.ReplaceAll(token.Value, "_", "")

		isBigInt := strings.Contains(valueStr, "n")
		valueStr = strings.TrimSuffix(valueStr, "n")

		if strings.HasPrefix(strings.ToLower(valueStr), "0x") {
			// Hex
			return nil, errors.New("not implemented: parseNumericLiteral - Hex")
		}

		if strings.HasPrefix(strings.ToLower(valueStr), "0b") {
			// Binary
			return nil, errors.New("not implemented: parseNumericLiteral - Binary")
		}

		if strings.HasPrefix(strings.ToLower(valueStr), "0o") {
			// Octal
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

		return &ast.NumericLiteralNode{
			Value: value,
		}, nil
	}

	return nil, nil
}

func parseArrayLiteral(parser *Parser) (ast.Node, error) {
	// ArrayLiteral[Yield, Await] :
	// [ Elision[opt] ]
	// [ ElementList[?Yield, ?Await] ]
	// [ ElementList[?Yield, ?Await] , Elision[opt] ]

	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.LeftBracket {
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

	token := CurrentToken(parser)
	if token == nil {
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
		return elementListItems, nil
	}

	// Consume `,` token
	ConsumeToken(parser)

	// Stop looking for elements if we hit a `]` token.
	token = CurrentToken(parser)
	if token.Type == lexer.RightBracket {
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

	return elementListItems, nil
}

func parseObjectLiteral(parser *Parser) (ast.Node, error) {
	// ObjectLiteral[Yield, Await] :
	// { }
	// { PropertyDefinitionList[?Yield, ?Await] }
	// { PropertyDefinitionList[?Yield, ?Await] , }

	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.LeftBrace {
		return nil, nil
	}

	// Consume `{` token
	ConsumeToken(parser)

	objectLiteral := &ast.ObjectLiteralNode{
		Properties: make([]ast.Node, 0),
	}

	propertyDefinitionList, err := parsePropertyDefinitionList(parser)
	if err != nil {
		return nil, err
	}

	if propertyDefinitionList != nil {
		objectLiteral.Properties = propertyDefinitionList

		token = CurrentToken(parser)
		if token != nil && token.Type == lexer.Comma {
			// Consume `,` token
			ConsumeToken(parser)
		}
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

	return objectLiteral, nil
}

func parsePropertyDefinitionList(parser *Parser) ([]ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	return nil, errors.New("not implemented: parsePropertyDefinitionList")
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

func CurrentToken(parser *Parser) *lexer.Token {
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
		if parser.LexerState.CurrentIndex-offset < 0 {
			return false
		}

		token := parser.LexerState.Tokens[parser.LexerState.CurrentIndex-offset]
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
	if parser.CurrentTokenIndex+1 == len(parser.LexerState.Tokens) {
		if !lexer.LexNextToken(parser.LexerState) {
			return nil
		}
	}

	return &parser.LexerState.Tokens[parser.CurrentTokenIndex+1]
}

func IsEOF(parser *Parser) bool {
	return parser.CurrentTokenIndex == len(parser.LexerState.Tokens) && lexer.IsEOF(parser.LexerState)
}
