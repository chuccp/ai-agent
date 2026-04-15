package value

import (
	"encoding/json"
	"net/url"
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

func (a *ArrayValue) AddAny(value any) *ArrayValue {
	a.Add(fromInterface(value))
	return a
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
func (a *ArrayValue) Has(value NodeValue) bool {
	for _, v := range a.values {
		if Equals(v, value) {
			return true
		}
	}
	return false
}
func (a *ArrayValue) HasString(value string) bool {
	for _, v := range a.values {
		if v.IsText() && v.AsText().Text == value {
			return true
		}
	}
	return false
}
func (a *ArrayValue) HasNumber(value float64) bool {
	for _, v := range a.values {
		if v.IsNumber() && v.AsNumber().Float64() == value {
			return true
		}
	}
	return false
}

func (a *ArrayValue) Find(value NodeValue) int {
	for i, v := range a.values {
		if Equals(v, value) {
			return i
		}
	}
	return -1
}
func (a *ArrayValue) Filter(fn func(index int, value NodeValue) bool) *ArrayValue {
	newArr := NewArrayValue()
	a.ForEach(func(index int, value NodeValue) bool {
		if fn(index, value) {
			newArr.Add(value)
		}
		return true
	})
	return newArr
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

// ToJSON 返回JSON字符串表示
func (a *ArrayValue) ToJSON() json.RawMessage {
	a.mu.RLock()
	defer a.mu.RUnlock()
	arr := make([]json.RawMessage, len(a.values))
	for i, v := range a.values {
		arr[i] = v.ToJSON()
	}
	data, _ := json.Marshal(arr)
	return data
}

func (a *ArrayValue) AsUrlsWithError() (urlsValue *UrlsValue, err error) {
	urlsValue = NewUrlsValue()
	a.ForEach(func(index int, value NodeValue) bool {
		u, err2 := url.Parse(value.AsText().Text)
		if err2 == nil {
			urlsValue.Add(*u)
		}
		if err == nil && err2 != nil {
			err = err2
		}
		return true
	})
	return urlsValue, err
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

func (a *ArrayValue) Clone() NodeValue {
	a.mu.RLock()
	defer a.mu.RUnlock()
	clone := NewArrayValue()
	for _, v := range a.values {
		clone.values = append(clone.values, cloneNodeValue(v))
	}
	return clone
}

// FindMaxByScore 根据评分函数找到得分最高的元素及其得分
// 返回值：元素、得分、是否找到
// scoreFunc: 评分函数，返回元素的得分值
func (a *ArrayValue) FindMaxByScore(scoreFunc func(value NodeValue) int) (NodeValue, int, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.values) == 0 {
		return nil, 0, false
	}

	maxScore := -1
	var maxValue NodeValue

	for _, v := range a.values {
		score := scoreFunc(v)
		if score > maxScore {
			maxScore = score
			maxValue = v
		}
	}

	return maxValue, maxScore, true
}
