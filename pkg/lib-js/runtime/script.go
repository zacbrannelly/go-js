package runtime

import (
	"fmt"
	"slices"

	"zbrannelly.dev/go-js/pkg/lib-js/parser"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

type Script struct {
	Realm      *Realm
	ScriptCode *ast.ScriptNode
	// TODO: [[LoadedModules]] for module support.
}

func ParseScript(sourceText string, realm *Realm) (*Script, error) {
	scriptNode, err := parser.ParseText(sourceText, ast.Script)
	if err != nil {
		return nil, err
	}

	if scriptNode == nil {
		return nil, fmt.Errorf("expected script node, got nil")
	}

	if scriptNode.GetNodeType() != ast.Script {
		return nil, fmt.Errorf("expected script node, got %s", ast.NodeTypeToString[scriptNode.GetNodeType()])
	}

	return &Script{
		Realm:      realm,
		ScriptCode: scriptNode.(*ast.ScriptNode),
	}, nil
}

func (s *Script) Evaluate(runtime *Runtime) *Completion {
	globalEnv := s.Realm.GlobalEnv
	scriptContext := &ExecutionContext{
		Function:            nil,
		Realm:               s.Realm,
		Script:              s,
		LexicalEnvironment:  globalEnv,
		VariableEnvironment: globalEnv,
		PrivateEnvironment:  globalEnv,
		VM:                  nil,
	}

	// Make the script the running context.
	runtime.PushExecutionContext(scriptContext)

	script := scriptContext.Script.ScriptCode
	result := GlobalDeclarationInstantiation(runtime, script, scriptContext.Realm.GlobalEnv)
	if result.Type != Normal {
		runtime.PopExecutionContext()
		return result
	}

	result = EvaluateScript(runtime, script)
	if result.Type == Normal && result.Value == nil {
		result = NewNormalCompletion(NewUndefinedValue())
	}

	runtime.PopExecutionContext()
	return result
}

func GlobalDeclarationInstantiation(runtime *Runtime, script *ast.ScriptNode, env *GlobalEnvironment) *Completion {
	// Get all lexical declarations.
	lexNames := LexicallyDeclaredNames(script)

	// Get all variable declarations.
	varNames := VarDeclaredNames(script)

	for _, name := range lexNames {
		// Check if there is already a lexical declaration for this name. If so throw a SyntaxError.
		if HasLexicalDeclaration(env, name) {
			return NewThrowCompletion(NewSyntaxError(fmt.Sprintf("Identifier '%s' has already been declared", name)))
		}

		// Check if there is already a "restricted global property" for this name (which includes var / function declarations). If so throw a SyntaxError.
		hasRestrictedGlobalPropertyCompletion := HasRestrictedGlobalProperty(env, name)
		if hasRestrictedGlobalPropertyCompletion.Type != Normal {
			return hasRestrictedGlobalPropertyCompletion
		}
		hasRestrictedGlobalProperty := hasRestrictedGlobalPropertyCompletion.Value.(*JavaScriptValue)
		if hasRestrictedGlobalProperty.Value.(*Boolean).Value {
			return NewThrowCompletion(NewSyntaxError(fmt.Sprintf("Identifier '%s' has already been declared", name)))
		}
	}

	for _, name := range varNames {
		// Check if there is already a lexical declaration for this name. If so throw a SyntaxError.
		if HasLexicalDeclaration(env, name) {
			return NewThrowCompletion(NewSyntaxError(fmt.Sprintf("Identifier '%s' has already been declared", name)))
		}
	}

	varDeclarations := VarScopedDeclarations(script)

	declaredFunctionNames := make([]string, 0)
	functionsToInitialize := make([]*ast.FunctionExpressionNode, 0)

	for i := len(varDeclarations) - 1; i >= 0; i-- {
		declaration := varDeclarations[i]
		declarationType := declaration.GetNodeType()

		if declarationType != ast.VariableDeclaration && declarationType != ast.BindingIdentifier && !IsForBinding(declaration) {
			if declarationType != ast.FunctionExpression {
				panic(fmt.Sprintf("Assert failed: Unexpected declaration type: %s", ast.NodeTypeToString[declarationType]))
			}

			functionExpression := declaration.(*ast.FunctionExpressionNode)
			functionName := functionExpression.GetName().(*ast.BindingIdentifierNode).Identifier

			if !slices.Contains(declaredFunctionNames, functionName) {
				// Check if the function is definable.
				definableCompletion := CanDeclareGlobalFunction(env, functionName)
				if definableCompletion.Type != Normal {
					return definableCompletion
				}
				if !definableCompletion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
					return NewThrowCompletion(NewTypeError(fmt.Sprintf("Function with name '%s' cannot be defined in this context", functionName)))
				}

				declaredFunctionNames = append(declaredFunctionNames, functionName)
				functionsToInitialize = append(functionsToInitialize, functionExpression)
			}
		}
	}

	declaredVarNames := make([]string, 0)

	for _, declaration := range varDeclarations {
		declarationType := declaration.GetNodeType()
		if declarationType != ast.VariableDeclaration && declarationType != ast.BindingIdentifier && !IsForBinding(declaration) {
			continue
		}

		boundNames := BoundNames(declaration)
		for _, name := range boundNames {
			if slices.Contains(declaredFunctionNames, name) {
				continue
			}

			definableCompletion := CanDeclareGlobalVar(env, name)
			if definableCompletion.Type != Normal {
				return definableCompletion
			}

			definableVal := definableCompletion.Value.(*JavaScriptValue)
			if definableVal.Type != TypeBoolean {
				panic("Assert failed: Expected a boolean value for CanDeclareGlobalVar.")
			}

			if !definableVal.Value.(*Boolean).Value {
				return NewThrowCompletion(NewTypeError(fmt.Sprintf("Variable with name '%s' cannot be defined in this context", name)))
			}

			if slices.Contains(declaredVarNames, name) {
				continue
			}
			declaredVarNames = append(declaredVarNames, name)
		}
	}

	// TODO: Annex B.3.2.2 has additional steps for web browsers.

	// Get all lexical declarations and create bindings for them (but don't initialize them).
	lexDeclarations := LexicallyScopedDeclarations(script)
	for _, declaration := range lexDeclarations {
		declarationBoundNames := BoundNames(declaration)
		for _, name := range declarationBoundNames {
			if IsConstantDeclaration(declaration) {
				completion := env.CreateImmutableBinding(name, true)
				if completion.Type != Normal {
					return completion
				}
			} else {
				completion := env.CreateMutableBinding(name, false)
				if completion.Type != Normal {
					return completion
				}
			}
		}
	}

	for _, function := range functionsToInitialize {
		boundNames := BoundNames(function)

		// Assert that there is only one bound name for the function.
		if len(boundNames) != 1 {
			panic(fmt.Sprintf("Assert failed: Unexpected number of bound names for function to initialize: %d", len(boundNames)))
		}

		// Create a Function object.
		functionObject := InstantiateFunctionObject(runtime, function, env, nil)

		// Create a binding for the function, and initialize it.
		functionName := boundNames[0]
		completion := env.CreateGlobalFunctionBinding(runtime, functionName, functionObject, false)
		if completion.Type != Normal {
			return completion
		}
	}

	for _, varName := range declaredVarNames {
		completion := env.CreateGlobalVarBinding(runtime, varName, false)
		if completion.Type != Normal {
			return completion
		}
	}

	return NewUnusedCompletion()
}

