package field

import "github.com/chuccp/ai-agent/types"

// ObjectField 对象字段
type ObjectField struct {
	*BaseField
	outputs []Field
}

// NewObjectField 创建对象字段
func NewObjectField(name string) *ObjectField {
	return &ObjectField{
		BaseField: NewBaseField(name, types.FieldTypeObject),
		outputs:   make([]Field, 0),
	}
}

// AddField 添加字段
func (f *ObjectField) AddField(field Field) *ObjectField {
	f.outputs = append(f.outputs, field)
	return f
}

// AddTextField 添加文本字段
func (f *ObjectField) AddTextField(fieldName, description string) *ObjectField {
	f.outputs = append(f.outputs, CreateTextField(fieldName, description))
	return f
}

// GetOutputs 获取所有输出字段
func (f *ObjectField) GetOutputs() []Field {
	return f.outputs
}

// SetOutputs 设置输出字段
func (f *ObjectField) SetOutputs(outputs []Field) {
	f.outputs = outputs
}