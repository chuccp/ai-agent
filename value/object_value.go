package value

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"emperror.dev/errors"
)

// 匹配 ${variable_name} 格式的正则表达式
var templateVarRegex = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// convertTemplateSyntax 将 ${variable} 格式转换为 Go template 的 {{.variable}} 格式
func convertTemplateSyntax(templateStr string) string {
	return templateVarRegex.ReplaceAllString(templateStr, "{{.$1}}")
}

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
func (o *ObjectValue) GetOrString(keys ...string) string {
	for _, key := range keys {
		v := o.Get(key)
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
	return ""
}

// GetUrlsValue 获取URL值
func (o *ObjectValue) GetUrls(key string) *UrlsValue {
	v := o.Get(key)
	if v == nil || !v.IsUrls() {
		return nil
	}
	return v.AsUrls()
}

// GetNumber 获取数值
func (o *ObjectValue) GetNumber(key string) float64 {
	v := o.Get(key)
	if v == nil || v.IsNull() {
		return 0
	}
	if v.IsNumber() {
		return v.AsNumber().Float64()
	}
	return 0
}

// GetBool 获取布尔值
func (o *ObjectValue) GetBool(key string) bool {
	v := o.Get(key)
	if v == nil || v.IsNull() {
		return false
	}
	if v.IsBool() {
		return v.AsBool().Value
	}
	return false
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

func (o *ObjectValue) PutUint(key string, value uint) {
	o.Put(key, NewNumberValue(float64(value)))
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
func (o *ObjectValue) AddAllIFNULL(other *ObjectValue) {
	if other == nil {
		return
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
func ParseStrObjectValue(data string) (*ObjectValue, error) {

	if data == "" {
		return NewObjectValue(), nil
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
	if templateStr == "" {
		return "", nil
	}

	// 将 ${variable} 格式转换为 {{.variable}} 格式
	convertedTemplate := convertTemplateSyntax(templateStr)

	tmpl, err := template.New("template").Parse(convertedTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, o.ToMap()); err != nil {
		return "", err
	}

	result := buf.String()

	// 检查是否仍有未解析的占位符
	if unresolved := findUnresolvedPlaceholders(result); len(unresolved) > 0 {
		return "", errors.New("unresolved placeholders found: " + strings.Join(unresolved, ", "))
	}

	return result, nil
}

// findUnresolvedPlaceholders 查找未解析的占位符
func findUnresolvedPlaceholders(text string) []string {
	var unresolved []string

	// 检查 Go template 格式 {{.variable}} 或 {{ variable }}
	goTemplateRegex := regexp.MustCompile(`\{\{\.?[a-zA-Z_][a-zA-Z0-9_]*\}\}`)
	goMatches := goTemplateRegex.FindAllString(text, -1)
	unresolved = append(unresolved, goMatches...)

	// 检查 <no value> 格式 (Go template 对缺失值的输出)
	if strings.Contains(text, "<no value>") {
		unresolved = append(unresolved, "<no value>")
	}

	return unresolved
}