func IsConstantDeclaration(node ast.Node) bool {
	if node.GetNodeType() == ast.LexicalBinding {
		return node.(*ast.LexicalBindingNode).Const
	}

	if node.GetNodeType() != ast.LexicalDeclaration {
		return false
	}

	binding := node.GetChildren()[0]
	return binding.(*ast.LexicalBindingNode).Const
}

func IsForBinding(node ast.Node) bool {
	if node.GetNodeType() != ast.BindingIdentifier && node.GetNodeType() != ast.ArrayBindingPattern && node.GetNodeType() != ast.ObjectBindingPattern {
		return false
	}

	if node.GetParent() == nil {
		return false
	}

	if node.GetParent().GetNodeType() != ast.ForInStatement && node.GetParent().GetNodeType() != ast.ForOfStatement {
		return false
	}

	return true
}

func HasLexicalDeclaration(env *GlobalEnvironment, name string) bool {
	return env.DeclarativeRecord.HasBinding(name)
}

func HasRestrictedGlobalProperty(env *GlobalEnvironment, name string) *Completion {
	propertyCompletion := env.ObjectRecord.BindingObject.GetOwnProperty(NewStringValue(name))
	if propertyCompletion.Type != Normal {
		return propertyCompletion
	}

	if propertyCompletion.Value == nil {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	propertyDesc := propertyCompletion.Value.(PropertyDescriptor)
	return NewNormalCompletion(NewBooleanValue(!propertyDesc.GetConfigurable()))
}

func LexicallyDeclaredNames(node ast.Node) []string {
	// Script
	if node.GetNodeType() == ast.Script {
		statementList := node.(*ast.ScriptNode).GetChildren()[0]
		return TopLevelLexicallyDeclaredNames(statementList)
	}

	// LabelledStatement
	if node.GetNodeType() == ast.LabelledStatement {
		labelledStatement := node.(*ast.LabelledStatementNode)
		labelledItem := labelledStatement.GetLabelledItem()
		if labelledItem != nil && labelledStatement.GetLabelledItem().GetNodeType() == ast.FunctionExpression {
			return BoundNames(labelledStatement.GetLabelledItem())
		}
		return []string{}
	}

	// Declaration
	if node.GetNodeType() == ast.LexicalDeclaration {
		return BoundNames(node)
	}
	if node.GetNodeType() == ast.ClassExpression && node.(*ast.ClassExpressionNode).Declaration {
		return BoundNames(node)
	}

	// StatementList
	if node.GetNodeType() == ast.StatementList {
		names := make([]string, 0)
		for _, child := range node.GetChildren() {
			names = append(names, LexicallyDeclaredNames(child)...)
		}
		return names
	}

	// SwitchStatement
	if node.GetNodeType() == ast.SwitchStatement {
		names := make([]string, 0)
		for _, child := range node.GetChildren() {
			if child.GetNodeType() == ast.StatementList {
				names = append(names, LexicallyDeclaredNames(child)...)
			}
		}
		return names
	}

	// TODO: Complete this syntax-directed operation (module / export nodes).

	return []string{}
}

func VarDeclaredNames(node ast.Node) []string {
	// TODO: Complete this syntax-directed operation.
	return []string{}
}

func TopLevelLexicallyDeclaredNames(node ast.Node) []string {
	if node.GetNodeType() == ast.StatementList {
		names := make([]string, 0)
		for _, child := range node.GetChildren() {
			names = append(names, TopLevelLexicallyDeclaredNames(child)...)
		}
		return names
	}

	if node.GetNodeType() == ast.LexicalDeclaration {
		return BoundNames(node)
	}

	if node.GetNodeType() == ast.ClassExpression && node.(*ast.ClassExpressionNode).Declaration {
		return BoundNames(node)
	}

	return []string{}
}

func BoundNames(node ast.Node) []string {
	if node.GetNodeType() == ast.BindingIdentifier {
		return []string{node.(*ast.BindingIdentifierNode).Identifier}
	}

	if node.GetNodeType() == ast.BindingElement {
		bindingElement := node.(*ast.BindingElementNode)
		return BoundNames(bindingElement.GetTarget())
	}

	if node.GetNodeType() == ast.BindingProperty {
		bindingProperty := node.(*ast.BindingPropertyNode)
		if bindingProperty.GetBindingElement() == nil {
			return BoundNames(bindingProperty.GetTarget())
		} else {
			return BoundNames(bindingProperty.GetBindingElement())
		}
	}

	// NOTE: This case was missing from the spec.
	if node.GetNodeType() == ast.BindingRestProperty {
		bindingRestProperty := node.(*ast.BindingRestNode)
		if bindingRestProperty.GetBindingPattern() == nil {
			return BoundNames(bindingRestProperty.GetIdentifier())
		} else {
			return BoundNames(bindingRestProperty.GetBindingPattern())
		}
	}

	if node.GetNodeType() == ast.ObjectBindingPattern {
		objectBindingPattern := node.(*ast.ObjectBindingPatternNode)
		names := make([]string, 0)
		for _, property := range objectBindingPattern.GetProperties() {
			names = append(names, BoundNames(property)...)
		}
		return names
	}

	if node.GetNodeType() == ast.ArrayBindingPattern {
		arrayBindingPattern := node.(*ast.ArrayBindingPatternNode)
		names := make([]string, 0)
		for _, element := range arrayBindingPattern.GetElements() {
			names = append(names, BoundNames(element)...)
		}
		return names
	}

	if node.GetNodeType() == ast.LexicalBinding {
		binding := node.(*ast.LexicalBindingNode)
		return BoundNames(binding.GetTarget())
	}

	if node.GetNodeType() == ast.LexicalDeclaration {
		names := make([]string, 0)
		for _, child := range node.GetChildren() {
			names = append(names, BoundNames(child)...)
		}
		return names
	}

	if node.GetNodeType() == ast.ClassExpression {
		classExpression := node.(*ast.ClassExpressionNode)

		// ClassDeclaration
		if classExpression.Declaration {
			if classExpression.GetName() == nil {
				return []string{"*default*"}
			}
			return BoundNames(classExpression.GetName())
		}

		// TODO: Handle ClassExpression.
	}

	if node.GetNodeType() == ast.VariableDeclaration {
		return BoundNames(node.GetChildren()[0])
	}

	if node.GetNodeType() == ast.FunctionExpression {
		functionExpression := node.(*ast.FunctionExpressionNode)
		if functionExpression.Declaration {
			return BoundNames(functionExpression.GetName())
		}
	}

	// TODO: Complete this syntax-directed operation.
	panic("Unhandled node type in BoundNames: " + ast.NodeTypeToString[node.GetNodeType()])
}

func VarScopedDeclarations(node ast.Node) []ast.Node {
	if node == nil {
		return []ast.Node{}
	}

	if node.GetNodeType() == ast.StatementList {
		parent := node.GetParent()

		// FunctionStatementList : StatementList
		if parent != nil && parent.GetNodeType() == ast.FunctionExpression {
			return TopLevelVarScopedDeclarations(node)
		}

		// ScriptBody : StatementList
		if parent != nil && parent.GetNodeType() == ast.Script {
			return TopLevelVarScopedDeclarations(node)
		}

		// ClassStaticBlockStatementList : StatementList
		if parent != nil && parent.GetNodeType() == ast.ClassStaticBlock {
			return TopLevelVarScopedDeclarations(node)
		}

		// StatementList : StatementList StatementListItem
		declarations := make([]ast.Node, 0)
		for _, child := range node.GetChildren() {
			declarations = append(declarations, VarScopedDeclarations(child)...)
		}
		return declarations
	}

	// VariableDeclarationList : VariableDeclaration
	if node.GetNodeType() == ast.VariableDeclaration {
		return []ast.Node{node}
	}

	// VariableDeclarationList : VariableDeclarationList VariableDeclaration
	if node.GetNodeType() == ast.VariableDeclarationList {
		declarations := make([]ast.Node, 0)
		for _, child := range node.GetChildren() {
			declarations = append(declarations, VarScopedDeclarations(child)...)
		}
		return declarations
	}

	// VariableStatement : var VariableDeclarationList ;
	// NOTE: No definition of thi code path in the spec, but our parser exports a VariableStatement with all the variable declarations, so we need to walk it.
	if node.GetNodeType() == ast.VariableStatement {
		return VarScopedDeclarations(node.GetChildren()[0])
	}

	// IfStatement
	if node.GetNodeType() == ast.IfStatement {
		declarations := make([]ast.Node, 0)
		ifStatement := node.(*ast.IfStatementNode)
		declarations = append(declarations, VarScopedDeclarations(ifStatement.GetTrueStatement())...)
		declarations = append(declarations, VarScopedDeclarations(ifStatement.GetElseStatement())...)
		return declarations
	}

	// DoWhileStatement
	if node.GetNodeType() == ast.DoWhileStatement {
		doWhileStatement := node.(*ast.DoWhileStatementNode)
		return VarScopedDeclarations(doWhileStatement.GetStatement())
	}

	// WhileStatement
	if node.GetNodeType() == ast.WhileStatement {
		whileStatement := node.(*ast.WhileStatementNode)
		return VarScopedDeclarations(whileStatement.GetStatement())
	}

	// ForStatement
	if node.GetNodeType() == ast.ForStatement {
		declarations := make([]ast.Node, 0)
		forStatement := node.(*ast.ForStatementNode)
		declarations = append(declarations, VarScopedDeclarations(forStatement.GetInitializer())...)
		declarations = append(declarations, VarScopedDeclarations(forStatement.GetBody())...)
		return declarations
	}

	// ForInStatement
	if node.GetNodeType() == ast.ForInStatement {
		declarations := make([]ast.Node, 0)
		forInStatement := node.(*ast.ForInStatementNode)
		declarations = append(declarations, VarScopedDeclarations(forInStatement.GetTarget())...)
		declarations = append(declarations, VarScopedDeclarations(forInStatement.GetBody())...)
		return declarations
	}

	// ForOfStatement
	if node.GetNodeType() == ast.ForOfStatement {
		declarations := make([]ast.Node, 0)
		forOfStatement := node.(*ast.ForOfStatementNode)
		declarations = append(declarations, VarScopedDeclarations(forOfStatement.GetTarget())...)
		declarations = append(declarations, VarScopedDeclarations(forOfStatement.GetBody())...)
		return declarations
	}

	// WithStatement
	if node.GetNodeType() == ast.WithStatement {
		withStatement := node.(*ast.WithStatementNode)
		return VarScopedDeclarations(withStatement.GetBody())
	}

	// SwitchStatement
	if node.GetNodeType() == ast.SwitchStatement {
		declarations := make([]ast.Node, 0)
		for _, child := range node.GetChildren() {
			if child.GetNodeType() == ast.StatementList {
				declarations = append(declarations, VarScopedDeclarations(child)...)
			}
		}
		return declarations
	}

	// LabelledStatement : LabelIdentifier : LabelledItem
	if node.GetNodeType() == ast.LabelledStatement {
		labelledStatement := node.(*ast.LabelledStatementNode)
		return VarScopedDeclarations(labelledStatement.GetLabelledItem())
	}

	// TryStatement
	if node.GetNodeType() == ast.TryStatement {
		tryStatement := node.(*ast.TryStatementNode)
		declarations := make([]ast.Node, 0)
		declarations = append(declarations, VarScopedDeclarations(tryStatement.GetBlock())...)
		declarations = append(declarations, VarScopedDeclarations(tryStatement.GetCatch())...)
		declarations = append(declarations, VarScopedDeclarations(tryStatement.GetFinally())...)
		return declarations
	}

	// ScriptBody : StatementList
	if node.GetNodeType() == ast.Script {
		script := node.(*ast.ScriptNode)
		return TopLevelVarScopedDeclarations(script.GetChildren()[0])
	}

	// TODO: Complete this syntax-directed operation (support module nodes).

	return []ast.Node{}
}

func TopLevelVarScopedDeclarations(node ast.Node) []ast.Node {
	if node == nil {
		return []ast.Node{}
	}

	if node.GetNodeType() == ast.StatementList {
		declarations := make([]ast.Node, 0)
		for _, child := range node.GetChildren() {
			declarations = append(declarations, TopLevelVarScopedDeclarations(child)...)
		}
		return declarations
	}

	if node.GetNodeType() == ast.LabelledStatement {
		labelledStatement := node.(*ast.LabelledStatementNode)
		labelledItem := labelledStatement.GetLabelledItem()

		// LabelledItem : FunctionDeclaration
		if labelledItem.GetNodeType() == ast.FunctionExpression && labelledItem.(*ast.FunctionExpressionNode).Declaration {
			return []ast.Node{labelledItem}
		}

		return VarScopedDeclarations(labelledItem)
	}

	// Declaration : HoistableDeclaration
	if node.GetNodeType() == ast.FunctionExpression && node.(*ast.FunctionExpressionNode).Declaration {
		return []ast.Node{node}
	}

	return VarScopedDeclarations(node)
}

func LexicallyScopedDeclarations(node ast.Node) []ast.Node {
	if node == nil {
		return []ast.Node{}
	}

	if node.GetNodeType() == ast.Block {
		return LexicallyScopedDeclarations(node.GetChildren()[0])
	}

	// StatementList : StatementList StatementListItem
	if node.GetNodeType() == ast.StatementList {
		// FunctionStatementList
		parent := node.GetParent()
		if parent != nil && parent.GetNodeType() == ast.FunctionExpression {
			return TopLevelLexicallyScopedDeclarations(node)
		}

		// ScriptBody : StatementList
		if parent != nil && parent.GetNodeType() == ast.Script {
			return TopLevelLexicallyScopedDeclarations(node)
		}

		// ClassStaticBlockStatementList : StatementList
		if parent != nil && parent.GetNodeType() == ast.ClassStaticBlock {
			return TopLevelLexicallyScopedDeclarations(node)
		}

		declarations := make([]ast.Node, 0)
		for _, child := range node.GetChildren() {
			declarations = append(declarations, LexicallyScopedDeclarations(child)...)
		}
		return declarations
	}

	if node.GetNodeType() == ast.LabelledStatement {
		labelledStatement := node.(*ast.LabelledStatementNode)
		labelledItem := labelledStatement.GetLabelledItem()

		// LabelledItem : FunctionDeclaration
		if labelledItem.GetNodeType() == ast.FunctionExpression && labelledItem.(*ast.FunctionExpressionNode).Declaration {
			return []ast.Node{labelledItem}
		}

		// LabelledItem : Statement
		return []ast.Node{}
	}

	// Declaration : LexicalDeclaration
	if node.GetNodeType() == ast.LexicalDeclaration {
		return []ast.Node{node}
	}

	// Declaration : HoistableDeclaration
	if node.GetNodeType() == ast.FunctionExpression && node.(*ast.FunctionExpressionNode).Declaration {
		return []ast.Node{node}
	}

	// Declaration : ClassDeclaration
	if node.GetNodeType() == ast.ClassExpression {
		classExpression := node.(*ast.ClassExpressionNode)
		if classExpression.Declaration {
			return []ast.Node{classExpression}
		}
	}

	if node.GetNodeType() == ast.Script {
		script := node.(*ast.ScriptNode)
		return TopLevelLexicallyScopedDeclarations(script.GetChildren()[0])
	}

	// NOTE: The spec contains cases for CaseBlock, CaseClause, and CaseDefault, but nothing for SwitchStatement.
	// Our parser exports these productions as just StatementList, so that should handle these cases.

	// TODO: Complete this syntax-directed operation (support module nodes).

	return []ast.Node{}
}

func TopLevelLexicallyScopedDeclarations(node ast.Node) []ast.Node {
	if node == nil {
		return []ast.Node{}
	}

	if node.GetNodeType() == ast.StatementList {
		declarations := make([]ast.Node, 0)
		for _, child := range node.GetChildren() {
			declarations = append(declarations, TopLevelLexicallyScopedDeclarations(child)...)
		}
		return declarations
	}

	// Declaration : HoistableDeclaration
	if node.GetNodeType() == ast.FunctionExpression && node.(*ast.FunctionExpressionNode).Declaration {
		return []ast.Node{}
	}

	if node.GetNodeType() == ast.ClassExpression && node.(*ast.ClassExpressionNode).Declaration {
		return []ast.Node{node}
	}

	if node.GetNodeType() == ast.LexicalDeclaration {
		return []ast.Node{node}
	}

	return []ast.Node{}
}
