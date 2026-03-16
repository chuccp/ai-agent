package out

import "github.com/chuccp/ai-agent/types"

// TextOutFormat 文本输出格式
type TextOutFormat struct {
	BaseOutFormat
}

// NewTextOutFormat 创建文本输出格式
func NewTextOutFormat() *TextOutFormat {
	return &TextOutFormat{
		BaseOutFormat: BaseOutFormat{outType: types.OutTypeText},
	}
}