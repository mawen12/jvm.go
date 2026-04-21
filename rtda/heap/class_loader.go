package heap

import (
	"fmt"

	"github.com/zxh0/jvm.go/classfile"
	"github.com/zxh0/jvm.go/classpath"
	"github.com/zxh0/jvm.go/vm"
	"github.com/zxh0/jvm.go/vmutils"
)

const (
	jlObjectClassName       = "java/lang/Object"
	jlClassClassName        = "java/lang/Class"
	jlStringClassName       = "java/lang/String"
	jlThreadClassName       = "java/lang/Thread"
	jlCloneableClassName    = "java/lang/Cloneable"
	ioSerializableClassName = "java/io/Serializable"
)

/*
class names:
    - primitive types: boolean, byte, int ...
    - primitive arrays: [Z, [B, [I ...
    - non-array classes: java/lang/Object ...
    - array classes: [Ljava/lang/Object; ...
*/

// the bootstrap class loader
// ClassLoader 负责加载类
type ClassLoader struct {
	rt *Runtime
	// 类路径
	classPath *classpath.ClassPath
	// 已加载的类文件，在实际的 JVM 中，ClassLoader 会持有一个指向 Metaspace 的引用，Metaspace 中保存已经 loaded 的类
	classMap map[string]*Class // loaded classes
	// 当实际加载完成后，输出日志
	verbose bool
	// some frequently used classes
	// 记录最频繁使用的类
	jlObjectClass       *Class // java/lang/Object
	jlClassClass        *Class // java/lang/Class
	jlStringClass       *Class // java/lang/String
	jlThreadClass       *Class // java/lang/Thread
	jlCloneableClass    *Class // java/lang/Cloneable
	ioSerializableClass *Class // java/io/Serializable
}

// newBootLoader 初始化 bootstrap 类加载器
func newBootLoader(cp *classpath.ClassPath, verbose bool) *ClassLoader {
	return &ClassLoader{
		classPath: cp,
		classMap:  map[string]*Class{},
		verbose:   verbose,
	}
}

// init 加载 Object, Class, Cloneable, Thread, String, 基本数据类型及其数组
func (loader *ClassLoader) init() {
	// 加载 java.lang.Object
	loader.jlObjectClass = loader.LoadClass(jlObjectClassName)
	// 加载 java.lang.Class
	loader.jlClassClass = loader.LoadClass(jlClassClassName)
	for _, class := range loader.classMap {
		if class.JClass == nil {
			class.JClass = loader.jlClassClass.NewObj()
			class.JClass.Extra = class
		}
	}
	// 加载 java.lang.Cloneable
	loader.jlCloneableClass = loader.LoadClass(jlCloneableClassName)
	// 加载 java.io.Serializable
	loader.ioSerializableClass = loader.LoadClass(ioSerializableClassName)
	// 加载 java.lang.Thread
	loader.jlThreadClass = loader.LoadClass(jlThreadClassName)
	// 加载 java.lang.String
	loader.jlStringClass = loader.LoadClass(jlStringClassName)
	// 加载原始数据类型
	loader.loadPrimitiveClasses()
	// 加载原始数据类型对应的数组
	loader.loadPrimitiveArrayClasses()
}

// loadPrimitiveClasses 加载原始数据类型
func (loader *ClassLoader) loadPrimitiveClasses() {
	// void -> java.lang.Void
	// boolean -> java.lang.Boolean
	// byte -> java.lang.Byte
	// char -> java.lang.Character
	// short -> java.lang.Short
	// int -> java.lang.Integer
	// long -> java.lang.Long
	// float -> java.lang.Float
	// double -> java.lang.Double
	for _, primitiveType := range primitiveTypes {
		loader.loadPrimitiveClass(primitiveType.Name)
	}
}
func (loader *ClassLoader) loadPrimitiveClass(className string) {
	class := &Class{Name: className}
	class.bootLoader = loader
	class.JClass = loader.jlClassClass.NewObj()
	class.JClass.Extra = class
	class.MarkFullyInitialized()
	loader.classMap[className] = class
}

// loadPrimitiveArrayClasses 加载原始数据类型对应的数组
func (loader *ClassLoader) loadPrimitiveArrayClasses() {
	// void -> java.lang.Void
	// boolean -> java.lang.Boolean
	// byte -> java.lang.Byte
	// char -> java.lang.Character
	// short -> java.lang.Short
	// int -> java.lang.Integer
	// long -> java.lang.Long
	// float -> java.lang.Float
	// double -> java.lang.Double
	for _, primitiveType := range primitiveTypes {
		loader.loadArrayClass(primitiveType.ArrayClassName)
	}
}
func (loader *ClassLoader) loadArrayClass(className string) *Class {
	class := &Class{Name: className}
	class.bootLoader = loader
	class.SuperClass = loader.jlObjectClass
	class.Interfaces = []*Class{loader.jlCloneableClass, loader.ioSerializableClass}
	class.JClass = loader.jlClassClass.NewObj()
	class.JClass.Extra = class
	createVtable(class)
	class.MarkFullyInitialized()
	loader.classMap[className] = class
	return class
}

func (loader *ClassLoader) getRefArrayClass(componentClass *Class) *Class {
	arrClassName := "[L" + componentClass.Name + ";"
	return loader.getRefArrayClassByName(arrClassName)
}
func (loader *ClassLoader) getRefArrayClassByName(arrClassName string) *Class {
	if arrClass, ok := loader.classMap[arrClassName]; ok {
		return arrClass
	}
	return loader.loadArrayClass(arrClassName)
}

