package util

import (
	"io"
)

// CopyStream 复制流
func CopyStream(src io.Reader, dst io.Writer) (int64, error) {
	return io.Copy(dst, src)
}

// CopyStreamN 复制指定数量的字节
func CopyStreamN(src io.Reader, dst io.Writer, n int64) (int64, error) {
	return io.CopyN(dst, src, n)
}

// ReadAll 读取所有内容
func ReadAll(src io.Reader) ([]byte, error) {
	return io.ReadAll(src)
}

// ReadString 读取为字符串
func ReadString(src io.Reader) (string, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteAll 写入所有内容
func WriteAll(dst io.Writer, data []byte) error {
	_, err := dst.Write(data)
	return err
}

// WriteString 写入字符串
func WriteString(dst io.Writer, s string) error {
	_, err := io.WriteString(dst, s)
	return err
}

// LimitReader 限制读取的长度
func LimitReader(src io.Reader, n int64) io.Reader {
	return io.LimitReader(src, n)
}

// TeeReader 创建TeeReader（同时读取和写入）
func TeeReader(src io.Reader, dst io.Writer) io.Reader {
	return io.TeeReader(src, dst)
}

// MultiReader 创建多读者
func MultiReader(readers ...io.Reader) io.Reader {
	return io.MultiReader(readers...)
}

// MultiWriter 创建多写者
func MultiWriter(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}

// Pipe 创建管道
func Pipe() (*io.PipeReader, *io.PipeWriter) {
	return io.Pipe()
}

// NopCloser 创建无操作的Close
func NopCloser(r io.Reader) io.ReadCloser {
	return io.NopCloser(r)
}