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

// 【段落 1】1955年美国缅因州小镇的雪夜木屋，克理躺在床上听壁炉柴火噼啪响，窗玻璃突然映出两道移动的黄色光轨。【段落 2】木屋前的积雪小径上，复古红色蒸汽火车喷着白雾停下，金属梯子“哐当”落在雪地里。【段落 3】戴蓝呢帽的车长握着黄铜怀表，对推开门的克理弯腰：“极地特快，开往北极的专列，请上车。”【段落 4】火车第一节车厢内，克理坐在绿色丝绒座椅上，对面扎金色马尾的艾莎正举着热可可杯哈气。【段落 5】1950年代风格的火车餐车里，车长推着银质餐车走过，给克理递来印着雪花纹的瓷杯。【段落 6】艾莎指着窗外掠过的雪松林：“听说北极有会飞的驯鹿，圣老就坐在驯鹿拉的雪橇上。”【段落 7】结冰的湖面上，火车车轮碾过冰面发出“咯吱”声，克理抓紧座椅扶手发抖。【段落 8】艾莎轻轻碰克理的手背：“别怕，车长说这列火车走过一百年了，从没出过事。”【段落 9】火车穿过极光笼罩的山谷时，车顶传来“笃笃”声，克理撩开窗帘看见鹿角剪影。【段落 10】车长打开车顶天窗，探身笑着喊：“北极驯鹿来迎接客人啦，它们是圣老的得力助手。”【段落 11】1950年代复古站台的北极冰原车站，下车的克理踩在咯吱作响的冰面上，看见远处灯火闪烁。【段落 12】冰雕广场上，穿红绒衣的精灵们举着星星灯奔跑，艾莎拉着克理躲进冰柱后的雪堆。【段落 13】水晶穹顶的精灵工坊内，圣老正弯腰检查雪橇铃铛，银白胡子上沾着细碎的雪花。【段落 14】圣老转身看见克理，摘下沾着金粉的眼镜：“勇敢的孩子，想要什么圣诞礼物？”【段落 15】返程火车上，克理发现口袋里的银铃不见了，急得眼眶发红，艾莎帮他翻遍大衣口袋。【段落 16】克理的木屋床头，圣诞树下躺着个系红缎带的小盒子，打开后银铃在月光下闪着柔光。
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
