package vmutils

import (
	"os"
	"strings"
)

// IsDir 检查路径是否指向一个目录
func IsDir(path string) bool {
	if fileInfo, err := os.Stat(path); err == nil {
		return fileInfo.IsDir()
	}
	return false
}

// IsExists 检查路径是否存在
func IsExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
		//} else if os.IsNotExist(err) {
		//	return false
	} else {
		return false
	}
}

// IsZipFile 检查文件是否为 zip 格式
func IsZipFile(name string) bool {
	return strings.HasSuffix(name, ".zip") ||
		strings.HasSuffix(name, ".ZIP")
}

// IsJarFile 检查文件是否为 jar 格式
func IsJarFile(name string) bool {
	return strings.HasSuffix(name, ".jar") ||
		strings.HasSuffix(name, ".JAR")
}

// IsJModFile 检查文件是否为 jmod 格式
func IsJModFile(name string) bool {
	return strings.HasSuffix(name, ".jmod")
}
