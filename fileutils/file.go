package fileutils

import (
	"fmt"
	"io/ioutil"
)

// ReadFile 读取文件内容并返回
func ReadFile(filepath string) (string, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	return string(content), nil
}
