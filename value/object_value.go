package value

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/util"
	"github.com/spf13/cast"
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
func (o *ObjectValue) IsEmpty() bool {
	return len(o.data) == 0
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
	if v == nil {
		log.Panic("GetString: "+key+" not found ", errors.New(key+" not found"))
	}
	if v.IsNull() {
		return ""
	}
	if v.IsText() {
		return v.AsText().Text
	}
	return v.String()
}

func (o *ObjectValue) GetStringOrDefault(key string, defaultValue string) string {
	v := o.Get(key)
	if v == nil {
		return defaultValue
	}
	if v.IsNull() {
		return defaultValue
	}
	if v.IsText() {
		return v.AsText().Text
	}
	return v.String()
}

func (o *ObjectValue) GetOrString(keys ...string) string {
	has := false
	for _, key := range keys {
		v := o.Get(key)
		if v != nil {
			has = true
		}
		if v != nil && !v.IsNull() {
			if v.IsText() {
				text := v.AsText().Text
				if len(text) == 0 {
					continue
				}
				return text
			}
			return v.String()
		}
	}
	if !has {
		keyStr := strings.Join(keys, " or ")
		log.Panic("GetOrString: "+keyStr+" not found  ", errors.New(keyStr+" not found"))

	}
	return ""
}

// GetUrlsValue 获取URL值
func (o *ObjectValue) GetUrls(key string) *UrlsValue {
	v := o.Get(key)
	if v == nil {
		log.Panic("GetUrls: "+key+" not found ", errors.New(key+" not found"))
	}
	if !v.IsUrls() {
		return nil
	}
	return v.AsUrls()
}

func (o *ObjectValue) GetResources(key string) *ResourcesValue {
	v := o.Get(key)

	if v == nil {
		log.Panic("GetResources: "+key+" not found ", errors.New(key+" not found"))
	}

	if v.IsResources() {
		return v.AsResources()
	}
	if v.IsNumber() {
		resources := NewResourcesValue()
		resources.Add(v.AsNumber().String())
		return resources
	}
	if v.IsText() {
		resources := NewResourcesValue()
		resources.Add(v.AsText().Text)
		return resources
	}
	if v.IsArray() {
		resources := NewResourcesValue()
		v.AsArray().ForEach(func(index int, v NodeValue) bool {
			if v.IsText() {
				resources.Add(v.AsText().Text)
			}
			if v.IsNumber() {
				resources.Add(v.AsNumber().String())
			}
			if v.IsResources() {
				resources.AddAll(v.AsResources())
			}
			return true
		})
		return resources
	}

	return nil
}

// GetNumber 获取数值
func (o *ObjectValue) GetNumber(key string) float64 {
	v := o.Get(key)

	if v == nil {
		log.Panic("GetNumber: "+key+" not found ", errors.New(key+" not found"))
	}

	if v.IsNull() {
		return 0
	}
	if v.IsNumber() {
		return v.AsNumber().Float64()
	}
	return 0
}

// GetNumberOrDefault 获取数值，如果不存在返回默认值
func (o *ObjectValue) GetNumberOrDefault(key string, defaultValue float64) float64 {
	v := o.Get(key)
	if v == nil {
		return defaultValue
	}
	if v.IsNull() {
		return defaultValue
	}
	if v.IsNumber() {
		return v.AsNumber().Float64()
	}
	return defaultValue
}

// GetInt 获取整数值
func (o *ObjectValue) GetInt(key string) int {
	v := o.Get(key)
	if v == nil {
		log.Panic("GetInt: "+key+" not found ", errors.New(key+" not found"))
	}
	if v.IsNull() {
		return 0
	}
	if v.IsNumber() {
		return int(v.AsNumber().Int64())
	}
	return 0
}
func (o *ObjectValue) GetUInt(key string) uint {
	v := o.Get(key)
	if v == nil {
		log.Panic("GetInt: "+key+" not found ", errors.New(key+" not found"))
	}
	if v.IsNull() {
		return 0
	}
	if v.IsNumber() {
		return uint(v.AsNumber().Uint64())
	}
	if v.IsResources() {
		ids := v.AsResources().AsUints()
		if len(ids) > 0 {
			return ids[0]
		}
	}
	if v.IsText() {
		return cast.ToUint(v.AsText().Text)
	}
	return 0
}

// GetIntOrDefault 获取整数值，如果不存在返回默认值
func (o *ObjectValue) GetIntOrDefault(key string, defaultValue int) int {
	v := o.Get(key)
	if v == nil {
		return defaultValue
	}
	if v.IsNull() {
		return defaultValue
	}
	if v.IsNumber() {
		return int(v.AsNumber().Int64())
	}
	return defaultValue
}

// GetBool 获取布尔值
func (o *ObjectValue) GetBool(key string) bool {
	v := o.Get(key)
	if v == nil {
		log.Panic("GetBool: "+key+" not found ", errors.New(key+" not found"))
	}
	if v.IsNull() {
		return false
	}
	if v.IsBool() {
		return v.AsBool().Value
	}
	return false
}

