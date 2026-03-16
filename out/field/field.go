package field

import "github.com/chuccp/ai-agent/types"

// Field 字段接口
type Field interface {
	GetName() string
	SetName(name string)
	GetDescription() string
	SetDescription(description string)
	GetFieldType() types.FieldType
}

// BaseField 字段基类
type BaseField struct {
	Name        string
	Description string
	fieldType   types.FieldType
}

// NewBaseField 创建字段
func NewBaseField(name string, fieldType types.FieldType) *BaseField {
	return &BaseField{
		Name:      name,
		fieldType: fieldType,
	}
}

// GetName 获取字段名
func (f *BaseField) GetName() string {
	return f.Name
}

// SetName 设置字段名
func (f *BaseField) SetName(name string) {
	f.Name = name
}

// GetDescription 获取描述
func (f *BaseField) GetDescription() string {
	return f.Description
}

// SetDescription 设置描述
func (f *BaseField) SetDescription(description string) {
	f.Description = description
}

// GetFieldType 获取字段类型
func (f *BaseField) GetFieldType() types.FieldType {
	return f.fieldType
}