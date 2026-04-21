package cpu

import (
	"fmt"

	"github.com/zxh0/jvm.go/instructions"
	"github.com/zxh0/jvm.go/instructions/base"
	"github.com/zxh0/jvm.go/rtda"
	"github.com/zxh0/jvm.go/rtda/heap"
	"github.com/zxh0/jvm.go/vm"
)

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

func Loop(thread *rtda.Thread) {
	// 读取线程对象
	threadObj := thread.JThread()
	// 通过反射读取 Thread#daemon 字段
	isDaemon := threadObj != nil && threadObj.GetFieldValue("daemon", "Z").IntValue() == 1
	// 检查是否为 daemon
	if !isDaemon {
		// 线程计数+1
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
		// 
		instr, nextPC := fetchInstruction(frame.Method, pc)
		frame.NextPC = nextPC

		// execute instruction
		instr.Execute(frame)
		if debug {
			_logInstruction(frame, instr)
		}
		if thread.IsStackEmpty() {
			break
		}
	}
}

// fetchInstruction 
func fetchInstruction(method *heap.Method, pc int) (base.Instruction, int) {
	// 检查方法指令是否存在
	if method.Instructions == nil {
		// 
		method.Instructions = instructions.Decode(method.Code)
	}

	instrs := method.Instructions.([]base.Instruction)
	instr := instrs[pc]

	// calc nextPC
	pc++
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

func _logInstruction(frame *rtda.Frame, instr base.Instruction) {
	thread := frame.Thread
	method := frame.Method
	className := method.Class.Name
	pc := thread.PC

	if method.IsStatic() {
		fmt.Printf("[instruction] thread:%p %v.%v() #%v %T %v\n",
			thread, className, method.Name, pc, instr, instr)
	} else {
		fmt.Printf("[instruction] thread:%p %v#%v() #%v %T %v\n",
			thread, className, method.Name, pc, instr, instr)
	}
}
