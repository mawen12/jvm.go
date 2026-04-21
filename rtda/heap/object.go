package heap

import (
	"fmt"
	"reflect"
	"sync"
)

// object 对象
/*
	Class 
		所属的 class 

	Fields
		底层是 heap.Slot[]，用于保存实例的字段值，在对象创建之后，便会执行初始化

	Extra
	
	Monitor
		每个对象都有一个 monitor，用于同步场景

	lock
		使用读写锁实现的同步	
*/
type Object struct {
	// 对象所属的类
	Class *Class
	// 字段，保存了字段值
	Fields  interface{} // []Slot for Object, []int32 for int[] ...
	Extra   interface{} // remember some important things from Golang
	Monitor *Monitor
	// 读写锁
	lock *sync.RWMutex // state lock
}

// newObj 根据指定 class 创建对象
func newObj(class *Class, fields, extra interface{}) *Object {

	return &Object{class, fields, extra, newMonitor(), &sync.RWMutex{}}
}

func (obj *Object) String() string {
	return fmt.Sprintf("{Object@%p class:%v extra:%v}",
		obj, obj.Class, obj.Extra)
}

// todo
// initFields 初始化实例字段
func (obj *Object) initFields() {
	// 读取该对象上预分配的，用于保存实例字段值的 slot 数组
	fields := obj.Fields.([]Slot)
	// 从当期前->父类...->java.lang.Object
	for class := obj.Class; class != nil; class = class.SuperClass {
		// 遍历类的字段
		for _, f := range class.Fields {
			// 检查是否为实例字段
			if !f.IsStatic() { // 实例
				// 为其赋予 null 值
				fields[f.SlotId] = EmptySlot // TODO
			}
		}
	}
}

// state lock
func (obj *Object) LockState()    { obj.lock.Lock() }
func (obj *Object) UnlockState()  { obj.lock.Unlock() }
func (obj *Object) RLockState()   { obj.lock.RLock() }
func (obj *Object) RUnlockState() { obj.lock.RUnlock() }

// reflection
// GetFieldValue 使用反射读取字段值，此处之所以是反射，是因为其并不是直接持有 Field 引用对象
func (obj *Object) GetFieldValue(fieldName, fieldDescriptor string) Slot {
	// 根据方法名称、方法描述符，从当前类一直找到父类，查找匹配的实例字段
	field := obj.Class.GetInstanceField(fieldName, fieldDescriptor)
	// 读取值
	return field.GetValue(obj)
}

// SetFieldValue 使用反射设置字段值，此处之所以是反射，是因为其并不是直接持有 Field 引用对象
func (obj *Object) SetFieldValue(fieldName, fieldDescriptor string, value Slot) {
	// 根据方法名称、方法描述符，从当前类一直找到父类，查找匹配的实例字段
	field := obj.Class.GetInstanceField(fieldName, fieldDescriptor)
	// 设置值
	field.PutValue(obj, value)
}

func (obj *Object) Clone() *Object {
	fields1 := reflect.ValueOf(obj.Fields)
	fields2 := reflect.MakeSlice(fields1.Type(), fields1.Len(), fields1.Len())
	reflect.Copy(fields2, fields1)
	var extra2 interface{} = nil // todo
	return newObj(obj.Class, fields2.Interface(), extra2)
}

func (obj *Object) GetGoClass() *Class {
	return obj.Extra.(*Class)
}
