package value

import (
	"encoding/json"
	"sync"
)

// ObjectValue 对象值
type ObjectValue struct {
	NodeValueBase
	data map[string]NodeValue
	mu   sync.RWMutex
}

// NewObjectValue 创建对象值
func NewObjectValue() *ObjectValue {
	return &ObjectValue{
		data: make(map[string]NodeValue),
	}
}

// NewObjectValueFromMap 从map创建对象值
func NewObjectValueFromMap(m map[string]interface{}) *ObjectValue {
	obj := NewObjectValue()
	for k, v := range m {
		obj.Put(k, fromInterface(v))
	}
	return obj
}

func (o *ObjectValue) IsObject() bool {
	return true
}

func (o *ObjectValue) AsObject() *ObjectValue {
	return o
}

// Get 获取值
func (o *ObjectValue) Get(key string) NodeValue {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.data[key]
}

// GetString 获取字符串值
func (o *ObjectValue) GetString(key string) string {
	v := o.Get(key)
	if v == nil || v.IsNull() {
		return ""
	}
	if v.IsText() {
		return v.AsText().Text
	}
	return v.String()
}

// GetObject 获取对象值
func (o *ObjectValue) GetObject(key string) *ObjectValue {
	v := o.Get(key)
	if v == nil || !v.IsObject() {
		return nil
	}
	return v.AsObject()
}

// GetArray 获取数组值
func (o *ObjectValue) GetArray(key string) *ArrayValue {
	v := o.Get(key)
	if v == nil || !v.IsArray() {
		return nil
	}
	return v.AsArray()
}

// Put 设置值
func (o *ObjectValue) Put(key string, value NodeValue) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.data[key] = value
}

// PutString 设置字符串值
func (o *ObjectValue) PutString(key, value string) {
	o.Put(key, NewTextValue(value))
}

// PutBool 设置布尔值
func (o *ObjectValue) PutBool(key string, value bool) {
	o.Put(key, NewBoolValue(value))
}

// PutNumber 设置数值
func (o *ObjectValue) PutNumber(key string, value float64) {
	o.Put(key, NewNumberValue(value))
}

// PutObject 设置对象值
func (o *ObjectValue) PutObject(key string, value *ObjectValue) {
	o.Put(key, value)
}

// PutArray 设置数组值
func (o *ObjectValue) PutArray(key string, value *ArrayValue) {
	o.Put(key, value)
}

// AddAll 添加所有值
func (o *ObjectValue) AddAll(other *ObjectValue) {
	if other == nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	for k, v := range other.data {
		o.data[k] = v
	}
}

// Clear 清空
func (o *ObjectValue) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.data = make(map[string]NodeValue)
}

// Keys 获取所有键
func (o *ObjectValue) Keys() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	keys := make([]string, 0, len(o.data))
	for k := range o.data {
		keys = append(keys, k)
	}
	return keys
}

// ForEach 遍历
func (o *ObjectValue) ForEach(fn func(key string, value NodeValue) bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for k, v := range o.data {
		if !fn(k, v) {
			break
		}
	}
}

// FindValue 查找值
func (o *ObjectValue) FindValue(path string) NodeValue {
	return FindValueByPath(o, path)
}

func (o *ObjectValue) String() string {
	return string(o.ToJSON())
}

func (o *ObjectValue) ToJSON() json.RawMessage {
	o.mu.RLock()
	defer o.mu.RUnlock()
	m := make(map[string]interface{})
	for k, v := range o.data {
		var val interface{}
		if err := json.Unmarshal(v.ToJSON(), &val); err == nil {
			m[k] = val
		}
	}
	data, _ := json.Marshal(m)
	return data
}

// ToMap 转换为map
func (o *ObjectValue) ToMap() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()
	m := make(map[string]interface{})
	for k, v := range o.data {
		var val interface{}
		if err := json.Unmarshal(v.ToJSON(), &val); err == nil {
			m[k] = val
		}
	}
	return m
}

// FromJSON 从JSON解析
func (o *ObjectValue) FromJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.data = make(map[string]NodeValue)
	for k, v := range m {
		o.data[k] = fromInterface(v)
	}
	return nil
}

// ParseObjectValue 从JSON解析对象值
func ParseObjectValue(data []byte) (*ObjectValue, error) {
	obj := NewObjectValue()
	if err := obj.FromJSON(data); err != nil {
		return nil, err
	}
	return obj, nil
}