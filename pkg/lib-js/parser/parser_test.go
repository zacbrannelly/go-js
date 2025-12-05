package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func expectScriptNodeAndGetChildren(t *testing.T, node ast.Node) []ast.Node {
	assert.Equal(t, ast.Script, node.GetNodeType())
	assert.Equal(t, 1, len(node.GetChildren()))

	child := node.GetChildren()[0]
	assert.NotNil(t, child)
	assert.Equal(t, ast.StatementList, child.GetNodeType())

	return child.GetChildren()
}

func parseScriptAndExpectNoErrors(t *testing.T, input string) []ast.Node {
	node, err := ParseText(input, ast.Script)
	if err != nil {
		t.Fatalf("Error parsing script: %v", err)
	}

	return expectScriptNodeAndGetChildren(t, node)
}

func expectNodeType[T ast.Node](t *testing.T, node ast.Node, expectedType ast.NodeType) T {
	assert.Equal(t, expectedType, node.GetNodeType(), "Expected %v, got %v", ast.NodeTypeToString[expectedType], ast.NodeTypeToString[node.GetNodeType()])
	return node.(T)
}

func expectSingleChild(t *testing.T, scriptBody []ast.Node) ast.Node {
	assert.Equal(t, 1, len(scriptBody), "Expected 1 child, got %v", len(scriptBody))
	return scriptBody[0]
}

func expectScriptValue[T ast.Node](t *testing.T, input string, expectedType ast.NodeType) T {
	scriptBody := parseScriptAndExpectNoErrors(t, input)
	node := expectSingleChild(t, scriptBody)
	return expectNodeType[T](t, node, expectedType)
}

// PrimaryExpression : this
func TestThisExpression(t *testing.T) {
	expectScriptValue[*ast.BasicNode](t, "this;", ast.ThisExpression)
}

// PrimaryExpression : IdentifierReference
func TestIdentifierReferenceExpression(t *testing.T) {
	identifierReference := expectScriptValue[*ast.IdentifierReferenceNode](t, "foo;", ast.IdentifierReference)
	assert.Equal(t, "foo", identifierReference.Identifier, "Expected identifier 'foo', got %s", identifierReference.Identifier)
}

// PrimaryExpression : Literal
func TestLiteralExpression(t *testing.T) {
	// NumericLiteral
	numericLiteral := expectScriptValue[*ast.NumericLiteralNode](t, "123;", ast.NumericLiteral)
	assert.Equal(t, float64(123), numericLiteral.Value, "Expected value 123, got %f", numericLiteral.Value)

	// NumericLiteral - Hex
	numericLiteral = expectScriptValue[*ast.NumericLiteralNode](t, "0xccf0fadf;", ast.NumericLiteral)
	assert.Equal(t, float64(0xccf0fadf), numericLiteral.Value, "Expected value 0xccf0fadf, got %f", numericLiteral.Value)

	// NumericLiteral - Binary
	numericLiteral = expectScriptValue[*ast.NumericLiteralNode](t, "0b11001100111100001111101011011111;", ast.NumericLiteral)
	assert.Equal(t, float64(0b11001100111100001111101011011111), numericLiteral.Value, "Expected value 0b11001100111100001111101011011111, got %f", numericLiteral.Value)

	// NumericLiteral - Octal
	numericLiteral = expectScriptValue[*ast.NumericLiteralNode](t, "0o7654321;", ast.NumericLiteral)
	assert.Equal(t, float64(0o7654321), numericLiteral.Value, "Expected value 0o7654321, got %f", numericLiteral.Value)

	// StringLiteral
	stringLiteral := expectScriptValue[*ast.StringLiteralNode](t, "\"foo\";", ast.StringLiteral)
	assert.Equal(t, "foo", stringLiteral.Value, "Expected value 'foo', got %s", stringLiteral.Value)

	// BooleanLiteral
	booleanLiteral := expectScriptValue[*ast.BooleanLiteralNode](t, "true;", ast.BooleanLiteral)
	assert.True(t, booleanLiteral.Value, "Expected value true, got %t", booleanLiteral.Value)

	// NullLiteral
	expectScriptValue[*ast.BasicNode](t, "null;", ast.NullLiteral)
}

// PrimaryExpression : ArrayLiteral
func TestArrayLiteralExpression(t *testing.T) {
	arrayLiteral := expectScriptValue[*ast.BasicNode](t, "[1, 2, 3];", ast.ArrayLiteral)
	assert.Equal(t, 3, len(arrayLiteral.Children), "Expected 3 elements, got %d", len(arrayLiteral.Children))

	for i, child := range arrayLiteral.Children {
		numericLiteral := expectNodeType[*ast.NumericLiteralNode](t, child, ast.NumericLiteral)
		assert.Equal(t, float64(i+1), numericLiteral.Value, "Expected value %d, got %f", i+1, numericLiteral.Value)
	}

	// Test elision
	arrayLiteral = expectScriptValue[*ast.BasicNode](t, "[1, , 3];", ast.ArrayLiteral)
	assert.Equal(t, 3, len(arrayLiteral.Children), "Expected 3 elements, got %d", len(arrayLiteral.Children))

	for i, child := range arrayLiteral.Children {
		if i == 1 {
			expectNodeType[*ast.BasicNode](t, child, ast.Elision)
		}
	}
}

// PrimaryExpression : ObjectLiteral
func TestObjectLiteralExpression(t *testing.T) {
	// TODO: This only works because of the parentheses around the object literal.
	// TODO: This needs to be fixed.
	script := `
		({
			foo: 1,
			bar: 2,
			"baz": 3,
			[identifier]: 4,
			...spread,
			method() {
				console.log("method");
			},
			5.5: 6,
		});
	`
	objectLiteral := expectScriptValue[*ast.ObjectLiteralNode](t, script, ast.ObjectLiteral)
	assert.Equal(t, 7, len(objectLiteral.GetProperties()), "Expected 7 elements, got %d", len(objectLiteral.GetProperties()))

	// Check the identifier properties
	for _, property := range objectLiteral.GetProperties()[:2] {
		propertyDefinition := expectNodeType[*ast.PropertyDefinitionNode](t, property, ast.PropertyDefinition)
		expectNodeType[*ast.IdentifierNameNode](t, propertyDefinition.GetKey(), ast.IdentifierName)
		expectNodeType[*ast.NumericLiteralNode](t, propertyDefinition.GetValue(), ast.NumericLiteral)
	}

	// Check the string literal properties
	bazProperty := objectLiteral.GetProperties()[2]
	bazPropertyDefinition := expectNodeType[*ast.PropertyDefinitionNode](
		t,
		bazProperty,
		ast.PropertyDefinition,
	)
	bazPropertyDefinitionKey := expectNodeType[*ast.StringLiteralNode](
		t,
		bazPropertyDefinition.GetKey(),
		ast.StringLiteral,
	)

	assert.Equal(t, "baz", bazPropertyDefinitionKey.Value, "Expected value 'baz', got %s", bazPropertyDefinitionKey.Value)

	// Check the identifier property
	identifierPropertyDefinition := expectNodeType[*ast.PropertyDefinitionNode](
		t,
		objectLiteral.GetProperties()[3],
		ast.PropertyDefinition,
	)
	identifierPropertyKey := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		identifierPropertyDefinition.GetKey(),
		ast.IdentifierReference,
	)

	assert.Equal(t, "identifier", identifierPropertyKey.Identifier, "Expected identifier 'identifier', got %s", identifierPropertyKey.Identifier)

	expectNodeType[*ast.NumericLiteralNode](
		t,
		identifierPropertyDefinition.GetValue(),
		ast.NumericLiteral,
	)

	// Check the spread property
	spreadElement := expectNodeType[*ast.SpreadElementNode](
		t, objectLiteral.GetProperties()[4], ast.SpreadElement,
	)
	identifierReference := expectNodeType[*ast.IdentifierReferenceNode](
		t, spreadElement.GetExpression(), ast.IdentifierReference,
	)

	assert.Equal(t, "spread", identifierReference.Identifier, "Expected identifier 'spread', got %s", identifierReference.Identifier)

	// Check the method property
	methodProperty := objectLiteral.GetProperties()[5]
	methodPropertyDefinition := expectNodeType[*ast.MethodDefinitionNode](
		t, methodProperty, ast.MethodDefinition,
	)
	methodName := expectNodeType[*ast.IdentifierNameNode](
		t,
		methodPropertyDefinition.GetName(),
		ast.IdentifierName,
	)
	assert.Equal(t, "method", methodName.Identifier, "Expected identifier 'method', got %s", methodName.Identifier)
	expectNodeType[*ast.StatementListNode](
		t,
		methodPropertyDefinition.GetBody(),
		ast.StatementList,
	)

	// Check the numeric literal property
	numericLiteralProperty := objectLiteral.GetProperties()[6]
	numericLiteralPropertyDefinition := expectNodeType[*ast.PropertyDefinitionNode](
		t,
		numericLiteralProperty,
		ast.PropertyDefinition,
	)
	expectNodeType[*ast.NumericLiteralNode](
		t,
		numericLiteralPropertyDefinition.GetValue(),
		ast.NumericLiteral,
	)
}

// PrimaryExpression : FunctionExpression
func TestFunctionExpression(t *testing.T) {
	// Test anonymous function expression
	functionExpression := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(function() {});",
		ast.FunctionExpression,
	)
	assert.Nil(t, functionExpression.GetName(), "Expected anonymous function, but got a named function")
	expectNodeType[*ast.StatementListNode](
		t,
		functionExpression.GetBody(),
		ast.StatementList,
	)

	// Test named function expression
	functionExpression = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(function foo() {});",
		ast.FunctionExpression,
	)
	functionName := expectNodeType[*ast.BindingIdentifierNode](
		t,
		functionExpression.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", functionName.Identifier, "Expected function name 'foo', got %s", functionName.Identifier)
	expectNodeType[*ast.StatementListNode](
		t,
		functionExpression.GetBody(),
		ast.StatementList,
	)
}

// PrimaryExpression : ClassExpression
func TestClassExpression(t *testing.T) {
	// Test anonymous class expression without heritage
	classExpression := expectScriptValue[*ast.ClassExpressionNode](
		t,
		"(class {});",
		ast.ClassExpression,
	)
	assert.False(t, classExpression.Declaration, "Expected expression, but got declaration")
	assert.Nil(t, classExpression.GetName(), "Expected anonymous class, but got a named class")
	assert.Nil(t, classExpression.GetHeritage(), "Expected no heritage, but got heritage")

	// Test named class expression without heritage
	classExpression = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"(class Foo {});",
		ast.ClassExpression,
	)
	assert.False(t, classExpression.Declaration, "Expected expression, but got declaration")
	className := expectNodeType[*ast.BindingIdentifierNode](
		t,
		classExpression.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "Foo", className.Identifier, "Expected class name 'Foo', got %s", className.Identifier)
	assert.Nil(t, classExpression.GetHeritage(), "Expected no heritage, but got heritage")

	// Test anonymous class expression with heritage
	classExpression = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"(class extends Bar {});",
		ast.ClassExpression,
	)
	assert.False(t, classExpression.Declaration, "Expected expression, but got declaration")
	assert.Nil(t, classExpression.GetName(), "Expected anonymous class, but got a named class")
	expectNodeType[*ast.IdentifierReferenceNode](
		t,
		classExpression.GetHeritage(),
		ast.IdentifierReference,
	)

	// Test named class expression with heritage
	classExpression = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"(class Foo extends Bar {});",
		ast.ClassExpression,
	)
	assert.False(t, classExpression.Declaration, "Expected expression, but got declaration")
	className = expectNodeType[*ast.BindingIdentifierNode](
		t,
		classExpression.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "Foo", className.Identifier, "Expected class name 'Foo', got %s", className.Identifier)
	expectNodeType[*ast.IdentifierReferenceNode](
		t,
		classExpression.GetHeritage(),
		ast.IdentifierReference,
	)
}

// PrimaryExpression : GeneratorExpression
func TestGeneratorExpression(t *testing.T) {
	// Test anonymous generator expression
	generatorExpression := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(function* () {});",
		ast.FunctionExpression,
	)
	assert.True(t, generatorExpression.Generator, "Expected generator function")
	assert.Nil(t, generatorExpression.GetName(), "Expected anonymous generator, but got a named generator")
	expectNodeType[*ast.StatementListNode](
		t,
		generatorExpression.GetBody(),
		ast.StatementList,
	)

	// Test named generator expression
	generatorExpression = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(function* gen() {});",
		ast.FunctionExpression,
	)
	assert.True(t, generatorExpression.Generator, "Expected generator function")
	generatorName := expectNodeType[*ast.BindingIdentifierNode](
		t,
		generatorExpression.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "gen", generatorName.Identifier, "Expected generator name 'gen', got %s", generatorName.Identifier)
	expectNodeType[*ast.StatementListNode](
		t,
		generatorExpression.GetBody(),
		ast.StatementList,
	)
}

// PrimaryExpression : AsyncFunctionExpression
func TestAsyncFunctionExpression(t *testing.T) {
	// Test anonymous async function expression
	asyncFunctionExpression := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(async function() {});",
		ast.FunctionExpression,
	)
	assert.True(t, asyncFunctionExpression.Async, "Expected async function")
	assert.Nil(t, asyncFunctionExpression.GetName(), "Expected anonymous async function, but got a named async function")
	expectNodeType[*ast.StatementListNode](
		t,
		asyncFunctionExpression.GetBody(),
		ast.StatementList,
	)

	// Test named async function expression
	asyncFunctionExpression = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(async function foo() {});",
		ast.FunctionExpression,
	)
	assert.True(t, asyncFunctionExpression.Async, "Expected async function")
	asyncFunctionName := expectNodeType[*ast.BindingIdentifierNode](
		t,
		asyncFunctionExpression.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", asyncFunctionName.Identifier, "Expected async function name 'foo', got %s", asyncFunctionName.Identifier)
	expectNodeType[*ast.StatementListNode](
		t,
		asyncFunctionExpression.GetBody(),
		ast.StatementList,
	)

	// Test function parameters
	functionWithParams := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(async function(a, b, c = 1, ...rest) {});",
		ast.FunctionExpression,
	)

	params := functionWithParams.GetParameters()
	assert.Equal(t, 4, len(params), "Expected 4 parameters, got %d", len(params))

	// Test basic parameters
	for i, param := range params[:2] {
		bindingElement := expectNodeType[*ast.BindingElementNode](t, param, ast.BindingElement)
		bindingIdentifier := expectNodeType[*ast.BindingIdentifierNode](t, bindingElement.GetTarget(), ast.BindingIdentifier)

		expectedName := string(rune('a' + i))
		assert.Equal(t, expectedName, bindingIdentifier.Identifier,
			"Expected parameter name '%s', got '%s'", expectedName, bindingIdentifier.Identifier)
	}

	// Test parameter with default value
	paramWithDefault := expectNodeType[*ast.BindingElementNode](t, params[2], ast.BindingElement)
	bindingIdentifier := expectNodeType[*ast.BindingIdentifierNode](t, paramWithDefault.GetTarget(), ast.BindingIdentifier)
	assert.Equal(t, "c", bindingIdentifier.Identifier, "Expected parameter name 'c', got '%s'", bindingIdentifier.Identifier)

	// TODO: Modify the parser to remove the BasicNode wrapper, and just return the NumericLiteralNode directly.
	initializer := expectNodeType[*ast.BasicNode](t, paramWithDefault.GetInitializer(), ast.Initializer)
	assert.Equal(t, 1, len(initializer.GetChildren()), "Expected 1 child, got %d", len(initializer.GetChildren()))

	numericLiteral := expectNodeType[*ast.NumericLiteralNode](t, initializer.GetChildren()[0], ast.NumericLiteral)
	assert.Equal(t, float64(1), numericLiteral.Value, "Expected default value 1, got %f", numericLiteral.Value)

	// Test rest parameter
	restParam := expectNodeType[*ast.BindingRestNode](t, params[3], ast.BindingRestProperty)
	restIdentifier := expectNodeType[*ast.BindingIdentifierNode](t, restParam.GetIdentifier(), ast.BindingIdentifier)
	assert.Equal(t, "rest", restIdentifier.Identifier, "Expected rest parameter name 'rest', got '%s'", restIdentifier.Identifier)
}

