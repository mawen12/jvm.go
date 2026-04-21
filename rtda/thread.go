package rtda

import (
	"fmt"
	"strings"
	"sync"

	"github.com/zxh0/jvm.go/rtda/heap"
	"github.com/zxh0/jvm.go/vm"
)

/*
JVM

	Thread
	  pc
	  Stack
	    Frame
	      LocalVars
	      OperandStack
*/

/*
PC

	程序计数器

stack

	线程栈，又被称为虚拟机栈，其中保存了已执行和当前正在执行的方法帧

jThread

	堆中的 Java 对象表示

lock

	同步锁

sleepingFlag

interruptedFlag

parkingFlag

unparkedFlag
*/
type Thread struct {
	PC              int // the address of the instruction currently being executed
	stack           *Stack
	frameCache      *FrameCache
	jThread         *heap.Object // java.lang.Thread
	lock            *sync.Mutex  // state lock
	ch              chan int
	sleepingFlag    bool
	interruptedFlag bool
	parkingFlag     bool // used by Unsafe
	unparkedFlag    bool // used by Unsafe
	VMOptions       *vm.Options
	JNIEnv          interface{}
	Runtime         *heap.Runtime
	// todo
}

// NewThread 创建线程
func NewThread(jThread *heap.Object, opts *vm.Options, rt *heap.Runtime) *Thread {
	// 使用参数配置的栈大小创建 stack 实例
	stack := newStack(uint(opts.ThreadStackSize))
	thread := &Thread{
		stack:     stack,
		jThread:   jThread,
		lock:      &sync.Mutex{},
		ch:        make(chan int),
		VMOptions: opts,
		Runtime:   rt,
	}
	thread.frameCache = newFrameCache(thread, 16) // todo
	return thread
}

// getters & setters
func (thread *Thread) JThread() *heap.Object {
	return thread.jThread
}

func (thread *Thread) IsStackEmpty() bool {
	return thread.stack.isEmpty()
}
func (thread *Thread) StackDepth() uint {
	return thread.stack.size
}

func (thread *Thread) CurrentFrame() *Frame {
	return thread.stack.top()
}
func (thread *Thread) TopFrame() *Frame {
	return thread.stack.top()
}
func (thread *Thread) TopFrameN(n uint) *Frame {
	return thread.stack.topN(n)
}

// PushFrame 将 frame 压入 stack
func (thread *Thread) PushFrame(frame *Frame) {
	// 将 frame 压入 stack
	thread.stack.push(frame)
}

// PopFrame 从栈中弹出帧
func (thread *Thread) PopFrame() *Frame {
	// 读取栈顶的帧
	top := thread.stack.pop()
	//
	for _, action := range top.onPopActions {
		action(top) // TODO
	}

	thread.frameCache.returnFrame(top)
	return top
}

// NewFrame 基于方法创建帧
func (thread *Thread) NewFrame(method *heap.Method) *Frame {
	// 检查是否为 native 方法
	if method.IsNative() {
		// 创建 naive 帧
		return newNativeFrame(thread, method)
	} else { // 创建 帧
		return thread.frameCache.borrowFrame(method)
		//return newFrame(thread, method)
	}
}

// InvokeMethodWithShim 调用shim方法，本质上是创建该方法的帧，并将帧压入线程栈中
func (thread *Thread) InvokeMethodWithShim(method *heap.Method, args []heap.Slot) {
	// 创建一个 shim 帧
	shimFrame := newShimFrame(thread, args)
	// 将帧压入线程的 stack
	thread.PushFrame(shimFrame)
	// 调用目标方法
	thread.InvokeMethod(method)
}

// InvokeMethod 调用方法，创建该方法对应的帧，然后将参数从当前的帧的操作数栈传递到新帧的本地变量表中
// 同时对于同步方法，还是使用 monitor 监控
func (thread *Thread) InvokeMethod(method *heap.Method) {
	//thread._logInvoke(thread.stack.size, method)
	// 读取栈顶的帧
	currentFrame := thread.CurrentFrame()
	// 基于方法创建新的帧
	newFrame := thread.NewFrame(method)
	// 将帧压入线程栈中
	thread.PushFrame(newFrame)
	// 将方法参数传递给帧的本地变量表的数组中
	if n := method.ParamSlotCount; n > 0 {
		// 从 from 的操作数栈中将参数值传递给 to 的 localVars
		_passArgs(currentFrame, newFrame, n)
	}

	// 检查是否为同步方法
	if method.IsSynchronized() { // 同步方法
		var monitor *heap.Monitor
		// 是否静态方法
		if method.IsStatic() { // 静态方法，使用类的监视器
			// 读取方法所在的 class 的类对象
			classObj := method.Class.JClass
			monitor = classObj.Monitor
		} else { // 实例方法，使用 this 的监视器
			// 从本地变量表中读取 this 对象
			thisObj := newFrame.GetThis()
			monitor = thisObj.Monitor
		}

		monitor.Enter(thread)
		newFrame.AppendOnPopAction(func(*Frame) {
			monitor.Exit(thread)
		})
	}
}

// _passArgs 从 from 的操作数栈中将参数值传递给 to 的 localVars
func _passArgs(from *Frame, to *Frame, argSlotsCount uint) {
	// 从 from 的操作数栈中读取 n slot
	args := from.PopTops(argSlotsCount)
	// 从 0 开始迭代
	for i := uint(0); i < argSlotsCount; i++ {
		// 将值设置到目标的 lovalVars 上
		to.SetLocalVar(i, args[i])
		// 将 from 的操作数栈上的值置 null
		args[i] = heap.EmptySlot
	}
}
func (thread *Thread) _logInvoke(stackSize uint, method *heap.Method) {
	space := strings.Repeat(" ", int(stackSize))
	className := method.Class.Name

	if method.IsStatic() {
		fmt.Printf("[method]%v thread:%p %v.%v()\n", space, thread, className, method.Name)
	} else {
		fmt.Printf("[method]%v thread:%p %v#%v()\n", space, thread, className, method.Name)
	}
}

func (thread *Thread) InitClass(class *heap.Class) {
	initClass(thread, class)
}

func (thread *Thread) HandleUncaughtException(ex *heap.Object) {
	thread.stack.clear()
	sysClass := thread.Runtime.BootLoader().LoadClass("java/lang/System")
	sysErr := sysClass.GetStaticValue("out", "Ljava/io/PrintStream;").Ref
	printStackTrace := ex.Class.GetInstanceMethod("printStackTrace", "(Ljava/io/PrintStream;)V")

	// call ex.printStackTrace(System.err)
	newFrame := thread.NewFrame(printStackTrace)
	newFrame.SetRefVar(0, ex)
	newFrame.SetRefVar(1, sysErr)
	thread.PushFrame(newFrame)

	//
	// printString := sysErr.Class().GetInstanceMethod("print", "(Ljava/lang/String;)V")
	// newFrame = thread.NewFrame(printString)
	// vars = newFrame.localVars
	// vars.SetRefVar(0, sysErr)
	// vars.SetRefVar(1, JSFromGoStr("Exception in thread \"main\" ", newFrame))
	// thread.PushFrame(newFrame)
}

// hack
func (thread *Thread) HackSetJThread(jThread *heap.Object) {
	thread.jThread = jThread
}
