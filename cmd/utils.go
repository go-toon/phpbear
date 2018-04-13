package cmd

import (
	"strings"
	"os"
	"io"
)

// 根据git项目地址获取项目名称
func getProjectName(gitUrl string) string {
	return subStr(gitUrl, strings.LastIndex(gitUrl, "/")+1, strings.LastIndex(gitUrl, "."))
}

// 截取字符串
func subStr(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)
	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < 0 || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}

// isDir 判断给定地址是否是目录, 如果不存在或者给定的是文件，返回false
func isDir(dir string) bool {
	f, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return f.IsDir()
}

// isFile 判断给定的path是否是文件，如果是目录或者不存在，返回false
func isFile(filePath string) bool {
	f, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return !f.IsDir()
}

// 判断文件或者目录是否存在
func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// 拷贝文件
func copyFile(source, dest string) bool {
	if source == "" || dest == "" {
		return false
	}

	// 大家资源
	sourceOpen, err := os.Open(source)
	if err != nil {
		return false
	}
	defer sourceOpen.Close()

	//只写模式打开文件，如果文件不存在则进行创建， 并赋予644的权限
	destOpen, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 6444)
	if err != nil {
		return false
	}
	defer destOpen.Close()

	//进行数据拷贝
	if _, err := io.Copy(destOpen, sourceOpen); err != nil {
		return false
	}

	return true
}
