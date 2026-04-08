package value

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// NodeValue 节点值接口
type NodeValue interface {
	IsObject() bool
	IsArray() bool
	IsText() bool
	IsBool() bool
	IsNumber() bool
	IsNull() bool
	IsUrls() bool
	IsFiles() bool
	IsStream() bool

	AsObject() *ObjectValue
	AsArray() *ArrayValue
	AsText() *TextValue
	AsBool() *BoolValue
	AsNumber() *NumberValue
	AsUrls() *UrlsValue
	AsFiles() *FilesValue
	AsStream() *StreamNodeValue

	FindValue(path string) NodeValue
	ToJSON() json.RawMessage
	String() string
}

// FindValueByPath 根据路径查找值
func FindValueByPath(current NodeValue, path string) NodeValue {
	if path == "" {
		return current
	}

	normalizedPath := path
	if strings.HasPrefix(path, "$.") {
		normalizedPath = path[2:]
	}

	segments := splitPath(normalizedPath)

	for _, segment := range segments {
		if current == nil {
			return nil
		}

		// 处理数组索引 [0], [1] 等
		if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
			if !current.IsArray() {
				return nil
			}
			indexStr := segment[1 : len(segment)-1]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil
			}
			current = current.AsArray().Get(index)
			continue
		}

		// 处理对象字段
		if !current.IsObject() {
			return nil
		}
		current = current.AsObject().Get(segment)
	}

	return current
}

// splitPath 分割路径
func splitPath(path string) []string {
	var segments []string
	var current strings.Builder

	for i := 0; i < len(path); i++ {
		c := path[i]
		if c == '.' {
			if current.Len() > 0 {
				segments = append(segments, current.String())
				current.Reset()
			}
			continue
		}

		if c == '[' {
			if current.Len() > 0 {
				segments = append(segments, current.String())
				current.Reset()
			}
			current.WriteByte(c)
			i++
			for i < len(path) && path[i] != ']' {
				current.WriteByte(path[i])
				i++
			}
			if i < len(path) {
				current.WriteByte(']')
			}
			segments = append(segments, current.String())
			current.Reset()
			continue
		}

		current.WriteByte(c)
	}

	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	return segments
}

// FromJSON 从JSON解析NodeValue
func FromJSON(data []byte) (NodeValue, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return fromInterface(v), nil
}

func fromInterface(v any) NodeValue {
	if v == nil {
		return NullValue
	}

	switch val := v.(type) {
	case NodeValue:
		return val
	case bool:
		return NewBoolValue(val)
	case float64:
		return NewNumberValue(val)
	case float32:
		return NewFloat32Value(val)
	case int:
		return NewIntValue(val)
	case int8:
		return NewInt8Value(val)
	case int16:
		return NewInt16Value(val)
	case int32:
		return NewInt32Value(val)
	case int64:
		return NewInt64Value(val)
	case uint:
		return NewUintValue(val)
	case uint8:
		return NewUint8Value(val)
	case uint16:
		return NewUint16Value(val)
	case uint32:
		return NewUint32Value(val)
	case uint64:
		return NewUint64Value(val)
	case string:
		return NewTextValue(val)
	case []any:
		arr := NewArrayValue()
		for _, item := range val {
			arr.Add(fromInterface(item))
		}
		return arr
	case map[string]any:
		obj := NewObjectValue()
		for k, item := range val {
			obj.Put(k, fromInterface(item))
		}
		return obj
	default:
		return NullValue
	}
}

// NodeValueBase 基础实现
type NodeValueBase struct{}

func (NodeValueBase) IsObject() bool             { return false }
func (NodeValueBase) IsArray() bool              { return false }
func (NodeValueBase) IsText() bool               { return false }
func (NodeValueBase) IsBool() bool               { return false }
func (NodeValueBase) IsNumber() bool             { return false }
func (NodeValueBase) IsNull() bool               { return false }
func (NodeValueBase) IsUrls() bool               { return false }
func (NodeValueBase) IsFiles() bool              { return false }
func (NodeValueBase) IsStream() bool             { return false }
func (NodeValueBase) AsObject() *ObjectValue     { panic("not an object") }
func (NodeValueBase) AsArray() *ArrayValue       { panic("not an array") }
func (NodeValueBase) AsText() *TextValue         { panic("not text") }
func (NodeValueBase) AsBool() *BoolValue         { panic("not bool") }
func (NodeValueBase) AsNumber() *NumberValue     { panic("not number") }
func (NodeValueBase) AsUrls() *UrlsValue         { panic("not urls") }
func (NodeValueBase) AsFiles() *FilesValue       { panic("not files") }
func (NodeValueBase) AsStream() *StreamNodeValue { panic("not a stream") }
func (NodeValueBase) FindValue(path string) NodeValue {
	return FindValueByPath(nil, path)
}
func (NodeValueBase) ToJSON() json.RawMessage {
	return json.RawMessage("null")
}
func (NodeValueBase) String() string {
	return "null"
}

