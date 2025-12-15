package runtime

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

type OpCode int

type OpEvaluateCallback func(runtime *Runtime, vm *ExecutionVM) *Completion

type OpEvaluateOperands struct {
	Expression             ast.Node
	NativeCallback         OpEvaluateCallback
	IgnoreAbruptCompletion bool
}

func EmitEvaluateExpression(expression ast.Node) Instruction {
	return Instruction{
		OpCode: OpEvaluate,
		Operand: OpEvaluateOperands{
			Expression: expression,
		},
	}
}

func EmitEvaluateNativeCallback(nativeCallback OpEvaluateCallback) Instruction {
	return Instruction{
		OpCode: OpEvaluate,
		Operand: OpEvaluateOperands{
			NativeCallback: nativeCallback,
		},
	}
}

type OpJumpOperands struct {
	Offset   int
	Absolute bool // Jump to an absolute instruction pointer.
}

func EmitJump(offset int) Instruction {
	return Instruction{
		OpCode: OpJump,
		Operand: OpJumpOperands{
			Offset: offset,
		},
	}
}

func EmitJumpIfTrue(offset int) Instruction {
	return Instruction{
		OpCode: OpJumpIfTrue,
		Operand: OpJumpOperands{
			Offset: offset,
		},
	}
}

func EmitJumpIfFalse(offset int) Instruction {
	return Instruction{
		OpCode: OpJumpIfFalse,
		Operand: OpJumpOperands{
			Offset: offset,
		},
	}
}

type YieldResult func(runtime *Runtime, vm *ExecutionVM) *JavaScriptValue

type OpYieldOperands struct {
	Result YieldResult
}

func EmitYield(result YieldResult) Instruction {
	return Instruction{
		OpCode: OpYield,
		Operand: OpYieldOperands{
			Result: result,
		},
	}
}

const (
	OpNop OpCode = iota

	// Evaluation.
	OpEvaluate // Evaluate an expression and push the result onto the evaluation stack.

	// Control Flow.
	OpJump
	OpJumpIfTrue
	OpJumpIfFalse

	// Suspending Execution.
	OpYield
)

type Instruction struct {
	OpCode  OpCode
	Operand any
}

type InstructionResult struct {
	Completion             *Completion
	Interrupt              bool
	IgnoreAbruptCompletion bool
}

type ExecutionVM struct {
	Instructions       []Instruction
	InstructionPointer int

	// Stack of values resulting from evaluation ops.
	LastEvaluationResult *Completion
	EvaluationStack      []*Completion

	// Scratch space that can be used by native functions to store temporary values.
	ScratchSpace map[string]any
}

func NewExecutionVM() *ExecutionVM {
	return &ExecutionVM{
		Instructions:       make([]Instruction, 0),
		InstructionPointer: 0,
		EvaluationStack:    make([]*Completion, 0),
		ScratchSpace:       make(map[string]any),
	}
}

func (vm *ExecutionVM) PopEvaluationStack() *Completion {
	completion := vm.EvaluationStack[len(vm.EvaluationStack)-1]
	vm.EvaluationStack = vm.EvaluationStack[:len(vm.EvaluationStack)-1]
	return completion
}

func (vm *ExecutionVM) PeekEvaluationStack() *Completion {
	if len(vm.EvaluationStack) == 0 {
		return nil
	}

	return vm.EvaluationStack[len(vm.EvaluationStack)-1]
}

func (vm *ExecutionVM) PushEvaluationStack(completion *Completion) {
	vm.EvaluationStack = append(vm.EvaluationStack, completion)
}

func Compile(runtime *Runtime, node ast.Node) []Instruction {
	if node.GetNodeType() == ast.Script {
		return CompileScript(runtime, node.(*ast.ScriptNode))
	}

	if node.GetNodeType() == ast.StatementList {
		return CompileStatementList(runtime, node.(*ast.StatementListNode))
	}

	if node.GetNodeType() == ast.IfStatement {
		return CompileIfStatement(runtime, node.(*ast.IfStatementNode))
	}

	// Default to evaluating the node as an expression.
	return []Instruction{
		EmitEvaluateExpression(node),
	}
}

