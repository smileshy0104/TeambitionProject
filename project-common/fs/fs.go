package fs

import "os"

// IsExist 判断所给路径文件/文件夹是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
