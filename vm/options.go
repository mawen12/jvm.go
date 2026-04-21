package vm

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	_1k = 1024
	_1m = _1k * _1k
	_1g = _1k * _1m
)

type Options struct {
	MainModule      string
	MainClass       string // 带有 Main 方法的类
	ClassPath       string
	ModulePath      string
	VerboseClass    bool
	VerboseModule   bool
	VerboseJNI      bool
	Xss             string // 线程栈的大小
	Xjre            string // jre 的路径
	XUseJavaHome    bool   // 是否通过 JAVA_HOME 来读取 jdk 的路径
	XDebugInstr     bool
	XCPUProfile     string
	AbsJavaHome     string // /path/to/jre 指向 jre 的完整路径
	AbsJreLib       string // /path/to/jre/lib 指向 jre/lib 的完整路径
	ThreadStackSize int    // 线程栈的大小，也是最大深入，被 Stack#maxSize 使用
}

// Init 初始化参数
func (options *Options) Init() {
	// 检查是否指定了 module
	if options.ModulePath != "" {
		// 使用了 module
		options.AbsJavaHome = getJavaHome13(options.Xjre)
	} else {
		// 读取 jre 的路径
		options.AbsJavaHome = getJavaHome8(options.Xjre, options.XUseJavaHome)
		// 指向 jre/lib 的完整路径
		options.AbsJreLib = filepath.Join(options.AbsJavaHome, "lib")
	}
	// 解析 Xss 作为线程栈的大小
	options.ThreadStackSize = parseXss(options.Xss)
}

func getJavaHome13(jreDir string) string {
	if absJH, err := filepath.Abs(jreDir); err != nil {
		panic(err) // TODO
	} else {
		return absJH
	}
}

// getJavaHome8 读取 JAVA_HOME 路径
func getJavaHome8(jreDir string, useOsEnv bool) string {
	jh := "./jre"
	if jreDir != "" { // 提供了 JRE 的路径，则直接使用
		jh = jreDir
	} else if useOsEnv { // 指定了通过环境变量读取 JAVA_HOME
		if jh = os.Getenv("JAVA_HOME"); jh == "" {
			panic("$JAVA_HOME not set!")
		}
	}

	if absJH, err := filepath.Abs(jh); err == nil {

		if strings.Contains(absJH, "jre") {
			return absJH
		} else {
			// 比如对于从环境变量读取的场景，就不会包含 jre，需要手动拼接
			return filepath.Join(absJH, "jre")
		}
	} else {
		panic(err) // TODO
	}
}

// -Xss<size>[g|G|m|M|k|K]
// 将 Xxx 设置的值转换为数字
func parseXss(size string) int {
	if size == "" {
		// 默认为 16kb
		return 16 * _1k
	}
	switch size[len(size)-1] {
	case 'g', 'G':
		// 将 G 转换为 k
		return parseSS(size[:len(size)-1], _1g)
	case 'm', 'M':
		// 将 M 转换为 K
		return parseSS(size[:len(size)-1], _1m)
	case 'k', 'K':
		// 将 k 转换为 k
		return parseSS(size[:len(size)-1], _1k)
	default:
		return parseSS(size, 1)
	}
}

func parseSS(size string, unit int) int {
	if i, err := strconv.Atoi(size); err != nil {
		panic(errors.New("invalid thread stack size: " + size))
	} else {
		return i * unit
	}
}
