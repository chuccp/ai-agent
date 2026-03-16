package value

import (
	"encoding/json"
	"sync"
)

// ArrayValue 数组值
type ArrayValue struct {
	NodeValueBase
	values []NodeValue
	mu     sync.RWMutex
}

// NewArrayValue 创建数组值
func NewArrayValue() *ArrayValue {
	return &ArrayValue{
		values: make([]NodeValue, 0),
	}
}

// NewArrayValueFromSlice 从切片创建数组值
func NewArrayValueFromSlice(values []NodeValue) *ArrayValue {
	arr := NewArrayValue()
	arr.values = append(arr.values, values...)
	return arr
}

func (a *ArrayValue) IsArray() bool {
	return true
}

func (a *ArrayValue) AsArray() *ArrayValue {
	return a
}

// Add 添加值
func (a *ArrayValue) Add(value NodeValue) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if value != nil {
		a.values = append(a.values, value)
	}
}

// AddText 添加文本值
func (a *ArrayValue) AddText(text string) {
	a.Add(NewTextValue(text))
}

// AddBool 添加布尔值
func (a *ArrayValue) AddBool(value bool) {
	a.Add(NewBoolValue(value))
}

// AddNumber 添加数值
func (a *ArrayValue) AddNumber(value float64) {
	a.Add(NewNumberValue(value))
}

// AddAll 添加所有值
func (a *ArrayValue) AddAll(other *ArrayValue) {
	if other == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	a.values = append(a.values, other.values...)
}

// Get 获取值
func (a *ArrayValue) Get(index int) NodeValue {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if index < 0 || index >= len(a.values) {
		return NullValue
	}
	return a.values[index]
}

// Size 获取大小
func (a *ArrayValue) Size() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.values)
}

// IsEmpty 是否为空
func (a *ArrayValue) IsEmpty() bool {
	return a.Size() == 0
}

// Values 获取所有值
func (a *ArrayValue) Values() []NodeValue {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]NodeValue, len(a.values))
	copy(result, a.values)
	return result
}

// StringValues 获取所有字符串值
func (a *ArrayValue) StringValues() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]string, 0, len(a.values))
	for _, v := range a.values {
		if v.IsText() {
			result = append(result, v.AsText().Text)
		}
	}
	return result
}

// ForEach 遍历
func (a *ArrayValue) ForEach(fn func(index int, value NodeValue) bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for i, v := range a.values {
		if !fn(i, v) {
			break
		}
	}
}

// FindValue 查找值
func (a *ArrayValue) FindValue(path string) NodeValue {
	return FindValueByPath(a, path)
}

func (a *ArrayValue) String() string {
	return string(a.ToJSON())
}

func (a *ArrayValue) ToJSON() json.RawMessage {
	a.mu.RLock()
	defer a.mu.RUnlock()
	arr := make([]interface{}, 0, len(a.values))
	for _, v := range a.values {
		var val interface{}
		if err := json.Unmarshal(v.ToJSON(), &val); err == nil {
			arr = append(arr, val)
		}
	}
	data, _ := json.Marshal(arr)
	return data
}

// FromJSON 从JSON解析
func (a *ArrayValue) FromJSON(data []byte) error {
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.values = make([]NodeValue, 0, len(arr))
	for _, v := range arr {
		a.values = append(a.values, fromInterface(v))
	}
	return nil
}

// ParseArrayValue 从JSON解析数组值
func ParseArrayValue(data []byte) (*ArrayValue, error) {
	arr := NewArrayValue()
	if err := arr.FromJSON(data); err != nil {
		return nil, err
	}
	return arr, nil
}