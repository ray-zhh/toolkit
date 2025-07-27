package fileutils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// ReadFile 读取文件内容并返回
func ReadFile(filepath string) (string, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	return string(content), nil
}

// WriteFile 写文件
func WriteFile(path string, content []byte) error {
	// 检查文件夹是否存在
	exist, err := FolderExistsOrCreate(filepath.Dir(path))
	if err != nil {
		return err
	}
	if !exist {
		return errors.New(fmt.Sprintf("create file fail is %v", path))
	}
	return ioutil.WriteFile(path, content, os.ModePerm)
}

// FileExists 文件是否存在
func FileExists(fileName string) (bool, error) {
	_, err := os.Stat(fileName)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

// FileModTime 获取文件最后修改日期
func FileModTime(filePath string) (time.Time, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	return fileInfo.ModTime(), nil
}

// FileSize 文件大小
func FileSize(fileName string) (int64, error) {
	stat, err := os.Stat(fileName)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func FileMd5(fileName string) (string, error) {
	// 文件全路径名
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	md5h := md5.New()
	_, err = io.Copy(md5h, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(md5h.Sum(nil)), nil
}
