package rtda

import (
	"github.com/zxh0/jvm.go/rtda/heap"
)

/**
 * 本地变量表，其中静态方法是从索引0开始保存，实例方法是从索引1开始，其索引0被用于保存this。
 * 保存内容：
 *   this 当前对象引用，仅在示例方法中
 *   params 方法参数，紧跟 this
 *   local variables 本地变量，紧跟 params
 *
 * 尽管 JVM 规范中说明了 long 和 double 这种64位的类型占用两个连续的slot，但是此处并未实现，
 * 而是讨巧的在底层的 heap.Slot 中使用 int64 ，使得heap.Slot支持64位，但是也代表了占用字节
 * 更多。
 *
 * 详见：https://docs.oracle.com/javase/specs/jvms/se8/html/jvms-2.html#jvms-2.6.1
 */
type LocalVars struct {
	slots []heap.Slot
}

// newLocalVars 创建一个指定大小的本地变量表
func newLocalVars(size uint) LocalVars {
	var slots []heap.Slot = nil
	if size > 0 {
		slots = make([]heap.Slot, size)
	}
	return LocalVars{slots: slots}
}

// GetIntVar 读取指定索引的 int 值
func (lv *LocalVars) GetIntVar(index uint) int32 {
	return lv.GetLocalVar(index).IntValue()
}

// SetIntVar 设置指定索引的 int 值
func (lv *LocalVars) SetIntVar(index uint, val int32) {
	// 底层使用 int slot
	lv.SetLocalVar(index, heap.NewIntSlot(val))
}

// GetLongVar 读取指定索引的 long 值
func (lv *LocalVars) GetLongVar(index uint) int64 {
	return lv.GetLocalVar(index).LongValue()
}

// SetLongVar 设置指定索引的 long 值
func (lv *LocalVars) SetLongVar(index uint, val int64) {
	// 底层使用 long slot
	lv.SetLocalVar(index, heap.NewLongSlot(val))
}

// GetFloatVar 读取指定索引的 float 值
func (lv *LocalVars) GetFloatVar(index uint) float32 {
	return lv.GetLocalVar(index).FloatValue()
}

// SetFloatVar 设置指定索引的 float 值
func (lv *LocalVars) SetFloatVar(index uint, val float32) {
	// 底层使用 float slot
	lv.SetLocalVar(index, heap.NewFloatSlot(val))
}

// GetDoubleVar 读取指定索引的 double 值
func (lv *LocalVars) GetDoubleVar(index uint) float64 {
	return lv.GetLocalVar(index).DoubleValue()
}

// SetDoubleVar 设置指定索引的 double 值
func (lv *LocalVars) SetDoubleVar(index uint, val float64) {
	// 底层使用 double slot
	lv.SetLocalVar(index, heap.NewDoubleSlot(val))
}

// GetRefVar 读取指定索引的引用类型值
func (lv *LocalVars) GetRefVar(index uint) *heap.Object {
	return lv.GetLocalVar(index).Ref
}

// SetRefVar 设置指定索引的引用类型值
func (lv *LocalVars) SetRefVar(index uint, ref *heap.Object) {
	// 底层使用 ref slot
	lv.SetLocalVar(index, heap.NewRefSlot(ref))
}

// GetLocalVar 直接读取指定索引的 slot
func (lv *LocalVars) GetLocalVar(index uint) heap.Slot {
	return lv.slots[index]
}
// SetLocalVar 直接设置指定索引的 slot
func (lv *LocalVars) SetLocalVar(index uint, slot heap.Slot) {
	lv.slots[index] = slot
}

// GetBooleanVar 读取指定索引的 bool 值
func (lv *LocalVars) GetBooleanVar(index uint) bool {
	// bool 底层使用 int 来实现
	return lv.GetIntVar(index) == 1
}

// GetThis 读取索引为0的引用对象值
// 方法仅在示例方法中才可被使用
func (lv *LocalVars) GetThis() *heap.Object {
	return lv.GetRefVar(0)
}

// clearLocalVars 清除本地变量表
func (lv *LocalVars) clearLocalVars() {
	for i := range lv.slots {
		// 将 slot 置为 null
		lv.slots[i] = heap.EmptySlot
	}
}

// DebugGetSlots 返回当前的本地变量中所有的 slots，用于 debug 场景
// 比如 IDEA debugger 界面上指定 Frame 的 Variables，就是 LocalVars
// 其中展示的值，就是 slot
func (lv *LocalVars) DebugGetSlots() []heap.Slot {
	return lv.slots
}
