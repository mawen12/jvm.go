package cpu

import (
	"fmt"

	"github.com/zxh0/jvm.go/instructions"
	"github.com/zxh0/jvm.go/instructions/base"
	"github.com/zxh0/jvm.go/rtda"
	"github.com/zxh0/jvm.go/rtda/heap"
	"github.com/zxh0/jvm.go/vm"
)

// 方法未被使用
func ExecMethod(thread *rtda.Thread, method *heap.Method, args []heap.Slot) heap.Slot {
	shimFrame := rtda.NewShimFrame(thread, args)
	thread.PushFrame(shimFrame)
	thread.InvokeMethod(method)

	debug := thread.VMOptions.XDebugInstr
	defer _catchErr(thread) // todo

	for {
		frame := thread.CurrentFrame()
		if frame == shimFrame {
			thread.PopFrame()
			if frame.IsStackEmpty() {
				return heap.EmptySlot
			} else {
				return frame.Pop()
			}
		}

		pc := frame.NextPC
		thread.PC = pc

		// fetch instruction
		instr, nextPC := fetchInstruction(frame.Method, pc)
		frame.NextPC = nextPC

		// execute instruction
		instr.Execute(frame)
		if debug {
			_logInstruction(frame, instr)
		}
	}
}

// Loop 对目标线程执行循环，直到没有方法可调用为止
func Loop(thread *rtda.Thread) {
	// 读取线程对象
	threadObj := thread.JThread()
	// 通过反射读取 Thread#daemon 字段
	isDaemon := threadObj != nil && threadObj.GetFieldValue("daemon", "Z").IntValue() == 1
	// 检查是否为 daemon
	if !isDaemon {
		// 非 daemon 线程计数+1
		nonDaemonThreadStart()
	}

	// 循环
	_loop(thread)

	// terminate thread
	threadObj = thread.JThread()
	threadObj.Monitor.NotifyAll()
	if !isDaemon {
		nonDaemonThreadStop()
	}
}

func _loop(thread *rtda.Thread) {
	debug := thread.VMOptions.XDebugInstr
	defer _catchErr(thread) // todo

	for {
		// 读取线程栈顶的帧
		frame := thread.CurrentFrame()
		// 读取帧上的程序计数器
		pc := frame.NextPC
		// 更新到线程上
		thread.PC = pc

		// fetch instruction
		// 从方法上读取指定程序计数器对应的指令，并对程序计数器按需增加，返回最新的指令和程序计数器
		instr, nextPC := fetchInstruction(frame.Method, pc)
		// 更新桢的程序计数器
		frame.NextPC = nextPC

		// execute instruction
		// 执行指令
		instr.Execute(frame)

		// 如果设置了 debug，则输出执行的指令和桢信息
		if debug {
			_logInstruction(frame, instr)
		}

		// 如果线程中的虚拟机栈没有可执行的方法，则退出循环
		if thread.IsStackEmpty() {
			break
		}
	}
}

// fetchInstruction 从方法上读取指定程序计数器对应的指令，并对程序计数器按需增加，返回最新的指令和程序计数器
func fetchInstruction(method *heap.Method, pc int) (base.Instruction, int) {

	// 检查方法指令是否已经解析
	if method.Instructions == nil { // 不存在，需要解析
		// TODO by mawen debug
		if method.Name == "main" {
			fmt.Println("Method ", method.Name, method.Descriptor, method.Class.Name)
		}

		// 将字节数组解码为指令数组
		method.Instructions = instructions.Decode(method.Code)
	}

	// 转换为指令接口数组
	instrs := method.Instructions.([]base.Instruction)
	// 根据程序计数器，读取对应位置的指令
	instr := instrs[pc]

	// calc nextPC
	// 程序计数器+1
	pc++
	// 遍历当前指令集，对于空指令的场景自增
	for pc < len(instrs) && instrs[pc] == nil {
		pc++
	}

	return instr, pc
}

// todo
func _catchErr(thread *rtda.Thread) {
	if r := recover(); r != nil {
		if err, ok := r.(vm.ClassNotFoundError); ok {
			thread.ThrowClassNotFoundException(err.Error())
			_loop(thread)
			return
		}

		_logFrames(thread)

		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("%v", r)
			panic(err.Error())
		} else {
			panic(err.Error())
		}
	}
}

func _logFrames(thread *rtda.Thread) {
	for !thread.IsStackEmpty() {
		frame := thread.PopFrame()
		method := frame.Method
		className := method.Class.Name
		lineNum := method.GetLineNumber(frame.NextPC)
		fmt.Printf(">> line:%4d pc:%4d %v.%v%v \n",
			lineNum, frame.NextPC, className, method.Name, method.Descriptor)
	}
}

// _logInstruction 输出执行的指令和桢信息
func _logInstruction(frame *rtda.Frame, instr base.Instruction) {
	thread := frame.Thread
	method := frame.Method
	className := method.Class.Name
	pc := thread.PC

	// 检查是否为静态方法
	if method.IsStatic() { // 静态方法
		fmt.Printf("[instruction] thread:%p %v.%v() #%v %T %v\n",
			thread, className, method.Name, pc, instr, instr)
	} else { // 实例方法
		fmt.Printf("[instruction] thread:%p %v#%v() #%v %T %v\n",
			thread, className, method.Name, pc, instr, instr)
	}
}
