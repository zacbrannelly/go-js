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

	// Flags
	AllowYield   bool
	AllowAwait   bool
	AllowReturn  bool
	AllowIn      bool
	AllowDefault bool

	allowYieldStack   []bool
	allowAwaitStack   []bool
	allowReturnStack  []bool
	allowInStack      []bool
	allowDefaultStack []bool
}

func (p *Parser) PushAllowIn(value bool) {
	p.allowInStack = append(p.allowInStack, p.AllowIn)
	p.AllowIn = value
}

func (p *Parser) PopAllowIn() {
	if len(p.allowInStack) == 0 {
		panic("allowInStack is empty")
	}

	p.AllowIn = p.allowInStack[len(p.allowInStack)-1]
	p.allowInStack = p.allowInStack[:len(p.allowInStack)-1]
}

func (p *Parser) PushAllowYield(value bool) {
	p.allowYieldStack = append(p.allowYieldStack, p.AllowYield)
	p.AllowYield = value
}

func (p *Parser) PopAllowYield() {
	if len(p.allowYieldStack) == 0 {
		panic("allowYieldStack is empty")
	}

	p.AllowYield = p.allowYieldStack[len(p.allowYieldStack)-1]
	p.allowYieldStack = p.allowYieldStack[:len(p.allowYieldStack)-1]
}

func (p *Parser) PushAllowAwait(value bool) {
	p.allowAwaitStack = append(p.allowAwaitStack, p.AllowAwait)
	p.AllowAwait = value
}

func (p *Parser) PopAllowAwait() {
	if len(p.allowAwaitStack) == 0 {
		panic("allowAwaitStack is empty")
	}

	p.AllowAwait = p.allowAwaitStack[len(p.allowAwaitStack)-1]
	p.allowAwaitStack = p.allowAwaitStack[:len(p.allowAwaitStack)-1]
}

func (p *Parser) PushAllowReturn(value bool) {
	p.allowReturnStack = append(p.allowReturnStack, p.AllowReturn)
	p.AllowReturn = value
}

func (p *Parser) PopAllowReturn() {
	if len(p.allowReturnStack) == 0 {
		panic("allowReturnStack is empty")
	}

	p.AllowReturn = p.allowReturnStack[len(p.allowReturnStack)-1]
	p.allowReturnStack = p.allowReturnStack[:len(p.allowReturnStack)-1]
}

func (p *Parser) PushAllowDefault(value bool) {
	p.allowDefaultStack = append(p.allowDefaultStack, p.AllowDefault)
	p.AllowDefault = value
}

func (p *Parser) PopAllowDefault() {
	if len(p.allowDefaultStack) == 0 {
		panic("allowDefaultStack is empty")
	}

	p.AllowDefault = p.allowDefaultStack[len(p.allowDefaultStack)-1]
	p.allowDefaultStack = p.allowDefaultStack[:len(p.allowDefaultStack)-1]
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

	parser.PushAllowReturn(false)
	parser.PushAllowYield(false)
	parser.PushAllowAwait(false)
	statementList, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowReturn()
	parser.PopAllowAwait()
	parser.PopAllowYield()

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
	declaration, declarationErr := parseDeclaration(parser)
	if declarationErr != nil {
		return nil, declarationErr
	}

	if declaration != nil {
		return declaration, nil
	}

	statement, statementErr := parseStatement(parser)
	if statementErr != nil {
		return nil, statementErr
	}

	if statement != nil {
		return statement, nil
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

	// VariableStatement
	variableStatement, variableStatementErr := parseVariableStatement(parser)
	if variableStatementErr != nil {
		return nil, variableStatementErr
	}

	if variableStatement != nil {
		return variableStatement, nil
	}

	// NOTE: Must be before parseExpressionStatement, to not clash.
	labelledStatement, labelledStatementErr := parseLabelledStatement(parser)
	if labelledStatementErr != nil {
		return nil, labelledStatementErr
	}

	if labelledStatement != nil {
		return labelledStatement, nil
	}

	expressionStatement, expressionStatementErr := parseExpressionStatement(parser)
	if expressionStatementErr != nil {
		return nil, expressionStatementErr
	}

	if expressionStatement != nil {
		return expressionStatement, nil
	}

	ifStatement, ifStatementErr := parseIfStatement(parser)
	if ifStatementErr != nil {
		return nil, ifStatementErr
	}

	if ifStatement != nil {
		return ifStatement, nil
	}

	breakableStatement, breakableStatementErr := parseBreakableStatement(parser)
	if breakableStatementErr != nil {
		return nil, breakableStatementErr
	}

	if breakableStatement != nil {
		return breakableStatement, nil
	}

	continueStatement, continueStatementErr := parseContinueStatement(parser)
	if continueStatementErr != nil {
		return nil, continueStatementErr
	}

	if continueStatement != nil {
		return continueStatement, nil
	}

	breakStatement, breakStatementErr := parseBreakStatement(parser)
	if breakStatementErr != nil {
		return nil, breakStatementErr
	}

	if breakStatement != nil {
		return breakStatement, nil
	}

	if parser.AllowReturn {
		returnStatement, returnStatementErr := parseReturnStatement(parser)
		if returnStatementErr != nil {
			return nil, returnStatementErr
		}

		if returnStatement != nil {
			return returnStatement, nil
		}
	}

	withStatement, withStatementErr := parseWithStatement(parser)
	if withStatementErr != nil {
		return nil, withStatementErr
	}

	if withStatement != nil {
		return withStatement, nil
	}

	throwStatement, throwStatementErr := parseThrowStatement(parser)
	if throwStatementErr != nil {
		return nil, throwStatementErr
	}

	if throwStatement != nil {
		return throwStatement, nil
	}

	tryStatement, tryStatementErr := parseTryStatement(parser)
	if tryStatementErr != nil {
		return nil, tryStatementErr
	}

	if tryStatement != nil {
		return tryStatement, nil
	}

	return nil, nil
}

func parseTryStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Try {
		return nil, nil
	}

	// Consume the `try` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	block, err := parseBlock(parser)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, fmt.Errorf("expected a block after the 'try' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	var catchNode ast.Node = nil

	if token.Type == lexer.Catch {
		// Consume the `catch` keyword
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		var catchTarget ast.Node = nil

		if token.Type == lexer.LeftParen {
			// Consume the `(` token
			ConsumeToken(parser)

			token = CurrentToken(parser)
			if token == nil {
				return nil, fmt.Errorf("unexpected EOF")
			}

			catchTarget, err = parseBindingIdentifier(parser)
			if err != nil {
				return nil, err
			}

			if catchTarget == nil {
				catchTarget, err = parseBindingPattern(parser)
				if err != nil {
					return nil, err
				}

				if catchTarget == nil {
					return nil, fmt.Errorf("expected a binding identifier or binding pattern after the '(' token")
				}
			}

			token = CurrentToken(parser)
			if token == nil {
				return nil, fmt.Errorf("unexpected EOF")
			}

			if token.Type != lexer.RightParen {
				return nil, fmt.Errorf("expected a ')' token after the binding identifier or binding pattern")
			}

			// Consume the `)` token
			ConsumeToken(parser)
		}

		catchBlock, err := parseBlock(parser)
		if err != nil {
			return nil, err
		}

		if catchBlock == nil {
			return nil, fmt.Errorf("expected a block after the 'catch' keyword")
		}

		catchNode = &ast.CatchNode{
			Target: catchTarget,
			Block:  catchBlock,
		}
	}

	token = CurrentToken(parser)
	if token == nil {
		return &ast.TryStatementNode{
			Block:   block,
			Catch:   catchNode,
			Finally: nil,
		}, nil
	}

	if token.Type == lexer.Finally {
		// Consume the `finally` keyword
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		finallyBlock, err := parseBlock(parser)
		if err != nil {
			return nil, err
		}

		if finallyBlock == nil {
			return nil, fmt.Errorf("expected a block after the 'finally' keyword")
		}

		return &ast.TryStatementNode{
			Block:   block,
			Catch:   catchNode,
			Finally: finallyBlock,
		}, nil
	}

	return &ast.TryStatementNode{
		Block:   block,
		Catch:   catchNode,
		Finally: nil,
	}, nil
}

func parseThrowStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Throw {
		return nil, nil
	}

	// Consume the `throw` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if HasLineTerminatorBeforeCurrentToken(parser) {
		return nil, fmt.Errorf("unexpected line terminator after the 'throw' keyword")
	}

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the 'throw' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("expected a ';' token after the expression")
	}

	// Consume the `;` token
	ConsumeToken(parser)

	return &ast.ThrowStatementNode{
		Expression: expression,
	}, nil
}

func parseLabelledStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	// TODO: Allow `await` and `yield` as identifiers if their respective flags are false.
	if token.Type != lexer.Identifier {
		return nil, nil
	}

	lookahead := LookaheadToken(parser)
	if lookahead == nil || lookahead.Type != lexer.TernaryColon {
		return nil, nil
	}

	// Consume the identifier token
	labelIdentifier := &ast.LabelIdentifierNode{
		Identifier: token.Value,
	}
	ConsumeToken(parser)
	CurrentToken(parser)

	// Consume the `:` token
	ConsumeToken(parser)
	token = CurrentToken(parser)

	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	item, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if item == nil {
		item, err = parseFunctionOrGeneratorExpression(parser, false)
		if err != nil {
			return nil, err
		}

		if item == nil {
			return nil, fmt.Errorf("expected a statement or function declaration after the label identifier")
		}

		if item.GetNodeType() != ast.FunctionExpression {
			return nil, fmt.Errorf("internal error: unsupported node type when parsing labelled statement")
		}

		functionExpression := item.(*ast.FunctionExpressionNode)

		if functionExpression.Name == nil || functionExpression.Arrow || functionExpression.Generator || functionExpression.Async {
			// Ensure the expression is an instance of FunctionDeclaration[~Default].
			return nil, fmt.Errorf("expected a statement or function declaration after the label identifier")
		}
	}

	return &ast.LabelledStatementNode{
		Label:        labelIdentifier,
		LabelledItem: item,
	}, nil
}

func parseContinueStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Continue {
		return nil, nil
	}

	// Consume the `continue` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.Semicolon {
		// Consume the `;` token
		ConsumeToken(parser)
		return &ast.ContinueStatementNode{}, nil
	}

	if HasLineTerminatorBeforeCurrentToken(parser) {
		return nil, fmt.Errorf("unexpected line terminator after the 'continue' keyword")
	}

	labelIdentifier, err := parseLabelIdentifier(parser)
	if err != nil {
		return nil, err
	}

	if labelIdentifier == nil {
		return nil, fmt.Errorf("expected a semicolon after the 'continue' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("expected a semicolon after the 'continue' keyword")
	}

	// Consume the `;` token
	ConsumeToken(parser)

	return &ast.ContinueStatementNode{
		Label: labelIdentifier,
	}, nil
}

func parseBreakStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Break {
		return nil, nil
	}

	// Consume the `break` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.Semicolon {
		// Consume the `;` token
		ConsumeToken(parser)
		return &ast.BreakStatementNode{}, nil
	}

	if HasLineTerminatorBeforeCurrentToken(parser) {
		return nil, fmt.Errorf("unexpected line terminator after the 'break' keyword")
	}

	labelIdentifier, err := parseLabelIdentifier(parser)
	if err != nil {
		return nil, err
	}

	if labelIdentifier == nil {
		return nil, fmt.Errorf("expected a semicolon after the 'break' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("expected a semicolon after the 'break' keyword")
	}

	// Consume the `;` token
	ConsumeToken(parser)

	return &ast.BreakStatementNode{
		Label: labelIdentifier,
	}, nil
}

func parseReturnStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Return {
		return nil, nil
	}

	// Consume the `return` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.Semicolon {
		// Consume the `;` token
		ConsumeToken(parser)
		return &ast.ReturnStatementNode{}, nil
	}

	if HasLineTerminatorBeforeCurrentToken(parser) {
		return nil, fmt.Errorf("unexpected line terminator after the 'return' keyword")
	}

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected a semicolon after the 'return' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("expected a semicolon after the 'return' keyword")
	}

	// Consume the `;` token
	ConsumeToken(parser)

	return &ast.ReturnStatementNode{
		Value: expression,
	}, nil
}