// PrimaryExpression : AsyncGeneratorExpression
func TestAsyncGeneratorExpression(t *testing.T) {
	// Test anonymous async generator expression
	asyncGeneratorExpression := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(async function* () {});",
		ast.FunctionExpression,
	)
	assert.True(t, asyncGeneratorExpression.Generator, "Expected generator function")
	assert.True(t, asyncGeneratorExpression.Async, "Expected async function")
	assert.Nil(t, asyncGeneratorExpression.GetName(), "Expected anonymous async generator, but got a named async generator")
	expectNodeType[*ast.StatementListNode](
		t,
		asyncGeneratorExpression.GetBody(),
		ast.StatementList,
	)

	// Test named async generator expression
	asyncGeneratorExpression = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(async function* gen() {});",
		ast.FunctionExpression,
	)
	assert.True(t, asyncGeneratorExpression.Generator, "Expected generator function")
	assert.True(t, asyncGeneratorExpression.Async, "Expected async function")
	generatorName := expectNodeType[*ast.BindingIdentifierNode](
		t,
		asyncGeneratorExpression.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "gen", generatorName.Identifier, "Expected generator name 'gen', got %s", generatorName.Identifier)
	expectNodeType[*ast.StatementListNode](
		t,
		asyncGeneratorExpression.GetBody(),
		ast.StatementList,
	)
}

// PrimaryExpression : RegularExpressionLiteral
func TestRegularExpressionLiteral(t *testing.T) {
	regularExpressionLiteral := expectScriptValue[*ast.RegularExpressionLiteralNode](
		t,
		"/foo/;",
		ast.RegularExpressionLiteral,
	)
	assert.Equal(
		t,
		"/foo/",
		regularExpressionLiteral.PatternAndFlags,
		"Expected regular expression 'foo', got %s",
		regularExpressionLiteral.PatternAndFlags,
	)

	// Test regular expression with flags
	regularExpressionLiteral = expectScriptValue[*ast.RegularExpressionLiteralNode](
		t,
		"/foo/gi;",
		ast.RegularExpressionLiteral,
	)
	assert.Equal(
		t,
		"/foo/gi",
		regularExpressionLiteral.PatternAndFlags,
		"Expected regular expression 'foo', got %s",
		regularExpressionLiteral.PatternAndFlags,
	)

	// Test complex regular expression (email address)
	regularExpressionLiteral = expectScriptValue[*ast.RegularExpressionLiteralNode](
		t,
		"/\\w+@\\w+\\.\\w+/;",
		ast.RegularExpressionLiteral,
	)
	assert.Equal(
		t,
		"/\\w+@\\w+\\.\\w+/",
		regularExpressionLiteral.PatternAndFlags,
		"Expected regular expression '/\\w+@\\w+\\.\\w+/', got %s",
		regularExpressionLiteral.PatternAndFlags,
	)
}

// PrimaryExpression : TemplateLiteral
func TestTemplateLiteral(t *testing.T) {
	// Test template literal with no substitutions
	templateLiteral := expectScriptValue[*ast.TemplateLiteralNode](
		t,
		"`simple template`;",
		ast.TemplateLiteral,
	)
	assert.Equal(t, 1, len(templateLiteral.Children), "Expected 1 child, got %d", len(templateLiteral.Children))
	stringLiteral := expectNodeType[*ast.StringLiteralNode](t, templateLiteral.Children[0], ast.StringLiteral)
	assert.Equal(t, "simple template", stringLiteral.Value, "Expected value 'simple template', got %s", stringLiteral.Value)

	// Test template literal with substitutions
	templateLiteral = expectScriptValue[*ast.TemplateLiteralNode](
		t,
		"`Hello ${name}, you are ${age} years old`;",
		ast.TemplateLiteral,
	)
	assert.Equal(t, 5, len(templateLiteral.Children), "Expected 5 children, got %d", len(templateLiteral.Children))

	// Check first string part
	firstPart := expectNodeType[*ast.StringLiteralNode](t, templateLiteral.Children[0], ast.StringLiteral)
	assert.Equal(t, "Hello ", firstPart.Value, "Expected value 'Hello ', got %s", firstPart.Value)

	// Check first substitution
	firstSub := expectNodeType[*ast.IdentifierReferenceNode](t, templateLiteral.Children[1], ast.IdentifierReference)
	assert.Equal(t, "name", firstSub.Identifier, "Expected identifier 'name', got %s", firstSub.Identifier)

	// Check middle string part
	middlePart := expectNodeType[*ast.StringLiteralNode](t, templateLiteral.Children[2], ast.StringLiteral)
	assert.Equal(t, ", you are ", middlePart.Value, "Expected value ', you are ', got %s", middlePart.Value)

	// Check second substitution
	secondSub := expectNodeType[*ast.IdentifierReferenceNode](t, templateLiteral.Children[3], ast.IdentifierReference)
	assert.Equal(t, "age", secondSub.Identifier, "Expected identifier 'age', got %s", secondSub.Identifier)

	// Check third string part
	thirdPart := expectNodeType[*ast.StringLiteralNode](t, templateLiteral.Children[4], ast.StringLiteral)
	assert.Equal(t, " years old", thirdPart.Value, "Expected value ' years old', got %s", thirdPart.Value)
}

// PrimaryExpression : ParenthesizedExpression
func TestParenthesizedExpression(t *testing.T) {
	// Test single expression
	numericLiteral := expectScriptValue[*ast.NumericLiteralNode](
		t,
		"(42);",
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(42), numericLiteral.Value, "Expected value 42, got %f", numericLiteral.Value)

	// Test nested expressions
	multiplicationExpression := expectScriptValue[*ast.MultiplicativeExpressionNode](
		t,
		"((1 + 2) * 3);",
		ast.MultiplicativeExpression,
	)

	// TODO: The (1 + 2) is currently being returned as a cover node, which is not what we want.
	// TODO: This needs to be fixed.
	coverParenNode := expectNodeType[*ast.BasicNode](t, multiplicationExpression.GetLeft(), ast.CoverParenthesizedExpressionAndArrowParameterList)
	additionExpression := expectNodeType[*ast.AdditiveExpressionNode](t, coverParenNode.GetChildren()[0], ast.AdditiveExpression)

	// Verify left side of addition
	leftNum := expectNodeType[*ast.NumericLiteralNode](t, additionExpression.GetLeft(), ast.NumericLiteral)
	assert.Equal(t, float64(1), leftNum.Value, "Expected value 1, got %f", leftNum.Value)

	// Verify right side of addition
	rightNum := expectNodeType[*ast.NumericLiteralNode](t, additionExpression.GetRight(), ast.NumericLiteral)
	assert.Equal(t, float64(2), rightNum.Value, "Expected value 2, got %f", rightNum.Value)

	// Verify right side of multiplication
	multiplyRight := expectNodeType[*ast.NumericLiteralNode](t, multiplicationExpression.GetRight(), ast.NumericLiteral)
	assert.Equal(t, float64(3), multiplyRight.Value, "Expected value 3, got %f", multiplyRight.Value)
}

// AssignmentExpression : ConditionalExpression
func TestConditionalExpression(t *testing.T) {
	// Test basic conditional expression
	conditionalExpression := expectScriptValue[*ast.ConditionalExpressionNode](
		t,
		"x ? 1 : 2;",
		ast.ConditionalExpression,
	)

	// Check condition
	condition := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		conditionalExpression.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", condition.Identifier, "Expected condition identifier 'x', got %s", condition.Identifier)

	// Check true expression
	trueExpr := expectNodeType[*ast.NumericLiteralNode](
		t,
		conditionalExpression.GetTrueExpr(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), trueExpr.Value, "Expected true expression value 1, got %f", trueExpr.Value)

	// Check false expression
	falseExpr := expectNodeType[*ast.NumericLiteralNode](
		t,
		conditionalExpression.GetFalseExpr(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(2), falseExpr.Value, "Expected false expression value 2, got %f", falseExpr.Value)
}

// AssignmentExpression : YieldExpression
func TestYieldExpression(t *testing.T) {
	functionExpression := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"(function* () { yield 1; });",
		ast.FunctionExpression,
	)
	assert.True(t, functionExpression.Generator, "Expected generator function")
	assert.Equal(t, 1, len(functionExpression.GetBody().GetChildren()), "Expected 1 child, got %d", len(functionExpression.GetBody().GetChildren()))

	// Check statement list
	statementList := expectNodeType[*ast.StatementListNode](t, functionExpression.GetBody(), ast.StatementList)
	assert.Equal(t, 1, len(statementList.GetChildren()), "Expected 1 child, got %d", len(statementList.GetChildren()))

	// Check yield expression
	yieldExpression := expectNodeType[*ast.YieldExpressionNode](t, statementList.GetChildren()[0], ast.YieldExpression)
	assert.Equal(t, 1, len(yieldExpression.GetChildren()), "Expected 1 child, got %d", len(yieldExpression.GetChildren()))

	// Check yield value
	numericLiteral := expectNodeType[*ast.NumericLiteralNode](t, yieldExpression.GetChildren()[0], ast.NumericLiteral)
	assert.Equal(t, float64(1), numericLiteral.Value, "Expected value 1, got %f", numericLiteral.Value)
}

// AssignmentExpression : ArrowFunction
func TestArrowFunction(t *testing.T) {
	arrowFunction := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"() => 1;",
		ast.FunctionExpression,
	)

	// Check arrow function
	assert.True(t, arrowFunction.Arrow, "Expected arrow function")
	assert.Equal(t, 1, len(arrowFunction.GetChildren()), "Expected 1 child, got %d", len(arrowFunction.GetChildren()))

	// Check return value
	numericLiteral := expectNodeType[*ast.NumericLiteralNode](t, arrowFunction.GetChildren()[0], ast.NumericLiteral)
	assert.Equal(t, float64(1), numericLiteral.Value, "Expected value 1, got %f", numericLiteral.Value)
}

// AssignmentExpression : AsyncArrowFunction
func TestAsyncArrowFunction(t *testing.T) {
	asyncArrowFunction := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"async () => 1;",
		ast.FunctionExpression,
	)

	// Check async arrow function
	assert.True(t, asyncArrowFunction.Async, "Expected async arrow function")
	assert.True(t, asyncArrowFunction.Arrow, "Expected arrow function")
	assert.Equal(t, 1, len(asyncArrowFunction.GetChildren()), "Expected 1 child, got %d", len(asyncArrowFunction.GetChildren()))

	// Check return value
	numericLiteral := expectNodeType[*ast.NumericLiteralNode](t, asyncArrowFunction.GetChildren()[0], ast.NumericLiteral)
	assert.Equal(t, float64(1), numericLiteral.Value, "Expected value 1, got %f", numericLiteral.Value)
}

