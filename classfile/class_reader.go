package classfile

import (
	"encoding/binary"
	"reflect"

	"github.com/zxh0/jvm.go/vmutils"
)

// ClassReader 基于 BytesReader 封装，且将读取的结果写入到 ClassFile。
type ClassReader struct {
	vmutils.BytesReader
	cf *ClassFile
}

// newClassReader 使用给定的数组构造，读取的方式顺序为：big-endian
func newClassReader(data []byte) ClassReader {
	br := vmutils.NewBytesReader(data, binary.BigEndian)
	return ClassReader{BytesReader: br}
}

// readUint16s 基于长度前缀法读取多个 uint16（2字节）
func (reader *ClassReader) readUint16s() []uint16 {
	// 读取首个 uint16，确定后续的 uint16 数量
	n := reader.ReadUint16()
	// 构造指定数量的数组
	s := make([]uint16, n)
	for i := range s {
		// 读取 uint16
		s[i] = reader.ReadUint16()
	}
	return s
}

// readFn: func(reader *ClassReader) XXX
func (reader *ClassReader) readTable(readFn interface{}) interface{} {
	n := int(reader.ReadUint16())

	itemType := reflect.TypeOf(readFn).Out(0)
	sliceType := reflect.SliceOf(itemType)
	s := reflect.MakeSlice(sliceType, n, n) // make([]x, n, n)

	readFnVal := reflect.ValueOf(readFn)
	args := []reflect.Value{reflect.ValueOf(reader)}

	for i := 0; i < n; i++ {
		x := readFnVal.Call(args)[0]
		s.Index(i).Set(x) // s[i] = x
	}

	return s.Interface()
}
