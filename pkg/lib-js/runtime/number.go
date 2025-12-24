package runtime

import (
	"math"
	"strconv"
)

type Number struct {
	Value float64
	NaN   bool
}

func NewNumberValue(value float64, nan bool) *JavaScriptValue {
	return NewJavaScriptValue(TypeNumber, &Number{
		Value: value,
		NaN:   nan,
	})
}

func NewNaNNumberValue() *JavaScriptValue {
	return NewNumberValue(0, true)
}

func NumberOp(left *Number, right *Number, op func(float64, float64) float64) *Number {
	if left.NaN || right.NaN {
		return &Number{
			Value: 0,
			NaN:   true,
		}
	}

	result := op(left.Value, right.Value)
	if math.IsNaN(result) {
		return &Number{
			Value: 0,
			NaN:   true,
		}
	}

	return &Number{
		Value: result,
		NaN:   false,
	}
}

func NumberAdd(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return a + b
	})
}

func NumberSub(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return a - b
	})
}

func NumberMul(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return a * b
	})
}

func NumberDiv(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return a / b
	})
}

func NumberExponentiate(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return math.Pow(a, b)
	})
}

func NumberRemainder(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return math.Mod(a, b)
	})
}

func NumberLeftShift(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		aInt := int(a)
		bInt := uint(b) % 32
		return float64(aInt << bInt)
	})
}

func NumberSignedRightShift(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		aInt := int(a)
		bInt := uint(b) % 32
		return float64(aInt >> bInt)
	})
}

func NumberUnsignedRightShift(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		aInt := uint(a)
		bInt := uint(b) % 32
		return float64(aInt >> bInt)
	})
}

func NumberBitwiseAnd(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return float64(int(a) & int(b))
	})
}

func NumberBitwiseOr(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return float64(int(a) | int(b))
	})
}

func NumberBitwiseXor(left *Number, right *Number) *Number {
	return NumberOp(left, right, func(a, b float64) float64 {
		return float64(int(a) ^ int(b))
	})
}

func NumberLessThan(left *Number, right *Number) *Completion {
	if left.NaN || right.NaN {
		return NewNormalCompletion(NewUndefinedValue())
	}
	return NewNormalCompletion(NewBooleanValue(left.Value < right.Value))
}

func NumberEqual(left *Number, right *Number) *Completion {
	if left.NaN || right.NaN {
		return NewNormalCompletion(NewBooleanValue(false))
	}
	return NewNormalCompletion(NewBooleanValue(left.Value == right.Value))
}

func NumberUnaryMinus(value *Number) *Number {
	if value.NaN {
		return &Number{
			Value: 0,
			NaN:   true,
		}
	}

	return &Number{
		Value: -value.Value,
		NaN:   false,
	}
}

func NumberBitwiseNot(value *Number) *Number {
	return &Number{
		Value: float64(^int(value.Value)),
		NaN:   false,
	}
}

func NumberSameValue(left *Number, right *Number) bool {
	if left.NaN && right.NaN {
		return true
	}

	if left.NaN || right.NaN {
		return false
	}

	return left.Value == right.Value
}

func NumberSameValueZero(left *Number, right *Number) bool {
	if left.NaN && right.NaN {
		return true
	}

	if left.NaN || right.NaN {
		return false
	}

	if math.Abs(left.Value) == 0 && math.Abs(right.Value) == 0 {
		return true
	}

	return left.Value == right.Value
}

func NumberToString(value *Number, radix int) *JavaScriptValue {
	// TODO: Implement Number::toString according to the spec, this is just a placeholder.
	if value.NaN {
		return NewStringValue("NaN")
	}

	valueFloat := value.Value
	valueInt := math.Trunc(valueFloat)

	if valueInt == valueFloat && radix == 10 {
		return NewStringValue(strconv.FormatInt(int64(valueInt), radix))
	}

	panic("TODO: Implement Number::toString for non-integer numbers.")
}
