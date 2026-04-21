package heap

import (
	"github.com/zxh0/jvm.go/classfile"
)

/*
对应类的常量池中的 Methodref
*/
type ConstantMethodRef struct {
	ConstantMemberRef
	ParamSlotCount uint
	resolved       *Method
	vslot          int
}

func newConstantMethodRef(class *Class, cf *classfile.ClassFile,
	cfRef classfile.ConstantMethodRefInfo) *ConstantMethodRef {

	ref := &ConstantMethodRef{vslot: -1}
	ref.ConstantMemberRef = newConstantMemberRef(class, cf, cfRef.ClassIndex, cfRef.NameAndTypeIndex)
	ref.ParamSlotCount = calcParamSlotCount(ref.descriptor)
	return ref
}

// GetMethod 读取静态/实例方法
func (ref *ConstantMethodRef) GetMethod(static bool) *Method {
	// 检查该常量是否已经解析
	if ref.resolved == nil { // 尚未解析
		// 检查是否为静态方法
		if static { // 静态
			ref.resolveStaticMethod()
		} else { // 实例方法
			ref.resolveSpecialMethod()
		}
	}
	return ref.resolved
}

// resolveStaticMethod 解析静态方法
func (ref *ConstantMethodRef) resolveStaticMethod() {
	method := ref.findMethod(true)
	if method != nil {
		ref.resolved = method
	} else {
		// todo
		panic("static method not found!")
	}
}

// resolveSpecialMethod 解析实例方法
func (ref *ConstantMethodRef) resolveSpecialMethod() {
	method := ref.findMethod(false)
	if method != nil {
		ref.resolved = method
		return
	}

	// todo
	// class := ref.cp.class.classLoader.LoadClass(ref.className)
	// if class.IsInterface() {
	// 	method = ref.findMethodInInterfaces(class)
	// 	if method != nil {
	// 		ref.method = method
	// 		return
	// 	}
	// }

	// todo
	panic("special method not found!")
}

// findMethod 查找静态/实例方法
func (ref *ConstantMethodRef) findMethod(isStatic bool) *Method {
	// 使用 bootstrap class loader 加载该常量所在的类
	class := ref.getBootLoader().LoadClass(ref.className)
	// 从类中通过方法名称、方法描述符、是否静态（access flag）来读取方法
	return class.getMethod(ref.name, ref.descriptor, isStatic)
}

// todo
/*func (mr *ConstantMethodref) findMethodInInterfaces(iface *Class) *Method {
	for _, m := range iface.methods {
		if !m.IsAbstract() {
			if m.name == mr.name && m.descriptor == mr.descriptor {
				return m
			}
		}
	}

	for _, superIface := range iface.interfaces {
		if m := mr.findMethodInInterfaces(superIface); m != nil {
			return m
		}
	}

	return nil
}*/

func (ref *ConstantMethodRef) GetVirtualMethod(obj *Object) *Method {
	if ref.vslot < 0 {
		ref.vslot = getVslot(obj.Class, ref.name, ref.descriptor)
	}
	if ref.vslot >= 0 {
		return obj.Class.vtable[ref.vslot]
	}

	// TODO: invoking private method ?
	//println("GetVirtualMethod:", ref.className, ref.name, ref.descriptor)
	class := ref.getBootLoader().LoadClass(ref.className)
	return class.getDeclaredMethod(ref.name, ref.descriptor, false)
}