// GetBoolOrDefault 获取布尔值，如果不存在返回默认值
func (o *ObjectValue) GetBoolOrDefault(key string, defaultValue bool) bool {
	v := o.Get(key)
	if v == nil {
		return defaultValue
	}
	if v.IsNull() {
		return defaultValue
	}
	if v.IsBool() {
		return v.AsBool().Value
	}
	return defaultValue
}

// GetObject 获取对象值
func (o *ObjectValue) GetObject(key string) *ObjectValue {
	v := o.Get(key)
	if v == nil {
		log.Panic("GetObject: "+key+" not found  ", errors.New(key+" not found"))
	}
	if !v.IsObject() {
		return nil
	}
	return v.AsObject()
}

// GetArray 获取数组值
func (o *ObjectValue) GetArray(key string) *ArrayValue {
	v := o.Get(key)
	if v == nil {
		log.Panic("GetArray: "+key+" not found  ", errors.New(key+" not found"))
	}
	if !v.IsArray() {
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
func (o *ObjectValue) PutAny(key string, value any) *ObjectValue {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.data[key] = fromInterface(value)
	return o
}

// PutString 设置字符串值
func (o *ObjectValue) PutString(key, value string) *ObjectValue {
	o.Put(key, NewTextValue(value))
	return o
}

// PutBool 设置布尔值
func (o *ObjectValue) PutBool(key string, value bool) *ObjectValue {
	o.Put(key, NewBoolValue(value))
	return o
}

// PutNumber 设置数值
func (o *ObjectValue) PutNumber(key string, value float64) *ObjectValue {
	o.Put(key, NewNumberValue(value))
	return o
}

func (o *ObjectValue) PutUint(key string, value uint) *ObjectValue {
	o.Put(key, NewNumberValue(float64(value)))
	return o
}

// PutObject 设置对象值
func (o *ObjectValue) PutObject(key string, value *ObjectValue) *ObjectValue {
	o.Put(key, value)
	return o
}

// PutArray 设置数组值
func (o *ObjectValue) PutArray(key string, value *ArrayValue) *ObjectValue {
	o.Put(key, value)
	return o
}

// AddAll 添加所有值
func (o *ObjectValue) AddAll(other *ObjectValue) *ObjectValue {
	if other == nil {
		return o
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	for k, v := range other.data {
		o.data[k] = v
	}
	return o
}
func (o *ObjectValue) AddAllIFNULL(other *ObjectValue) *ObjectValue {
	if other == nil {
		return o
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	for k, v := range other.data {
		v0, ok := o.data[k]
		if ok && v0 != nil {
			continue
		}
		o.data[k] = v
	}
	return o
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
func (o *ObjectValue) Clone() NodeValue {
	o.mu.RLock()
	defer o.mu.RUnlock()
	clone := NewObjectValue()
	for k, v := range o.data {
		clone.data[k] = cloneNodeValue(v)
	}
	return clone
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

// ToJSON 返回JSON字符串表示
func (o *ObjectValue) ToJSON() json.RawMessage {
	o.mu.RLock()
	defer o.mu.RUnlock()
	m := make(map[string]json.RawMessage)
	for k, v := range o.data {
		m[k] = v.ToJSON()
	}
	data, _ := json.Marshal(m)
	return data
}

// ToMap 转换为map
func (o *ObjectValue) ToMap() map[string]NodeValue {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.data
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
func ParseObjectString(data string) (*ObjectValue, error) {
	return ParseObjectValue([]byte(data))
}
func ParseStrObjectValue(data string) (*ObjectValue, error) {
	if util.IsBlank(data) {
		return nil, errors.New("data is empty")
	}
	return ParseObjectValue([]byte(data))
}

// ExecuteTemplate 执行模板，使用标准 {{.variable}} 格式
//func (o *ObjectValue) ExecuteTemplate(templateStr string) (string, error) {
//	if templateStr == "" {
//		return "", nil
//	}
//
//	tmpl, err := template.New("template").Parse(templateStr)
//	if err != nil {
//		return "", err
//	}
//
//	var buf bytes.Buffer
//	if err := tmpl.Execute(&buf, o.ToMap()); err != nil {
//		return "", err
//	}
//
//	return buf.String(), nil
//}

// ExecuteTemplateWithDollarFormat 执行模板，支持 ${variable} 格式自动转换为 {{.variable}} 格式
// 如果结果中仍包含未解析的占位符，则返回错误
func (o *ObjectValue) ExecuteTemplateWithDollarFormat(templateStr string) (string, error) {
	data := o.ToMap()
	dataValue := make(map[string]any)
	for k, v := range data {
		dataValue[k] = v
	}
	return util.ExecuteTemplateWithDollarFormat(dataValue, templateStr)
}

func (o *ObjectValue) HasKey(s string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	_, ok := o.data[s]
	return ok
}

type OptionsValue struct {
	*ObjectValue
}

func NewOptionsValue() *OptionsValue {
	return &OptionsValue{
		ObjectValue: NewObjectValue(),
	}
}
