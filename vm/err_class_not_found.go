package vm

// ClassNotFoundError 当 ClassLoader 尝试从其管理的路径中加载某个 Class 得到的结果为 null 时抛出
type ClassNotFoundError struct {
	name string
}

func NewClassNotFoundError(name string) ClassNotFoundError {
	return ClassNotFoundError{name}
}

func (err ClassNotFoundError) Error() string {
	return err.name
}
