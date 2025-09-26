package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"zbrannelly.dev/go-js/cmd/parser/ast"
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
	assert.Nil(t, classExpression.GetName(), "Expected anonymous class, but got a named class")
	assert.Nil(t, classExpression.GetHeritage(), "Expected no heritage, but got heritage")

	// Test named class expression without heritage
	classExpression = expectScriptValue[*ast.ClassExpressionNode](
		t,
		"(class Foo {});",
		ast.ClassExpression,
	)
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
	templateLiteral := expectScriptValue[*ast.BasicNode](
		t,
		"`simple template`;",
		ast.TemplateLiteral,
	)
	assert.Equal(t, 1, len(templateLiteral.Children), "Expected 1 child, got %d", len(templateLiteral.Children))
	stringLiteral := expectNodeType[*ast.StringLiteralNode](t, templateLiteral.Children[0], ast.StringLiteral)
	assert.Equal(t, "simple template", stringLiteral.Value, "Expected value 'simple template', got %s", stringLiteral.Value)

	// Test template literal with substitutions
	templateLiteral = expectScriptValue[*ast.BasicNode](
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
