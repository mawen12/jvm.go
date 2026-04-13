package vmutils

import (
	"io/ioutil"
	"path/filepath"
)

// 代表一个目录
type Dir struct {
	absPath string
}

func NewDir(path string) (*Dir, error) {
	if absPath, err := filepath.Abs(path); err != nil {
		return nil, err
	} else {
		return &Dir{absPath: absPath}, nil
	}
}

func (dir *Dir) AbsPath() string {
	return dir.absPath
}

func (dir *Dir) ReadFile(filename string) ([]byte, error) {
	// 拼接完整的文件路径
	absFilename := filepath.Join(dir.absPath, filename)
	// 读取文件内容
	if data, err := ioutil.ReadFile(absFilename); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}
