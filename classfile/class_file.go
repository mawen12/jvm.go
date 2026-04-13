package classfile

import (
	"fmt"

	"github.com/zxh0/jvm.go/vmutils"
)

/*
	ClassFile {
	    u4             magic;
	    u2             minor_version;
	    u2             major_version;
	    u2             constant_pool_count;
	    cp_info        constant_pool[constant_pool_count-1];
	    u2             access_flags;
	    u2             this_class;
	    u2             super_class;
	    u2             interfaces_count;
	    u2             interfaces[interfaces_count];
	    u2             fields_count;
	    field_info     fields[fields_count];
	    u2             methods_count;
	    method_info    methods[methods_count];
	    u2             attributes_count;
	    attribute_info attributes[attributes_count];
	}
*/
type ClassFile struct {
	//magic      uint32
	MinorVersion uint16
	MajorVersion uint16
	ConstantPool []ConstantInfo
	AccessFlags  uint16
	ThisClass    uint16
	SuperClass   uint16
	Interfaces   []uint16
	Fields       []MemberInfo
	Methods      []MemberInfo
	AttributeTable
}

// read 读取类信息
func (cf *ClassFile) read(reader *ClassReader) {
	reader.cf = cf
	// 读取并校验 magic
	cf.readAndCheckMagic(reader)
	// 读取并校验 版本号
	cf.readAndCheckVersions(reader)
	// 读取常量池
	cf.ConstantPool = readConstantPool(reader)
	// 访问标识符
	cf.AccessFlags = reader.ReadUint16()
	// 当前类的索引
	cf.ThisClass = reader.ReadUint16()
	// 父类的索引
	cf.SuperClass = reader.ReadUint16()
	// 接口的多个索引
	cf.Interfaces = reader.readUint16s()
	// 读取字段
	cf.Fields = readMembers(reader)
	// 读取方法
	cf.Methods = readMembers(reader)
	// 读取属性
	cf.AttributeTable = readAttributes(reader)
}

// readAndCheckMagic 检查接下来的4字节内容是否为：0xCAFEBABE
func (cf *ClassFile) readAndCheckMagic(reader *ClassReader) {
	// 读取 4字节 作为 magic
	magic := reader.ReadUint32()
	// 模数固定为：0xCAFEBABE
	if magic != 0xCAFEBABE {
		panic("Bad magic!") // TODO
	}
}

// readAndCheckVersions 读取主版本和次版本号，检查版本是否合法
func (cf *ClassFile) readAndCheckVersions(reader *ClassReader) {
	// 读取 2字节 作为次版本号
	cf.MinorVersion = reader.ReadUint16()
	// 读取 2字节 作为主版本号
	cf.MajorVersion = reader.ReadUint16()

	switch cf.MajorVersion {
	case 45: // JDK 1.1
		return
	case 46, 47, 48, 49, 50, 51, 52,
		53, 54, 55, 56, 57: // JDK 1.2 -> JDK13
		if cf.MinorVersion == 0 {
			return
		}
	}
	// 不支持的版本
	panic("java.lang.UnsupportedClassVersionError!")
}

// GetThisClassName 读取当前类名
func (cf *ClassFile) GetThisClassName() string {
	return cf.GetClassName(cf.ThisClass)
}

// GetThisClassName 读取父类名
func (cf *ClassFile) GetSuperClassName() string {
	return cf.GetClassName(cf.SuperClass)
}

// GetInterfaceNames 读取接口名
func (cf *ClassFile) GetInterfaceNames() []string {
	return cf.GetClassNames(cf.Interfaces)
}

// GetNameAndType 根据索引读取名称和类型
func (cf *ClassFile) GetNameAndType(cpIndex uint16) (name, _type string) {
	if cpIndex > 0 {
		ntInfo := cf.getConstantInfo(cpIndex).(ConstantNameAndTypeInfo)
		// 读取 name
		name = cf.GetUTF8(ntInfo.NameIndex)
		// 读取 type
		_type = cf.GetUTF8(ntInfo.DescriptorIndex)
	}
	return
}

// GetClassName 根据索引从常量池中读取类名
func (cf *ClassFile) GetClassName(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	// 读取类名的字节数组表示
	classInfo := cf.getConstantInfo(cpIndex).(ConstantClassInfo)
	// 转换为 string
	return cf.GetUTF8(classInfo.NameIndex)
}

// GetPackageName 根据索引从常量池中读取包名
func (cf *ClassFile) GetPackageName(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	// 读取包名的字节数组表示
	pkgInfo := cf.getConstantInfo(cpIndex).(ConstantPackageInfo)
	// 转换为 string
	return cf.GetUTF8(pkgInfo.NameIndex)
}

// GetModuleName 根据索引从常量池中读取模块名
func (cf *ClassFile) GetModuleName(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	// 读取模块名的字节数组表示
	modInfo := cf.getConstantInfo(cpIndex).(ConstantModuleInfo)
	// 转换为 string
	return cf.GetUTF8(modInfo.NameIndex)
}

// GetClassNames 根据一组索引读取一组类名
func (cf *ClassFile) GetClassNames(cpIndexes []uint16) []string {
	ss := make([]string, len(cpIndexes))
	for i, cpIndex := range cpIndexes {
		ss[i] = cf.GetClassName(cpIndex)
	}
	return ss
}

// GetModuleNames 根据一组索引读取一组模块名
func (cf *ClassFile) GetModuleNames(cpIndexes []uint16) []string {
	ss := make([]string, len(cpIndexes))
	for i, cpIndex := range cpIndexes {
		ss[i] = cf.GetModuleName(cpIndex)
	}
	return ss
}

// GetRawUTF8 从常量池中读取指定索引的值，并转换为 string
func (cf *ClassFile) GetRawUTF8(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	// 从常量池中读取指定索引的值
	rawBytes := cf.getConstantInfo(cpIndex).([]byte)
	// 转换为 string
	return string(rawBytes)
}
func (cf *ClassFile) GetUTF8(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	bytes := cf.getConstantInfo(cpIndex).([]byte)
	return vmutils.DecodeMUTF8(bytes)
}

// getConstantInfo 从常量池中读取指定索引的值
func (cf *ClassFile) getConstantInfo(cpIndex uint16) ConstantInfo {
	// 从常量池中读取指定索引的值
	if cpInfo := cf.ConstantPool[cpIndex]; cpInfo == nil {
		panic(fmt.Errorf("invalid constant pool index: %d", cpIndex))
	} else {
		return cpInfo
	}
}
