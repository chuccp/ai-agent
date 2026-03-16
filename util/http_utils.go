package util

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadFile 下载文件到指定路径
func DownloadFile(url, path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return NewHTTPError(resp.StatusCode, resp.Status)
	}

	// 创建目标文件
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, resp.Body)
	return err
}

// DownloadFileToTemp 下载文件到临时目录
func DownloadFileToTemp(url, pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		os.Remove(file.Name())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(file.Name())
		return "", NewHTTPError(resp.StatusCode, resp.Status)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(file.Name())
		return "", err
	}

	return file.Name(), nil
}

// HTTPError HTTP错误
type HTTPError struct {
	StatusCode int
	Status     string
}

// NewHTTPError 创建HTTP错误
func NewHTTPError(statusCode int, status string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Status:     status,
	}
}

// Error 实现error接口
func (e *HTTPError) Error() string {
	return e.Status
}

// GetHTTPStatus 获取HTTP状态码
func (e *HTTPError) GetHTTPStatus() int {
	return e.StatusCode
}

// GetHTTPContent 获取HTTP内容
func GetHTTPContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHTTPError(resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// GetHTTPString 获取HTTP内容为字符串
func GetHTTPString(url string) (string, error) {
	data, err := GetHTTPContent(url)
	if err != nil {
		return "", err
	}
	return string(data), nil
}