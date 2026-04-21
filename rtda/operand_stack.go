package rtda

import (
	"github.com/zxh0/jvm.go/rtda/heap"
)

/**
 * 操作数栈，被帧持有，其中保存了
 *
 * 详见：https://docs.oracle.com/javase/specs/jvms/se8/html/jvms-2.html#jvms-2.6.2
 */
type OperandStack struct {
	size  uint // TODO: change to int
	slots []heap.Slot
}

// newOperandStackWithSlots 使用 slots 创建一个操作数栈
func newOperandStackWithSlots(slots []heap.Slot) OperandStack {
	return OperandStack{
		size:  uint(len(slots)),
		slots: slots,
	}
}

// newOperandStack 使用指定大小创建操作数栈
func newOperandStack(size uint) OperandStack {
	var slots []heap.Slot = nil
	if size > 0 {
		slots = make([]heap.Slot, size)
	}
	return OperandStack{size: 0, slots: slots}
}

// IsStackEmpty 检查操作数栈是否为空
func (stack *OperandStack) IsStackEmpty() bool {
	return stack.size == 0
}

// PushNull 向操作数栈中推送一个空 Slot
func (stack *OperandStack) PushNull() {
	stack.Push(heap.EmptySlot)
}

// PushBoolean 向操作数栈中推送一个 bool 
func (stack *OperandStack) PushBoolean(val bool) {
	// 上层的 bool，底层使用 int 表示
	// 1: true, 0: false
	if val {
		stack.PushInt(1)
	} else {
		stack.PushInt(0)
	}
}

// PopBoolean 从操作数栈取出栈顶的 bool 值
func (stack *OperandStack) PopBoolean() bool {
	// 由于 bool 底层使用 int，因此采用 PopInt
	return stack.PopInt() == 1
}

// PushInt 向操作数栈中推送一个 int
func (stack *OperandStack) PushInt(val int32) {
	// 创建 int slot
	stack.Push(heap.NewIntSlot(val))
}

// PopInt 从操作数栈取出栈顶的 int 值
func (stack *OperandStack) PopInt() int32 {
	return stack.Pop().IntValue()
}

// long consumes two slots
// PushLong 向操作数栈中推送一个 long
func (stack *OperandStack) PushLong(val int64) {
	// 创建 long slot
	stack.Push(heap.NewLongSlot(val))
	// long 是 64 位的，消耗两个 slot
	stack.size++
}

// PopLong 从操作数栈取出栈顶的 long 值
func (stack *OperandStack) PopLong() int64 {
	// long 是 64 位的，消耗两个 slot
	stack.size--
	return stack.Pop().LongValue()
}

// PushFloat 向操作数栈中推送一个 float
func (stack *OperandStack) PushFloat(val float32) {
	// 创建 float slot
	stack.Push(heap.NewFloatSlot(val))
}

// PopFloat 从操作数栈取出栈顶的 float 值
func (stack *OperandStack) PopFloat() float32 {
	return stack.Pop().FloatValue()
}

// double consumes two slots
// PushDouble 向操作数栈中推送一个 double
func (stack *OperandStack) PushDouble(val float64) {
	// 创建 double slot
	stack.Push(heap.NewDoubleSlot(val))
	// double 是 64 位的，消耗两个 slot
	stack.size++
}

// PopDouble 从操作数栈取出栈顶的 double 值
func (stack *OperandStack) PopDouble() float64 {
	// double 是 64 位的，消耗两个 slot
	stack.size--
	return stack.Pop().DoubleValue()
}

// PushRef 向操作数栈中推送一个 应用对象
func (stack *OperandStack) PushRef(ref *heap.Object) {
	// 创建 ref slot
	stack.Push(heap.NewRefSlot(ref))
}

// PopRef 从操作数栈取出栈顶的 ref 值
func (stack *OperandStack) PopRef() *heap.Object {
	return stack.Pop().Ref
}

// Push 向操作数栈中推送一个 slot
func (stack *OperandStack) Push(slot heap.Slot) {
	// 将 slot 放到指定的 size
	stack.slots[stack.size] = slot
	// 将 size + 1
	stack.size++
}

// Pop 从操作数栈中取出一个 slot
func (stack *OperandStack) Pop() heap.Slot {
	// 将 size -1
	stack.size--
	// 读取指定的 size 的 slot
	top := stack.slots[stack.size]
	// 将该位置的 slot 置为 null，帮助 GC
	stack.slots[stack.size] = heap.EmptySlot // help GC
	return top
}

// PushL 向操作数栈中推送一个 slot，并指定值长度是否为 64 位
func (stack *OperandStack) PushL(slot heap.Slot, isLongOrDouble bool) {
	stack.Push(slot)
	// 如果是 64 位，则需要占用两个 slot
	if isLongOrDouble {
		stack.size++
	}
}

// PopL 从操作数栈中取出一个 slot，并指定值长度是否为 64 位
func (stack *OperandStack) PopL(isLongOrDouble bool) heap.Slot {
	// 如果是 64 位，则需要占用两个 slot
	if isLongOrDouble {
		stack.size--
	}
	return stack.Pop()
}

// PopTops 弹出 n 个 slot，每个都是 32 位
func (stack *OperandStack) PopTops(n uint) []heap.Slot {
	// 计算弹出后的 size
	start := stack.size - n
	end := stack.size
	// 读取该区间内的 slots
	top := stack.slots[start:end]
	// 计算弹出后的 size
	stack.size -= n
	return top
}

// TopRef 弹出第 n 个的引用对象的 slot
func (stack *OperandStack) TopRef(n uint) *heap.Object {
	return stack.slots[stack.size-1-n].Ref
}

// ClearStack 清理操作数栈
func (stack *OperandStack) ClearStack() {
	// 大小置为0
	stack.size = 0
	for i := range stack.slots {
		// 每个 slot 都置为 null
		stack.slots[i] = heap.EmptySlot
	}
}

// only used by native methods
// HackSetSlots
func (stack *OperandStack) HackSetSlots(slots []heap.Slot) { // TODO
	stack.slots = slots
	stack.size = uint(len(slots))
}

// DebugGetSlots 返回当前的 stack 中所有的 slots，用于 debug 场景
func (stack *OperandStack) DebugGetSlots() []heap.Slot {
	return stack.slots[:stack.size]
}
