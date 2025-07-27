package fileutils

import (
	"io/ioutil"
	"os"
)

// FolderExists 判断文件夹是否存在
func FolderExists(dirname string) (bool, error) {
	// 获取文件夹信息
	fileInfo, err := os.Stat(dirname)
	// 文件夹存在
	if err == nil && fileInfo.IsDir() {
		return true, nil
	}
	// 判断文件夹是否不存在
	if os.IsNotExist(err) {
		return false, nil
	}
	// 返回其他错误信息
	return false, err
}

// FolderExistsOrCreate 文件夹存在不存在就创建
func FolderExistsOrCreate(dirname string) (bool, error) {
	fileInfo, err := os.Stat(dirname)
	// 文件夹存在
	if err == nil && fileInfo.IsDir() {
		return true, nil
	}
	// 判断文件夹是否不存在
	if os.IsNotExist(err) {
		err = os.MkdirAll(dirname, 0755) // 创建文件夹
		if err != nil {
			return false, err
		}
		return true, nil
	}
	// 返回其他错误信息
	return false, err
}

// ListFiles 获取目录下所有文件
func ListFiles(path string, recursive bool) ([]string, error) {
	var ret = make([]string, 0)
	// 读取全部目录
	catalog, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, value := range catalog {
		// 保存当前目录
		crtPath := path + string(os.PathSeparator) + value.Name()
		ret = append(ret, crtPath)

		// 递归获取
		if recursive && value.IsDir() {
			folder, err := ListFiles(crtPath, recursive)
			if err != nil {
				return nil, err
			}
			ret = append(ret, folder...)
		}
	}
	return ret, nil
}
