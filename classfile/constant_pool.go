package classfile

// readConstantPool 使用长度前缀读取常量池
func readConstantPool(reader *ClassReader) []ConstantInfo {
	// 读取常量池的数量，使用 uint16（2字节)存储，取出时转换为4字节的int
	cpCount := int(reader.ReadUint16())
	// 构造对应的数据
	cp := make([]ConstantInfo, cpCount)

	// The constant_pool table is indexed from 1 to constant_pool_count - 1.
	// 常量池表的索引从 1 到 count - 1，而非从0开始
	for i := 1; i < cpCount; i++ {
		// 循环读取常量池的信息
		cp[i] = readConstantInfo(reader)
		// http://docs.oracle.com/javase/specs/jvms/se8/html/jvms-4.html#jvms-4.4.5
		// All 8-byte constants take up two entries in the constant_pool table of the class file.
		// If a CONSTANT_Long_info or CONSTANT_Double_info structure is the item in the constant_pool
		// table at index n, then the next usable item in the pool is located at index n+2.
		// The constant_pool index n+1 must be valid but is considered unusable.
		switch cp[i].(type) {
		// 对于 64 位，会使用两个位置保存，这在数据库的设计中也是如此。
		// 因为常量池的每个元素大小固定为 32-bit，即 4字节。
		case int64, float64:
			i++
		}
	}

	return cp
}