// 确保 NullNodeValue 等类型实现 NodeValue 接口
var _ NodeValue = (*NullNodeValue)(nil)
var _ NodeValue = (*TextValue)(nil)
var _ NodeValue = (*BoolValue)(nil)
var _ NodeValue = (*NumberValue)(nil)
var _ NodeValue = (*ObjectValue)(nil)
var _ NodeValue = (*ArrayValue)(nil)
var _ NodeValue = (*UrlsValue)(nil)
var _ NodeValue = (*FilesValue)(nil)
var _ NodeValue = (*StreamNodeValue)(nil)

// StreamNodeValue 流节点值
type StreamNodeValue struct {
	NodeValueBase
	ch   chan NodeValue
	done bool
	mu   sync.Mutex
}

// NewStreamNodeValue 创建流节点值
func NewStreamNodeValue() *StreamNodeValue {
	return &StreamNodeValue{
		ch: make(chan NodeValue, 100),
	}
}

func (s *StreamNodeValue) IsStream() bool {
	return true
}

func (s *StreamNodeValue) AsStream() *StreamNodeValue {
	return s
}

// Send 发送值
func (s *StreamNodeValue) Send(v NodeValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return
	}
	s.ch <- v
}

// Close 关闭流
func (s *StreamNodeValue) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return
	}
	s.done = true
	close(s.ch)
}

// Channel 获取通道
func (s *StreamNodeValue) Channel() <-chan NodeValue {
	return s.ch
}

// IsDone 是否完成
func (s *StreamNodeValue) IsDone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.done
}

// String 返回字符串表示
func (s *StreamNodeValue) String() string {
	return "stream"
}

// ToJSON 返回JSON表示，将流中的所有值收集为数组
func (s *StreamNodeValue) ToJSON() json.RawMessage {
	arr := s.Collect()
	return arr.ToJSON()
}

// Collect 收集所有值到数组
func (s *StreamNodeValue) Collect() *ArrayValue {
	arr := NewArrayValue()
	for v := range s.ch {
		arr.Add(v)
	}
	return arr
}

// NullNodeValue 空值
type NullNodeValue struct {
	NodeValueBase
}

func (n *NullNodeValue) IsNull() bool {
	return true
}

func (n *NullNodeValue) String() string {
	return "null"
}

func (n *NullNodeValue) ToJSON() json.RawMessage {
	return json.RawMessage("null")
}

var NullValue = &NullNodeValue{}

// TextValue 文本值
type TextValue struct {
	NodeValueBase
	Text string
}

func NewTextValue(text string) *TextValue {
	return &TextValue{Text: text}
}

func (t *TextValue) IsText() bool {
	return true
}
func (t *TextValue) Append(text string) *TextValue {
	t.Text += text
	return t
}
func (t *TextValue) AsText() *TextValue {
	return t
}

func (t *TextValue) String() string {
	return t.Text
}

func (t *TextValue) ToJSON() json.RawMessage {
	data, _ := json.Marshal(t.Text)
	return data
}

func (t *TextValue) Equals(other NodeValue) bool {
	if other == nil {
		return false
	}
	if other.IsText() {
		return t.Text == other.AsText().Text
	}
	return false
}

// BoolValue 布尔值
type BoolValue struct {
	NodeValueBase
	Value bool
}

func NewBoolValue(value bool) *BoolValue {
	return &BoolValue{Value: value}
}

func (b *BoolValue) IsBool() bool {
	return true
}

func (b *BoolValue) AsBool() *BoolValue {
	return b
}

func (b *BoolValue) String() string {
	return fmt.Sprintf("%v", b.Value)
}

func (b *BoolValue) ToJSON() json.RawMessage {
	if b.Value {
		return json.RawMessage("true")
	}
	return json.RawMessage("false")
}

// NumberKind 表示数字的具体类型
type NumberKind int

