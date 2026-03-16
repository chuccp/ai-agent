package value

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

	AsObject() *ObjectValue
	AsArray() *ArrayValue
	AsText() *TextValue
	AsBool() *BoolValue
	AsNumber() *NumberValue
	AsUrls() *UrlsValue
	AsFiles() *FilesValue

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

func fromInterface(v interface{}) NodeValue {
	if v == nil {
		return NullValue
	}

	switch val := v.(type) {
	case bool:
		return NewBoolValue(val)
	case float64:
		return NewNumberValue(val)
	case string:
		return NewTextValue(val)
	case []interface{}:
		arr := NewArrayValue()
		for _, item := range val {
			arr.Add(fromInterface(item))
		}
		return arr
	case map[string]interface{}:
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

func (NodeValueBase) IsObject() bool       { return false }
func (NodeValueBase) IsArray() bool        { return false }
func (NodeValueBase) IsText() bool         { return false }
func (NodeValueBase) IsBool() bool         { return false }
func (NodeValueBase) IsNumber() bool       { return false }
func (NodeValueBase) IsNull() bool         { return false }
func (NodeValueBase) IsUrls() bool         { return false }
func (NodeValueBase) IsFiles() bool        { return false }
func (NodeValueBase) AsObject() *ObjectValue { panic("not an object") }
func (NodeValueBase) AsArray() *ArrayValue  { panic("not an array") }
func (NodeValueBase) AsText() *TextValue    { panic("not text") }
func (NodeValueBase) AsBool() *BoolValue    { panic("not bool") }
func (NodeValueBase) AsNumber() *NumberValue { panic("not number") }
func (NodeValueBase) AsUrls() *UrlsValue    { panic("not urls") }
func (NodeValueBase) AsFiles() *FilesValue  { panic("not files") }
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
	data, _ := json.Marshal(b.Value)
	return data
}

// NumberValue 数值
type NumberValue struct {
	NodeValueBase
	Value float64
}

func NewNumberValue(value float64) *NumberValue {
	return &NumberValue{Value: value}
}

func (n *NumberValue) IsNumber() bool {
	return true
}

func (n *NumberValue) AsNumber() *NumberValue {
	return n
}

func (n *NumberValue) String() string {
	return fmt.Sprintf("%v", n.Value)
}

func (n *NumberValue) ToJSON() json.RawMessage {
	data, _ := json.Marshal(n.Value)
	return data
}

func (n *NumberValue) Int() int {
	return int(n.Value)
}

func (n *NumberValue) Int64() int64 {
	return int64(n.Value)
}

func (n *NumberValue) Float64() float64 {
	return n.Value
}