func parseLabelIdentifier(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if parser.AllowYield && token.Type == lexer.Yield {
		return nil, nil
	}

	if token.Type != lexer.Identifier && token.Type != lexer.Await && token.Type != lexer.Yield {
		return nil, nil
	}

	// If [Await = false], allow `await` as an identifier.
	if parser.AllowAwait && token.Type == lexer.Await {
		return nil, nil
	}

	// If [Yield = false], allow `yield` as an identifier.
	if parser.AllowYield && token.Type == lexer.Yield {
		return nil, nil
	}

	// Consume the identifier token
	ConsumeToken(parser)

	return &ast.LabelIdentifierNode{
		Identifier: token.Value,
	}, nil
}

func parseWithStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.With {
		return nil, nil
	}

	// Consume the `with` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'with' keyword")
	}

	// Consume the `(` token
	ConsumeToken(parser)

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the '(' token")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the expression")
	}

	// Consume the `)` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	statement, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if statement == nil {
		return nil, fmt.Errorf("expected a statement after the ')' token")
	}

	return &ast.WithStatementNode{
		Expression: expression,
		Body:       statement,
	}, nil
}

func parseExpressionStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type == lexer.LeftBrace || token.Type == lexer.Function || token.Type == lexer.Class {
		return nil, nil
	}

	lookahead := LookaheadToken(parser)

	// TODO: Figure out how to detect "No line terminator after async keyword" with lookahead involved.
	if token.Type == lexer.Identifier && token.Value == "async" && lookahead != nil && lookahead.Type == lexer.Function {
		return nil, nil
	}

	if token.Type == lexer.Identifier && token.Value == "let" && lookahead != nil && lookahead.Type == lexer.LeftBracket {
		return nil, nil
	}

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, nil
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("expected a ';' token after the expression")
	}

	// Consume the semicolon token.
	ConsumeToken(parser)

	return expression, nil
}

func parseIfStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.If {
		return nil, nil
	}

	// Consume the `if` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'if' keyword")
	}

	// Consume the `(` token
	ConsumeToken(parser)

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the '(' token")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the expression")
	}

	// Consume the `)` token
	ConsumeToken(parser)

	trueStatement, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if trueStatement == nil {
		return nil, fmt.Errorf("expected a statement after the ')' token")
	}

	token = CurrentToken(parser)

	if token == nil || token.Type != lexer.Else {
		return &ast.IfStatementNode{
			Parent:        nil,
			Children:      []ast.Node{},
			Condition:     expression,
			TrueStatement: trueStatement,
			ElseStatement: nil,
		}, nil
	}

	// Consume the `else` keyword
	ConsumeToken(parser)

	elseStatement, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if elseStatement == nil {
		return nil, fmt.Errorf("expected a statement after the 'else' keyword")
	}

	return &ast.IfStatementNode{
		Condition:     expression,
		TrueStatement: trueStatement,
		ElseStatement: elseStatement,
	}, nil
}

func parseBreakableStatement(parser *Parser) (ast.Node, error) {
	iterationStatement, err := parseIterationStatement(parser)
	if err != nil {
		return nil, err
	}

	if iterationStatement != nil {
		return iterationStatement, nil
	}

	switchStatement, err := parseSwitchStatement(parser)
	if err != nil {
		return nil, err
	}

	if switchStatement != nil {
		return switchStatement, nil
	}

	return nil, nil
}

func parseIterationStatement(parser *Parser) (ast.Node, error) {
	doWhileStatement, err := parseDoWhileStatement(parser)
	if err != nil {
		return nil, err
	}

	if doWhileStatement != nil {
		return doWhileStatement, nil
	}

	whileStatement, err := parseWhileStatement(parser)
	if err != nil {
		return nil, err
	}

	if whileStatement != nil {
		return whileStatement, nil
	}

	forStatement, err := parseForStatementOrForInOfStatement(parser)
	if err != nil {
		return nil, err
	}

	if forStatement != nil {
		return forStatement, nil
	}

	return nil, nil
}

func parseSwitchStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Switch {
		return nil, nil
	}

	// Consume the `switch` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'switch' keyword")
	}

	// Consume the `(` token
	ConsumeToken(parser)

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the '(' token")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the expression")
	}

	// Consume the `)` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftBrace {
		return nil, fmt.Errorf("expected a '{' token after the ')' token")
	}

	// Consume the `{` token
	ConsumeToken(parser)

	switchStatement := &ast.SwitchStatementNode{
		Children: make([]ast.Node, 0),
		Target:   expression,
	}

	consumedDefaultCase := false

	for {
		switchCase, statementList, err := parseSwitchCase(parser)
		if err != nil {
			return nil, err
		}

		if switchCase != nil {
			ast.AddChild(switchStatement, switchCase)
			if statementList != nil {
				ast.AddChild(switchStatement, statementList)
			}
			continue
		}

		if consumedDefaultCase {
			break
		}

		switchDefault, statementList, err := parseSwitchDefault(parser)
		if err != nil {
			return nil, err
		}

		if switchDefault != nil {
			ast.AddChild(switchStatement, switchDefault)
			if statementList != nil {
				ast.AddChild(switchStatement, statementList)
			}
			consumedDefaultCase = true
			continue
		}

		break
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightBrace {
		return nil, fmt.Errorf("expected a '}' token after the switch statement")
	}

	// Consume the `}` token
	ConsumeToken(parser)

	return switchStatement, nil
}

func parseSwitchCase(parser *Parser) (ast.Node, ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil, nil
	}

	if token.Type != lexer.Case {
		return nil, nil, nil
	}

	// Consume the `case` keyword
	ConsumeToken(parser)

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, nil, fmt.Errorf("expected an expression after the 'case' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.TernaryColon {
		return nil, nil, fmt.Errorf("expected a ':' token after the expression")
	}

	// Consume the `:` token
	ConsumeToken(parser)

	statementList, err := parseStatementList(parser)
	if err != nil {
		return nil, nil, err
	}

	return &ast.SwitchCaseNode{
		Expression: expression,
	}, statementList, nil
}

