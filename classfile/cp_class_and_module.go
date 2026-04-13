package classfile

/*
CONSTANT_Class_info {
    u1 tag;
    u2 name_index;
}

CONSTANT_Module_info {
    u1 tag;
    u2 name_index;
}

CONSTANT_Package_info {
    u1 tag;
    u2 name_index;
}
*/

type ConstantClassInfo constantWithNameIdx
type ConstantModuleInfo constantWithNameIdx
type ConstantPackageInfo constantWithNameIdx

type constantWithNameIdx struct {
	NameIndex uint16
}

func readConstantClassInfo(reader *ClassReader) ConstantClassInfo {
	// 这是 class 信息
	return ConstantClassInfo(readConstantWithNameIdx(reader))
}
func readConstantModuleInfo(reader *ClassReader) ConstantModuleInfo {
	// 这是 module 信息
	return ConstantModuleInfo(readConstantWithNameIdx(reader))
}
func readConstantPackageInfo(reader *ClassReader) ConstantPackageInfo {
	// 这是 package 信息
	return ConstantPackageInfo(readConstantWithNameIdx(reader))
}

func readConstantWithNameIdx(reader *ClassReader) constantWithNameIdx {
	return constantWithNameIdx{
		// 读取 uint16 2字节的内容
		NameIndex: reader.ReadUint16(),
	}
}
