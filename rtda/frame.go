package rtda

import (
	"github.com/zxh0/jvm.go/rtda/heap"
)

type OnPopAction func(popped *Frame)

/**
 * stack frame 栈帧
 *
 * 在线程调用方法时创建，为线程私有
 *
 * 详见：https://docs.oracle.com/javase/specs/jvms/se8/html/jvms-2.html#jvms-2.6
 */

/*
LocalVars

	本地变量表，帧的核心

OperandStack

	操作数栈，帧的核心

lower

	线程栈被实现为链表，底层通过帧来指向上一个帧来实现的

Thread

	该帧所属的线程

Method

	该帧所对应的方法

maxLocals

	本地变量表的最大长度

maxStack

	操作数栈的最大深度

NextPC

	下一个程序计数器

onPopActions

	在帧从线程栈中弹出时触发的回调，
	在帧注册到线程栈中时设置，对应到开始调用方法时，
	比如 invokeSpecial/invokeVirtual 等触发 Thread#invokeMethod 时，开始创建新的帧
*/
type Frame struct {
	// 本地变量表，保存操作该帧需要的本地变量
	LocalVars
	// 操作栈，是一个后进先出的队列
	OperandStack
	// 由于 frame 以链表实现，因此其指向下一帧
	lower *Frame // stack is implemented as linked list
	// 所属的线程
	Thread *Thread
	// 所属的方法
	Method *heap.Method
	// 最大本地变量表
	maxLocals uint
	// 最大
	maxStack uint
	// 该帧调用后的下一个指令
	NextPC       int // the next instruction after the call
	onPopActions []OnPopAction
}

// TODO
func NewFrame(maxLocals, maxStack int) *Frame {
	return &Frame{
		LocalVars:    newLocalVars(uint(maxLocals)),
		OperandStack: newOperandStack(uint(maxStack)),
	}
}

// newFrame 创建新的帧
func newFrame(thread *Thread, method *heap.Method) *Frame {
	return &Frame{
		Thread:    thread,           // 所属线程
		Method:    method,           // 关联方法
		maxLocals: method.MaxLocals, // 本地变量表最大大小
		maxStack:  method.MaxStack,  // 操作数栈最大大小
		// 创建对应大小的本地变量表
		LocalVars: newLocalVars(method.MaxLocals),
		// 创建对应大小的方法的操作栈
		OperandStack: newOperandStack(method.MaxStack),
	}
}

func (frame *Frame) reset(method *heap.Method) {
	frame.Method = method
	frame.NextPC = 0
	frame.lower = nil
	frame.onPopActions = nil
	if frame.maxLocals > 0 {
		frame.clearLocalVars()
	}
	if frame.maxStack > 0 {
		frame.ClearStack()
	}
}

func (frame *Frame) RevertNextPC() {
	frame.NextPC = frame.Thread.PC
}
func (frame *Frame) AppendOnPopAction(action OnPopAction) {
	frame.onPopActions = append(frame.onPopActions, action)
}

func (frame *Frame) Load(idx uint, isLongOrDouble bool) {
	slot := frame.GetLocalVar(idx)
	frame.Push(slot)
	if isLongOrDouble {
		frame.PushNull()
	}
}
func (frame *Frame) Store(idx uint, isLongOrDouble bool) {
	if isLongOrDouble {
		frame.Pop()
	}
	slot := frame.Pop()
	frame.SetLocalVar(idx, slot)
}

// shortcuts
func (frame *Frame) GetRuntime() *heap.Runtime {
	return frame.Thread.Runtime // TODO
}
func (frame *Frame) GetBootLoader() *heap.ClassLoader {
	return frame.Thread.Runtime.BootLoader()
}
func (frame *Frame) GetClass() *heap.Class {
	return frame.Method.Class
}
func (frame *Frame) GetConstantPool() heap.ConstantPool {
	return frame.Method.Class.ConstantPool
}

// todo
func (frame *Frame) GetClassLoader() *heap.ClassLoader {
	return frame.Thread.Runtime.BootLoader()
}