const (
	KindInt NumberKind = iota
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindFloat32
	KindFloat64
)

// NumberValue 数值，支持多种数字类型
type NumberValue struct {
	NodeValueBase
	kind  NumberKind
	value interface{} // 存储实际值
}

// 构造函数 - 各类型
func NewNumberValue(value float64) *NumberValue {
	return &NumberValue{kind: KindFloat64, value: value}
}

func NewIntValue(value int) *NumberValue {
	return &NumberValue{kind: KindInt, value: value}
}

func NewInt8Value(value int8) *NumberValue {
	return &NumberValue{kind: KindInt8, value: value}
}

func NewInt16Value(value int16) *NumberValue {
	return &NumberValue{kind: KindInt16, value: value}
}

func NewInt32Value(value int32) *NumberValue {
	return &NumberValue{kind: KindInt32, value: value}
}

func NewInt64Value(value int64) *NumberValue {
	return &NumberValue{kind: KindInt64, value: value}
}

func NewUintValue(value uint) *NumberValue {
	return &NumberValue{kind: KindUint, value: value}
}

func NewUint8Value(value uint8) *NumberValue {
	return &NumberValue{kind: KindUint8, value: value}
}

func NewUint16Value(value uint16) *NumberValue {
	return &NumberValue{kind: KindUint16, value: value}
}

func NewUint32Value(value uint32) *NumberValue {
	return &NumberValue{kind: KindUint32, value: value}
}

func NewUint64Value(value uint64) *NumberValue {
	return &NumberValue{kind: KindUint64, value: value}
}

func NewFloat32Value(value float32) *NumberValue {
	return &NumberValue{kind: KindFloat32, value: value}
}

func NewFloat64Value(value float64) *NumberValue {
	return &NumberValue{kind: KindFloat64, value: value}
}

func (n *NumberValue) IsNumber() bool {
	return true
}

func (n *NumberValue) AsNumber() *NumberValue {
	return n
}

// Kind 返回数字的具体类型
func (n *NumberValue) Kind() NumberKind {
	return n.kind
}

// IsInt 系列方法 - 判断是否为特定类型
func (n *NumberValue) IsInt() bool     { return n.kind == KindInt }
func (n *NumberValue) IsInt8() bool    { return n.kind == KindInt8 }
func (n *NumberValue) IsInt16() bool   { return n.kind == KindInt16 }
func (n *NumberValue) IsInt32() bool   { return n.kind == KindInt32 }
func (n *NumberValue) IsInt64() bool   { return n.kind == KindInt64 }
func (n *NumberValue) IsUint() bool    { return n.kind == KindUint }
func (n *NumberValue) IsUint8() bool   { return n.kind == KindUint8 }
func (n *NumberValue) IsUint16() bool  { return n.kind == KindUint16 }
func (n *NumberValue) IsUint32() bool  { return n.kind == KindUint32 }
func (n *NumberValue) IsUint64() bool  { return n.kind == KindUint64 }
func (n *NumberValue) IsFloat32() bool { return n.kind == KindFloat32 }
func (n *NumberValue) IsFloat64() bool { return n.kind == KindFloat64 }

// IsInteger 判断是否为整数类型（有符号或无符号）
func (n *NumberValue) IsInteger() bool {
	return n.kind >= KindInt && n.kind <= KindUint64
}

// IsFloat 判断是否为浮点类型
func (n *NumberValue) IsFloat() bool {
	return n.kind == KindFloat32 || n.kind == KindFloat64
}

// IsSigned 判断是否为有符号整数类型
func (n *NumberValue) IsSigned() bool {
	return n.kind >= KindInt && n.kind <= KindInt64
}

// IsUnsigned 判断是否为无符号整数类型
func (n *NumberValue) IsUnsigned() bool {
	return n.kind >= KindUint && n.kind <= KindUint64
}

// Value 获取原始值
func (n *NumberValue) Value() interface{} {
	return n.value
}

// String 返回字符串表示
func (n *NumberValue) String() string {
	return fmt.Sprintf("%v", n.value)
}

// ToJSON 返回JSON表示
func (n *NumberValue) ToJSON() json.RawMessage {
	data, _ := json.Marshal(n.value)
	return data
}

