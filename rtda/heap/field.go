package heap

import (
	"github.com/zxh0/jvm.go/classfile"
)

type Field struct {
	ClassMember
	// 是否为 long or double，本质是其是否为 64 为的
	IsLongOrDouble bool
	// 常量值索引，指向常量池
	ConstValueIndex uint16
	SlotId          uint
	_type           *Class
}

func newField(class *Class, cf *classfile.ClassFile, cfMember classfile.MemberInfo) *Field {
	field := &Field{}
	field.Class = class
	field.copyMemberData(cf, cfMember)
	//
	field.IsLongOrDouble = field.Descriptor == "J" || field.Descriptor == "D"
	field.ConstValueIndex = cfMember.GetConstantValueIndex()
	return field
}

// GetValue 读取字段值
func (field *Field) GetValue(ref *Object) Slot {
	// 读取对象上的slot数组
	fields := ref.Fields.([]Slot)
	// 根据 slotId 读取值
	return fields[field.SlotId]
}

// PutValue 设置字段值
func (field *Field) PutValue(ref *Object, val Slot) {
	// 读取对象上的slot数组
	fields := ref.Fields.([]Slot)
	// 将给定 slotId 的索引位置写入值
	fields[field.SlotId] = val
}

// GetStaticValue 读取静态字段值
func (field *Field) GetStaticValue() Slot {
	return field.Class.StaticFieldSlots[field.SlotId]
}

// PutStaticValue 设置静态字段值
func (field *Field) PutStaticValue(val Slot) {
	field.Class.StaticFieldSlots[field.SlotId] = val
}

// reflection
func (field *Field) Type() *Class {
	if field._type == nil {
		field._type = field.resolveType()
	}
	return field._type
}
func (field *Field) resolveType() *Class {
	bootLoader := field.Class.bootLoader
	className := getClassName(field.Descriptor)
	return bootLoader.LoadClass(className)
}
