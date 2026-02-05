package domain

type Value struct {
	val float64
}

func NewValue(v float64) Value {
	return Value{val: v}
}

func (v Value) Float64() float64 {
	return v.val
}
