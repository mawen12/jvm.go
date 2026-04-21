package references

import (
	"github.com/zxh0/jvm.go/instructions/base"
	"github.com/zxh0/jvm.go/rtda"
	"github.com/zxh0/jvm.go/rtda/heap"
)

// Invoke instance method;
// special handling for superclass, private, and instance initialization method invocations

/*

	invoke instance method



	详见：https://docs.oracle.com/javase/specs/jvms/se8/html/jvms-4.html#jvms-4.10.1.9.invokespecial
*/
type InvokeSpecial struct{ base.Index16Instruction }

func (instr *InvokeSpecial) Execute(frame *rtda.Frame) {
	// 读取帧所在方法的类的常量池
	cp := frame.GetConstantPool()
	// 从常量池中读取索引对应的常量信息
	k := cp.GetConstant(instr.Index)
	// 检查是否为 MethodRef
	if kMethodRef, ok := k.(*heap.ConstantMethodRef); ok {
		// 
		method := kMethodRef.GetMethod(false)
		frame.Thread.InvokeMethod(method)
	} else {
		method := k.(*heap.ConstantInterfaceMethodRef).GetMethod(false)
		frame.Thread.InvokeMethod(method)
	}
}