func ExecuteVM(runtime *Runtime, vm *ExecutionVM) *Completion {
	for vm.InstructionPointer < len(vm.Instructions) {
		instruction := vm.Instructions[vm.InstructionPointer]
		vm.InstructionPointer++

		result := EvaluateInstruction(runtime, vm, instruction)

		vm.LastEvaluationResult = result.Completion
		if vm.LastEvaluationResult != nil {
			vm.PushEvaluationStack(vm.LastEvaluationResult)
		}

		if result.Interrupt {
			break
		}

		if !result.IgnoreAbruptCompletion && result.Completion != nil && result.Completion.Type != Normal {
			break
		}
	}

	return vm.PopEvaluationStack()
}

func EvaluateInstruction(runtime *Runtime, vm *ExecutionVM, instruction Instruction) InstructionResult {
	switch instruction.OpCode {
	case OpNop:
		return InstructionResult{}
	case OpEvaluate:
		return EvaluateOpEvaluate(runtime, vm, instruction)
	case OpJump:
		return EvaluateOpJump(runtime, vm, instruction)
	case OpJumpIfTrue:
		return EvaluateOpJumpIfTrue(runtime, vm, instruction)
	case OpJumpIfFalse:
		return EvaluateOpJumpIfFalse(runtime, vm, instruction)
	case OpYield:
		return EvaluateOpYield(runtime, vm, instruction)
	}

	panic(fmt.Sprintf("Unknown instruction code: %d", instruction.OpCode))
}

func EvaluateOpEvaluate(runtime *Runtime, vm *ExecutionVM, instruction Instruction) InstructionResult {
	operand := instruction.Operand.(OpEvaluateOperands)

	var completion *Completion
	if operand.Expression != nil {
		completion = Evaluate(runtime, operand.Expression)
	} else if operand.NativeCallback != nil {
		completion = operand.NativeCallback(runtime, vm)
	} else {
		panic("Assert failed: Invalid evaluate operand.")
	}

	return InstructionResult{
		Completion:             completion,
		IgnoreAbruptCompletion: operand.IgnoreAbruptCompletion,
	}
}

func EvaluateOpJump(runtime *Runtime, vm *ExecutionVM, instruction Instruction) InstructionResult {
	operand := instruction.Operand.(OpJumpOperands)
	offset := operand.Offset

	if operand.Absolute {
		vm.InstructionPointer = offset
	} else {
		vm.InstructionPointer += offset
	}

	return InstructionResult{}
}

func EvaluateOpJumpIfTrue(runtime *Runtime, vm *ExecutionVM, instruction Instruction) InstructionResult {
	operand := instruction.Operand.(OpJumpOperands)
	offset := operand.Offset

	completion := vm.PopEvaluationStack()
	if completion.Type != Normal {
		return InstructionResult{
			Completion: completion,
		}
	}

	boolValue := completion.Value.(*JavaScriptValue)
	if boolValue.Type != TypeBoolean {
		panic("Assert failed: OpJumpIfFalse expects the last completion to be a boolean.")
	}

	if boolValue.Value.(*Boolean).Value {
		if operand.Absolute {
			vm.InstructionPointer = offset
		} else {
			vm.InstructionPointer += offset
		}
	}

	return InstructionResult{}
}

func EvaluateOpJumpIfFalse(runtime *Runtime, vm *ExecutionVM, instruction Instruction) InstructionResult {
	operand := instruction.Operand.(OpJumpOperands)
	offset := operand.Offset

	completion := vm.PopEvaluationStack()
	if completion.Type != Normal {
		return InstructionResult{
			Completion: completion,
		}
	}

	boolValue := completion.Value.(*JavaScriptValue)
	if boolValue.Type != TypeBoolean {
		panic("Assert failed: OpJumpIfFalse expects the last completion to be a boolean.")
	}

	if !boolValue.Value.(*Boolean).Value {
		if operand.Absolute {
			vm.InstructionPointer = offset
		} else {
			vm.InstructionPointer += offset
		}
	}

	return InstructionResult{}
}

// Instruction should have the semantics of GeneratorYield in the spec.
func EvaluateOpYield(runtime *Runtime, vm *ExecutionVM, instruction Instruction) InstructionResult {
	operand := instruction.Operand.(OpYieldOperands)
	result := operand.Result(runtime, vm)

	genContext := runtime.GetRunningExecutionContext()
	generator := genContext.Generator

	generator.GeneratorState = GeneratorStateSuspendedYield

	// Remove the generator's context from the execution context stack.
	runtime.PopExecutionContext()

	// Interrupt the current execution loop and return the result.
	return InstructionResult{
		Completion: NewNormalCompletion(result),
		Interrupt:  true,
	}
}
