package rtda

import (
	"strings"

	"github.com/zxh0/jvm.go/rtda/heap"
)

var (
	_invokeNativeIReturn = []byte{0xfe, 0xac}
	_invokeNativeLReturn = []byte{0xfe, 0xad}
	_invokeNativeFReturn = []byte{0xfe, 0xae}
	_invokeNativeDReturn = []byte{0xfe, 0xaf}
	_invokeNativeAReturn = []byte{0xfe, 0xb0}
	_invokeNativeReturn  = []byte{0xfe, 0xb1}
)

// newNativeFrame 创建一个代表 naive 方法的帧
func newNativeFrame(thread *Thread, method *heap.Method) *Frame {
	frame := &Frame{}
	frame.Thread = thread
	frame.Method = method
	// 创建本地变量表，大小为参数格式，naive 方法总是 static 的
	frame.LocalVars = newLocalVars(method.ParamSlotCount) // todo
	// 创建大小=4操作栈
	frame.OperandStack = newOperandStack(4)               // todo

	if method.Code == nil {
		method.Code = getHackCode(method.Descriptor) // hack!
	}

	return frame
}

func getHackCode(methodDescriptor string) []byte {
	rParenIdx := strings.IndexByte(methodDescriptor, ')')
	switch methodDescriptor[rParenIdx+1] {
	case 'V':
		return _invokeNativeReturn
	case 'L', '[':
		return _invokeNativeAReturn
	case 'D':
		return _invokeNativeDReturn
	case 'F':
		return _invokeNativeFReturn
	case 'J':
		return _invokeNativeLReturn
	default:
		return _invokeNativeIReturn
	}
}
