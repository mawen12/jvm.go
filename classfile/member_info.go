package classfile

/*
	field_info {
	    u2             access_flags;
	    u2             name_index;
	    u2             descriptor_index;
	    u2             attributes_count;
	    attribute_info attributes[attributes_count];
	}

	method_info {
	    u2             access_flags;
	    u2             name_index;
	    u2             descriptor_index;
	    u2             attributes_count;
	    attribute_info attributes[attributes_count];
	}
*/
type MemberInfo struct {
	AccessFlags     uint16
	NameIndex       uint16
	DescriptorIndex uint16
	AttributeTable
}

// read field or method table
func readMembers(reader *ClassReader) []MemberInfo {
	return reader.readTable(readMember).([]MemberInfo)
}

func readMember(reader *ClassReader) MemberInfo {
	return MemberInfo{
		// 读取 uint16 2字节的内容，作为 access_flags
		AccessFlags: reader.ReadUint16(),
		// 读取 uint16 2字节的内容，作为 name 索引
		NameIndex: reader.ReadUint16(),
		// 读取 uint16 2字节的内容，作为 descriptor 索引
		DescriptorIndex: reader.ReadUint16(),
		AttributeTable:  readAttributes(reader),
	}
}