// 类型安全的获取方法 - 返回原始类型值
func (n *NumberValue) Int() int {
	switch n.kind {
	case KindInt:
		return n.value.(int)
	case KindInt8:
		return int(n.value.(int8))
	case KindInt16:
		return int(n.value.(int16))
	case KindInt32:
		return int(n.value.(int32))
	case KindInt64:
		return int(n.value.(int64))
	case KindUint:
		return int(n.value.(uint))
	case KindUint8:
		return int(n.value.(uint8))
	case KindUint16:
		return int(n.value.(uint16))
	case KindUint32:
		return int(n.value.(uint32))
	case KindUint64:
		return int(n.value.(uint64))
	case KindFloat32:
		return int(n.value.(float32))
	case KindFloat64:
		return int(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Int8() int8 {
	switch n.kind {
	case KindInt:
		return int8(n.value.(int))
	case KindInt8:
		return n.value.(int8)
	case KindInt16:
		return int8(n.value.(int16))
	case KindInt32:
		return int8(n.value.(int32))
	case KindInt64:
		return int8(n.value.(int64))
	case KindUint:
		return int8(n.value.(uint))
	case KindUint8:
		return int8(n.value.(uint8))
	case KindUint16:
		return int8(n.value.(uint16))
	case KindUint32:
		return int8(n.value.(uint32))
	case KindUint64:
		return int8(n.value.(uint64))
	case KindFloat32:
		return int8(n.value.(float32))
	case KindFloat64:
		return int8(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Int16() int16 {
	switch n.kind {
	case KindInt:
		return int16(n.value.(int))
	case KindInt8:
		return int16(n.value.(int8))
	case KindInt16:
		return n.value.(int16)
	case KindInt32:
		return int16(n.value.(int32))
	case KindInt64:
		return int16(n.value.(int64))
	case KindUint:
		return int16(n.value.(uint))
	case KindUint8:
		return int16(n.value.(uint8))
	case KindUint16:
		return int16(n.value.(uint16))
	case KindUint32:
		return int16(n.value.(uint32))
	case KindUint64:
		return int16(n.value.(uint64))
	case KindFloat32:
		return int16(n.value.(float32))
	case KindFloat64:
		return int16(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Int32() int32 {
	switch n.kind {
	case KindInt:
		return int32(n.value.(int))
	case KindInt8:
		return int32(n.value.(int8))
	case KindInt16:
		return int32(n.value.(int16))
	case KindInt32:
		return n.value.(int32)
	case KindInt64:
		return int32(n.value.(int64))
	case KindUint:
		return int32(n.value.(uint))
	case KindUint8:
		return int32(n.value.(uint8))
	case KindUint16:
		return int32(n.value.(uint16))
	case KindUint32:
		return int32(n.value.(uint32))
	case KindUint64:
		return int32(n.value.(uint64))
	case KindFloat32:
		return int32(n.value.(float32))
	case KindFloat64:
		return int32(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Int64() int64 {
	switch n.kind {
	case KindInt:
		return int64(n.value.(int))
	case KindInt8:
		return int64(n.value.(int8))
	case KindInt16:
		return int64(n.value.(int16))
	case KindInt32:
		return int64(n.value.(int32))
	case KindInt64:
		return n.value.(int64)
	case KindUint:
		return int64(n.value.(uint))
	case KindUint8:
		return int64(n.value.(uint8))
	case KindUint16:
		return int64(n.value.(uint16))
	case KindUint32:
		return int64(n.value.(uint32))
	case KindUint64:
		return int64(n.value.(uint64))
	case KindFloat32:
		return int64(n.value.(float32))
	case KindFloat64:
		return int64(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Uint() uint {
	switch n.kind {
	case KindInt:
		return uint(n.value.(int))
	case KindInt8:
		return uint(n.value.(int8))
	case KindInt16:
		return uint(n.value.(int16))
	case KindInt32:
		return uint(n.value.(int32))
	case KindInt64:
		return uint(n.value.(int64))
	case KindUint:
		return n.value.(uint)
	case KindUint8:
		return uint(n.value.(uint8))
	case KindUint16:
		return uint(n.value.(uint16))
	case KindUint32:
		return uint(n.value.(uint32))
	case KindUint64:
		return uint(n.value.(uint64))
	case KindFloat32:
		return uint(n.value.(float32))
	case KindFloat64:
		return uint(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Uint8() uint8 {
	switch n.kind {
	case KindInt:
		return uint8(n.value.(int))
	case KindInt8:
		return uint8(n.value.(int8))
	case KindInt16:
		return uint8(n.value.(int16))
	case KindInt32:
		return uint8(n.value.(int32))
	case KindInt64:
		return uint8(n.value.(int64))
	case KindUint:
		return uint8(n.value.(uint))
	case KindUint8:
		return n.value.(uint8)
	case KindUint16:
		return uint8(n.value.(uint16))
	case KindUint32:
		return uint8(n.value.(uint32))
	case KindUint64:
		return uint8(n.value.(uint64))
	case KindFloat32:
		return uint8(n.value.(float32))
	case KindFloat64:
		return uint8(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Uint16() uint16 {
	switch n.kind {
	case KindInt:
		return uint16(n.value.(int))
	case KindInt8:
		return uint16(n.value.(int8))
	case KindInt16:
		return uint16(n.value.(int16))
	case KindInt32:
		return uint16(n.value.(int32))
	case KindInt64:
		return uint16(n.value.(int64))
	case KindUint:
		return uint16(n.value.(uint))
	case KindUint8:
		return uint16(n.value.(uint8))
	case KindUint16:
		return n.value.(uint16)
	case KindUint32:
		return uint16(n.value.(uint32))
	case KindUint64:
		return uint16(n.value.(uint64))
	case KindFloat32:
		return uint16(n.value.(float32))
	case KindFloat64:
		return uint16(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Uint32() uint32 {
	switch n.kind {
	case KindInt:
		return uint32(n.value.(int))
	case KindInt8:
		return uint32(n.value.(int8))
	case KindInt16:
		return uint32(n.value.(int16))
	case KindInt32:
		return uint32(n.value.(int32))
	case KindInt64:
		return uint32(n.value.(int64))
	case KindUint:
		return uint32(n.value.(uint))
	case KindUint8:
		return uint32(n.value.(uint8))
	case KindUint16:
		return uint32(n.value.(uint16))
	case KindUint32:
		return n.value.(uint32)
	case KindUint64:
		return uint32(n.value.(uint64))
	case KindFloat32:
		return uint32(n.value.(float32))
	case KindFloat64:
		return uint32(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Uint64() uint64 {
	switch n.kind {
	case KindInt:
		return uint64(n.value.(int))
	case KindInt8:
		return uint64(n.value.(int8))
	case KindInt16:
		return uint64(n.value.(int16))
	case KindInt32:
		return uint64(n.value.(int32))
	case KindInt64:
		return uint64(n.value.(int64))
	case KindUint:
		return uint64(n.value.(uint))
	case KindUint8:
		return uint64(n.value.(uint8))
	case KindUint16:
		return uint64(n.value.(uint16))
	case KindUint32:
		return uint64(n.value.(uint32))
	case KindUint64:
		return n.value.(uint64)
	case KindFloat32:
		return uint64(n.value.(float32))
	case KindFloat64:
		return uint64(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Float32() float32 {
	switch n.kind {
	case KindInt:
		return float32(n.value.(int))
	case KindInt8:
		return float32(n.value.(int8))
	case KindInt16:
		return float32(n.value.(int16))
	case KindInt32:
		return float32(n.value.(int32))
	case KindInt64:
		return float32(n.value.(int64))
	case KindUint:
		return float32(n.value.(uint))
	case KindUint8:
		return float32(n.value.(uint8))
	case KindUint16:
		return float32(n.value.(uint16))
	case KindUint32:
		return float32(n.value.(uint32))
	case KindUint64:
		return float32(n.value.(uint64))
	case KindFloat32:
		return n.value.(float32)
	case KindFloat64:
		return float32(n.value.(float64))
	default:
		return 0
	}
}

func (n *NumberValue) Float64() float64 {
	switch n.kind {
	case KindInt:
		return float64(n.value.(int))
	case KindInt8:
		return float64(n.value.(int8))
	case KindInt16:
		return float64(n.value.(int16))
	case KindInt32:
		return float64(n.value.(int32))
	case KindInt64:
		return float64(n.value.(int64))
	case KindUint:
		return float64(n.value.(uint))
	case KindUint8:
		return float64(n.value.(uint8))
	case KindUint16:
		return float64(n.value.(uint16))
	case KindUint32:
		return float64(n.value.(uint32))
	case KindUint64:
		return float64(n.value.(uint64))
	case KindFloat32:
		return float64(n.value.(float32))
	case KindFloat64:
		return n.value.(float64)
	default:
		return 0
	}
}
func (n *NumberValue) Equals(other *NumberValue) bool {
	if other == nil {
		return false
	}
	// 如果类型相同，直接比较原始值
	if n.kind == other.kind {
		return n.value == other.value
	}
	// 不同类型比较：统一转为 Float64 进行数值比较
	return n.Float64() == other.Float64()
}

// ParseNumber 从字符串解析数值
func ParseNumber(s string) (*NumberValue, error) {
	// 先尝试解析为整数
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return NewInt64Value(i), nil
	}
	// 尝试解析为浮点数
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return NewFloat64Value(f), nil
	}
	return nil, fmt.Errorf("cannot parse %q as number", s)
}

// MustNumber 从字符串解析数值，失败返回0
func MustNumber(s string) *NumberValue {
	n, err := ParseNumber(s)
	if err != nil {
		return NewIntValue(0)
	}
	return n
}

// Equals 判断两个NodeValue是否相等
func Equals(a, b NodeValue) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 类型不同则不相等
	if a.IsText() != b.IsText() ||
		a.IsNumber() != b.IsNumber() ||
		a.IsBool() != b.IsBool() ||
		a.IsNull() != b.IsNull() ||
		a.IsObject() != b.IsObject() ||
		a.IsArray() != b.IsArray() {
		return false
	}

	switch {
	case a.IsText():
		return a.AsText().Text == b.AsText().Text
	case a.IsNumber():
		// 数值比较：先比较类型，再比较值
		na, nb := a.AsNumber(), b.AsNumber()
		if na.Kind() != nb.Kind() {
			return na.Float64() == nb.Float64()
		}
		switch na.Kind() {
		case KindInt:
			return na.Int() == nb.Int()
		case KindInt8:
			return na.Int8() == nb.Int8()
		case KindInt16:
			return na.Int16() == nb.Int16()
		case KindInt32:
			return na.Int32() == nb.Int32()
		case KindInt64:
			return na.Int64() == nb.Int64()
		case KindUint:
			return na.Uint() == nb.Uint()
		case KindUint8:
			return na.Uint8() == nb.Uint8()
		case KindUint16:
			return na.Uint16() == nb.Uint16()
		case KindUint32:
			return na.Uint32() == nb.Uint32()
		case KindUint64:
			return na.Uint64() == nb.Uint64()
		case KindFloat32:
			return na.Float32() == nb.Float32()
		case KindFloat64:
			return na.Float64() == nb.Float64()
		default:
			return false
		}
	case a.IsBool():
		return a.AsBool().Value == b.AsBool().Value
	case a.IsNull():
		return true
	default:
		return false
	}
}

// Clone 克隆NodeValue
func Clone(v NodeValue) NodeValue {
	if v == nil {
		return nil
	}

	switch {
	case v.IsText():
		return NewTextValue(v.AsText().Text)
	case v.IsNumber():
		nv := v.AsNumber()
		// 克隆时保留原始类型
		switch nv.Kind() {
		case KindInt:
			return NewIntValue(nv.Int())
		case KindInt8:
			return NewInt8Value(nv.Int8())
		case KindInt16:
			return NewInt16Value(nv.Int16())
		case KindInt32:
			return NewInt32Value(nv.Int32())
		case KindInt64:
			return NewInt64Value(nv.Int64())
		case KindUint:
			return NewUintValue(nv.Uint())
		case KindUint8:
			return NewUint8Value(nv.Uint8())
		case KindUint16:
			return NewUint16Value(nv.Uint16())
		case KindUint32:
			return NewUint32Value(nv.Uint32())
		case KindUint64:
			return NewUint64Value(nv.Uint64())
		case KindFloat32:
			return NewFloat32Value(nv.Float32())
		case KindFloat64:
			return NewFloat64Value(nv.Float64())
		default:
			return NewFloat64Value(nv.Float64())
		}
	case v.IsBool():
		return NewBoolValue(v.AsBool().Value)
	case v.IsNull():
		return NullValue
	case v.IsObject():
		obj := v.AsObject()
		newObj := NewObjectValue()
		obj.ForEach(func(key string, val NodeValue) bool {
			newObj.Put(key, Clone(val))
			return true
		})
		return newObj
	case v.IsArray():
		arr := v.AsArray()
		newArr := NewArrayValue()
		arr.ForEach(func(index int, val NodeValue) bool {
			newArr.Add(Clone(val))
			return true
		})
		return newArr
	default:
		return NullValue
	}
}
