package classfile

/*
	CONSTANT_NameAndType_info {
	    u1 tag;
	    u2 name_index;
	    u2 descriptor_index;
	}
*/
type ConstantNameAndTypeInfo struct {
	NameIndex       uint16
	DescriptorIndex uint16
}

func readConstantNameAndTypeInfo(reader *ClassReader) ConstantNameAndTypeInfo {
	return ConstantNameAndTypeInfo{
		// 读取 uint16 2字节的内容，作为 name 索引
		NameIndex: reader.ReadUint16(),
		// 读取 uint16 2字节的内容，作为 descriptor 索引
		DescriptorIndex: reader.ReadUint16(),
	}
}