func (loader *ClassLoader) JLObjectClass() *Class {
	return loader.jlObjectClass
}
func (loader *ClassLoader) JLClassClass() *Class {
	return loader.jlClassClass
}
func (loader *ClassLoader) JLStringClass() *Class {
	return loader.jlStringClass
}
func (loader *ClassLoader) JLThreadClass() *Class {
	return loader.jlThreadClass
}

// todo
func (loader *ClassLoader) GetPrimitiveClass(name string) *Class {
	return loader.getClass(name)
}

func (loader *ClassLoader) FindLoadedClass(name string) *Class {
	if class, ok := loader.classMap[name]; ok {
		return class
	}
	return nil
}

// todo dangerous
func (loader *ClassLoader) getClass(name string) *Class {
	if class, ok := loader.classMap[name]; ok {
		return class
	}
	panic("class not loaded! " + name)
}

// LoadClass 通过类名加载类
func (loader *ClassLoader) LoadClass(name string) *Class {
	// 检查是否已经加载过，加载过就直接返回
	if class, ok := loader.classMap[name]; ok {
		// already loaded
		return class
	} else if name[0] == '[' { // 如果首字节为 [，代表是数组
		// array class
		return loader.getRefArrayClassByName(name)
	} else {
		// 执行原始的类加载流程
		return loader.reallyLoadClass(name)
	}
}

// reallyLoadClass 执行原始的类加载流程
func (loader *ClassLoader) reallyLoadClass(name string) *Class {
	// 读取类的文件信息
	cpEntry, data := loader.readClassData(name)
	// 将文件加载到内存中，并解析为 Class
	class := loader.loadClass(name, data)
	// 添加 Class 的来源
	class.LoadedFrom = cpEntry

	// 输出
	if loader.verbose {
		fmt.Printf("[Loaded %s from %s]\n", name, cpEntry)
	}

	return class
}

// readClassData 定位文件，并读取为 []byte
func (loader *ClassLoader) readClassData(name string) (classpath.Entry, []byte) {
	// 读取类的文件 entry 和字节数组
	cpEntry, classData := loader.classPath.ReadClass(name)
	// 如果读取不到，则抛出 ClassNotFoundError
	if classData == nil {
		panic(vm.NewClassNotFoundError(vmutils.SlashToDot(name)))
	}

	return cpEntry, classData
}

// parseClassData 解析并将结果转换为 Class
func (loader *ClassLoader) parseClassData(name string, data []byte) *Class {
	// 将字节数组解析为 classFile
	cf, err := classfile.Parse(data)
	if err != nil {
		// todo
		panic("failed to parse class file: " + name + "! " + err.Error())
	}

	// 将 classFile -> Class
	return newClass(cf)
}

func (loader *ClassLoader) loadClass(name string, data []byte) *Class {
	// 解析并将结果转换为 Class
	class := loader.parseClassData(name, data)
	//
	hackClass(class)
	loader.resolveSuperClass(class)
	loader.resolveInterfaces(class)
	calcStaticFieldSlotIds(class)
	calcInstanceFieldSlotIds(class)
	createVtable(class)
	prepare(class)
	// todo
	class.bootLoader = loader
	loader.classMap[name] = class

	if loader.jlClassClass != nil {
		class.JClass = loader.jlClassClass.NewObj()
		class.JClass.Extra = class
	}

	return class
}

// todo
func hackClass(class *Class) {
	if class.Name == "java/lang/ClassLoader" {
		loadLibrary := class.GetStaticMethod("loadLibrary", "(Ljava/lang/Class;Ljava/lang/String;Z)V")
		loadLibrary.Code = []byte{0xb1} // return void
	}
}

// todo
// resolveSuperClass 解决父类
func (loader *ClassLoader) resolveSuperClass(class *Class) {
	if class.superClassName != "" {
		// 加载父类
		class.SuperClass = loader.LoadClass(class.superClassName)
	}
}

// resolveInterfaces 解决接口
func (loader *ClassLoader) resolveInterfaces(class *Class) {
	interfaceCount := len(class.interfaceNames)
	if interfaceCount > 0 {
		class.Interfaces = make([]*Class, interfaceCount)
		for i, interfaceName := range class.interfaceNames {
			// 加载接口
			class.Interfaces[i] = loader.LoadClass(interfaceName)
		}
	}
}

func calcStaticFieldSlotIds(class *Class) {
	slotId := uint(0)
	for _, field := range class.Fields {
		if field.IsStatic() {
			field.SlotId = slotId
			slotId++
		}
	}
	class.staticFieldCount = slotId
}

func calcInstanceFieldSlotIds(class *Class) {
	slotId := uint(0)
	if class.superClassName != "" {
		slotId = class.SuperClass.instanceFieldCount
	}
	for _, field := range class.Fields {
		if !field.IsStatic() {
			field.SlotId = slotId
			slotId++
		}
	}
	class.instanceFieldCount = slotId
}

func prepare(class *Class) {
	class.StaticFieldSlots = make([]Slot, class.staticFieldCount)
	for _, field := range class.Fields {
		if field.IsStatic() {
			class.StaticFieldSlots[field.SlotId] = EmptySlot // TODO
		}
	}
}

// todo
func (loader *ClassLoader) DefineClass(name string, data []byte) *Class {
	return loader.loadClass(name, data)
}
