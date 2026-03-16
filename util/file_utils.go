package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir 检查是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile 检查是否为文件
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Mkdirs 创建目录
func Mkdirs(path string) error {
	return os.MkdirAll(path, 0755)
}

// CreateFile 创建文件（包含父目录）
func CreateFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return os.Create(path)
}

// DeleteFile 删除文件
func DeleteFile(path string) error {
	return os.Remove(path)
}

// DeleteDir 删除目录（递归）
func DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// ReadFile 读取文件内容
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadFileString 读取文件内容为字符串
func ReadFileString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile 写入文件
func WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// WriteFileString 写入字符串到文件
func WriteFileString(path, content string) error {
	return WriteFile(path, []byte(content))
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, sourceFile)
	return err
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

// GetExtension 获取文件扩展名
func GetExtension(path string) string {
	ext := filepath.Ext(path)
	if len(ext) > 0 && ext[0] == '.' {
		return ext[1:]
	}
	return ext
}

// GetNameWithoutExtension 获取不带扩展名的文件名
func GetNameWithoutExtension(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	if ext != "" {
		return name[:len(name)-len(ext)]
	}
	return name
}

// GetFileSize 获取文件大小
func GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return -1
	}
	return info.Size()
}

// HumanReadableSize 将字节数转换为可读格式
func HumanReadableSize(bytes int64) string {
	if bytes < 0 {
		return "N/A"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(bytes)
	unitIndex := 0

	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

// ListFiles 列出目录下的文件
func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files, nil
}

// ListFilesWithExt 列出目录下指定扩展名的文件
func ListFilesWithExt(dir, ext string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ext) {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// WalkDir 遍历目录
func WalkDir(root string, fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return fn(path, info)
	})
}

// TempFile 创建临时文件
func TempFile(pattern string) (*os.File, error) {
	return os.CreateTemp("", pattern)
}

// TempDir 创建临时目录
func TempDir(pattern string) (string, error) {
	return os.MkdirTemp("", pattern)
}