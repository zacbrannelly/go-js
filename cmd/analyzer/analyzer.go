package analyzer

import (
	"slices"

	"zbrannelly.dev/go-js/cmd/parser/ast"
)

type StaticAnalyzer struct {
	rootNode   ast.Node
	goalSymbol ast.NodeType
}

type StaticAnalyzerError struct {
	Message string
}

func (e *StaticAnalyzerError) Error() string {
	return e.Message
}

func NewStaticAnalyzer() *StaticAnalyzer {
	return &StaticAnalyzer{}
}

func (a *StaticAnalyzer) Analyze(rootNode ast.Node, goalSymbol ast.NodeType) []StaticAnalyzerError {
	a.rootNode = rootNode
	a.goalSymbol = goalSymbol

	switch goalSymbol {
	case ast.Script:
		return analyzeScript(a)
	}

	return nil
}

func analyzeScript(analyzer *StaticAnalyzer) []StaticAnalyzerError {
	// ==== Lexical productions ====
	// TODO: 12.7.1.1 Early Errors for IdentifierStart :: \ UnicodeEscapeSequence and IdentifierPart :: \ UnicodeEscapeSequence
	// TODO: This lexical production needs to be supported in the lexer first.

	// TODO: 12.9.3.1 Early Errors for NumericLiteral :: LegacyOctalIntegerLiteral, DecimalIntegerLiteral :: NonOctalDecimalIntegerLiteral
	// TODO: This production isn't really supported in the lexer, unclear how to support it. Come back to this.

	// TODO: 12.9.4.1 Early Errors for EscapeSequence :: LegacyOctalEscapeSequence, EscapeSequence :: NonOctalDecimalEscapeSequence

	errors := make([]StaticAnalyzerError, 0)
	ast.Walk(analyzer.rootNode, func(node ast.Node) {
		switch node.GetNodeType() {
		case ast.IdentifierReference:
			identifierReferenceErrors := analyzeIdentifierReference(node)
			if identifierReferenceErrors != nil {
				errors = append(errors, identifierReferenceErrors...)
			}
		case ast.BindingIdentifier:
			bindingIdentifierErrors := analyzeBindingIdentifier(node)
			if bindingIdentifierErrors != nil {
				errors = append(errors, bindingIdentifierErrors...)
			}
		case ast.LabelIdentifier:
			labelIdentifierErrors := analyzeLabelIdentifier(node)
			if labelIdentifierErrors != nil {
				errors = append(errors, labelIdentifierErrors...)
			}
		}
	})

	return errors
}

func analyzeIdentifierReference(node ast.Node) []StaticAnalyzerError {
	if node.GetNodeType() != ast.IdentifierReference {
		return nil
	}

	errors := make([]StaticAnalyzerError, 0)

	identifierReference := node.(*ast.IdentifierReferenceNode)
	isStrictMode := IsStrictMode(node)

	// 13.1.1 Early Errors
	// IdentifierReference : Identifier
	if isStrictMode && slices.Contains([]string{"implements", "interface", "let", "package", "private", "protected", "public", "static", "yield"}, identifierReference.Identifier) {
		errors = append(errors, StaticAnalyzerError{Message: "In strict mode, '" + identifierReference.Identifier + "' is a reserved word"})
	}

	return errors
}

func analyzeBindingIdentifier(node ast.Node) []StaticAnalyzerError {
	if node.GetNodeType() != ast.BindingIdentifier {
		return nil
	}

	errors := make([]StaticAnalyzerError, 0)

	bindingIdentifier := node.(*ast.BindingIdentifierNode)
	isStrictMode := IsStrictMode(node)

	// 13.1.1 Early Errors
	// BindingIdentifier : Identifier
	if isStrictMode && (bindingIdentifier.Identifier == "arguments" || bindingIdentifier.Identifier == "eval") {
		errors = append(errors, StaticAnalyzerError{Message: "In strict mode, '" + bindingIdentifier.Identifier + "' is a reserved word"})
	}

	// 13.1.1 Early Errors
	// BindingIdentifier : "yield"
	// BindingIdentifier : Identifier
	if isStrictMode && slices.Contains([]string{"implements", "interface", "let", "package", "private", "protected", "public", "static", "yield"}, bindingIdentifier.Identifier) {
		errors = append(errors, StaticAnalyzerError{Message: "In strict mode, '" + bindingIdentifier.Identifier + "' is a reserved word, cannot be used as an identifier"})
	}

	return errors
}

func analyzeLabelIdentifier(node ast.Node) []StaticAnalyzerError {
	if node.GetNodeType() != ast.LabelIdentifier {
		return nil
	}

	errors := make([]StaticAnalyzerError, 0)

	labelIdentifier := node.(*ast.LabelIdentifierNode)
	isStrictMode := IsStrictMode(node)

	// 13.1.1 Early Errors
	// LabelIdentifier : Identifier
	if isStrictMode && slices.Contains([]string{"implements", "interface", "let", "package", "private", "protected", "public", "static", "yield"}, labelIdentifier.Identifier) {
		errors = append(errors, StaticAnalyzerError{Message: "In strict mode, '" + labelIdentifier.Identifier + "' is a reserved word, cannot be used as a label"})
	}

	return errors
}

func IsStrictMode(node ast.Node) bool {
	if node == nil {
		panic("node is nil")
	}

	if node.GetNodeType() == ast.Script {
		// Get the Directive Prologue of the script.
		statementList := node.GetChildren()[0].(*ast.StatementListNode)
		prologue := GetDirectivePrologue(statementList)

		// Check if the prologue contains "use strict".
		return ContainsDirective(prologue, "use strict")
	}

	script := ast.FindAncestor(node, ast.Script)
	if script != nil {
		// Is the script in strict mode? Then everything is strict mode.
		statementList := script.GetChildren()[0].(*ast.StatementListNode)
		prologue := GetDirectivePrologue(statementList)
		scriptIsStrictMode := ContainsDirective(prologue, "use strict")
		if scriptIsStrictMode {
			return true
		}

		// Class code is always strict mode.
		isClassCode := ast.IsDescendantOf(node, ast.ClassExpression)
		if isClassCode {
			return true
		}

		isPartOfFunction := ast.IsDescendantOfOneOf(
			node,
			[]ast.NodeType{
				ast.FunctionExpression,
				ast.MethodDefinition,
			},
		)

		// If the code is part of a function, check if the function is in strict mode.
		if isPartOfFunction {
			statementList := ast.FindAncestor(node, ast.StatementList)
			if statementList != nil {
				prologue := GetDirectivePrologue(statementList.(*ast.StatementListNode))
				return ContainsDirective(prologue, "use strict")
			}
		}
	}

	// TODO: Module code is always strict mode.
	// TODO: Eval code is strict if it contains the "use strict" directive, or the caller is strict.

	return false
}

func GetDirectivePrologue(node *ast.StatementListNode) []string {
	prologue := []string{}
	for _, statementListItem := range node.GetChildren() {
		if statementListItem.GetNodeType() == ast.StringLiteral {
			prologue = append(prologue, statementListItem.(*ast.StringLiteralNode).Value)
		} else {
			break
		}
	}
	return prologue
}

func ContainsDirective(prologue []string, directive string) bool {
	return slices.Contains(prologue, directive)
}
