package out

import (
	"encoding/json"

	"github.com/chuccp/ai-agent/out/field"
	"github.com/chuccp/ai-agent/types"
)

// JsonOutFormat JSON输出格式
type JsonOutFormat struct {
	BaseOutFormat
	outputs     []field.Field
	exampleJSON string
	isAuto      bool
}

// NewJsonOutFormat 创建JSON输出格式
func NewJsonOutFormat() *JsonOutFormat {
	return &JsonOutFormat{
		BaseOutFormat: BaseOutFormat{outType: types.OutTypeJSON},
		outputs:       make([]field.Field, 0),
	}
}

// Auto 创建自动JSON输出格式
func AutoJsonOutFormat() *JsonOutFormat {
	return &JsonOutFormat{
		BaseOutFormat: BaseOutFormat{outType: types.OutTypeJSON},
		isAuto:        true,
	}
}

// OfExampleJSON 从示例JSON创建
func OfExampleJSON(exampleJSON string) *JsonOutFormat {
	return &JsonOutFormat{
		BaseOutFormat: BaseOutFormat{outType: types.OutTypeJSON},
		exampleJSON:   exampleJSON,
	}
}

// IsAuto 是否自动模式
func (f *JsonOutFormat) IsAuto() bool {
	return f.isAuto
}

// AddField 添加字段
func (f *JsonOutFormat) AddField(field field.Field) *JsonOutFormat {
	f.outputs = append(f.outputs, field)
	return f
}

// AddTextField 添加文本字段
func (f *JsonOutFormat) AddTextField(name, description string) *JsonOutFormat {
	f.outputs = append(f.outputs, field.CreateTextField(name, description))
	return f
}

// GetOutputs 获取输出字段
func (f *JsonOutFormat) GetOutputs() []field.Field {
	return f.outputs
}

// ToExampleJSON 转换为示例JSON
func (f *JsonOutFormat) ToExampleJSON() string {
	if f.isAuto {
		return ""
	}
	if f.exampleJSON != "" {
		return "EXAMPLE JSON OUTPUT:\n" + f.exampleJSON
	}

	result := make(map[string]interface{})
	for _, fld := range f.outputs {
		result[fld.GetName()] = f.getJSONElement(fld)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return "EXAMPLE JSON OUTPUT:\n" + string(data)
}

// getJSONElement 获取字段的JSON元素
func (f *JsonOutFormat) getJSONElement(fld field.Field) interface{} {
	switch fld.GetFieldType() {
	case types.FieldTypeText:
		return fld.GetDescription()

	case types.FieldTypeObject:
		if objField, ok := fld.(*field.ObjectField); ok {
			result := make(map[string]interface{})
			for _, of := range objField.GetOutputs() {
				result[of.GetName()] = of.GetDescription()
			}
			return result
		}

	case types.FieldTypeArray:
		if arrField, ok := fld.(*field.ArrayField); ok {
			if len(arrField.GetDescriptions()) > 0 {
				return arrField.GetDescriptions()
			}
			if len(arrField.GetOutputs()) > 0 {
				result := make(map[string]interface{})
				for _, af := range arrField.GetOutputs() {
					result[af.GetName()] = af.GetDescription()
				}
				return []interface{}{result}
			}
		}
	}

	return nil
}

// Builder JSON输出格式构建器
type JsonOutFormatBuilder struct {
	jsonOut *JsonOutFormat
}

// NewJsonOutFormatBuilder 创建JSON输出格式构建器
func NewJsonOutFormatBuilder() *JsonOutFormatBuilder {
	return &JsonOutFormatBuilder{
		jsonOut: NewJsonOutFormat(),
	}
}

// AddField 添加字段
func (b *JsonOutFormatBuilder) AddField(fld field.Field) *JsonOutFormatBuilder {
	b.jsonOut.AddField(fld)
	return b
}

// AddTextField 添加文本字段
func (b *JsonOutFormatBuilder) AddTextField(name, description string) *JsonOutFormatBuilder {
	b.jsonOut.AddTextField(name, description)
	return b
}

// Build 构建
func (b *JsonOutFormatBuilder) Build() *JsonOutFormat {
	return b.jsonOut
}