package field

import "github.com/chuccp/ai-agent/types"

// ArrayField 数组字段
type ArrayField struct {
	*ObjectField
	descriptions []string
}

// NewArrayField 创建数组字段
func NewArrayField(name string) *ArrayField {
	return &ArrayField{
		ObjectField:  NewObjectField(name),
		descriptions: make([]string, 0),
	}
}

// CreateArrayField 创建数组字段
func CreateArrayField(name string, descriptions ...string) *ArrayField {
	af := NewArrayField(name)
	if len(descriptions) > 0 {
		af.Description = descriptions[0]
		af.descriptions = append(af.descriptions, descriptions...)
	}
	return af
}

// GetDescriptions 获取描述列表
func (f *ArrayField) GetDescriptions() []string {
	return f.descriptions
}

// SetDescriptions 设置描述列表
func (f *ArrayField) SetDescriptions(descriptions []string) {
	f.descriptions = descriptions
}

// GetFieldType 获取字段类型
func (f *ArrayField) GetFieldType() types.FieldType {
	return types.FieldTypeArray
}