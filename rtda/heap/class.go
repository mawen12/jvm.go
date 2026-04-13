package heap

import (
	"sync"

	"github.com/zxh0/jvm.go/classfile"
	"github.com/zxh0/jvm.go/classpath"
	"github.com/zxh0/jvm.go/vmutils"
)

// initialization state
// 初始化状态
// 0 -> 类已被 verified 和 prepared
// 1 -> 类正在被 initialized
// 2 -> 类已经被 initialized
// 3 -> 类处于错误状态，可能是因为初始化尝试失败了
const (
	_notInitialized   = 0 // This Class object is verified and prepared but not initialized.
	_beingInitialized = 1 // This Class object is being initialized by some particular thread T.
	_fullyInitialized = 2 // This Class object is fully initialized and ready for use.
	_initFailed       = 3 // This Class object is in an erroneous state, perhaps because initialization was attempted and failed.
)

type EnclosingMethod struct {
	ClassName        string
	MethodName       string
	MethodDescriptor string
}

// name, superClassName and interfaceNames are all binary names(jvms8-4.2.1)
type Class struct {
	// 类的访问标志，比如 public/static/final...
	classfile.AccessFlags
	// 常量池，用于 .class 的字符常量解析
	ConstantPool
	// 当前类名
	Name string // thisClassName
	// 父类名
	superClassName string
	// 接口类名
	interfaceNames []string
	// 原始文件
	SourceFile string
	// 签名
	Signature string
	// 注解数据
	AnnotationData []byte // RuntimeVisibleAnnotations_attribute
	// 该类是在哪个内部类中构造的
	// public class Outer {
	// 	 public void test() {
	// 		Runnable r = new Runnable() {
	//			@Override
	//			public void run() {
	//				System.out.println("hello");
	//			}
	//		}
	//	 }
	// }
	// 会生成 Outer$1.class，对应的信息为：Out test ()V
	EnclosingMethod *EnclosingMethod
	// 类中声明的字段
	Fields []*Field
	// 类中声明的方法
	Methods []*Method
	// 实例字段数量
	instanceFieldCount uint
	// 静态字段数据
	staticFieldCount uint
	// 静态字段插槽，保存了静态字段对应的值
	StaticFieldSlots []Slot
	// 虚拟方法表
	vtable []*Method // virtual method table
	// java.lang.Class 示例
	JClass *Object // java.lang.Class instance
	// 父类引用
	SuperClass *Class
	// 接口引用
	Interfaces []*Class
	// 该 Class 对象的原始来源，就是其 .class 文件
	LoadedFrom classpath.Entry
	// 状态
	initState int
	InitCond  *sync.Cond
	// 负责执行初始的线程
	initThread uintptr
	// 加载该类的类加载器
	bootLoader *ClassLoader // TODO
}

func (class *Class) String() string {
	return "{Class name:" + class.Name + "}"
}

// todo
// NameJlsFormat 将类名从 / 转换为 .，比如 java/lang/Object -> java.lang.Object
func (class *Class) NameJlsFormat() string {
	return vmutils.SlashToDot(class.Name)
}

func (class *Class) InitializationNotStarted() bool {
	return class.initState < _beingInitialized // todo
}
func (class *Class) IsBeingInitialized() (bool, uintptr) {
	return class.initState == _beingInitialized, class.initThread
}
func (class *Class) IsFullyInitialized() bool {
	return class.initState == _fullyInitialized
}
func (class *Class) IsInitializationFailed() bool {
	return class.initState == _initFailed
}
func (class *Class) MarkBeingInitialized(thread uintptr) {
	class.initState = _beingInitialized
	class.initThread = thread
}
func (class *Class) MarkFullyInitialized() {
	class.initState = _fullyInitialized
}

// getField 读取字段
func (class *Class) getField(name, descriptor string, isStatic bool) *Field {
	// 从当前类->父类->父类... -> java.lang.Object
	for k := class; k != nil; k = k.SuperClass {
		// 遍历字段
		for _, field := range k.Fields {
			// 检查条件是否匹配
			if field.IsStatic() == isStatic &&
				field.Name == name &&
				field.Descriptor == descriptor {

				return field
			}
		}
	}
	// todo
	return nil
}

// getMethod 读取方法
func (class *Class) getMethod(name, descriptor string, isStatic bool) *Method {
	// 从当前类->父类->父类... -> java.lang.Object
	for k := class; k != nil; k = k.SuperClass {
		// 遍历方法
		for _, method := range k.Methods {
			// 检查条件是否匹配
			if method.IsStatic() == isStatic &&
				method.Name == name &&
				method.Descriptor == descriptor {

				return method
			}
		}
	}
	// todo
	return nil
}

// getDeclaredMethod 读取当前类中声明的方法
func (class *Class) getDeclaredMethod(name, descriptor string, isStatic bool) *Method {
	for _, method := range class.Methods {
		if method.IsStatic() == isStatic &&
			method.Name == name &&
			method.Descriptor == descriptor {

			return method
		}
	}
	return nil
}

// GetStaticField 读取静态字段
func (class *Class) GetStaticField(name, descriptor string) *Field {
	return class.getField(name, descriptor, true)
}

// GetInstanceField 读取实例字段
func (class *Class) GetInstanceField(name, descriptor string) *Field {
	return class.getField(name, descriptor, false)
}

// GetStaticMethod 读取静态方法
func (class *Class) GetStaticMethod(name, descriptor string) *Method {
	return class.getMethod(name, descriptor, true)
}

// GetInstanceMethod 读取实例方法
func (class *Class) GetInstanceMethod(name, descriptor string) *Method {
	return class.getMethod(name, descriptor, false)
}

// GetMainMethod 读取 main 方法
func (class *Class) GetMainMethod() *Method {
	return class.GetStaticMethod(mainMethodName, mainMethodDesc)
}

func (class *Class) GetClinitMethod() *Method {
	return class.getDeclaredMethod(clinitMethodName, clinitMethodDesc, true)
}

func (class *Class) NewObjWithExtra(extra interface{}) *Object {
	obj := class.NewObj()
	obj.Extra = extra
	return obj
}

// NewObj 初始化对象
func (class *Class) NewObj() *Object {
	// 要初始化示例字段
	if class.instanceFieldCount > 0 {
		// 将示例字段保存到 slot
		fields := make([]Slot, class.instanceFieldCount)
		obj := newObj(class, fields, nil)
		obj.initFields()
		return obj
	} else {
		return newObj(class, nil, nil)
	}
}
func (class *Class) NewArray(count uint) *Object {
	return newRefArray(class, count)
}

func (class *Class) isJlObject() bool {
	return class == class.bootLoader.jlObjectClass
}
func (class *Class) isJlCloneable() bool {
	return class == class.bootLoader.jlCloneableClass
}
func (class *Class) isJioSerializable() bool {
	return class == class.bootLoader.ioSerializableClass
}

// reflection
func (class *Class) GetStaticValue(fieldName, fieldDescriptor string) Slot {
	field := class.GetStaticField(fieldName, fieldDescriptor)
	return field.GetStaticValue()
}
func (class *Class) SetStaticValue(fieldName, fieldDescriptor string, value Slot) {
	field := class.GetStaticField(fieldName, fieldDescriptor)
	field.PutStaticValue(value)
}

func (class *Class) AsObj() *Object {
	return &Object{Fields: class.StaticFieldSlots}
}
