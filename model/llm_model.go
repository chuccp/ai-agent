package model

import (
	"github.com/chuccp/ai-agent/out"
	"github.com/chuccp/ai-agent/value"
)

// LLMModel LLM模型接口
type LLMModel interface {
	// ChatCompletions 聊天完成
	ChatCompletions(resourcesValue *value.ResourcesValue, prompt, text string, outFormat out.OutFormat, streamValue *value.StreamNodeValue) (value.NodeValue, error)
}
