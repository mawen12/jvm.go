package classfile

/*
	CONSTANT_InvokeDynamic_info {
	    u1 tag;
	    u2 bootstrap_method_attr_index;
	    u2 name_and_type_index;
	}
*/
type ConstantInvokeDynamicInfo struct {
	BootstrapMethodAttrIndex uint16
	NameAndTypeIndex         uint16
}

func readConstantInvokeDynamicInfo(reader *ClassReader) ConstantInvokeDynamicInfo {
	return ConstantInvokeDynamicInfo{
		// 读取 uint16 2字节的内容，作为 bootstrap_method_attr 索引
		BootstrapMethodAttrIndex: reader.ReadUint16(),
		// 读取 uint16 2字节的内容，作为 name_and_type 索引
		NameAndTypeIndex: reader.ReadUint16(),
	}
}

/*
	CONSTANT_MethodHandle_info {
	    u1 tag;
	    u1 reference_kind;
	    u2 reference_index;
	}
*/
type ConstantMethodHandleInfo struct {
	ReferenceKind  uint8
	ReferenceIndex uint16
}

func readConstantMethodHandleInfo(reader *ClassReader) ConstantMethodHandleInfo {
	return ConstantMethodHandleInfo{
		// 读取 uint8 1字节的内容，作为 reference_kind
		ReferenceKind: reader.ReadUint8(),
		// 读取 uint8 1字节的内容，作为 reference 索引
		ReferenceIndex: reader.ReadUint16(),
	}
}

/*
	CONSTANT_MethodType_info {
	    u1 tag;
	    u2 descriptor_index;
	}
*/
type ConstantMethodTypeInfo struct {
	DescriptorIndex uint16
}

func readConstantMethodTypeInfo(reader *ClassReader) ConstantMethodTypeInfo {
	return ConstantMethodTypeInfo{
		// // 读取 uint16 2字节的内容，作为 descriptor 索引
		DescriptorIndex: reader.ReadUint16(),
	}
}
