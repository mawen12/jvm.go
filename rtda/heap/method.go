package heap

import (
	"strings"

	"github.com/zxh0/jvm.go/classfile"
)

const (
	// main 方法名
	mainMethodName = "main"
	// main 方法描述，参数是 []string，且无返回值
	mainMethodDesc   = "([Ljava/lang/String;)V"
	clinitMethodName = "<clinit>"
	clinitMethodDesc = "()V"
	constructorName  = "<init>"
)

type MethodData struct {
	MaxStack                uint
	MaxLocals               uint
	Code                    []byte
	exceptionTable          []classfile.ExceptionTableEntry
	lineNumberTable         []classfile.LineNumberTableEntry
	ParameterAnnotationData []byte // RuntimeVisibleParameterAnnotations_attribute
	AnnotationDefaultData   []byte // AnnotationDefault_attribute
}

/*
代表 Java 对象中的方法

	instance initialization method
	class method
	interface initialization method

format:

	method_info {
		u2	access_flags;
		u2	name_index; // 方法名称在类的常量池中的索引
		u2	descriptor_index; // 方法描述在类的常量池中的索引
		u2	attributes_count;
		attribute_info
	}

该 type 从 MemberInfo 解析而来。

详见：https://docs.oracle.com/javase/specs/jvms/se8/html/jvms-4.html#jvms-4.6

ParamCount

	参数数量

ParamSlotCount

	用于保存参数值的 slot 数量

	
*/
type Method struct {
	ClassMember
	MethodData
	MethodDescriptor
	ParamCount     uint
	ParamSlotCount uint
	Slot           uint
	exIndexTable   []uint16    // TODO: rename
	Instructions   interface{} // []instructions.Instruction
}

func newMethod(class *Class, cf *classfile.ClassFile, cfMember classfile.MemberInfo, slot uint) *Method {
	method := &Method{}
	method.Class = class
	method.Slot = slot
	method.copyMemberData(cf, cfMember)
	method.copyAttributes(cfMember)
	method.parseDescriptor()
	return method
}

func (method *Method) copyAttributes(cfMember classfile.MemberInfo) {
	if codeAttr, found := cfMember.GetCodeAttribute(); found {
		method.exIndexTable = cfMember.GetExceptionIndexTable()
		method.MaxStack = uint(codeAttr.MaxStack)
		method.MaxLocals = uint(codeAttr.MaxLocals)
		method.Code = codeAttr.Code
		method.exceptionTable = codeAttr.ExceptionTable
		method.lineNumberTable = codeAttr.GetLineNumberTable()
	}
	method.ParameterAnnotationData = cfMember.GetRuntimeVisibleParameterAnnotationsAttributeData()
	method.AnnotationDefaultData = cfMember.GetAnnotationDefaultAttributeData()
}

func (method *Method) parseDescriptor() {
	method.MethodDescriptor = parseMethodDescriptor(method.Descriptor)
	method.ParamCount = uint(len(method.ParameterTypes))
	method.ParamSlotCount = method.getParamSlotCount()
	if !method.IsStatic() {
		method.ParamSlotCount++
	}
}

// IsVoidReturnType 检查方法是否无返回值
func (method *Method) IsVoidReturnType() bool {
	// 方法描述符以 )V 结尾
	return strings.HasSuffix(method.Descriptor, ")V")
}

// IsConstructor 检查是否为实例构造器方法
func (method *Method) IsConstructor() bool {
	// 非静态方法，且方法名为 <init>
	return !method.IsStatic() && method.Name == constructorName
}

// IsClinit 检查是否类构造构造器方法
func (method *Method) IsClinit() bool {
	// 静态，且方法名为 <clinit> 且方法描述符为 ()V
	return method.IsStatic() &&
		method.Name == clinitMethodName &&
		method.Descriptor == clinitMethodDesc
}

// IsRegisterNatives 是否为 registerNatives 方法
func (method *Method) IsRegisterNatives() bool {
	// 静态，且方法名为 registerNatives 且方法描述符为 ()V
	return method.IsStatic() &&
		method.Name == "registerNatives" &&
		method.Descriptor == "()V"
}

// IsInitIDs 是否为 initIDs 方法
func (method *Method) IsInitIDs() bool {
	// 静态，且方法名为 initIDs 且方法描述符为 ()V
	return method.IsStatic() &&
		method.Name == "initIDs" &&
		method.Descriptor == "()V"
}

func (method *Method) FindExceptionHandler(exClass *Class, pc int) int {
	for _, handler := range method.exceptionTable {
		// jvms: The start_pc is inclusive and end_pc is exclusive
		if pc >= int(handler.StartPc) && pc < int(handler.EndPc) {
			if handler.CatchType == 0 {
				// catch all
				return int(handler.HandlerPc)
			}

			catchType := method.Class.GetConstantClass(uint(handler.CatchType))
			if catchType.GetClass() == exClass ||
				catchType.GetClass().isSuperClassOf(exClass) {

				return int(handler.HandlerPc)
			}
		}
	}
	return -1
}

func (method *Method) GetLineNumber(pc int) int {
	if method.IsNative() {
		return -2
	}
	for i := len(method.lineNumberTable) - 1; i >= 0; i-- {
		entry := method.lineNumberTable[i]
		if pc >= int(entry.StartPC) {
			return int(entry.LineNumber)
		}
	}
	return -1
}
