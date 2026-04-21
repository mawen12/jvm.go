package heap

import (
	"math"
)

// 代表 null 的 slot
var EmptySlot = Slot{0, nil}

type Slot struct {
	Val int64 // big enough to hold any primitive value
	// 保存引用对象
	Ref *Object
}

// NewIntSlot 创建基础类型 int 的 slot
func NewIntSlot(n int32) Slot {
	return Slot{Val: int64(n)}
}

// NewLongSlot 创建基础类型 long 的 slot
func NewLongSlot(n int64) Slot {
	return Slot{Val: n}
}

// NewFloatSlot 创建基础类型 float32 的 slot
func NewFloatSlot(n float32) Slot {
	return Slot{Val: int64(math.Float32bits(n))}
}

// NewDoubleSlot 创建基础类型 double 的 slot
func NewDoubleSlot(n float64) Slot {
	return Slot{Val: int64(math.Float64bits(n))}
}

// NewRefSlot 创建引用对象的 slot
func NewRefSlot(ref *Object) Slot {
	return Slot{Ref: ref}
}

// IntValue 读取 int 类型的 slot 值
func (slot Slot) IntValue() int32 {
	return int32(slot.Val)
}

// LongValue 读取 long 类型的 slot 值
func (slot Slot) LongValue() int64 {
	return slot.Val
}

// FloatValue 读取 float 类型的 slot 值
func (slot Slot) FloatValue() float32 {
	return math.Float32frombits(uint32(slot.Val))
}

// DoubleValue 读取 double 类型的 slot 值
func (slot Slot) DoubleValue() float64 {
	return math.Float64frombits(uint64(slot.Val))
}

// TODO
func NewHackSlot(x interface{}) Slot {
	return NewRefSlot(&Object{Extra: x})
}
func (slot Slot) GetHack() interface{} {
	return slot.Ref.Extra
}
