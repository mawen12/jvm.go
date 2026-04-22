package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/zxh0/jvm.go/classpath"
	"github.com/zxh0/jvm.go/cpu"
	"github.com/zxh0/jvm.go/module"
	_ "github.com/zxh0/jvm.go/native/all"
	"github.com/zxh0/jvm.go/rtda"
	"github.com/zxh0/jvm.go/rtda/heap"
	"github.com/zxh0/jvm.go/vm"
	"github.com/zxh0/jvm.go/vmutils"
)

const usage = `
Usage: {java} [options] class [args...]
    or {java} [options] -m <module>[/<mainclass>] [args...]
       {java} [options] --module <module>[/<mainclass>] [args...]
`

var (
	// 输出版本并退出
	versionFlag bool
	// 输出帮助信息并退出
	helpFlag                 bool
	listModulesFlag          bool
	showModuleResolutionFlag bool
)

// 对应 java 命令 go run ./cmd/java --Xjre /opt/java/jdk1.8.0_451.jdk/Contents/Home/jre HelloWorld
func main() {
	// 解析参数
	opts, args := parseOptions()

	if helpFlag || opts.MainClass == "" {
		printUsage()
	} else if versionFlag {
		printVersion()
	} else if listModulesFlag {
		listModules(opts)
	} else if opts.MainModule != "" {
		startJVM13(opts, args)
	} else {
		// 启动 JVM8
		startJVM8(opts, args)
	}
}

// parseOptions 将参数解析到 Options，并完成初始化操作，将剩下的参数解析到 MainClas 和 args
func parseOptions() (*vm.Options, []string) {
	options := &vm.Options{}
	flag.BoolVar(&versionFlag, "version", false, "Displays version information and exit.")
	flag.BoolVar(&helpFlag, "help", false, "Displays usage information and exit.")
	flag.BoolVar(&helpFlag, "h", false, "Displays usage information and exit.")
	flag.BoolVar(&helpFlag, "?", false, "Displays usage information and exit.")
	flag.BoolVar(&listModulesFlag, "list-modules", false, "Lists observable modules and exit.")
	flag.BoolVar(&showModuleResolutionFlag, "show-module-resolution", false, "Shows module resolution output during startup.")
	flag.StringVar(&options.Xss, "Xss", "", "Sets the thread stack size.")
	flag.StringVar(&options.MainModule, "module", "", "Specifies main module & main class.")
	flag.StringVar(&options.MainModule, "m", "", "Specifies main module & main class.")
	flag.StringVar(&options.ClassPath, "classpath", "", "Specifies a list of directories, JAR files, and ZIP archives to search for class files.")
	flag.StringVar(&options.ClassPath, "cp", "", "Specifies a list of directories, JAR files, and ZIP archives to search for class files.")
	flag.StringVar(&options.ModulePath, "module-path", "", "Specifies module path.") // TODO
	flag.StringVar(&options.ModulePath, "p", "", "Specifies module path.")           // TODO
	flag.BoolVar(&options.VerboseClass, "verbose:class", false, "Displays information about each class loaded.")
	flag.BoolVar(&options.VerboseModule, "verbose:module", false, "Displays information about the modules in use.")
	flag.BoolVar(&options.VerboseJNI, "verbose:jni", false, "Displays information about the use of native methods and other Java Native Interface (JNI) activity.")
	flag.StringVar(&options.Xjre, "Xjre", "", "Specifies JRE path.")
	flag.BoolVar(&options.XUseJavaHome, "XuseJavaHome", false, "Uses JAVA_HOME to find JRE path.")
	flag.BoolVar(&options.XDebugInstr, "Xdebug:instr", false, "Displays executed instructions.")
	flag.StringVar(&options.XCPUProfile, "Xprofile:cpu", "", "")
	flag.Parse()

	// 读取去除 flag 的参数
	args := flag.Args()
	options.Init()

	if mm := options.MainModule; mm != "" {
		if idx := strings.IndexByte(mm, '/'); idx >= 0 {
			options.MainModule = mm[:idx]
			options.MainClass = mm[idx+1:]
		}
	} else if len(args) > 0 {
		// 使用第一个参数作为 MainClass，比如 HelloWorld
		options.MainClass = args[0]
		// 其余的作为 main 方法的启动参数
		args = args[1:]
	}

	return options, args
}

func printUsage() {
	fmt.Println(strings.ReplaceAll(strings.TrimSpace(usage), "{java}", os.Args[0]))
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Println("jvm.go 0.1.8.0")
}

func listModules(opts *vm.Options) {
	mp := module.ParseModulePath(opts)
	for _, m := range mp {
		info := m.GetInfo()
		fmt.Printf("%s@%s\n", info.Name, info.Version)
	}

	fmt.Println("----------")
	x := module.CheckDeps(mp, opts.MainModule)
	for _, m := range x {
		info := m.GetInfo()
		fmt.Printf("%s@%s\n", info.Name, info.Version)
	}
}

func startJVM13(opts *vm.Options, args []string) {
	// TODO
}

func startJVM8(opts *vm.Options, args []string) {
	if opts.XCPUProfile != "" {
		f, err := os.Create(opts.XCPUProfile)
		if err != nil {
			panic(err)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// 创建 main 线程
	mainThread := createMainThread(opts, args)
	// 对主线程执行循环，直到没有方法可调用为止
	cpu.Loop(mainThread)
	//  当 非 daemon 线程数量 > 0 时，阻塞等待
	cpu.KeepAlive()
}

// createMainThread 初始化 bootstrap class loader, 创建 main 线程
func createMainThread(opts *vm.Options, args []string) *rtda.Thread {
	// 解析 jdk 和 classpath 路径
	cp := classpath.Parse(opts)
	// 初始化 Boot 类加载器，初始化 Runtime，加载 Object, Class, Cloneable, Thread, String, 基本数据类型及其数组
	rt := heap.NewRuntime(cp, opts.VerboseClass)

	// 将main从 . -> /，方便查找路径
	mainClass := vmutils.DotToSlash(opts.MainClass)
	// 构造 slot
	bootArgs := []heap.Slot{heap.NewHackSlot(mainClass), heap.NewHackSlot(args)}
	// 创建 main 线程
	mainThread := rtda.NewThread(nil, opts, rt)
	// 调用 shim 的方法
	mainThread.InvokeMethodWithShim(rtda.ShimBootstrapMethod, bootArgs)
	return mainThread
}
