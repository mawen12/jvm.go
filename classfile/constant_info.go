package classfile

import (
	"fmt"
)

// Constant pool tags
// 常量池标志映射表：
// utf8 -> 1 (JDK 1.0.2)
// integer -> 3 (JDK 1.0.2)
// float -> 4 (JDK 1.0.2)
// long -> 5 (JDK 1.0.2)
// double -> 6 (JDK 1.0.2)
// class -> 7 (JDK 1.0.2)
// string -> 8 (JDK 1.0.2)
// field ref -> 9 (JDK 1.0.2)
// method ref -> 10 (JDK 1.0.2)
// interface method ref -> 11 (JDK 1.0.2)
// name and type -> 12 (JDK 1.0.2)
// method handle -> 13 (JDK 1.7) 方法反射的替代方案
// method type -> 18 (JDK 1.7) 方法反射的辅助
// invoke dynamic -> 19 (JDK 1.7) 方法反射的辅助
// module -> 19 (JDK 9)
// package -> 20 (JDK 9)
// dynamic -> 17 (JDK 11)
const (
	ConstantUtf8               = 1  // Java 1.0.2
	ConstantInteger            = 3  // Java 1.0.2
	ConstantFloat              = 4  // Java 1.0.2
	ConstantLong               = 5  // Java 1.0.2
	ConstantDouble             = 6  // Java 1.0.2
	ConstantClass              = 7  // Java 1.0.2
	ConstantString             = 8  // Java 1.0.2
	ConstantFieldRef           = 9  // Java 1.0.2
	ConstantMethodRef          = 10 // Java 1.0.2
	ConstantInterfaceMethodRef = 11 // Java 1.0.2
	ConstantNameAndType        = 12 // Java 1.0.2
	ConstantMethodHandle       = 15 // Java 7
	ConstantMethodType         = 16 // Java 7
	ConstantInvokeDynamic      = 18 // Java 7
	ConstantModule             = 19 // Java 9
	ConstantPackage            = 20 // Java 9
	ConstantDynamic            = 17 // Java 11
)

/*
	cp_info {
	    u1 tag;
	    u1 info[];
	}
*/
type ConstantInfo interface{}

// readConstantInfo 读取常量池的信息
func readConstantInfo(reader *ClassReader) ConstantInfo {
	// 读取 1字节，u1 的标志位
	tag := reader.ReadUint8()
	switch tag {
	case ConstantInteger:
		return readConstantIntegerInfo(reader)
	case ConstantFloat:
		return readConstantFloatInfo(reader)
	case ConstantLong:
		return readConstantLongInfo(reader)
	case ConstantDouble:
		return readConstantDoubleInfo(reader)
	case ConstantUtf8:
		return readConstantUtf8Info(reader)
	case ConstantString:
		return readConstantStringInfo(reader)
	case ConstantClass:
		return readConstantClassInfo(reader)
	case ConstantModule:
		return readConstantModuleInfo(reader)
	case ConstantPackage:
		return readConstantPackageInfo(reader)
	case ConstantFieldRef:
		return readConstantFieldRefInfo(reader)
	case ConstantMethodRef:
		return readConstantMethodRefInfo(reader)
	case ConstantInterfaceMethodRef:
		return readConstantInterfaceMethodRefInfo(reader)
	case ConstantNameAndType:
		return readConstantNameAndTypeInfo(reader)
	case ConstantMethodType:
		return readConstantMethodTypeInfo(reader)
	case ConstantMethodHandle:
		return readConstantMethodHandleInfo(reader)
	case ConstantInvokeDynamic:
		return readConstantInvokeDynamicInfo(reader)
	default: // TODO
		panic(fmt.Errorf("invalid constant pool tag: %d", tag))
	}
}
