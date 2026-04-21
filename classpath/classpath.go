package classpath

import (
	"path/filepath"
	"strings"

	"github.com/zxh0/jvm.go/vm"
)

type ClassPath struct {
	entries []Entry
}

// Parse 解析 jdk 和 classpath 路径
func Parse(opts *vm.Options) *ClassPath {
	cp := &ClassPath{}
	// 解析 jdk/lib 和 jdk/lib/ext 的路径
	cp.parseBootAndExtClassPath(opts.AbsJavaHome)
	// 解析用户指定的 classpath
	cp.parseUserClassPath(opts.ClassPath)
	return cp
}

// parseBootAndExtClassPath 解析 jdk/lib 和 jdk/lib/ext 的路径
func (cp *ClassPath) parseBootAndExtClassPath(absJavaHome string) {
	// jre/lib/*
	jreLibPath := filepath.Join(absJavaHome, "lib", "*")
	cp.entries = append(cp.entries, spreadWildcardEntry(jreLibPath)...)

	// jre/lib/ext/*
	jreExtPath := filepath.Join(absJavaHome, "lib", "ext", "*")
	cp.entries = append(cp.entries, spreadWildcardEntry(jreExtPath)...)
}

// parseUserClassPath 解析用户指定的 classpath
func (cp *ClassPath) parseUserClassPath(cpOption string) {
	if cpOption == "" {
		cpOption = "."
	}
	cp.entries = append(cp.entries, parsePath(cpOption)...)
}

// className: fully/qualified/ClassName
// ReadClass 通过全限定类名读取类信息，比如 com.github.mawen12.HelloWorld
func (cp *ClassPath) ReadClass(className string) (Entry, []byte) {
	// 添加 .class 后缀
	className = className + ".class"
	// 迭代该类路径下的 entry
	for _, entry := range cp.entries {
		//
		if data, err := entry.readClass(className); err == nil {
			return entry, data
		}
	}
	return nil, nil
}

func IsBootClassPath(entry Entry, absJreLib string) bool {
	if entry == nil {
		// todo
		return true
	}

	return strings.HasPrefix(entry.String(), absJreLib)
}
