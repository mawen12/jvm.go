package classfile

import (
	"fmt"
)

// Parse 将一个 class 文件的字节表示，读取为 ClassFile
func Parse(classData []byte) (cf *ClassFile, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	cf = &ClassFile{}
	// 构造读取器
	cr := newClassReader(classData)
	cf.read(&cr)
	return
}
