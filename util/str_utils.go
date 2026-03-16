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

// IsEmpty 判断字符串是否为空
func IsEmpty(s string) bool {
	return s == ""
}

// IsNotEmpty 判断字符串是否不为空
func IsNotEmpty(s string) bool {
	return s != ""
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

// Equals 判断两个字符串是否相等（区分大小写）
func Equals(str1, str2 string) bool {
	return str1 == str2
}

// EqualsIgnoreCase 判断两个字符串是否相等（不区分大小写）
func EqualsIgnoreCase(str1, str2 string) bool {
	return strings.EqualFold(str1, str2)
}

// EqualsAny 判断字符串是否等于任意一个目标字符串（区分大小写）
func EqualsAny(str string, targets ...string) bool {
	for _, target := range targets {
		if str == target {
			return true
		}
	}
	return false
}

// EqualsAnyIgnoreCase 判断字符串是否等于任意一个目标字符串（不区分大小写）
func EqualsAnyIgnoreCase(str string, targets ...string) bool {
	for _, target := range targets {
		if strings.EqualFold(str, target) {
			return true
		}
	}
	return false
}

// Contains 检查字符串是否包含子串
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ContainsIgnoreCase 判断字符串是否包含子串（不区分大小写）
func ContainsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// ContainsAny 判断字符串是否包含任意一个子串（区分大小写）
func ContainsAny(str string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}

// ContainsAnyIgnoreCase 判断字符串是否包含任意一个子串（不区分大小写）
func ContainsAnyIgnoreCase(str string, substrs ...string) bool {
	lowerStr := strings.ToLower(str)
	for _, substr := range substrs {
		if strings.Contains(lowerStr, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// ContainsAll 判断字符串是否包含所有子串（区分大小写）
func ContainsAll(str string, substrs ...string) bool {
	for _, substr := range substrs {
		if !strings.Contains(str, substr) {
			return false
		}
	}
	return true
}

// ContainsAllIgnoreCase 判断字符串是否包含所有子串（不区分大小写）
func ContainsAllIgnoreCase(str string, substrs ...string) bool {
	lowerStr := strings.ToLower(str)
	for _, substr := range substrs {
		if !strings.Contains(lowerStr, strings.ToLower(substr)) {
			return false
		}
	}
	return true
}

// StartsWith 判断字符串是否以指定前缀开头
func StartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// EndsWith 判断字符串是否以指定后缀结尾
func EndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// Trim 去除字符串首尾空白字符
func Trim(str string) string {
	return strings.TrimSpace(str)
}

// DefaultIfEmpty 如果字符串为空则返回默认值
func DefaultIfEmpty(str, defaultValue string) string {
	if str == "" {
		return defaultValue
	}
	return str
}

// DefaultIfBlank 如果字符串为空或全是空白则返回默认值
func DefaultIfBlank(str, defaultValue string) string {
	if IsBlank(str) {
		return defaultValue
	}
	return str
}

// Substring 截取子字符串（支持负数索引，-1 表示最后一个字符）
func Substring(str string, start, end int) string {
	if str == "" {
		return ""
	}
	length := len(str)
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start >= end {
		return ""
	}
	return str[start:end]
}

// Left 获取字符串左边指定长度的子串
func Left(str string, length int) string {
	if str == "" || length <= 0 {
		return ""
	}
	if length >= len(str) {
		return str
	}
	return str[:length]
}

// Right 获取字符串右边指定长度的子串
func Right(str string, length int) string {
	if str == "" || length <= 0 {
		return ""
	}
	if length >= len(str) {
		return str
	}
	return str[len(str)-length:]
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

// IsAnyBlank 判断是否有任意一个字符串为空或全是空白
func IsAnyBlank(strs ...string) bool {
	for _, str := range strs {
		if IsBlank(str) {
			return true
		}
	}
	return false
}

// IsAllBlank 判断所有字符串是否都为空或全是空白
func IsAllBlank(strs ...string) bool {
	for _, str := range strs {
		if IsNotBlank(str) {
			return false
		}
	}
	return true
}

// IsNoneBlank 判断所有字符串都不为空且不全是空白
func IsNoneBlank(strs ...string) bool {
	for _, str := range strs {
		if IsBlank(str) {
			return false
		}
	}
	return true
}

// RemoveBracketAndContent 移除括号【】及其内容
func RemoveBracketAndContent(str string) string {
	if str == "" {
		return str
	}
	result := str
	for {
		start := strings.Index(result, "【")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "】")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

// ExtractParagraphList 提取段落列表，按换行符分割并去除空白
func ExtractParagraphList(content string) []string {
	if content == "" {
		return nil
	}

	paragraphs := strings.Split(content, "\n")
	result := make([]string, 0)
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}