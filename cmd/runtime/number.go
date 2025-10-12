package runtime

import "math"

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