func parseSwitchDefault(parser *Parser) (ast.Node, ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil, nil
	}

	if token.Type != lexer.Default {
		return nil, nil, nil
	}

	// Consume the `default` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.TernaryColon {
		return nil, nil, fmt.Errorf("expected a ':' token after the 'default' keyword")
	}

	// Consume the `:` token
	ConsumeToken(parser)

	statementList, err := parseStatementList(parser)
	if err != nil {
		return nil, nil, err
	}

	return &ast.BasicNode{
		NodeType: ast.SwitchDefault,
		Parent:   nil,
	}, statementList, nil
}

func parseDoWhileStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Do {
		return nil, nil
	}

	// Consume the `do` keyword
	ConsumeToken(parser)

	statement, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if statement == nil {
		return nil, fmt.Errorf("expected a statement after the 'do' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.While {
		return nil, fmt.Errorf("expected a 'while' keyword after the statement")
	}

	// Consume the `while` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'while' keyword")
	}

	// Consume the `(` token
	ConsumeToken(parser)

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the '(' token")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the expression")
	}

	// Consume the `)` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil || token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("expected a ';' token after the expression")
	}

	// Consume the `;` token
	ConsumeToken(parser)

	return &ast.DoWhileStatementNode{
		Parent:    nil,
		Children:  []ast.Node{},
		Condition: expression,
		Statement: statement,
	}, nil
}

func parseWhileStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.While {
		return nil, nil
	}

	// Consume the `while` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'while' keyword")
	}

	// Consume the `(` token
	ConsumeToken(parser)

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the '(' token")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the expression")
	}

	// Consume the `)` token
	ConsumeToken(parser)

	statement, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if statement == nil {
		return nil, fmt.Errorf("expected a statement after the ')' token")
	}

	return &ast.WhileStatementNode{
		Condition: expression,
		Statement: statement,
	}, nil
}

func parseForStatementOrForInOfStatement(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.For {
		return nil, nil
	}

	// Consume the `for` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if parser.AllowAwait && token.Type == lexer.Await {
		// Consume the `await` keyword
		ConsumeToken(parser)
		return parseForAwaitStatementAfterForAwaitKeywords(parser)
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'for' keyword")
	}

	// Consume the `(` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.Var {
		// Consume the `var` keyword
		ConsumeToken(parser)

		variableDeclarationList, err := parseVariableDeclarationList(parser)
		if err != nil {
			return nil, err
		}

		if variableDeclarationList == nil {
			return nil, fmt.Errorf("expected a variable declaration list after the 'var' keyword")
		}

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type == lexer.Semicolon {
			// Consume the semicolon token.
			ConsumeToken(parser)

			return parseForStatementAfterInitializer(parser, variableDeclarationList)
		}

		if token.Type != lexer.In && token.Type != lexer.Identifier {
			return nil, fmt.Errorf("expected a 'in' or 'of' keyword after the variable declaration list")
		}

		if token.Type == lexer.Identifier && token.Value != "of" {
			return nil, fmt.Errorf("expected a 'of' keyword after the variable declaration list")
		}

		if len(variableDeclarationList.GetChildren()) > 1 || len(variableDeclarationList.GetChildren()[0].GetChildren()) > 1 {
			return nil, fmt.Errorf("expected a single variable declaration")
		}

		// Consume the `in` or `of` keyword
		ConsumeToken(parser)

		// Extract the binding identifier / binding pattern from the variable declaration list
		declaration := variableDeclarationList.GetChildren()[0]

		if token.Type == lexer.In {
			return parseForInStatementAfterInKeyword(parser, declaration)
		}
		return parseForOfStatementAfterOfKeyword(parser, declaration)
	}

	if token.Type == lexer.Const || (token.Type == lexer.Identifier && token.Value == "let") {
		// Consume the `const` or `let` keyword
		ConsumeToken(parser)

		isConst := token.Type == lexer.Const
		isBindingPattern := false

		targetNode, err := parseBindingIdentifier(parser)
		if err != nil {
			return nil, err
		}

		if targetNode == nil {
			isBindingPattern = true
			targetNode, err = parseBindingPattern(parser)
			if err != nil {
				return nil, err
			}
		}

		if targetNode == nil {
			return nil, fmt.Errorf("expected a binding identifier or binding pattern after the 'const' or 'let' keyword")
		}

		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type == lexer.Semicolon || token.Type == lexer.Comma {
			// Binding pattern must have an initializer, if ForStatement path.
			if isBindingPattern && initializer == nil {
				return nil, fmt.Errorf("expected an initializer after the binding pattern")
			}
		}

		if token.Type == lexer.Semicolon {
			// Consume the semicolon token.
			ConsumeToken(parser)

			lexicalDeclaration := &ast.BasicNode{
				NodeType: ast.LexicalDeclaration,
				Parent:   nil,
				Children: make([]ast.Node, 0),
			}
			lexicalBinding := &ast.LexicalBindingNode{
				Parent:      nil,
				Children:    make([]ast.Node, 0),
				Target:      targetNode,
				Initializer: initializer,
				Const:       isConst,
			}
			ast.AddChild(lexicalDeclaration, lexicalBinding)

			return parseForStatementAfterInitializer(parser, lexicalDeclaration)
		}

		if token.Type == lexer.Comma {
			lexicalDeclaration := &ast.BasicNode{
				NodeType: ast.LexicalDeclaration,
				Parent:   nil,
				Children: make([]ast.Node, 0),
			}

			lexicalBinding := &ast.LexicalBindingNode{
				Target:      targetNode,
				Initializer: initializer,
				Const:       isConst,
			}
			ast.AddChild(lexicalDeclaration, lexicalBinding)

			for {
				token = CurrentToken(parser)
				if token == nil {
					return nil, fmt.Errorf("unexpected EOF")
				}

				if token.Type != lexer.Comma {
					break
				}

				// Consume the comma token.
				ConsumeToken(parser)

				lexicalBinding, err := parseLexicalBinding(parser, isConst)
				if err != nil {
					return nil, err
				}

				if lexicalBinding == nil {
					return nil, fmt.Errorf("expected a lexical binding after the comma")
				}

				ast.AddChild(lexicalDeclaration, lexicalBinding)
			}

			token = CurrentToken(parser)
			if token == nil {
				return nil, fmt.Errorf("unexpected EOF")
			}

			if token.Type != lexer.Semicolon {
				return nil, fmt.Errorf("expected a semicolon token after the lexical declaration")
			}

			// Consume the semicolon token.
			ConsumeToken(parser)

			return parseForStatementAfterInitializer(parser, lexicalDeclaration)
		}

		if token.Type == lexer.In {
			// Consume the `in` keyword
			ConsumeToken(parser)

			lexicalBinding := &ast.LexicalBindingNode{
				Target:      targetNode,
				Initializer: initializer,
				Const:       isConst,
			}
			return parseForInStatementAfterInKeyword(parser, lexicalBinding)
		}

		if token.Type != lexer.Identifier || token.Value != "of" {
			return nil, fmt.Errorf("expected an 'in' or 'of' keyword after the lexical binding")
		}

		// Consume the `of` keyword
		ConsumeToken(parser)

		lexicalBinding := &ast.LexicalBindingNode{
			Target:      targetNode,
			Initializer: initializer,
			Const:       isConst,
		}
		return parseForOfStatementAfterOfKeyword(parser, lexicalBinding)
	}

	// [+In = false]
	parser.PushAllowIn(false)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.Semicolon {
		// Consume the semicolon token.
		ConsumeToken(parser)

		return parseForStatementAfterInitializer(parser, expression)
	}
	if token.Type != lexer.In && token.Type != lexer.Identifier {
		return nil, fmt.Errorf("expected an 'in' or 'of' keyword after the expression")
	}

	if token.Type == lexer.Identifier && token.Value != "of" {
		return nil, fmt.Errorf("expected an 'in' or 'of' keyword after the expression")
	}

	// Consume the `in` or `of` keyword
	ConsumeToken(parser)

	// TODO: Implement syntax-directed operation `AssignmentTargetType` for Expression node.
	// TODO: If `AssignmentTargetType` of `expression` is `SIMPLE`, then it can be used for the ForIn/ForOf statement.

	// TODO: If `expression` is ObjectLiteral or ArrayLiteral, then it needs to follow AssignmentPattern production.

	if token.Type == lexer.In {
		return parseForInStatementAfterInKeyword(parser, expression)
	}

	return parseForOfStatementAfterOfKeyword(parser, expression)
}

