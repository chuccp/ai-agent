package out

import (
	"github.com/chuccp/ai-agent/types"
)

// OutFormat 输出格式接口
type OutFormat interface {
	GetOutType() types.OutType
	IsJSONOut() bool
	IsTextOut() bool
}

// BaseOutFormat 基础输出格式
type BaseOutFormat struct {
	outType types.OutType
}

// GetOutType 获取输出类型
func (f *BaseOutFormat) GetOutType() types.OutType {
	return f.outType
}

// IsJSONOut 是否JSON输出
func (f *BaseOutFormat) IsJSONOut() bool {
	return f.outType == types.OutTypeJSON
}

// IsTextOut 是否文本输出
func (f *BaseOutFormat) IsTextOut() bool {
	return f.outType == types.OutTypeText
}