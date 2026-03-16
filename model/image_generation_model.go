package model

import (
	"github.com/chuccp/ai-agent/value"
)

// ImageGenerationModel 图片生成模型接口
type ImageGenerationModel interface {
	// Generate 生成图片
	Generate(prompt string, maxNumber int, scale string) (*value.UrlsValue, error)
}