func parseForInStatementAfterInKeyword(parser *Parser, declaration ast.Node) (ast.Node, error) {
	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the 'in' keyword")
	}

	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the expression")
	}

	// Consume the `)` token
	ConsumeToken(parser)

	body, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if body == nil {
		return nil, fmt.Errorf("expected a statement after the ')' token")
	}

	return &ast.ForInStatementNode{
		Target:   declaration,
		Iterable: expression,
		Body:     body,
	}, nil
}

func parseForOfStatementAfterOfKeyword(parser *Parser, declaration ast.Node) (ast.Node, error) {
	// [+In = true]
	parser.PushAllowIn(true)
	assignmentExpression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if assignmentExpression == nil {
		return nil, fmt.Errorf("expected an assignment expression after the 'of' keyword")
	}

	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the assignment expression")
	}

	// Consume the `)` token
	ConsumeToken(parser)

	body, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if body == nil {
		return nil, fmt.Errorf("expected a statement after the ')' token")
	}

	return &ast.ForOfStatementNode{
		Target:   declaration,
		Iterable: assignmentExpression,
		Body:     body,
	}, nil
}

func parseForStatementAfterInitializer(parser *Parser, initializer ast.Node) (ast.Node, error) {
	// [+In = true]
	parser.PushAllowIn(true)
	condition, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("expected a semicolon token after the condition expression")
	}

	// Consume the semicolon token.
	ConsumeToken(parser)

	// [+In = true]
	parser.PushAllowIn(true)
	updateExpression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.RightParen {
		return nil, fmt.Errorf("expected a ')' token after the update expression")
	}

	// Consume the `)` token
	ConsumeToken(parser)

	body, err := parseStatement(parser)
	if err != nil {
		return nil, err
	}

	if body == nil {
		return nil, fmt.Errorf("expected a statement after the ')' token")
	}

	return &ast.ForStatementNode{
		Initializer: initializer,
		Condition:   condition,
		Update:      updateExpression,
		Body:        body,
	}, nil
}

