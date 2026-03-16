package util

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

// IsBlank 检查字符串是否为空白
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotBlank 检查字符串是否非空白
func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

// GenerateUUID 生成UUID
func GenerateUUID() string {
	id := uuid.New()
	return strings.ReplaceAll(id.String(), "-", "")
}

// GenerateID 生成ID
func GenerateID(prefix string) string {
	return prefix + "_" + GenerateUUID()
}

// MD5 计算MD5
func MD5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// Contains 检查字符串是否包含子串
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Join 连接字符串
func Join(sep string, parts ...string) string {
	return strings.Join(parts, sep)
}

// Split 分割字符串
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// SplitN 分割字符串，限制数量
func SplitN(s, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

// AppendSlice 追加切片
func AppendSlice[T any](slice []T, items ...T) []T {
	return append(slice, items...)
}