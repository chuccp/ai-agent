package field

import "github.com/chuccp/ai-agent/types"

// TextField 文本字段
type TextField struct {
	*BaseField
}

// NewTextField 创建文本字段
func NewTextField(name string) *TextField {
	return &TextField{
		BaseField: NewBaseField(name, types.FieldTypeText),
	}
}

// CreateTextField 创建文本字段并设置描述
func CreateTextField(name, description string) Field {
	tf := NewTextField(name)
	tf.Description = description
	return tf
}