func parseForAwaitStatementAfterForAwaitKeywords(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the 'for await' keywords")
	}

	// Consume the `(` token
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.Var {
		// Consume the `var` token
		ConsumeToken(parser)

		targetNode, err := parseBindingIdentifier(parser)
		if err != nil {
			return nil, err
		}

		if targetNode == nil {
			targetNode, err = parseBindingPattern(parser)
			if err != nil {
				return nil, err
			}
		}

		if targetNode == nil {
			return nil, fmt.Errorf("expected a binding identifier or binding pattern after the 'var' keyword")
		}

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Identifier || token.Value != "of" {
			return nil, fmt.Errorf("expected an 'of' keyword after the binding identifier or binding pattern")
		}

		// Consume the `of` keyword
		ConsumeToken(parser)

		forOfStatement, err := parseForOfStatementAfterOfKeyword(parser, targetNode)
		if err != nil {
			return nil, err
		}

		if forOfStatement == nil {
			return nil, fmt.Errorf("expected a for of statement after the 'of' keyword")
		}

		// Mark the for of statement as await.
		forOfStatement.(*ast.ForOfStatementNode).Await = true

		return forOfStatement, nil
	}

	if token.Type == lexer.Const || (token.Type == lexer.Identifier && token.Value == "let") {
		// Consume the `const` or `let` keyword
		ConsumeToken(parser)

		isConst := token.Type == lexer.Const

		targetNode, err := parseBindingIdentifier(parser)
		if err != nil {
			return nil, err
		}

		if targetNode == nil {
			targetNode, err = parseBindingPattern(parser)
			if err != nil {
				return nil, err
			}
		}

		if targetNode == nil {
			return nil, fmt.Errorf("expected a binding identifier or binding pattern after the 'const' or 'let' keyword")
		}

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Identifier || token.Value != "of" {
			return nil, fmt.Errorf("expected an 'of' keyword after the binding identifier or binding pattern")
		}

		// Consume the `of` keyword
		ConsumeToken(parser)

		targetNode = &ast.LexicalBindingNode{
			Target: targetNode,
			Const:  isConst,
		}

		forOfStatement, err := parseForOfStatementAfterOfKeyword(parser, targetNode)
		if err != nil {
			return nil, err
		}

		if forOfStatement == nil {
			return nil, fmt.Errorf("expected a for of statement after the 'of' keyword")
		}

		// Mark the for of statement as await.
		forOfStatement.(*ast.ForOfStatementNode).Await = true

		return forOfStatement, nil
	}

	targetNode, err := parseLeftHandSideExpression(parser)
	if err != nil {
		return nil, err
	}

	// TODO: Implement syntax-directed operation `AssignmentTargetType` for LeftHandSideExpression node.
	// TODO: If `AssignmentTargetType` of `targetNode` is `SIMPLE`, then it can be used for the ForOf statement.

	if targetNode == nil {
		return nil, fmt.Errorf("expected an expression before the 'of' keyword")
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Identifier || token.Value != "of" {
		return nil, fmt.Errorf("expected an 'of' keyword after the expression")
	}

	// Consume the `of` keyword
	ConsumeToken(parser)

	forOfStatement, err := parseForOfStatementAfterOfKeyword(parser, targetNode)
	if err != nil {
		return nil, err
	}

	if forOfStatement == nil {
		return nil, fmt.Errorf("expected a for of statement after the 'of' keyword")
	}

	// Mark the for of statement as await.
	forOfStatement.(*ast.ForOfStatementNode).Await = true

	return forOfStatement, nil
}

func parseLexicalDeclaration(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if token.Type != lexer.Const && token.Type != lexer.Identifier {
		return nil, nil
	}

	if token.Type == lexer.Identifier && token.Value != "let" {
		return nil, nil
	}

	// Consume the const or let token.
	ConsumeToken(parser)

	isConst := token.Type == lexer.Const

	lexicalDeclaration := &ast.BasicNode{
		NodeType: ast.LexicalDeclaration,
		Parent:   nil,
		Children: make([]ast.Node, 0),
	}

	lexicalBinding, err := parseLexicalBinding(parser, isConst)
	if err != nil {
		return nil, err
	}

	if lexicalBinding == nil {
		return nil, fmt.Errorf("expected a lexical binding after the const or let keyword")
	}

	ast.AddChild(lexicalDeclaration, lexicalBinding)

	for {
		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Comma {
			break
		}

		// Consume the comma token.
		ConsumeToken(parser)

		lexicalBinding, err := parseLexicalBinding(parser, isConst)
		if err != nil {
			return nil, err
		}

		if lexicalBinding == nil {
			return nil, fmt.Errorf("expected a lexical binding after the comma")
		}

		ast.AddChild(lexicalDeclaration, lexicalBinding)
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.Semicolon {
		return nil, fmt.Errorf("expected a semicolon token after the lexical declaration")
	}

	// Consume the semicolon token.
	ConsumeToken(parser)

	return lexicalDeclaration, nil
}

func parseLexicalBinding(parser *Parser, isConst bool) (ast.Node, error) {
	targetNode, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}

	if targetNode != nil {
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}

		return &ast.LexicalBindingNode{
			Target:      targetNode,
			Initializer: initializer,
			Const:       isConst,
		}, nil
	}

	targetNode, err = parseBindingPattern(parser)
	if err != nil {
		return nil, err
	}

	if targetNode == nil {
		return nil, nil
	}

	initializer, err := parseInitializer(parser)
	if err != nil {
		return nil, err
	}

	if initializer == nil {
		return nil, fmt.Errorf("expected an initializer after the binding pattern")
	}

	return &ast.LexicalBindingNode{
		Target:      targetNode,
		Initializer: initializer,
		Const:       isConst,
	}, nil
}

