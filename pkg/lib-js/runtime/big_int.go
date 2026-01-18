package runtime

import (
	"math"
	"math/big"
)

type BigInt struct {
	Value *big.Int
}

func NewBigIntValue(value *big.Int) *JavaScriptValue {
	return NewJavaScriptValue(TypeBigInt, &BigInt{
		Value: value,
	})
}

func NumberToBigInt(runtime *Runtime, value *Number) *Completion {
	if value.NaN {
		return NewThrowCompletion(NewRangeError(runtime, "Cannot convert NaN to a BigInt"))
	}

	if math.Floor(value.Value) != value.Value {
		return NewThrowCompletion(NewRangeError(runtime, "Cannot convert non-integer number to a BigInt"))
	}

	return NewNormalCompletion(NewBigIntValue(big.NewInt(int64(value.Value))))
}

func StringToBigInt(runtime *Runtime, value *String) *Completion {
	panic("TODO: Implement String to BigInt conversion.")
}
