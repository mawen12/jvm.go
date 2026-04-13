package heap

import (
	"github.com/zxh0/jvm.go/classfile"
)

// ClassMember 类成员通用结构，比如 field 和 method 就属于类成员
type ClassMember struct {
	// 标识符
	classfile.AccessFlags
	// 名称，字段名或方法名
	Name string
	// 描述符
	Descriptor     string
	Signature      string
	AnnotationData []byte // RuntimeVisibleAnnotations_attribute
	// 所属的 Class
	Class *Class
}

func (m *ClassMember) copyMemberData(cf *classfile.ClassFile, cfMember classfile.MemberInfo) {
	m.AccessFlags = classfile.AccessFlags(cfMember.AccessFlags)
	m.Name = cf.GetUTF8(cfMember.NameIndex)
	m.Descriptor = cf.GetUTF8(cfMember.DescriptorIndex)
	m.Signature = cf.GetUTF8(cfMember.GetSignatureIndex())
	m.AnnotationData = cfMember.GetRuntimeVisibleAnnotationsAttributeData()
}