// AssignmentExpression : LeftHandSideExpression [Operators] AssignmentExpression
func TestAssignmentExpression(t *testing.T) {
	// Test basic assignment
	assignmentExpression := expectScriptValue[*ast.AssignmentExpressionNode](
		t,
		"a = b;",
		ast.AssignmentExpression,
	)

	// Check target
	target := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		assignmentExpression.GetTarget(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", target.Identifier, "Expected target identifier 'a', got %s", target.Identifier)

	// Check value
	value := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		assignmentExpression.GetValue(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", value.Identifier, "Expected value identifier 'b', got %s", value.Identifier)

	// Test compound assignment operators
	operators := []string{
		"*=", "/=", "%=", "+=", "-=", "<<=", ">>=", ">>>=", "&=", "^=", "|=", "**=",
	}

	for _, op := range operators {
		assignmentExpression = expectScriptValue[*ast.AssignmentExpressionNode](
			t,
			"a "+op+" b;",
			ast.AssignmentExpression,
		)
		assert.Equal(t, op, assignmentExpression.Operator.Value, "Expected operator '%s', got %s", op, assignmentExpression.Operator.Value)

		// Check target
		target = expectNodeType[*ast.IdentifierReferenceNode](
			t,
			assignmentExpression.GetTarget(),
			ast.IdentifierReference,
		)
		assert.Equal(t, "a", target.Identifier, "Expected target identifier 'a', got %s", target.Identifier)

		// Check value
		value = expectNodeType[*ast.IdentifierReferenceNode](
			t,
			assignmentExpression.GetValue(),
			ast.IdentifierReference,
		)
		assert.Equal(t, "b", value.Identifier, "Expected value identifier 'b', got %s", value.Identifier)
	}

	// Test logical assignment operators
	logicalOperators := []string{"&&=", "||=", "??="}

	for _, op := range logicalOperators {
		assignmentExpression = expectScriptValue[*ast.AssignmentExpressionNode](
			t,
			"a "+op+" b;",
			ast.AssignmentExpression,
		)
		assert.Equal(t, op, assignmentExpression.Operator.Value, "Expected operator '%s', got %s", op, assignmentExpression.Operator.Value)

		// Check target
		target = expectNodeType[*ast.IdentifierReferenceNode](
			t,
			assignmentExpression.GetTarget(),
			ast.IdentifierReference,
		)
		assert.Equal(t, "a", target.Identifier, "Expected target identifier 'a', got %s", target.Identifier)

		// Check value
		value = expectNodeType[*ast.IdentifierReferenceNode](
			t,
			assignmentExpression.GetValue(),
			ast.IdentifierReference,
		)
		assert.Equal(t, "b", value.Identifier, "Expected value identifier 'b', got %s", value.Identifier)
	}
}

// ShortCircuitExpression : CoalesceExpression
func TestCoalesceExpression(t *testing.T) {
	// Test basic coalesce expression
	coalesceExpression := expectScriptValue[*ast.CoalesceExpressionNode](
		t,
		"a ?? b;",
		ast.CoalesceExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		coalesceExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		coalesceExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test left association
	coalesceExpression = expectScriptValue[*ast.CoalesceExpressionNode](
		t,
		"a ?? b ?? c;",
		ast.CoalesceExpression,
	)

	// Check outer left operand (which should be another coalesce expression)
	leftCoalesce := expectNodeType[*ast.CoalesceExpressionNode](
		t,
		coalesceExpression.GetLeft(),
		ast.CoalesceExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftCoalesce.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftCoalesce.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		coalesceExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// ShortCircuitExpression : LogicalORExpression
func TestLogicalORExpression(t *testing.T) {
	// Test basic logical OR expression
	logicalORExpression := expectScriptValue[*ast.LogicalORExpressionNode](
		t,
		"a || b;",
		ast.LogicalORExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		logicalORExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		logicalORExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test left association
	logicalORExpression = expectScriptValue[*ast.LogicalORExpressionNode](
		t,
		"a || b || c;",
		ast.LogicalORExpression,
	)

	// Check outer left operand (which should be another logical OR expression)
	leftLogicalOR := expectNodeType[*ast.LogicalORExpressionNode](
		t,
		logicalORExpression.GetLeft(),
		ast.LogicalORExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftLogicalOR.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftLogicalOR.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		logicalORExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// LogicalORExpression : LogicalANDExpression
func TestLogicalANDExpression(t *testing.T) {
	// Test basic logical AND expression
	logicalANDExpression := expectScriptValue[*ast.LogicalANDExpressionNode](
		t,
		"a && b;",
		ast.LogicalANDExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		logicalANDExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		logicalANDExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test left association
	logicalANDExpression = expectScriptValue[*ast.LogicalANDExpressionNode](
		t,
		"a && b && c;",
		ast.LogicalANDExpression,
	)

	// Check outer left operand (which should be another logical AND expression)
	leftLogicalAND := expectNodeType[*ast.LogicalANDExpressionNode](
		t,
		logicalANDExpression.GetLeft(),
		ast.LogicalANDExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftLogicalAND.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftLogicalAND.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		logicalANDExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// LogicalANDExpression : BitwiseORExpression
func TestBitwiseORExpression(t *testing.T) {
	// Test basic bitwise OR expression
	bitwiseORExpression := expectScriptValue[*ast.BitwiseORExpressionNode](
		t,
		"a | b;",
		ast.BitwiseORExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseORExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseORExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test left association
	bitwiseORExpression = expectScriptValue[*ast.BitwiseORExpressionNode](
		t,
		"a | b | c;",
		ast.BitwiseORExpression,
	)

	// Check outer left operand (which should be another bitwise OR expression)
	leftBitwiseOR := expectNodeType[*ast.BitwiseORExpressionNode](
		t,
		bitwiseORExpression.GetLeft(),
		ast.BitwiseORExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftBitwiseOR.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftBitwiseOR.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseORExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// BitwiseORExpression : BitwiseXORExpression
func TestBitwiseXORExpression(t *testing.T) {
	// Test basic bitwise XOR expression
	bitwiseXORExpression := expectScriptValue[*ast.BitwiseXORExpressionNode](
		t,
		"a ^ b;",
		ast.BitwiseXORExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseXORExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseXORExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test left association
	bitwiseXORExpression = expectScriptValue[*ast.BitwiseXORExpressionNode](
		t,
		"a ^ b ^ c;",
		ast.BitwiseXORExpression,
	)

	// Check outer left operand (which should be another bitwise XOR expression)
	leftBitwiseXOR := expectNodeType[*ast.BitwiseXORExpressionNode](
		t,
		bitwiseXORExpression.GetLeft(),
		ast.BitwiseXORExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftBitwiseXOR.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftBitwiseXOR.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseXORExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// BitwiseXORExpression : BitwiseANDExpression
func TestBitwiseANDExpression(t *testing.T) {
	// Test basic bitwise AND expression
	bitwiseANDExpression := expectScriptValue[*ast.BitwiseANDExpressionNode](
		t,
		"a & b;",
		ast.BitwiseANDExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseANDExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseANDExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test left association
	bitwiseANDExpression = expectScriptValue[*ast.BitwiseANDExpressionNode](
		t,
		"a & b & c;",
		ast.BitwiseANDExpression,
	)

	// Check outer left operand (which should be another bitwise AND expression)
	leftBitwiseAND := expectNodeType[*ast.BitwiseANDExpressionNode](
		t,
		bitwiseANDExpression.GetLeft(),
		ast.BitwiseANDExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftBitwiseAND.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftBitwiseAND.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		bitwiseANDExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// BitwiseANDExpression : EqualityExpression
func TestEqualityExpression(t *testing.T) {
	// Test equality (==) operator
	equalityExpression := expectScriptValue[*ast.EqualityExpressionNode](
		t,
		"a == b;",
		ast.EqualityExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		equalityExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		equalityExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test inequality (!=) operator
	equalityExpression = expectScriptValue[*ast.EqualityExpressionNode](
		t,
		"a != b;",
		ast.EqualityExpression,
	)
	assert.Equal(t, "!=", equalityExpression.Operator.Value, "Expected operator '!=', got %s", equalityExpression.Operator.Value)

	// Test strict equality (===) operator
	equalityExpression = expectScriptValue[*ast.EqualityExpressionNode](
		t,
		"a === b;",
		ast.EqualityExpression,
	)
	assert.Equal(t, "===", equalityExpression.Operator.Value, "Expected operator '===', got %s", equalityExpression.Operator.Value)

	// Test strict inequality (!==) operator
	equalityExpression = expectScriptValue[*ast.EqualityExpressionNode](
		t,
		"a !== b;",
		ast.EqualityExpression,
	)
	assert.Equal(t, "!==", equalityExpression.Operator.Value, "Expected operator '!==', got %s", equalityExpression.Operator.Value)

	// Test left association
	equalityExpression = expectScriptValue[*ast.EqualityExpressionNode](
		t,
		"a == b == c;",
		ast.EqualityExpression,
	)

	// Check outer left operand (which should be another equality expression)
	leftEquality := expectNodeType[*ast.EqualityExpressionNode](
		t,
		equalityExpression.GetLeft(),
		ast.EqualityExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftEquality.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftEquality.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		equalityExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// EqualityExpression : RelationalExpression
func TestRelationalExpression(t *testing.T) {
	// Test less than operator
	relationalExpression := expectScriptValue[*ast.RelationalExpressionNode](
		t,
		"a < b;",
		ast.RelationalExpression,
	)
	assert.Equal(t, "<", relationalExpression.Operator.Value, "Expected operator '<', got %s", relationalExpression.Operator.Value)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		relationalExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		relationalExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test greater than operator
	relationalExpression = expectScriptValue[*ast.RelationalExpressionNode](
		t,
		"a > b;",
		ast.RelationalExpression,
	)
	assert.Equal(t, ">", relationalExpression.Operator.Value, "Expected operator '>', got %s", relationalExpression.Operator.Value)

	// Test less than or equal operator
	relationalExpression = expectScriptValue[*ast.RelationalExpressionNode](
		t,
		"a <= b;",
		ast.RelationalExpression,
	)
	assert.Equal(t, "<=", relationalExpression.Operator.Value, "Expected operator '<=', got %s", relationalExpression.Operator.Value)

	// Test greater than or equal operator
	relationalExpression = expectScriptValue[*ast.RelationalExpressionNode](
		t,
		"a >= b;",
		ast.RelationalExpression,
	)
	assert.Equal(t, ">=", relationalExpression.Operator.Value, "Expected operator '>=', got %s", relationalExpression.Operator.Value)

	// Test instanceof operator
	relationalExpression = expectScriptValue[*ast.RelationalExpressionNode](
		t,
		"a instanceof b;",
		ast.RelationalExpression,
	)
	assert.Equal(t, "instanceof", relationalExpression.Operator.Value, "Expected operator 'instanceof', got %s", relationalExpression.Operator.Value)

	// Test in operator
	relationalExpression = expectScriptValue[*ast.RelationalExpressionNode](
		t,
		"a in b;",
		ast.RelationalExpression,
	)
	assert.Equal(t, "in", relationalExpression.Operator.Value, "Expected operator 'in', got %s", relationalExpression.Operator.Value)

	// Test left association
	relationalExpression = expectScriptValue[*ast.RelationalExpressionNode](
		t,
		"a < b < c;",
		ast.RelationalExpression,
	)

	// Check outer left operand (which should be another relational expression)
	leftRelational := expectNodeType[*ast.RelationalExpressionNode](
		t,
		relationalExpression.GetLeft(),
		ast.RelationalExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftRelational.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftRelational.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		relationalExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// RelationalExpression : ShiftExpression
func TestShiftExpression(t *testing.T) {
	// Test left shift operator
	shiftExpression := expectScriptValue[*ast.ShiftExpressionNode](
		t,
		"a << b;",
		ast.ShiftExpression,
	)
	assert.Equal(t, "<<", shiftExpression.Operator.Value, "Expected operator '<<', got %s", shiftExpression.Operator.Value)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		shiftExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		shiftExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test right shift operator
	shiftExpression = expectScriptValue[*ast.ShiftExpressionNode](
		t,
		"a >> b;",
		ast.ShiftExpression,
	)
	assert.Equal(t, ">>", shiftExpression.Operator.Value, "Expected operator '>>', got %s", shiftExpression.Operator.Value)

	// Test unsigned right shift operator
	shiftExpression = expectScriptValue[*ast.ShiftExpressionNode](
		t,
		"a >>> b;",
		ast.ShiftExpression,
	)
	assert.Equal(t, ">>>", shiftExpression.Operator.Value, "Expected operator '>>>', got %s", shiftExpression.Operator.Value)

	// Test left association
	shiftExpression = expectScriptValue[*ast.ShiftExpressionNode](
		t,
		"a << b << c;",
		ast.ShiftExpression,
	)

	// Check outer left operand (which should be another shift expression)
	leftShift := expectNodeType[*ast.ShiftExpressionNode](
		t,
		shiftExpression.GetLeft(),
		ast.ShiftExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftShift.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftShift.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		shiftExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// ShiftExpression : AdditiveExpression
func TestAdditiveExpression(t *testing.T) {
	// Test addition operator
	additiveExpression := expectScriptValue[*ast.AdditiveExpressionNode](
		t,
		"a + b;",
		ast.AdditiveExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		additiveExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		additiveExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test subtraction operator
	additiveExpression = expectScriptValue[*ast.AdditiveExpressionNode](
		t,
		"a - b;",
		ast.AdditiveExpression,
	)
	assert.Equal(t, "-", additiveExpression.Operator.Value, "Expected operator '-', got %s", additiveExpression.Operator.Value)

	// Test left association
	additiveExpression = expectScriptValue[*ast.AdditiveExpressionNode](
		t,
		"a + b + c;",
		ast.AdditiveExpression,
	)

	// Check outer left operand (which should be another additive expression)
	leftAdditive := expectNodeType[*ast.AdditiveExpressionNode](
		t,
		additiveExpression.GetLeft(),
		ast.AdditiveExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftAdditive.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftAdditive.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		additiveExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// AdditiveExpression : MultiplicativeExpression
func TestMultiplicativeExpression(t *testing.T) {
	// Test multiplication operator
	multiplicativeExpression := expectScriptValue[*ast.MultiplicativeExpressionNode](
		t,
		"a * b;",
		ast.MultiplicativeExpression,
	)
	assert.Equal(t, "*", multiplicativeExpression.Operator.Value, "Expected operator '*', got %s", multiplicativeExpression.Operator.Value)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		multiplicativeExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		multiplicativeExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test division operator
	multiplicativeExpression = expectScriptValue[*ast.MultiplicativeExpressionNode](
		t,
		"(a / b);",
		ast.MultiplicativeExpression,
	)
	assert.Equal(t, "/", multiplicativeExpression.Operator.Value, "Expected operator '/', got %s", multiplicativeExpression.Operator.Value)

	// Test modulo operator
	multiplicativeExpression = expectScriptValue[*ast.MultiplicativeExpressionNode](
		t,
		"a % b;",
		ast.MultiplicativeExpression,
	)
	assert.Equal(t, "%", multiplicativeExpression.Operator.Value, "Expected operator '%', got %s", multiplicativeExpression.Operator.Value)

	// Test left association
	multiplicativeExpression = expectScriptValue[*ast.MultiplicativeExpressionNode](
		t,
		"a * b * c;",
		ast.MultiplicativeExpression,
	)

	// Check outer left operand (which should be another multiplicative expression)
	leftMultiplicative := expectNodeType[*ast.MultiplicativeExpressionNode](
		t,
		multiplicativeExpression.GetLeft(),
		ast.MultiplicativeExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftMultiplicative.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected inner left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		leftMultiplicative.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected inner right operand identifier 'b', got %s", rightOperand.Identifier)

	// Check outer right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		multiplicativeExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected outer right operand identifier 'c', got %s", rightOperand.Identifier)
}

// MultiplicativeExpression : ExponentiationExpression
func TestExponentiationExpression(t *testing.T) {
	// Test basic exponentiation expression
	exponentiationExpression := expectScriptValue[*ast.ExponentiationExpressionNode](
		t,
		"a ** b;",
		ast.ExponentiationExpression,
	)

	// Check left operand
	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		exponentiationExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand
	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		exponentiationExpression.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)

	// Test right association
	exponentiationExpression = expectScriptValue[*ast.ExponentiationExpressionNode](
		t,
		"a ** b ** c;",
		ast.ExponentiationExpression,
	)

	// Check left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		exponentiationExpression.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	// Check right operand (which should be another exponentiation expression)
	rightExponentiation := expectNodeType[*ast.ExponentiationExpressionNode](
		t,
		exponentiationExpression.GetRight(),
		ast.ExponentiationExpression,
	)

	// Check inner left operand
	leftOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		rightExponentiation.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", leftOperand.Identifier, "Expected inner left operand identifier 'b', got %s", leftOperand.Identifier)

	// Check inner right operand
	rightOperand = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		rightExponentiation.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "c", rightOperand.Identifier, "Expected inner right operand identifier 'c', got %s", rightOperand.Identifier)
}

// ExponentiationExpression : UnaryExpression
func TestUnaryExpression(t *testing.T) {
	// Test delete operator
	unaryExpression := expectScriptValue[*ast.UnaryExpressionNode](
		t,
		"delete foo;",
		ast.UnaryExpression,
	)
	assert.Equal(t, "delete", unaryExpression.Operator.Value, "Expected operator 'delete', got %s", unaryExpression.Operator.Value)
	expectNodeType[*ast.IdentifierReferenceNode](
		t,
		unaryExpression.GetValue(),
		ast.IdentifierReference,
	)

	// Test void operator
	unaryExpression = expectScriptValue[*ast.UnaryExpressionNode](
		t,
		"void 0;",
		ast.UnaryExpression,
	)
	assert.Equal(t, "void", unaryExpression.Operator.Value, "Expected operator 'void', got %s", unaryExpression.Operator.Value)
	expectNodeType[*ast.NumericLiteralNode](
		t,
		unaryExpression.GetValue(),
		ast.NumericLiteral,
	)

	// Test typeof operator
	unaryExpression = expectScriptValue[*ast.UnaryExpressionNode](
		t,
		"typeof x;",
		ast.UnaryExpression,
	)
	assert.Equal(t, "typeof", unaryExpression.Operator.Value, "Expected operator 'typeof', got %s", unaryExpression.Operator.Value)
	expectNodeType[*ast.IdentifierReferenceNode](
		t,
		unaryExpression.GetValue(),
		ast.IdentifierReference,
	)

	// Test unary plus operator
	unaryExpression = expectScriptValue[*ast.UnaryExpressionNode](
		t,
		"+x;",
		ast.UnaryExpression,
	)
	assert.Equal(t, "+", unaryExpression.Operator.Value, "Expected operator '+', got %s", unaryExpression.Operator.Value)
	expectNodeType[*ast.IdentifierReferenceNode](
		t,
		unaryExpression.GetValue(),
		ast.IdentifierReference,
	)

	// Test unary minus operator
	unaryExpression = expectScriptValue[*ast.UnaryExpressionNode](
		t,
		"-x;",
		ast.UnaryExpression,
	)
	assert.Equal(t, "-", unaryExpression.Operator.Value, "Expected operator '-', got %s", unaryExpression.Operator.Value)
	expectNodeType[*ast.IdentifierReferenceNode](
		t,
		unaryExpression.GetValue(),
		ast.IdentifierReference,
	)

	// Test bitwise NOT operator
	unaryExpression = expectScriptValue[*ast.UnaryExpressionNode](
		t,
		"~x;",
		ast.UnaryExpression,
	)
	assert.Equal(t, "~", unaryExpression.Operator.Value, "Expected operator '~', got %s", unaryExpression.Operator.Value)
	expectNodeType[*ast.IdentifierReferenceNode](
		t,
		unaryExpression.GetValue(),
		ast.IdentifierReference,
	)

	// Test logical NOT operator
	unaryExpression = expectScriptValue[*ast.UnaryExpressionNode](
		t,
		"!x;",
		ast.UnaryExpression,
	)
	assert.Equal(t, "!", unaryExpression.Operator.Value, "Expected operator '!', got %s", unaryExpression.Operator.Value)
	expectNodeType[*ast.IdentifierReferenceNode](
		t,
		unaryExpression.GetValue(),
		ast.IdentifierReference,
	)
}

// UnaryExpression : AwaitExpression
func TestAwaitExpression(t *testing.T) {
	// Test await expression
	functionExpression := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		`async function foo() {
			await derp;
		}`,
		ast.FunctionExpression,
	)

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		functionExpression.GetBody(),
		ast.StatementList,
	)
	awaitExpression := expectNodeType[*ast.AwaitExpressionNode](
		t,
		statementList.GetChildren()[0],
		ast.AwaitExpression,
	)

	// Check await expression value
	identifierReference := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		awaitExpression.GetExpression(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "derp", identifierReference.Identifier, "Expected identifier 'derp', got %s", identifierReference.Identifier)
}

// UnaryExpression : UpdateExpression
func TestUpdateExpression(t *testing.T) {
	// Test prefix increment operator
	updateExpression := expectScriptValue[*ast.UpdateExpressionNode](
		t,
		"++x;",
		ast.UpdateExpression,
	)
	assert.Equal(t, "++", updateExpression.Operator.Value, "Expected operator '++', got %s", updateExpression.Operator.Value)
	identifierReference := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		updateExpression.GetValue(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", identifierReference.Identifier, "Expected identifier 'x', got %s", identifierReference.Identifier)

	// Test prefix decrement operator
	updateExpression = expectScriptValue[*ast.UpdateExpressionNode](
		t,
		"--x;",
		ast.UpdateExpression,
	)
	assert.Equal(t, "--", updateExpression.Operator.Value, "Expected operator '--', got %s", updateExpression.Operator.Value)
	identifierReference = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		updateExpression.GetValue(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", identifierReference.Identifier, "Expected identifier 'x', got %s", identifierReference.Identifier)

	// Test postfix increment operator
	updateExpression = expectScriptValue[*ast.UpdateExpressionNode](
		t,
		"x++;",
		ast.UpdateExpression,
	)
	assert.Equal(t, "++", updateExpression.Operator.Value, "Expected operator '++', got %s", updateExpression.Operator.Value)
	identifierReference = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		updateExpression.GetValue(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", identifierReference.Identifier, "Expected identifier 'x', got %s", identifierReference.Identifier)

	// Test postfix decrement operator
	updateExpression = expectScriptValue[*ast.UpdateExpressionNode](
		t,
		"x--;",
		ast.UpdateExpression,
	)
	assert.Equal(t, "--", updateExpression.Operator.Value, "Expected operator '--', got %s", updateExpression.Operator.Value)
	identifierReference = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		updateExpression.GetValue(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", identifierReference.Identifier, "Expected identifier 'x', got %s", identifierReference.Identifier)
}

// LeftHandSideExpression : NewExpression
func TestNewExpression(t *testing.T) {
	// Test new operator with constructor
	newExpression := expectScriptValue[*ast.NewExpressionNode](
		t,
		"new Foo();",
		ast.NewExpression,
	)
	constructor := expectNodeType[*ast.CallExpressionNode](
		t,
		newExpression.GetConstructor(),
		ast.CallExpression,
	)
	callee := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		constructor.GetCallee(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "Foo", callee.Identifier, "Expected constructor 'Foo', got %s", callee.Identifier)

	// Test new operator with constructor and arguments
	newExpression = expectScriptValue[*ast.NewExpressionNode](
		t,
		"new Foo(1, 'bar');",
		ast.NewExpression,
	)
	constructor = expectNodeType[*ast.CallExpressionNode](
		t,
		newExpression.GetConstructor(),
		ast.CallExpression,
	)
	callee = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		constructor.GetCallee(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "Foo", callee.Identifier, "Expected constructor 'Foo', got %s", callee.Identifier)

	// Check arguments
	arguments := constructor.GetArguments()
	assert.Equal(t, 2, len(arguments), "Expected 2 arguments, got %d", len(arguments))

	// Check first argument
	firstArg := expectNodeType[*ast.NumericLiteralNode](
		t,
		arguments[0],
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), firstArg.Value, "Expected first argument value 1, got %f", firstArg.Value)

	// Check second argument
	secondArg := expectNodeType[*ast.StringLiteralNode](
		t,
		arguments[1],
		ast.StringLiteral,
	)
	assert.Equal(t, "bar", secondArg.Value, "Expected second argument value 'bar', got %s", secondArg.Value)
}

// NewExpression : MemberExpression
func TestMemberExpression(t *testing.T) {
	// Test member expression with computed property
	memberExpression := expectScriptValue[*ast.MemberExpressionNode](
		t,
		"foo[bar];",
		ast.MemberExpression,
	)

	// Check object
	object := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		memberExpression.GetObject(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", object.Identifier, "Expected object identifier 'foo', got %s", object.Identifier)

	// Check property
	property := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		memberExpression.GetProperty(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", property.Identifier, "Expected property identifier 'bar', got %s", property.Identifier)

	// Test member expression with dot notation
	memberExpression = expectScriptValue[*ast.MemberExpressionNode](
		t,
		"foo.bar;",
		ast.MemberExpression,
	)

	// Check object
	object = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		memberExpression.GetObject(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", object.Identifier, "Expected object identifier 'foo', got %s", object.Identifier)
	assert.Equal(t, "bar", memberExpression.PropertyIdentifier, "Expected property identifier 'bar', got %s", memberExpression.PropertyIdentifier)

	// Test member expression with template literal
	templateLiteral := expectScriptValue[*ast.TemplateLiteralNode](
		t,
		"foo.bar`template`;",
		ast.TemplateLiteral,
	)
	memberExpression = expectNodeType[*ast.MemberExpressionNode](
		t,
		templateLiteral.GetTagFunctionRef(),
		ast.MemberExpression,
	)

	// Check object
	object = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		memberExpression.GetObject(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", object.Identifier, "Expected object identifier 'foo', got %s", object.Identifier)
	assert.Equal(t, "bar", memberExpression.PropertyIdentifier, "Expected property identifier 'bar', got %s", memberExpression.PropertyIdentifier)

	// Test super property
	memberExpression = expectScriptValue[*ast.MemberExpressionNode](
		t,
		"super.foo;",
		ast.MemberExpression,
	)
	assert.True(t, memberExpression.Super, "Expected super property")
	assert.Equal(t, "foo", memberExpression.PropertyIdentifier, "Expected property identifier 'foo', got %s", memberExpression.PropertyIdentifier)

	// Test member expression with constructor call
	newExpression := expectScriptValue[*ast.NewExpressionNode](
		t,
		"new foo.bar();",
		ast.NewExpression,
	)
	constructor := expectNodeType[*ast.CallExpressionNode](
		t,
		newExpression.GetConstructor(),
		ast.CallExpression,
	)
	callee := expectNodeType[*ast.MemberExpressionNode](
		t,
		constructor.GetCallee(),
		ast.MemberExpression,
	)
	identifierReference := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		callee.GetObject(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", identifierReference.Identifier, "Expected object identifier 'foo', got %s", identifierReference.Identifier)
	assert.Equal(t, "bar", callee.PropertyIdentifier, "Expected property identifier 'bar', got %s", callee.PropertyIdentifier)
}

// LeftHandSideExpression : CallExpression
func TestCallExpression(t *testing.T) {
	// Test basic call expression
	callExpression := expectScriptValue[*ast.CallExpressionNode](
		t,
		"foo(a, b);",
		ast.CallExpression,
	)

	// Check callee
	callee := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		callExpression.GetCallee(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", callee.Identifier, "Expected callee identifier 'foo', got %s", callee.Identifier)

	// Check arguments
	arguments := callExpression.GetArguments()
	assert.Equal(t, 2, len(arguments), "Expected 2 arguments, got %d", len(arguments))

	// Check first argument
	firstArg := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		arguments[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", firstArg.Identifier, "Expected first argument identifier 'a', got %s", firstArg.Identifier)

	// Check second argument
	secondArg := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		arguments[1],
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", secondArg.Identifier, "Expected second argument identifier 'b', got %s", secondArg.Identifier)

	// Test call expression with member expression callee
	callExpression = expectScriptValue[*ast.CallExpressionNode](
		t,
		"foo.bar(a);",
		ast.CallExpression,
	)

	// Check member expression callee
	memberExpression := expectNodeType[*ast.MemberExpressionNode](
		t,
		callExpression.GetCallee(),
		ast.MemberExpression,
	)

	// Check object
	object := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		memberExpression.GetObject(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", object.Identifier, "Expected object identifier 'foo', got %s", object.Identifier)
	assert.Equal(t, "bar", memberExpression.PropertyIdentifier, "Expected property identifier 'bar', got %s", memberExpression.PropertyIdentifier)

	// Check argument
	arguments = callExpression.GetArguments()
	assert.Equal(t, 1, len(arguments), "Expected 1 argument, got %d", len(arguments))
	arg := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		arguments[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", arg.Identifier, "Expected argument identifier 'a', got %s", arg.Identifier)

	// Test super call
	callExpression = expectScriptValue[*ast.CallExpressionNode](
		t,
		"super(a, b);",
		ast.CallExpression,
	)
	assert.True(t, callExpression.Super, "Expected super call")

	// Check arguments
	arguments = callExpression.GetArguments()
	assert.Equal(t, 2, len(arguments), "Expected 2 arguments, got %d", len(arguments))

	// Check first argument
	firstArg = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		arguments[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", firstArg.Identifier, "Expected first argument identifier 'a', got %s", firstArg.Identifier)

	// Check second argument
	secondArg = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		arguments[1],
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", secondArg.Identifier, "Expected second argument identifier 'b', got %s", secondArg.Identifier)

	// Test import call
	importCall := expectScriptValue[*ast.BasicNode](
		t,
		"import(foo, bar);",
		ast.ImportCall,
	)
	assert.Equal(t, 2, len(importCall.GetChildren()), "Expected 2 children, got %d", len(importCall.GetChildren()))

	// Check first argument
	firstArg = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		importCall.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", firstArg.Identifier, "Expected first argument identifier 'foo', got %s", firstArg.Identifier)

	// Check second argument
	secondArg = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		importCall.GetChildren()[1],
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", secondArg.Identifier, "Expected second argument identifier 'bar', got %s", secondArg.Identifier)
}

// LeftHandSideExpression : OptionalExpression
func TestOptionalExpression(t *testing.T) {
	// Test basic optional expression with member expression
	optionalExpression := expectScriptValue[*ast.OptionalExpressionNode](
		t,
		"foo?.bar;",
		ast.OptionalExpression,
	)

	// Check member expression
	memberExpression := expectNodeType[*ast.MemberExpressionNode](
		t,
		optionalExpression.GetExpression(),
		ast.MemberExpression,
	)

	// Check object
	object := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		memberExpression.GetObject(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", object.Identifier, "Expected object identifier 'foo', got %s", object.Identifier)
	assert.Equal(t, "bar", memberExpression.PropertyIdentifier, "Expected property identifier 'bar', got %s", memberExpression.PropertyIdentifier)

	// Test optional expression with call expression
	callExpression := expectScriptValue[*ast.CallExpressionNode](
		t,
		"foo?.bar(a);",
		ast.CallExpression,
	)

	// Check optional expression
	optionalExpression = expectNodeType[*ast.OptionalExpressionNode](
		t,
		callExpression.GetCallee(),
		ast.OptionalExpression,
	)

	memberExpression = expectNodeType[*ast.MemberExpressionNode](
		t,
		optionalExpression.GetExpression(),
		ast.MemberExpression,
	)

	// Check object
	object = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		memberExpression.GetObject(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", object.Identifier, "Expected object identifier 'foo', got %s", object.Identifier)
	assert.Equal(t, "bar", memberExpression.PropertyIdentifier, "Expected property identifier 'bar', got %s", memberExpression.PropertyIdentifier)

	// Test left association
	optionalExpression = expectScriptValue[*ast.OptionalExpressionNode](
		t,
		"foo?.bar?.baz;",
		ast.OptionalExpression,
	)

	// Check outer member expression
	memberExpression = expectNodeType[*ast.MemberExpressionNode](
		t,
		optionalExpression.GetExpression(),
		ast.MemberExpression,
	)

	// Check inner optional expression
	innerOptional := expectNodeType[*ast.OptionalExpressionNode](
		t,
		memberExpression.GetObject(),
		ast.OptionalExpression,
	)

	// Check inner member expression
	innerMember := expectNodeType[*ast.MemberExpressionNode](
		t,
		innerOptional.GetExpression(),
		ast.MemberExpression,
	)

	// Check object
	object = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		innerMember.GetObject(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", object.Identifier, "Expected object identifier 'foo', got %s", object.Identifier)
	assert.Equal(t, "bar", innerMember.PropertyIdentifier, "Expected property identifier 'bar', got %s", innerMember.PropertyIdentifier)
	assert.Equal(t, "baz", memberExpression.PropertyIdentifier, "Expected property identifier 'baz', got %s", memberExpression.PropertyIdentifier)
}

// Statement : BlockStatement
func TestBlockStatement(t *testing.T) {
	blockStatement := expectScriptValue[*ast.BasicNode](
		t,
		"{ foo; bar; }",
		ast.Block,
	)
	assert.Equal(t, 1, len(blockStatement.GetChildren()), "Expected 2 children, got %d", len(blockStatement.GetChildren()))

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		blockStatement.GetChildren()[0],
		ast.StatementList,
	)

	// Check first child
	firstChild := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", firstChild.Identifier, "Expected first child identifier 'foo', got %s", firstChild.Identifier)

	// Check second child
	secondChild := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[1],
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", secondChild.Identifier, "Expected second child identifier 'bar', got %s", secondChild.Identifier)
}

// Statement : VariableStatement
func TestVariableStatement(t *testing.T) {
	// Test basic variable declaration with single identifier
	variableStatement := expectScriptValue[*ast.BasicNode](
		t,
		"var foo;",
		ast.VariableStatement,
	)

	declarationList := expectNodeType[*ast.BasicNode](
		t,
		variableStatement.GetChildren()[0],
		ast.VariableDeclarationList,
	)

	declaration := expectNodeType[*ast.BasicNode](
		t,
		declarationList.GetChildren()[0],
		ast.VariableDeclaration,
	)

	identifier := expectNodeType[*ast.BindingIdentifierNode](
		t,
		declaration.GetChildren()[0],
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", identifier.Identifier, "Expected identifier 'foo', got %s", identifier.Identifier)

	// Test variable declaration with initializer
	variableStatement = expectScriptValue[*ast.BasicNode](
		t,
		"var foo = 42;",
		ast.VariableStatement,
	)

	declarationList = expectNodeType[*ast.BasicNode](
		t,
		variableStatement.GetChildren()[0],
		ast.VariableDeclarationList,
	)

	declaration = expectNodeType[*ast.BasicNode](
		t,
		declarationList.GetChildren()[0],
		ast.VariableDeclaration,
	)

	identifier = expectNodeType[*ast.BindingIdentifierNode](
		t,
		declaration.GetChildren()[0],
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", identifier.Identifier, "Expected identifier 'foo', got %s", identifier.Identifier)

	initializer := expectNodeType[*ast.BasicNode](
		t,
		declaration.GetChildren()[1],
		ast.Initializer,
	)
	numericLiteral := expectNodeType[*ast.NumericLiteralNode](
		t,
		initializer.GetChildren()[0],
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(42), numericLiteral.Value, "Expected numeric value 42, got %f", numericLiteral.Value)

	// Test multiple variable declarations
	variableStatement = expectScriptValue[*ast.BasicNode](
		t,
		"var foo = 42, bar;",
		ast.VariableStatement,
	)

	declarationList = expectNodeType[*ast.BasicNode](
		t,
		variableStatement.GetChildren()[0],
		ast.VariableDeclarationList,
	)
	assert.Equal(t, 2, len(declarationList.GetChildren()), "Expected 2 declarations, got %d", len(declarationList.GetChildren()))

	// Check first declaration with initializer
	firstDeclaration := expectNodeType[*ast.BasicNode](
		t,
		declarationList.GetChildren()[0],
		ast.VariableDeclaration,
	)
	identifier = expectNodeType[*ast.BindingIdentifierNode](
		t,
		firstDeclaration.GetChildren()[0],
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", identifier.Identifier, "Expected identifier 'foo', got %s", identifier.Identifier)

	// Check second declaration without initializer
	secondDeclaration := expectNodeType[*ast.BasicNode](
		t,
		declarationList.GetChildren()[1],
		ast.VariableDeclaration,
	)
	identifier = expectNodeType[*ast.BindingIdentifierNode](
		t,
		secondDeclaration.GetChildren()[0],
		ast.BindingIdentifier,
	)
	assert.Equal(t, "bar", identifier.Identifier, "Expected identifier 'bar', got %s", identifier.Identifier)

	// Test variable declaration with object binding pattern
	variableStatement = expectScriptValue[*ast.BasicNode](
		t,
		"var {a, b} = obj;",
		ast.VariableStatement,
	)

	declarationList = expectNodeType[*ast.BasicNode](
		t,
		variableStatement.GetChildren()[0],
		ast.VariableDeclarationList,
	)

	declaration = expectNodeType[*ast.BasicNode](
		t,
		declarationList.GetChildren()[0],
		ast.VariableDeclaration,
	)

	bindingPattern := expectNodeType[*ast.ObjectBindingPatternNode](
		t,
		declaration.GetChildren()[0],
		ast.ObjectBindingPattern,
	)

	properties := bindingPattern.GetProperties()
	assert.Equal(t, 2, len(properties), "Expected 2 properties, got %d", len(properties))

	// Test variable declaration with array binding pattern
	variableStatement = expectScriptValue[*ast.BasicNode](
		t,
		"var [x, y] = arr;",
		ast.VariableStatement,
	)

	declarationList = expectNodeType[*ast.BasicNode](
		t,
		variableStatement.GetChildren()[0],
		ast.VariableDeclarationList,
	)

	declaration = expectNodeType[*ast.BasicNode](
		t,
		declarationList.GetChildren()[0],
		ast.VariableDeclaration,
	)

	arrayPattern := expectNodeType[*ast.ArrayBindingPatternNode](
		t,
		declaration.GetChildren()[0],
		ast.ArrayBindingPattern,
	)

	elements := arrayPattern.GetElements()
	assert.Equal(t, 2, len(elements), "Expected 2 elements, got %d", len(elements))
}

// Statement : EmptyStatement
func TestEmptyStatement(t *testing.T) {
	expectScriptValue[*ast.BasicNode](
		t,
		";",
		ast.EmptyStatement,
	)
}

// Statement : IfStatement
func TestIfStatement(t *testing.T) {
	// Test basic if statement
	ifStatement := expectScriptValue[*ast.IfStatementNode](
		t,
		"if (x) y;",
		ast.IfStatement,
	)

	condition := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		ifStatement.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", condition.Identifier, "Expected condition identifier 'x', got %s", condition.Identifier)

	trueStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		ifStatement.GetTrueStatement(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "y", trueStatement.Identifier, "Expected true statement identifier 'y', got %s", trueStatement.Identifier)

	assert.Nil(t, ifStatement.GetElseStatement(), "Expected no else statement")

	// Test if-else statement
	ifStatement = expectScriptValue[*ast.IfStatementNode](
		t,
		"if (x) y; else z;",
		ast.IfStatement,
	)

	condition = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		ifStatement.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", condition.Identifier, "Expected condition identifier 'x', got %s", condition.Identifier)

	trueStatement = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		ifStatement.GetTrueStatement(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "y", trueStatement.Identifier, "Expected true statement identifier 'y', got %s", trueStatement.Identifier)

	elseStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		ifStatement.GetElseStatement(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "z", elseStatement.Identifier, "Expected else statement identifier 'z', got %s", elseStatement.Identifier)

	// Test nested if-else statement
	ifStatement = expectScriptValue[*ast.IfStatementNode](
		t,
		"if (x) if (y) a; else b;",
		ast.IfStatement,
	)

	condition = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		ifStatement.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", condition.Identifier, "Expected condition identifier 'x', got %s", condition.Identifier)

	nestedIf := expectNodeType[*ast.IfStatementNode](
		t,
		ifStatement.GetTrueStatement(),
		ast.IfStatement,
	)

	nestedCondition := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		nestedIf.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "y", nestedCondition.Identifier, "Expected nested condition identifier 'y', got %s", nestedCondition.Identifier)

	nestedTrue := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		nestedIf.GetTrueStatement(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", nestedTrue.Identifier, "Expected nested true statement identifier 'a', got %s", nestedTrue.Identifier)

	nestedElse := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		nestedIf.GetElseStatement(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", nestedElse.Identifier, "Expected nested else statement identifier 'b', got %s", nestedElse.Identifier)
}

// IterationStatement : DoWhileStatement
func TestDoWhileStatement(t *testing.T) {
	// Test basic do-while statement
	doWhileStatement := expectScriptValue[*ast.DoWhileStatementNode](
		t,
		"do foo; while (bar);",
		ast.DoWhileStatement,
	)

	statement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		doWhileStatement.GetStatement(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", statement.Identifier, "Expected statement identifier 'foo', got %s", statement.Identifier)

	condition := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		doWhileStatement.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", condition.Identifier, "Expected condition identifier 'bar', got %s", condition.Identifier)

	// Test do-while with block statement
	doWhileStatement = expectScriptValue[*ast.DoWhileStatementNode](
		t,
		"do { foo; bar; } while (baz);",
		ast.DoWhileStatement,
	)

	blockStatement := expectNodeType[*ast.BasicNode](
		t,
		doWhileStatement.GetStatement(),
		ast.Block,
	)

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		blockStatement.GetChildren()[0],
		ast.StatementList,
	)

	firstStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", firstStatement.Identifier, "Expected first statement identifier 'foo', got %s", firstStatement.Identifier)

	secondStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[1],
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", secondStatement.Identifier, "Expected second statement identifier 'bar', got %s", secondStatement.Identifier)

	condition = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		doWhileStatement.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "baz", condition.Identifier, "Expected condition identifier 'baz', got %s", condition.Identifier)
}

// IterationStatement : WhileStatement
func TestWhileStatement(t *testing.T) {
	// Test basic while statement
	whileStatement := expectScriptValue[*ast.WhileStatementNode](
		t,
		"while (foo) bar;",
		ast.WhileStatement,
	)

	condition := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		whileStatement.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", condition.Identifier, "Expected condition identifier 'foo', got %s", condition.Identifier)

	statement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		whileStatement.GetStatement(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", statement.Identifier, "Expected statement identifier 'bar', got %s", statement.Identifier)

	// Test while with block statement
	whileStatement = expectScriptValue[*ast.WhileStatementNode](
		t,
		"while (foo) { bar; baz; }",
		ast.WhileStatement,
	)

	condition = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		whileStatement.GetCondition(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", condition.Identifier, "Expected condition identifier 'foo', got %s", condition.Identifier)

	blockStatement := expectNodeType[*ast.BasicNode](
		t,
		whileStatement.GetStatement(),
		ast.Block,
	)

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		blockStatement.GetChildren()[0],
		ast.StatementList,
	)

	firstStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", firstStatement.Identifier, "Expected first statement identifier 'bar', got %s", firstStatement.Identifier)

	secondStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[1],
		ast.IdentifierReference,
	)
	assert.Equal(t, "baz", secondStatement.Identifier, "Expected second statement identifier 'baz', got %s", secondStatement.Identifier)
}

// IterationStatement : ForStatement
func TestForStatement(t *testing.T) {
	// Test basic for statement
	forStatement := expectScriptValue[*ast.ForStatementNode](
		t,
		"for (i = 0; i < 10; i++) { foo; }",
		ast.ForStatement,
	)

	// Check initializer
	initializer := expectNodeType[*ast.AssignmentExpressionNode](
		t,
		forStatement.GetInitializer(),
		ast.AssignmentExpression,
	)

	initTarget := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		initializer.GetTarget(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "i", initTarget.Identifier, "Expected initializer target identifier 'i', got %s", initTarget.Identifier)

	initValue := expectNodeType[*ast.NumericLiteralNode](
		t,
		initializer.GetValue(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(0), initValue.Value, "Expected initializer value 0, got %f", initValue.Value)

	// Check condition
	condition := expectNodeType[*ast.RelationalExpressionNode](
		t,
		forStatement.GetCondition(),
		ast.RelationalExpression,
	)

	condLeft := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		condition.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "i", condLeft.Identifier, "Expected condition left identifier 'i', got %s", condLeft.Identifier)

	condRight := expectNodeType[*ast.NumericLiteralNode](
		t,
		condition.GetRight(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(10), condRight.Value, "Expected condition right value 10, got %f", condRight.Value)

	// Check update
	update := expectNodeType[*ast.UpdateExpressionNode](
		t,
		forStatement.GetUpdate(),
		ast.UpdateExpression,
	)
	assert.Equal(t, "++", update.Operator.Value, "Expected update operator '++', got %s", update.Operator.Value)

	updateValue := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		update.GetValue(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "i", updateValue.Identifier, "Expected update identifier 'i', got %s", updateValue.Identifier)

	// Check body
	block := expectNodeType[*ast.BasicNode](
		t,
		forStatement.GetBody(),
		ast.Block,
	)

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		block.GetChildren()[0],
		ast.StatementList,
	)

	statement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", statement.Identifier, "Expected statement identifier 'foo', got %s", statement.Identifier)

	// Test for statement with empty expressions
	forStatement = expectScriptValue[*ast.ForStatementNode](
		t,
		"for (;;) { }",
		ast.ForStatement,
	)

	assert.Nil(t, forStatement.GetInitializer(), "Expected nil initializer")
	assert.Nil(t, forStatement.GetCondition(), "Expected nil condition")
	assert.Nil(t, forStatement.GetUpdate(), "Expected nil update")
}

// IterationStatement : ForInStatement
func TestForInStatement(t *testing.T) {
	// Test for-in statement with variable declaration
	forInStatement := expectScriptValue[*ast.ForInStatementNode](
		t,
		"for (var x in obj) { foo; }",
		ast.ForInStatement,
	)

	// Check target (variable declaration)
	varDecl := expectNodeType[*ast.BasicNode](
		t,
		forInStatement.GetTarget(),
		ast.VariableDeclaration,
	)

	binding := expectNodeType[*ast.BindingIdentifierNode](
		t,
		varDecl.GetChildren()[0],
		ast.BindingIdentifier,
	)
	assert.Equal(t, "x", binding.Identifier, "Expected binding identifier 'x', got %s", binding.Identifier)

	// Check iterable
	iterable := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		forInStatement.GetIterable(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "obj", iterable.Identifier, "Expected iterable identifier 'obj', got %s", iterable.Identifier)

	// Check body
	block := expectNodeType[*ast.BasicNode](
		t,
		forInStatement.GetBody(),
		ast.Block,
	)

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		block.GetChildren()[0],
		ast.StatementList,
	)

	statement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", statement.Identifier, "Expected statement identifier 'foo', got %s", statement.Identifier)

	// Test for-in statement with identifier reference
	forInStatement = expectScriptValue[*ast.ForInStatementNode](
		t,
		"for (x in obj) { foo; }",
		ast.ForInStatement,
	)

	// Check target (identifier reference)
	target := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		forInStatement.GetTarget(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", target.Identifier, "Expected target identifier 'x', got %s", target.Identifier)

	// Check iterable
	iterable = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		forInStatement.GetIterable(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "obj", iterable.Identifier, "Expected iterable identifier 'obj', got %s", iterable.Identifier)

	// Check body
	block = expectNodeType[*ast.BasicNode](
		t,
		forInStatement.GetBody(),
		ast.Block,
	)

	statementList = expectNodeType[*ast.StatementListNode](
		t,
		block.GetChildren()[0],
		ast.StatementList,
	)

	statement = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", statement.Identifier, "Expected statement identifier 'foo', got %s", statement.Identifier)

	// Test for-in statement with lexical declarations (let and const)
	testLexicalForIn := func(keyword string, isConst bool) {
		forInStatement := expectScriptValue[*ast.ForInStatementNode](
			t,
			fmt.Sprintf("for (%s x in obj) { foo; }", keyword),
			ast.ForInStatement,
		)

		// Check target (lexical binding)
		lexicalBinding := expectNodeType[*ast.LexicalBindingNode](
			t,
			forInStatement.GetTarget(),
			ast.LexicalBinding,
		)
		assert.Equal(t, isConst, lexicalBinding.Const, "Expected %s binding, got %s", keyword, map[bool]string{true: "const", false: "let"}[!isConst])

		bindingIdentifier := expectNodeType[*ast.BindingIdentifierNode](
			t,
			lexicalBinding.GetTarget(),
			ast.BindingIdentifier,
		)
		assert.Equal(t, "x", bindingIdentifier.Identifier, "Expected binding identifier 'x', got %s", bindingIdentifier.Identifier)

		// Check iterable
		iterable := expectNodeType[*ast.IdentifierReferenceNode](
			t,
			forInStatement.GetIterable(),
			ast.IdentifierReference,
		)
		assert.Equal(t, "obj", iterable.Identifier, "Expected iterable identifier 'obj', got %s", iterable.Identifier)

		// Check body
		block := expectNodeType[*ast.BasicNode](
			t,
			forInStatement.GetBody(),
			ast.Block,
		)

		statementList := expectNodeType[*ast.StatementListNode](
			t,
			block.GetChildren()[0],
			ast.StatementList,
		)

		statement := expectNodeType[*ast.IdentifierReferenceNode](
			t,
			statementList.GetChildren()[0],
			ast.IdentifierReference,
		)
		assert.Equal(t, "foo", statement.Identifier, "Expected statement identifier 'foo', got %s", statement.Identifier)
	}

	testLexicalForIn("let", false)
	testLexicalForIn("const", true)
}

// IterationStatement : ForOfStatement
func TestForOfStatement(t *testing.T) {
	// Test basic for-of statement
	forOfStatement := expectScriptValue[*ast.ForOfStatementNode](
		t,
		"for (let x of arr) { foo; }",
		ast.ForOfStatement,
	)

	// Check target (lexical binding)
	lexicalBinding := expectNodeType[*ast.LexicalBindingNode](
		t,
		forOfStatement.GetTarget(),
		ast.LexicalBinding,
	)
	assert.False(t, lexicalBinding.Const, "Expected let binding, got const")

	bindingIdentifier := expectNodeType[*ast.BindingIdentifierNode](
		t,
		lexicalBinding.GetTarget(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "x", bindingIdentifier.Identifier, "Expected binding identifier 'x', got %s", bindingIdentifier.Identifier)

	// Check iterable
	iterable := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		forOfStatement.GetIterable(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "arr", iterable.Identifier, "Expected iterable identifier 'arr', got %s", iterable.Identifier)

	// Check body
	block := expectNodeType[*ast.BasicNode](
		t,
		forOfStatement.GetBody(),
		ast.Block,
	)

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		block.GetChildren()[0],
		ast.StatementList,
	)

	statement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", statement.Identifier, "Expected statement identifier 'foo', got %s", statement.Identifier)

	// Test for-await-of statement
	functionExpression := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		`
		async function foo() {
			for await (let x of arr) { foo; }
		}
		`,
		ast.FunctionExpression,
	)

	statementList = expectNodeType[*ast.StatementListNode](
		t,
		functionExpression.GetBody(),
		ast.StatementList,
	)

	forOfStatement = expectNodeType[*ast.ForOfStatementNode](
		t,
		statementList.GetChildren()[0],
		ast.ForOfStatement,
	)
	assert.True(t, forOfStatement.Await, "Expected await flag to be true")

	// Check target (lexical binding)
	lexicalBinding = expectNodeType[*ast.LexicalBindingNode](
		t,
		forOfStatement.GetTarget(),
		ast.LexicalBinding,
	)
	assert.False(t, lexicalBinding.Const, "Expected let binding, got const")

	bindingIdentifier = expectNodeType[*ast.BindingIdentifierNode](
		t,
		lexicalBinding.GetTarget(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "x", bindingIdentifier.Identifier, "Expected binding identifier 'x', got %s", bindingIdentifier.Identifier)

	// Check iterable
	iterable = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		forOfStatement.GetIterable(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "arr", iterable.Identifier, "Expected iterable identifier 'arr', got %s", iterable.Identifier)

	// Check body
	block = expectNodeType[*ast.BasicNode](
		t,
		forOfStatement.GetBody(),
		ast.Block,
	)

	statementList = expectNodeType[*ast.StatementListNode](
		t,
		block.GetChildren()[0],
		ast.StatementList,
	)

	statement = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", statement.Identifier, "Expected statement identifier 'foo', got %s", statement.Identifier)
}

// BreakableStatement : SwitchStatement
func TestSwitchStatement(t *testing.T) {
	// Test basic switch statement
	switchStatement := expectScriptValue[*ast.SwitchStatementNode](
		t,
		"switch (x) { case 1: foo; case 2: bar; default: baz; }",
		ast.SwitchStatement,
	)

	// Check target
	target := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		switchStatement.GetTarget(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", target.Identifier, "Expected target identifier 'x', got %s", target.Identifier)

	// Check cases and their statement lists
	children := switchStatement.GetChildren()
	assert.Equal(t, 6, len(children), "Expected 6 children (3 cases + 3 statement lists), got %d", len(children))

	// Check first case and its statement list
	firstCase := expectNodeType[*ast.SwitchCaseNode](
		t,
		children[0],
		ast.SwitchCase,
	)
	caseExpression := expectNodeType[*ast.NumericLiteralNode](
		t,
		firstCase.GetExpression(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), caseExpression.Value, "Expected case expression value 1, got %f", caseExpression.Value)

	firstStatementList := expectNodeType[*ast.StatementListNode](
		t,
		children[1],
		ast.StatementList,
	)
	firstStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		firstStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", firstStatement.Identifier, "Expected statement identifier 'foo', got %s", firstStatement.Identifier)

	// Check second case and its statement list
	secondCase := expectNodeType[*ast.SwitchCaseNode](
		t,
		children[2],
		ast.SwitchCase,
	)
	caseExpression = expectNodeType[*ast.NumericLiteralNode](
		t,
		secondCase.GetExpression(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(2), caseExpression.Value, "Expected case expression value 2, got %f", caseExpression.Value)

	secondStatementList := expectNodeType[*ast.StatementListNode](
		t,
		children[3],
		ast.StatementList,
	)
	secondStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		secondStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", secondStatement.Identifier, "Expected statement identifier 'bar', got %s", secondStatement.Identifier)

	// Check default case and its statement list
	defaultCase := expectNodeType[*ast.BasicNode](
		t,
		children[4],
		ast.SwitchDefault,
	)
	assert.NotNil(t, defaultCase, "Expected default case to not be nil")

	defaultStatementList := expectNodeType[*ast.StatementListNode](
		t,
		children[5],
		ast.StatementList,
	)
	defaultStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		defaultStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "baz", defaultStatement.Identifier, "Expected statement identifier 'baz', got %s", defaultStatement.Identifier)

	// Test switch statement with block statements
	switchStatement = expectScriptValue[*ast.SwitchStatementNode](
		t,
		"switch (x) { case 1: { foo; bar; } default: { baz; } }",
		ast.SwitchStatement,
	)

	// Check target
	target = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		switchStatement.GetTarget(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "x", target.Identifier, "Expected target identifier 'x', got %s", target.Identifier)

	// Check cases and their statement lists
	children = switchStatement.GetChildren()
	assert.Equal(t, 4, len(children), "Expected 4 children (2 cases + 2 statement lists), got %d", len(children))

	// Check case with block statement list
	caseWithBlock := expectNodeType[*ast.SwitchCaseNode](
		t,
		children[0],
		ast.SwitchCase,
	)
	caseExpression = expectNodeType[*ast.NumericLiteralNode](
		t,
		caseWithBlock.GetExpression(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), caseExpression.Value, "Expected case expression value 1, got %f", caseExpression.Value)

	caseStatementList := expectNodeType[*ast.StatementListNode](
		t,
		children[1],
		ast.StatementList,
	)
	assert.Equal(t, 1, len(caseStatementList.GetChildren()), "Expected 1 statement in case statement list, got %d", len(caseStatementList.GetChildren()))
	blockStatement := expectNodeType[*ast.BasicNode](
		t,
		caseStatementList.GetChildren()[0],
		ast.Block,
	)
	blockStatementList := expectNodeType[*ast.StatementListNode](
		t,
		blockStatement.GetChildren()[0],
		ast.StatementList,
	)
	assert.Equal(t, 2, len(blockStatementList.GetChildren()), "Expected 2 statements in block, got %d", len(blockStatementList.GetChildren()))

	firstBlockStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		blockStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", firstBlockStatement.Identifier, "Expected first statement identifier 'foo', got %s", firstBlockStatement.Identifier)
	secondBlockStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		blockStatementList.GetChildren()[1],
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", secondBlockStatement.Identifier, "Expected second statement identifier 'bar', got %s", secondBlockStatement.Identifier)

	// Check default with block statement list
	defaultWithBlock := expectNodeType[*ast.BasicNode](
		t,
		children[2],
		ast.SwitchDefault,
	)
	assert.NotNil(t, defaultWithBlock, "Expected default case to not be nil")

	caseStatementList = expectNodeType[*ast.StatementListNode](
		t,
		children[3],
		ast.StatementList,
	)
	assert.Equal(t, 1, len(caseStatementList.GetChildren()), "Expected 1 statement in case statement list, got %d", len(caseStatementList.GetChildren()))
	blockStatement = expectNodeType[*ast.BasicNode](
		t,
		caseStatementList.GetChildren()[0],
		ast.Block,
	)
	defaultBlockStatementList := expectNodeType[*ast.StatementListNode](
		t,
		blockStatement.GetChildren()[0],
		ast.StatementList,
	)
	assert.Equal(t, 1, len(defaultBlockStatementList.GetChildren()), "Expected 1 statement in block, got %d", len(defaultBlockStatementList.GetChildren()))

	defaultBlockStatement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		defaultBlockStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "baz", defaultBlockStatement.Identifier, "Expected statement identifier 'baz', got %s", defaultBlockStatement.Identifier)
}

// Statement : ContinueStatement
func TestContinueStatement(t *testing.T) {
	// Test continue statement without label
	continueStatement := expectScriptValue[*ast.ContinueStatementNode](
		t,
		"continue;",
		ast.ContinueStatement,
	)
	assert.Nil(t, continueStatement.GetLabel(), "Expected no label, got a label")

	// Test continue statement with label
	continueStatement = expectScriptValue[*ast.ContinueStatementNode](
		t,
		"continue foo;",
		ast.ContinueStatement,
	)

	label := expectNodeType[*ast.LabelIdentifierNode](
		t,
		continueStatement.GetLabel(),
		ast.LabelIdentifier,
	)
	assert.Equal(t, "foo", label.Identifier, "Expected label identifier 'foo', got %s", label.Identifier)
}

// Statement : ReturnStatement
func TestReturnStatement(t *testing.T) {
	// Test return statement without value
	functionExpr := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"function foo() { return; }",
		ast.FunctionExpression,
	)

	body := expectNodeType[*ast.StatementListNode](
		t,
		functionExpr.GetBody(),
		ast.StatementList,
	)

	returnStatement := expectNodeType[*ast.ReturnStatementNode](
		t,
		body.GetChildren()[0],
		ast.ReturnStatement,
	)
	assert.Nil(t, returnStatement.GetValue(), "Expected no return value, got a value")

	// Test return statement with value
	functionExpr = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"function foo() { return 42; }",
		ast.FunctionExpression,
	)

	body = expectNodeType[*ast.StatementListNode](
		t,
		functionExpr.GetBody(),
		ast.StatementList,
	)

	returnStatement = expectNodeType[*ast.ReturnStatementNode](
		t,
		body.GetChildren()[0],
		ast.ReturnStatement,
	)

	value := expectNodeType[*ast.NumericLiteralNode](
		t,
		returnStatement.GetValue(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(42), value.Value, "Expected return value 42, got %f", value.Value)

	// Test return statement with expression
	functionExpr = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"function foo() { return a + b; }",
		ast.FunctionExpression,
	)

	body = expectNodeType[*ast.StatementListNode](
		t,
		functionExpr.GetBody(),
		ast.StatementList,
	)

	returnStatement = expectNodeType[*ast.ReturnStatementNode](
		t,
		body.GetChildren()[0],
		ast.ReturnStatement,
	)

	addExpr := expectNodeType[*ast.AdditiveExpressionNode](
		t,
		returnStatement.GetValue(),
		ast.AdditiveExpression,
	)

	leftOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		addExpr.GetLeft(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "a", leftOperand.Identifier, "Expected left operand identifier 'a', got %s", leftOperand.Identifier)

	rightOperand := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		addExpr.GetRight(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "b", rightOperand.Identifier, "Expected right operand identifier 'b', got %s", rightOperand.Identifier)
}

// Statement : WithStatement
func TestWithStatement(t *testing.T) {
	// Test basic with statement
	withStatement := expectScriptValue[*ast.WithStatementNode](
		t,
		"with (obj) { foo; }",
		ast.WithStatement,
	)

	// Check expression
	expression := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		withStatement.GetExpression(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "obj", expression.Identifier, "Expected expression identifier 'obj', got %s", expression.Identifier)

	// Check body
	block := expectNodeType[*ast.BasicNode](
		t,
		withStatement.GetBody(),
		ast.Block,
	)

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		block.GetChildren()[0],
		ast.StatementList,
	)

	statement := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		statementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", statement.Identifier, "Expected statement identifier 'foo', got %s", statement.Identifier)
}

// Statement : LabelledStatement
func TestLabelledStatement(t *testing.T) {
	// Test basic labelled statement
	labelledStatement := expectScriptValue[*ast.LabelledStatementNode](
		t,
		"myLabel: foo;",
		ast.LabelledStatement,
	)

	// Check label
	label := expectNodeType[*ast.LabelIdentifierNode](
		t,
		labelledStatement.GetLabel(),
		ast.LabelIdentifier,
	)
	assert.Equal(t, "myLabel", label.Identifier, "Expected label identifier 'myLabel', got %s", label.Identifier)

	// Check labelled item
	labelledItem := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		labelledStatement.GetLabelledItem(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", labelledItem.Identifier, "Expected labelled item identifier 'foo', got %s", labelledItem.Identifier)

	// Test nested labelled statement
	nestedLabelledStatement := expectScriptValue[*ast.LabelledStatementNode](
		t,
		"outer: inner: foo;",
		ast.LabelledStatement,
	)

	// Check outer label
	outerLabel := expectNodeType[*ast.LabelIdentifierNode](
		t,
		nestedLabelledStatement.GetLabel(),
		ast.LabelIdentifier,
	)
	assert.Equal(t, "outer", outerLabel.Identifier, "Expected outer label identifier 'outer', got %s", outerLabel.Identifier)

	// Check inner labelled statement
	innerLabelledStatement := expectNodeType[*ast.LabelledStatementNode](
		t,
		nestedLabelledStatement.GetLabelledItem(),
		ast.LabelledStatement,
	)

	// Check inner label
	innerLabel := expectNodeType[*ast.LabelIdentifierNode](
		t,
		innerLabelledStatement.GetLabel(),
		ast.LabelIdentifier,
	)
	assert.Equal(t, "inner", innerLabel.Identifier, "Expected inner label identifier 'inner', got %s", innerLabel.Identifier)

	// Check inner labelled item
	innerLabelledItem := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		innerLabelledStatement.GetLabelledItem(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", innerLabelledItem.Identifier, "Expected inner labelled item identifier 'foo', got %s", innerLabelledItem.Identifier)
}

// Statement : ThrowStatement
func TestThrowStatement(t *testing.T) {
	// Test basic throw statement
	throwStatement := expectScriptValue[*ast.ThrowStatementNode](
		t,
		"throw foo;",
		ast.ThrowStatement,
	)

	// Check expression
	expression := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		throwStatement.GetExpression(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", expression.Identifier, "Expected expression identifier 'foo', got %s", expression.Identifier)

	// Test throw statement with literal
	throwStatement = expectScriptValue[*ast.ThrowStatementNode](
		t,
		"throw 'error';",
		ast.ThrowStatement,
	)

	// Check string literal expression
	stringLiteral := expectNodeType[*ast.StringLiteralNode](
		t,
		throwStatement.GetExpression(),
		ast.StringLiteral,
	)
	assert.Equal(t, "error", stringLiteral.Value, "Expected string literal value 'error', got %s", stringLiteral.Value)
}

// Statement : TryStatement
func TestTryStatement(t *testing.T) {
	// Test try statement with catch block
	tryStatement := expectScriptValue[*ast.TryStatementNode](
		t,
		"try { foo; } catch (err) { bar; }",
		ast.TryStatement,
	)

	// Check try block
	tryBlock := expectNodeType[*ast.BasicNode](
		t,
		tryStatement.GetBlock(),
		ast.Block,
	)
	tryStatementList := expectNodeType[*ast.StatementListNode](
		t,
		tryBlock.GetChildren()[0],
		ast.StatementList,
	)
	tryExpression := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		tryStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", tryExpression.Identifier, "Expected try block identifier 'foo', got %s", tryExpression.Identifier)

	// Check catch block
	catchBlock := expectNodeType[*ast.CatchNode](
		t,
		tryStatement.GetCatch(),
		ast.Catch,
	)

	// Check catch parameter
	catchParam := expectNodeType[*ast.BindingIdentifierNode](
		t,
		catchBlock.GetTarget(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "err", catchParam.Identifier, "Expected catch parameter 'err', got %s", catchParam.Identifier)

	// Check catch body
	catchBody := expectNodeType[*ast.BasicNode](
		t,
		catchBlock.GetBlock(),
		ast.Block,
	)
	catchStatementList := expectNodeType[*ast.StatementListNode](
		t,
		catchBody.GetChildren()[0],
		ast.StatementList,
	)
	catchExpression := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		catchStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", catchExpression.Identifier, "Expected catch block identifier 'bar', got %s", catchExpression.Identifier)

	// Test try statement with finally block
	tryStatement = expectScriptValue[*ast.TryStatementNode](
		t,
		"try { foo; } finally { baz; }",
		ast.TryStatement,
	)

	// Check try block
	tryBlock = expectNodeType[*ast.BasicNode](
		t,
		tryStatement.GetBlock(),
		ast.Block,
	)
	tryStatementList = expectNodeType[*ast.StatementListNode](
		t,
		tryBlock.GetChildren()[0],
		ast.StatementList,
	)
	tryExpression = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		tryStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", tryExpression.Identifier, "Expected try block identifier 'foo', got %s", tryExpression.Identifier)

	// Check finally block
	finallyBlock := expectNodeType[*ast.BasicNode](
		t,
		tryStatement.GetFinally(),
		ast.Block,
	)
	finallyStatementList := expectNodeType[*ast.StatementListNode](
		t,
		finallyBlock.GetChildren()[0],
		ast.StatementList,
	)
	finallyExpression := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		finallyStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "baz", finallyExpression.Identifier, "Expected finally block identifier 'baz', got %s", finallyExpression.Identifier)

	// Test try statement with catch and finally blocks
	tryStatement = expectScriptValue[*ast.TryStatementNode](
		t,
		"try { foo; } catch (err) { bar; } finally { baz; }",
		ast.TryStatement,
	)

	// Check try block
	tryBlock = expectNodeType[*ast.BasicNode](
		t,
		tryStatement.GetBlock(),
		ast.Block,
	)
	tryStatementList = expectNodeType[*ast.StatementListNode](
		t,
		tryBlock.GetChildren()[0],
		ast.StatementList,
	)
	tryExpression = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		tryStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "foo", tryExpression.Identifier, "Expected try block identifier 'foo', got %s", tryExpression.Identifier)

	// Check catch block
	catchBlock = expectNodeType[*ast.CatchNode](
		t,
		tryStatement.GetCatch(),
		ast.Catch,
	)

	// Check catch parameter
	catchParam = expectNodeType[*ast.BindingIdentifierNode](
		t,
		catchBlock.GetTarget(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "err", catchParam.Identifier, "Expected catch parameter 'err', got %s", catchParam.Identifier)

	// Check catch body
	catchBody = expectNodeType[*ast.BasicNode](
		t,
		catchBlock.GetBlock(),
		ast.Block,
	)
	catchStatementList = expectNodeType[*ast.StatementListNode](
		t,
		catchBody.GetChildren()[0],
		ast.StatementList,
	)
	catchExpression = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		catchStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "bar", catchExpression.Identifier, "Expected catch block identifier 'bar', got %s", catchExpression.Identifier)

	// Check finally block
	finallyBlock = expectNodeType[*ast.BasicNode](
		t,
		tryStatement.GetFinally(),
		ast.Block,
	)
	finallyStatementList = expectNodeType[*ast.StatementListNode](
		t,
		finallyBlock.GetChildren()[0],
		ast.StatementList,
	)
	finallyExpression = expectNodeType[*ast.IdentifierReferenceNode](
		t,
		finallyStatementList.GetChildren()[0],
		ast.IdentifierReference,
	)
	assert.Equal(t, "baz", finallyExpression.Identifier, "Expected finally block identifier 'baz', got %s", finallyExpression.Identifier)
}

// Statement : DebuggerStatement
func TestDebuggerStatement(t *testing.T) {
	// Test basic debugger statement
	expectScriptValue[*ast.BasicNode](
		t,
		"debugger;",
		ast.DebuggerStatement,
	)
}

// Declaration : FunctionDeclaration
func TestFunctionDeclaration(t *testing.T) {
	// Test basic function declaration
	functionDeclaration := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"function foo() { }",
		ast.FunctionExpression,
	)

	// Check name
	name := expectNodeType[*ast.BindingIdentifierNode](
		t,
		functionDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", name.Identifier, "Expected function name 'foo', got %s", name.Identifier)

	// Check empty parameters
	assert.Empty(t, functionDeclaration.GetParameters(), "Expected no parameters, got %d", len(functionDeclaration.GetParameters()))

	// Check empty body
	body := expectNodeType[*ast.StatementListNode](
		t,
		functionDeclaration.GetBody(),
		ast.StatementList,
	)
	assert.Empty(t, body.GetChildren(), "Expected empty body, got %d statements", len(body.GetChildren()))

	// Test function declaration with parameters
	functionDeclaration = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"function bar(a, b) { return a + b; }",
		ast.FunctionExpression,
	)

	// Check name
	name = expectNodeType[*ast.BindingIdentifierNode](
		t,
		functionDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "bar", name.Identifier, "Expected function name 'bar', got %s", name.Identifier)

	// Check parameters
	parameters := functionDeclaration.GetParameters()
	assert.Equal(t, 2, len(parameters), "Expected 2 parameters, got %d", len(parameters))

	// Check first parameter
	param1 := expectNodeType[*ast.BindingElementNode](
		t,
		parameters[0],
		ast.BindingElement,
	)
	paramIdentifier := expectNodeType[*ast.BindingIdentifierNode](
		t,
		param1.GetTarget(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "a", paramIdentifier.Identifier, "Expected first parameter 'a', got %s", paramIdentifier.Identifier)

	// Check second parameter
	param2 := expectNodeType[*ast.BindingElementNode](
		t,
		parameters[1],
		ast.BindingElement,
	)
	paramIdentifier = expectNodeType[*ast.BindingIdentifierNode](
		t,
		param2.GetTarget(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "b", paramIdentifier.Identifier, "Expected second parameter 'b', got %s", paramIdentifier.Identifier)

	// TODO: Test [+Default] path when that's implemented
}

// Declaration : GeneratorDeclaration
func TestGeneratorDeclaration(t *testing.T) {
	// Test basic generator declaration
	generatorDeclaration := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"function* foo() { }",
		ast.FunctionExpression,
	)
	assert.True(t, generatorDeclaration.Generator, "Expected generator function")
	assert.False(t, generatorDeclaration.Async, "Expected generator function, not async function")
	nameIdentifier := expectNodeType[*ast.BindingIdentifierNode](
		t,
		generatorDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", nameIdentifier.Identifier, "Expected generator name 'foo', got %s", nameIdentifier.Identifier)

	// Test with yield statement
	generatorDeclaration = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"function* foo() { yield 1; }",
		ast.FunctionExpression,
	)
	assert.True(t, generatorDeclaration.Generator, "Expected generator function")
	assert.False(t, generatorDeclaration.Async, "Expected generator function, not async function")
	nameIdentifier = expectNodeType[*ast.BindingIdentifierNode](
		t,
		generatorDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", nameIdentifier.Identifier, "Expected generator name 'foo', got %s", nameIdentifier.Identifier)

	// Check body contains yield statement
	body := expectNodeType[*ast.StatementListNode](
		t,
		generatorDeclaration.GetBody(),
		ast.StatementList,
	)
	assert.Equal(t, 1, len(body.GetChildren()), "Expected 1 statement in body, got %d", len(body.GetChildren()))

	yieldExpression := expectNodeType[*ast.YieldExpressionNode](
		t,
		body.GetChildren()[0],
		ast.YieldExpression,
	)

	argument := expectNodeType[*ast.NumericLiteralNode](
		t,
		yieldExpression.GetExpression(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), argument.Value, "Expected yield argument value 1, got %v", argument.Value)

	// TODO: Test [+Default] path when that's implemented
}

// Declaration : AsyncFunctionDeclaration
func TestAsyncFunctionDeclaration(t *testing.T) {
	// Test basic async function declaration
	asyncFunctionDeclaration := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"async function foo() { }",
		ast.FunctionExpression,
	)
	assert.True(t, asyncFunctionDeclaration.Async, "Expected async function")
	assert.False(t, asyncFunctionDeclaration.Generator, "Expected async function, not generator function")
	nameIdentifier := expectNodeType[*ast.BindingIdentifierNode](
		t,
		asyncFunctionDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", nameIdentifier.Identifier, "Expected async function name 'foo', got %s", nameIdentifier.Identifier)

	// Test with await statement
	asyncFunctionDeclaration = expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"async function foo() { await 1; }",
		ast.FunctionExpression,
	)
	assert.True(t, asyncFunctionDeclaration.Async, "Expected async function")
	assert.False(t, asyncFunctionDeclaration.Generator, "Expected async function, not generator function")
	nameIdentifier = expectNodeType[*ast.BindingIdentifierNode](
		t,
		asyncFunctionDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", nameIdentifier.Identifier, "Expected async function name 'foo', got %s", nameIdentifier.Identifier)

	// Check body contains await statement
	body := expectNodeType[*ast.StatementListNode](
		t,
		asyncFunctionDeclaration.GetBody(),
		ast.StatementList,
	)
	assert.Equal(t, 1, len(body.GetChildren()), "Expected 1 statement in body, got %d", len(body.GetChildren()))

	awaitExpression := expectNodeType[*ast.AwaitExpressionNode](
		t,
		body.GetChildren()[0],
		ast.AwaitExpression,
	)

	argument := expectNodeType[*ast.NumericLiteralNode](
		t,
		awaitExpression.GetExpression(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), argument.Value, "Expected await argument value 1, got %v", argument.Value)

	// TODO: Test [+Default] path when that's implemented
}

// Declaration : AsyncGeneratorDeclaration
func TestAsyncGeneratorDeclaration(t *testing.T) {
	// Test basic async generator with both an await and yield statement
	asyncGeneratorDeclaration := expectScriptValue[*ast.FunctionExpressionNode](
		t,
		"async function* foo() { await 1; yield 2; }",
		ast.FunctionExpression,
	)
	assert.True(t, asyncGeneratorDeclaration.Async, "Expected async generator")
	assert.True(t, asyncGeneratorDeclaration.Generator, "Expected generator function")
	nameIdentifier := expectNodeType[*ast.BindingIdentifierNode](
		t,
		asyncGeneratorDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "foo", nameIdentifier.Identifier, "Expected async generator name 'foo', got %s", nameIdentifier.Identifier)

	// Check body contains await and yield statements
	body := expectNodeType[*ast.StatementListNode](
		t,
		asyncGeneratorDeclaration.GetBody(),
		ast.StatementList,
	)
	assert.Equal(t, 2, len(body.GetChildren()), "Expected 2 statements in body, got %d", len(body.GetChildren()))

	awaitExpression := expectNodeType[*ast.AwaitExpressionNode](
		t,
		body.GetChildren()[0],
		ast.AwaitExpression,
	)
	numericLiteral := expectNodeType[*ast.NumericLiteralNode](
		t,
		awaitExpression.GetExpression(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), numericLiteral.Value, "Expected await argument value 1, got %v", numericLiteral.Value)

	yieldExpression := expectNodeType[*ast.YieldExpressionNode](
		t,
		body.GetChildren()[1],
		ast.YieldExpression,
	)
	numericLiteral = expectNodeType[*ast.NumericLiteralNode](
		t,
		yieldExpression.GetExpression(),
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(2), numericLiteral.Value, "Expected yield argument value 2, got %v", numericLiteral.Value)

	// TODO: Test [+Default] path when that's implemented
}

// Declaration : ClassDeclaration
func TestClassDeclaration(t *testing.T) {
	// Test basic class declaration
	classDeclaration := expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Foo { }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check name
	name := expectNodeType[*ast.BindingIdentifierNode](
		t,
		classDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "Foo", name.Identifier, "Expected class name 'Foo', got %s", name.Identifier)

	// Check empty elements
	assert.Empty(t, classDeclaration.GetElements(), "Expected no elements, got %d", len(classDeclaration.GetElements()))

	// Test class declaration with heritage
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Bar extends Foo { }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check name
	name = expectNodeType[*ast.BindingIdentifierNode](
		t,
		classDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "Bar", name.Identifier, "Expected class name 'Bar', got %s", name.Identifier)

	// Check heritage
	heritage := expectNodeType[*ast.IdentifierReferenceNode](
		t,
		classDeclaration.GetHeritage(),
		ast.IdentifierReference,
	)
	assert.Equal(t, "Foo", heritage.Identifier, "Expected heritage identifier 'Foo', got %s", heritage.Identifier)

	// Check empty elements
	assert.Empty(t, classDeclaration.GetElements(), "Expected no elements, got %d", len(classDeclaration.GetElements()))

	// Test class declaration with method
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Baz { method() { } }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check name
	name = expectNodeType[*ast.BindingIdentifierNode](
		t,
		classDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "Baz", name.Identifier, "Expected class name 'Baz', got %s", name.Identifier)

	// Check method definition
	elements := classDeclaration.GetElements()
	assert.Equal(t, 1, len(elements), "Expected 1 element, got %d", len(elements))

	methodDefinition := expectNodeType[*ast.MethodDefinitionNode](
		t,
		elements[0],
		ast.MethodDefinition,
	)

	methodName := expectNodeType[*ast.IdentifierNameNode](
		t,
		methodDefinition.GetName(),
		ast.IdentifierName,
	)
	assert.Equal(t, "method", methodName.Identifier, "Expected method name 'method', got %s", methodName.Identifier)

	// Test class declaration with static method
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Qux { static method() { } }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check name
	name = expectNodeType[*ast.BindingIdentifierNode](
		t,
		classDeclaration.GetName(),
		ast.BindingIdentifier,
	)
	assert.Equal(t, "Qux", name.Identifier, "Expected class name 'Qux', got %s", name.Identifier)

	// Check method definition
	elements = classDeclaration.GetElements()
	assert.Equal(t, 1, len(elements), "Expected 1 element, got %d", len(elements))

	methodDefinition = expectNodeType[*ast.MethodDefinitionNode](
		t,
		elements[0],
		ast.MethodDefinition,
	)
	assert.True(t, methodDefinition.Static, "Expected static method")
	methodName = expectNodeType[*ast.IdentifierNameNode](
		t,
		methodDefinition.GetName(),
		ast.IdentifierName,
	)
	assert.Equal(t, "method", methodName.Identifier, "Expected method name 'method', got %s", methodName.Identifier)

	// Test with field declaration
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Qux { field = 1; }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check field
	elements = classDeclaration.GetElements()
	assert.Equal(t, 1, len(elements), "Expected 1 element, got %d", len(elements))

	fieldDefinition := expectNodeType[*ast.PropertyDefinitionNode](
		t,
		elements[0],
		ast.PropertyDefinition,
	)
	fieldName := expectNodeType[*ast.IdentifierNameNode](
		t,
		fieldDefinition.GetKey(),
		ast.IdentifierName,
	)
	assert.Equal(t, "field", fieldName.Identifier, "Expected field name 'field', got %s", fieldName.Identifier)
	fieldInitializer := expectNodeType[*ast.BasicNode](
		t,
		fieldDefinition.GetValue(),
		ast.Initializer,
	)
	fieldValue := expectNodeType[*ast.NumericLiteralNode](
		t,
		fieldInitializer.GetChildren()[0],
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), fieldValue.Value, "Expected field value 1, got %v", fieldValue.Value)

	// Test with private field
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Qux { #field = 1; }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check field
	elements = classDeclaration.GetElements()
	assert.Equal(t, 1, len(elements), "Expected 1 element, got %d", len(elements))

	fieldDefinition = expectNodeType[*ast.PropertyDefinitionNode](
		t,
		elements[0],
		ast.PropertyDefinition,
	)
	fieldName = expectNodeType[*ast.IdentifierNameNode](
		t,
		fieldDefinition.GetKey(),
		ast.IdentifierName,
	)
	assert.Equal(t, "#field", fieldName.Identifier, "Expected field name 'field', got %s", fieldName.Identifier)
	fieldInitializer = expectNodeType[*ast.BasicNode](
		t,
		fieldDefinition.GetValue(),
		ast.Initializer,
	)
	fieldValue = expectNodeType[*ast.NumericLiteralNode](
		t,
		fieldInitializer.GetChildren()[0],
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), fieldValue.Value, "Expected field value 1, got %v", fieldValue.Value)

	// Test with getter
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Qux { get field() { } }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check field
	elements = classDeclaration.GetElements()
	assert.Equal(t, 1, len(elements), "Expected 1 element, got %d", len(elements))

	methodDefinition = expectNodeType[*ast.MethodDefinitionNode](
		t,
		elements[0],
		ast.MethodDefinition,
	)
	methodName = expectNodeType[*ast.IdentifierNameNode](
		t,
		methodDefinition.GetName(),
		ast.IdentifierName,
	)
	assert.Equal(t, "field", methodName.Identifier, "Expected method name 'field', got %s", methodName.Identifier)
	assert.True(t, methodDefinition.Getter, "Expected getter method")

	// Test with setter
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Qux { set field(value) { } }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check field
	elements = classDeclaration.GetElements()
	assert.Equal(t, 1, len(elements), "Expected 1 element, got %d", len(elements))

	methodDefinition = expectNodeType[*ast.MethodDefinitionNode](
		t,
		elements[0],
		ast.MethodDefinition,
	)
	methodName = expectNodeType[*ast.IdentifierNameNode](
		t,
		methodDefinition.GetName(),
		ast.IdentifierName,
	)
	assert.Equal(t, "field", methodName.Identifier, "Expected method name 'field', got %s", methodName.Identifier)
	assert.True(t, methodDefinition.Setter, "Expected setter method")

	// Test with static field
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Qux { static field = 1; }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check field
	elements = classDeclaration.GetElements()
	assert.Equal(t, 1, len(elements), "Expected 1 element, got %d", len(elements))

	fieldDefinition = expectNodeType[*ast.PropertyDefinitionNode](
		t,
		elements[0],
		ast.PropertyDefinition,
	)
	assert.True(t, fieldDefinition.Static, "Expected static field")

	fieldName = expectNodeType[*ast.IdentifierNameNode](
		t,
		fieldDefinition.GetKey(),
		ast.IdentifierName,
	)
	assert.Equal(t, "field", fieldName.Identifier, "Expected field name 'field', got %s", fieldName.Identifier)
	fieldInitializer = expectNodeType[*ast.BasicNode](
		t,
		fieldDefinition.GetValue(),
		ast.Initializer,
	)
	fieldValue = expectNodeType[*ast.NumericLiteralNode](
		t,
		fieldInitializer.GetChildren()[0],
		ast.NumericLiteral,
	)
	assert.Equal(t, float64(1), fieldValue.Value, "Expected field value 1, got %v", fieldValue.Value)

	// Test static block
	classDeclaration = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"class Qux { static { this.foo = 1; } }",
		ast.ClassExpression,
	)
	assert.True(t, classDeclaration.Declaration, "Expected class declaration")

	// Check static block
	elements = classDeclaration.GetElements()
	assert.Equal(t, 1, len(elements), "Expected 1 element, got %d", len(elements))

	staticBlock := expectNodeType[*ast.ClassStaticBlockNode](
		t,
		elements[0],
		ast.ClassStaticBlock,
	)

	statementList := expectNodeType[*ast.StatementListNode](
		t,
		staticBlock.GetBody(),
		ast.StatementList,
	)
	assert.Equal(t, 1, len(statementList.GetChildren()), "Expected 1 statement, got %d", len(statementList.GetChildren()))

	// TODO: Test [+Default] path when that's implemented
}

// Declaration : LexicalDeclaration
func TestLexicalDeclaration(t *testing.T) {
	testLexicalDeclaration := func(keyword string, isConst bool) {
		// Test basic lexical declaration
		lexicalDeclaration := expectScriptValue[*ast.BasicNode](
			t,
			fmt.Sprintf("%s x = 1;", keyword),
			ast.LexicalDeclaration,
		)

		// Check binding
		lexicalBinding := expectNodeType[*ast.LexicalBindingNode](
			t,
			lexicalDeclaration.GetChildren()[0],
			ast.LexicalBinding,
		)
		assert.Equal(t, isConst, lexicalBinding.Const, "Expected %s binding, got %s", keyword, map[bool]string{true: "const", false: "let"}[!isConst])

		// Check target
		bindingIdentifier := expectNodeType[*ast.BindingIdentifierNode](
			t,
			lexicalBinding.GetTarget(),
			ast.BindingIdentifier,
		)
		assert.Equal(t, "x", bindingIdentifier.Identifier, "Expected binding identifier 'x', got %s", bindingIdentifier.Identifier)

		// Check initializer
		initializer := expectNodeType[*ast.BasicNode](
			t,
			lexicalBinding.GetInitializer(),
			ast.Initializer,
		)
		numericLiteral := expectNodeType[*ast.NumericLiteralNode](
			t,
			initializer.GetChildren()[0],
			ast.NumericLiteral,
		)
		assert.Equal(t, float64(1), numericLiteral.Value, "Expected initializer value 1, got %f", numericLiteral.Value)

		// Test multiple bindings
		lexicalDeclaration = expectScriptValue[*ast.BasicNode](
			t,
			fmt.Sprintf("%s x = 1, y = 2;", keyword),
			ast.LexicalDeclaration,
		)

		// Check first binding
		lexicalBinding = expectNodeType[*ast.LexicalBindingNode](
			t,
			lexicalDeclaration.GetChildren()[0],
			ast.LexicalBinding,
		)
		assert.Equal(t, isConst, lexicalBinding.Const, "Expected %s binding, got %s", keyword, map[bool]string{true: "const", false: "let"}[!isConst])

		bindingIdentifier = expectNodeType[*ast.BindingIdentifierNode](
			t,
			lexicalBinding.GetTarget(),
			ast.BindingIdentifier,
		)
		assert.Equal(t, "x", bindingIdentifier.Identifier, "Expected binding identifier 'x', got %s", bindingIdentifier.Identifier)

		initializer = expectNodeType[*ast.BasicNode](
			t,
			lexicalBinding.GetInitializer(),
			ast.Initializer,
		)
		numericLiteral = expectNodeType[*ast.NumericLiteralNode](
			t,
			initializer.GetChildren()[0],
			ast.NumericLiteral,
		)
		assert.Equal(t, float64(1), numericLiteral.Value, "Expected initializer value 1, got %f", numericLiteral.Value)

		// Check second binding
		lexicalBinding = expectNodeType[*ast.LexicalBindingNode](
			t,
			lexicalDeclaration.GetChildren()[1],
			ast.LexicalBinding,
		)
		assert.Equal(t, isConst, lexicalBinding.Const, "Expected %s binding, got %s", keyword, map[bool]string{true: "const", false: "let"}[!isConst])

		bindingIdentifier = expectNodeType[*ast.BindingIdentifierNode](
			t,
			lexicalBinding.GetTarget(),
			ast.BindingIdentifier,
		)
		assert.Equal(t, "y", bindingIdentifier.Identifier, "Expected binding identifier 'y', got %s", bindingIdentifier.Identifier)

		initializer = expectNodeType[*ast.BasicNode](
			t,
			lexicalBinding.GetInitializer(),
			ast.Initializer,
		)
		numericLiteral = expectNodeType[*ast.NumericLiteralNode](
			t,
			initializer.GetChildren()[0],
			ast.NumericLiteral,
		)
		assert.Equal(t, float64(2), numericLiteral.Value, "Expected initializer value 2, got %f", numericLiteral.Value)
	}

	testLexicalDeclaration("let", false)
	testLexicalDeclaration("const", true)
}