func parseDeclaration(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	asyncFunctionDeclaration, err := parseAsyncFunctionOrGeneratorExpression(parser)
	if err != nil {
		return nil, err
	}

	if asyncFunctionDeclaration != nil {
		if asyncFunctionDeclaration.GetNodeType() != ast.FunctionExpression {
			return nil, fmt.Errorf("internal error: unsupported node type when parsing declaration")
		}

		// Name is not required if [Default = true]
		if !parser.AllowDefault && asyncFunctionDeclaration.(*ast.FunctionExpressionNode).Name == nil {
			return nil, fmt.Errorf("expected a binding identifier after the function keyword")
		}
		return asyncFunctionDeclaration, nil
	}

	functionDeclaration, err := parseFunctionOrGeneratorExpression(parser, false)
	if err != nil {
		return nil, err
	}

	if functionDeclaration != nil {
		if functionDeclaration.GetNodeType() != ast.FunctionExpression {
			return nil, fmt.Errorf("internal error: unsupported node type when parsing declaration")
		}
		// Name is not required if [Default = true]
		if !parser.AllowDefault && functionDeclaration.(*ast.FunctionExpressionNode).Name == nil {
			return nil, fmt.Errorf("expected a binding identifier after the function keyword")
		}
		return functionDeclaration, nil
	}

	classDeclaration, err := parseClassExpression(parser)
	if err != nil {
		return nil, err
	}

	if classDeclaration != nil {
		if classDeclaration.GetNodeType() != ast.ClassExpression {
			return nil, fmt.Errorf("internal error: unsupported node type when parsing declaration")
		}
		// Name is not required if [Default = true]
		if !parser.AllowDefault && classDeclaration.(*ast.ClassExpressionNode).Name == nil {
			return nil, fmt.Errorf("expected a binding identifier after the class keyword")
		}
		return classDeclaration, nil
	}

	// [+In = true]
	parser.PushAllowIn(true)
	lexicalDeclaration, err := parseLexicalDeclaration(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if lexicalDeclaration != nil {
		return lexicalDeclaration, nil
	}

	return nil, nil
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

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.RightBrace {
		// Consume the right brace token.
		ConsumeToken(parser)
		return &ast.BasicNode{
			NodeType: ast.Block,
			Parent:   nil,
			Children: make([]ast.Node, 0),
		}, nil
	}

	statementList, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

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

	// [+In = true]
	parser.PushAllowIn(true)
	variableDeclarationList, err := parseVariableDeclarationList(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

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

	// Allow await and yield as identifiers.
	if token.Type != lexer.Identifier && token.Type != lexer.Await && token.Type != lexer.Yield {
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
		// [+In = true]
		parser.PushAllowIn(true)
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

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
		// [+In = true]
		parser.PushAllowIn(true)
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

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

				body, err := parseArrowFunctionConciseBody(parser, false)
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

			return nil, fmt.Errorf("expected the arrow operator after the parameters")
		}

		// ArrowFunction : BindingIdentifier => ConciseBody[?Yield, ?Await]
		if token != nil && token.Type == lexer.ArrowOperator && conditionalExpression.GetNodeType() == ast.IdentifierReference {
			// Consume `=>` token
			ConsumeToken(parser)

			body, err := parseArrowFunctionConciseBody(parser, false)
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

					body, err := parseArrowFunctionConciseBody(parser, true /* Async = true */)
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

					body, err := parseArrowFunctionConciseBody(parser, true /* Async = true */)
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

func parseArrowFunctionConciseBody(parser *Parser, async bool) (ast.Node, error) {
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
		parser.PushAllowReturn(true)
		parser.PushAllowYield(false)
		parser.PushAllowAwait(async)
		body, err := parseStatementList(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowReturn()
		parser.PopAllowYield()
		parser.PopAllowAwait()

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

	parser.PushAllowAwait(async)
	assignmentExpression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowAwait()

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

	// [+In = true]
	parser.PushAllowIn(true)
	assignmentExpression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

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
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	if parser.AllowIn && token.Type == lexer.PrivateIdentifier {
		lookaheadToken := LookaheadToken(parser)

		// [+In] PrivateIdentifier in ShiftExpression[?Yield, ?Await]
		if lookaheadToken != nil && lookaheadToken.Type == lexer.In {
			// Consume the private identifier.
			ConsumeToken(parser)

			identifier := &ast.IdentifierReferenceNode{
				Parent:     nil,
				Children:   make([]ast.Node, 0),
				Identifier: token.Value,
			}

			// Consume the `in` keyword.
			CurrentToken(parser)
			ConsumeToken(parser)

			shiftExpression, err := parseShiftExpression(parser)
			if err != nil {
				return nil, err
			}

			if shiftExpression == nil {
				return nil, fmt.Errorf("expected a shift expression after the private identifier")
			}

			return &ast.RelationalExpressionNode{
				Left:  identifier,
				Right: shiftExpression,
				Operator: lexer.Token{
					Type: lexer.In,
				},
			}, nil
		}
	}

	operators := make([]lexer.TokenType, len(lexer.RelationalOperators))
	copy(operators, lexer.RelationalOperators)

	// [+In] RelationalExpression in ShiftExpression
	if parser.AllowIn {
		operators = append(operators, lexer.In)
	}

	return parseOperatorExpression(
		parser,
		operators,
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

	if parser.AllowAwait && token.Type == lexer.Await {
		// Consume the `await` keyword
		ConsumeToken(parser)

		unaryExpression, err := parseUnaryExpression(parser)
		if err != nil {
			return nil, err
		}

		if unaryExpression == nil {
			return nil, fmt.Errorf("expected a value expression after the await operator")
		}

		return &ast.AwaitExpressionNode{
			Expression: unaryExpression,
		}, nil
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
		// [+In = true]
		parser.PushAllowIn(true)
		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

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

	// [+In = true]
	parser.PushAllowIn(true)
	assignmentExpression, err := parseAssignmentExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

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

			// [+In = true]
			parser.PushAllowIn(true)
			expression, err := parseExpression(parser)
			if err != nil {
				return nil, err
			}
			parser.PopAllowIn()

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

	if token.Type != lexer.Identifier && token.Type != lexer.Await && token.Type != lexer.Yield {
		return nil, nil
	}

	// If [Await = false], allow `await` as an identifier.
	if parser.AllowAwait && token.Type == lexer.Await {
		return nil, nil
	}

	// If [Yield = false], allow `yield` as an identifier.
	if parser.AllowYield && token.Type == lexer.Yield {
		return nil, nil
	}

	// Consume the identifier token.
	ConsumeToken(parser)

	return &ast.IdentifierReferenceNode{
		Identifier: token.Value,
	}, nil
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

		// [+In = true]
		parser.PushAllowIn(true)
		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

		if assignmentExpression == nil {
			return nil, fmt.Errorf("expected an assignment expression after the '...' token")
		}

		elementListItems = append(elementListItems, &ast.SpreadElementNode{
			Expression: assignmentExpression,
		})
	} else {
		// [+In = true]
		parser.PushAllowIn(true)
		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

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
		// [+In = true]
		parser.PushAllowIn(true)
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

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

		// [+In = true]
		parser.PushAllowIn(true)
		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

		if assignmentExpression == nil {
			return nil, fmt.Errorf("expected an assignment expression after the ':' token")
		}

		return &ast.PropertyDefinitionNode{
			Key:   propertyName,
			Value: assignmentExpression,
		}, nil
	}

	// PropertyDefinition : ... AssignmentExpression
	if token.Type == lexer.Spread {
		// Consume `...` token
		ConsumeToken(parser)

		// [+In = true]
		parser.PushAllowIn(true)
		assignmentExpression, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

		if assignmentExpression == nil {
			return nil, fmt.Errorf("expected an assignment expression after the '...' token")
		}

		return &ast.SpreadElementNode{
			Expression: assignmentExpression,
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

		// [+In = true]
		parser.PushAllowIn(true)
		computedPropertyName, err := parseAssignmentExpression(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

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

	parser.PushAllowAwait(await)
	parser.PushAllowYield(yield)
	formalParameters, err := parseFormalParameters(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowAwait()
	parser.PopAllowYield()

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
	parser.PushAllowReturn(true)
	parser.PushAllowAwait(await)
	parser.PushAllowYield(yield)
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowReturn()
	parser.PopAllowAwait()
	parser.PopAllowYield()

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

	parser.PushAllowReturn(true)
	parser.PushAllowYield(false)
	parser.PushAllowAwait(false)
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowReturn()
	parser.PopAllowAwait()
	parser.PopAllowYield()

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

	parser.PushAllowReturn(true)
	parser.PushAllowYield(false)
	parser.PushAllowAwait(false)
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowReturn()
	parser.PopAllowAwait()
	parser.PopAllowYield()

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

	parser.PushAllowYield(false)
	parser.PushAllowAwait(false)
	formalParameter, err := parseBindingElement(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowAwait()
	parser.PopAllowYield()

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

	parser.PushAllowReturn(true)
	parser.PushAllowYield(false)
	parser.PushAllowAwait(false)
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowReturn()
	parser.PopAllowAwait()
	parser.PopAllowYield()

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

	parser.PushAllowAwait(async)
	parser.PushAllowYield(isGenerator)
	bindingIdentifier, err := parseBindingIdentifier(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowAwait()
	parser.PopAllowYield()

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type != lexer.LeftParen {
		return nil, fmt.Errorf("expected a '(' token after the function keyword")
	}

	// Consume `(` token
	ConsumeToken(parser)

	parser.PushAllowAwait(async)
	parser.PushAllowYield(isGenerator)
	formalParameters, err := parseFormalParameters(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowAwait()
	parser.PopAllowYield()

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

	parser.PushAllowReturn(true)
	parser.PushAllowAwait(async)
	parser.PushAllowYield(isGenerator)
	functionBody, err := parseStatementList(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowReturn()
	parser.PopAllowAwait()
	parser.PopAllowYield()

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

		// [+In = true]
		parser.PushAllowIn(true)
		initializer, err := parseInitializer(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

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

		parser.PushAllowReturn(false)
		parser.PushAllowAwait(true)
		parser.PushAllowYield(false)
		body, err := parseStatementList(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowReturn()
		parser.PopAllowAwait()
		parser.PopAllowYield()

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

		// [+In = true]
		parser.PushAllowIn(true)
		expression, err := parseExpression(parser)
		if err != nil {
			return nil, err
		}
		parser.PopAllowIn()

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
		Cover:    true,
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

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

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

	if token.Type != lexer.Super {
		return nil, nil
	}

	// Consume `super` keyword
	ConsumeToken(parser)

	token = CurrentToken(parser)
	if token == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if token.Type == lexer.Dot {
		// Consume `.` token
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Identifier {
			return nil, fmt.Errorf("expected an identifier after the '.' token")
		}

		// Consume the identifier token
		ConsumeToken(parser)

		return &ast.MemberExpressionNode{
			Object:             nil,
			Property:           nil,
			PropertyIdentifier: token.Value,
			Super:              true,
		}, nil
	}

	if token.Type != lexer.LeftBracket {
		return nil, fmt.Errorf("expected a '.' or '[' token after the 'super' keyword")
	}

	// Consume `[` token
	ConsumeToken(parser)

	// [+In = true]
	parser.PushAllowIn(true)
	expression, err := parseExpression(parser)
	if err != nil {
		return nil, err
	}
	parser.PopAllowIn()

	if expression == nil {
		return nil, fmt.Errorf("expected an expression after the '[' token")
	}

	token = CurrentToken(parser)
	if token == nil || token.Type != lexer.RightBracket {
		return nil, fmt.Errorf("expected a ']' token after the expression")
	}

	// Consume `]` token
	ConsumeToken(parser)

	return &ast.MemberExpressionNode{
		Object:             nil,
		Property:           expression,
		PropertyIdentifier: "",
		Super:              true,
	}, nil
}

func parseMetaProperty(parser *Parser) (ast.Node, error) {
	token := CurrentToken(parser)
	if token == nil {
		return nil, nil
	}

	lookaheadToken := LookaheadToken(parser)
	if token.Type == lexer.Import && lookaheadToken != nil && lookaheadToken.Type == lexer.Dot {
		// Consume `import` keyword
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Dot {
			return nil, fmt.Errorf("expected a '.' token after the 'import' keyword")
		}

		// Consume `.` token
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Identifier {
			return nil, fmt.Errorf("expected an identifier after the '.' token")
		}

		if token.Value != "meta" {
			return nil, fmt.Errorf("expected 'meta' keyword after the '.' token")
		}

		// Consume the `meta` keyword
		ConsumeToken(parser)

		return &ast.BasicNode{
			NodeType: ast.ImportMeta,
			Parent:   nil,
			Children: make([]ast.Node, 0),
		}, nil
	}

	if token.Type == lexer.New && lookaheadToken != nil && lookaheadToken.Type == lexer.Dot {
		// Consume `new` keyword
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Dot {
			return nil, fmt.Errorf("expected a '.' token after the 'new' keyword")
		}

		// Consume `.` token
		ConsumeToken(parser)

		token = CurrentToken(parser)
		if token == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if token.Type != lexer.Identifier {
			return nil, fmt.Errorf("expected an identifier after the '.' token")
		}

		if token.Value != "target" {
			return nil, fmt.Errorf("expected 'target' keyword after the '.' token")
		}

		// Consume the `target` keyword
		ConsumeToken(parser)

		return &ast.BasicNode{
			NodeType: ast.NewTarget,
			Parent:   nil,
			Children: make([]ast.Node, 0),
		}, nil
	}

	return nil, nil
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
		opNode.SetOperator(*token)

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
