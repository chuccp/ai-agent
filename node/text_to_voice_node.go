package node

import (
	"emperror.dev/errors"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// TextToVoiceFunction 文字转语音函数
type TextToVoiceFunction func(state *State, text string, voice string, speed float64) (value.NodeValue, error)

// TextToVoiceNode 文字转语音节点
type TextToVoiceNode struct {
	*BaseNode
	textToVoiceFunction TextToVoiceFunction
	textTemplate        string
	voice               string
	speed               float64
	cacheEnabled        bool
}

// NewTextToVoiceNode 创建文字转语音节点
func NewTextToVoiceNode(id string) *TextToVoiceNode {
	return &TextToVoiceNode{
		BaseNode:     NewBaseNode(id, types.NodeTypeSingle),
		voice:        "",
		speed:        1.0,
		cacheEnabled: true,
	}
}

// SetTextToVoiceFunction 设置文字转语音函数
func (n *TextToVoiceNode) SetTextToVoiceFunction(fn TextToVoiceFunction) *TextToVoiceNode {
	n.textToVoiceFunction = fn
	return n
}

// SetTextTemplate 设置文本模板
func (n *TextToVoiceNode) SetTextTemplate(template string) *TextToVoiceNode {
	n.textTemplate = template
	return n
}

// SetVoice 设置语音类型
func (n *TextToVoiceNode) SetVoice(voice string) *TextToVoiceNode {
	n.voice = voice
	return n
}

// SetSpeed 设置语音速度
func (n *TextToVoiceNode) SetSpeed(speed float64) *TextToVoiceNode {
	n.speed = speed
	return n
}

// SetCacheEnabled 设置是否启用缓存
func (n *TextToVoiceNode) SetCacheEnabled(enabled bool) *TextToVoiceNode {
	n.cacheEnabled = enabled
	return n
}

// Exec 执行节点
func (n *TextToVoiceNode) Exec(state *State) (value.NodeValue, error) {
	// 解析输入
	nodeValue, err := n.ParseValuesFromWithError(state, n.ValuesFrom)
	if err != nil {
		return nil, err
	}

	// 执行模板
	text, err := nodeValue.ExecuteTemplateWithDollarFormat(n.textTemplate)
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, errors.New(n.ID + " text is empty")
	}

	// 解析参数
	voice := n.resolveVoice(state)
	speed := n.resolveSpeed(state)

	// 构建缓存键
	cacheKey := text + voice + string(rune(speed))

	// 检查缓存
	if n.cacheEnabled && state.IsCacheEnabled() {
		cachedResult, err := state.GetCache(cacheKey)
		if err == nil && cachedResult != nil && !cachedResult.IsNull() {
			return cachedResult, nil
		}
	}

	// 执行文字转语音函数
	if n.textToVoiceFunction == nil {
		return nil, errors.New(n.ID + " textToVoiceFunction is nil")
	}

	state.SetStatusType(types.NodeStatusRunning)
	result, err := n.textToVoiceFunction(state, text, voice, speed)
	if err != nil {
		state.SetStatusType(types.NodeStatusFailed)
		return nil, err
	}

	// 处理结果状态
	if result == nil || result.IsNull() {
		state.SetStatusType(types.NodeStatusRunning)
		return nil, nil
	}

	// 保存到缓存
	if n.cacheEnabled && state.IsCacheEnabled() {
		_ = state.SaveCache(cacheKey, result)
	}

	state.SetStatusType(types.NodeStatusSucceeded)
	return result, nil
}

// resolveVoice 解析voice参数
func (n *TextToVoiceNode) resolveVoice(state *State) string {
	if n.voice != "" {
		return n.voice
	}
	return state.GetParameterString("voice", n.voice)
}

// resolveSpeed 解析speed参数
func (n *TextToVoiceNode) resolveSpeed(state *State) float64 {
	if n.speed != 0 {
		return n.speed
	}
	return float64(state.GetParameterInt("speed", int(n.speed)))
}

// GetNodeGraph 获取节点图
func (n *TextToVoiceNode) GetNodeGraph() *graph.NodeGraph {
	return graph.NewNodeGraph(n.ID, "TextToVoiceNode", n.ValuesFrom)
}

// TextToVoiceNodeBuilder 文字转语音节点构建器
type TextToVoiceNodeBuilder struct {
	node *TextToVoiceNode
}

// NewTextToVoiceNodeBuilder 创建文字转语音节点构建器
func NewTextToVoiceNodeBuilder(id string) *TextToVoiceNodeBuilder {
	return &TextToVoiceNodeBuilder{
		node: NewTextToVoiceNode(id),
	}
}

// TextToVoiceFunction 设置文字转语音函数
func (b *TextToVoiceNodeBuilder) TextToVoiceFunction(fn TextToVoiceFunction) *TextToVoiceNodeBuilder {
	b.node.SetTextToVoiceFunction(fn)
	return b
}

// TextTemplate 设置文本模板
func (b *TextToVoiceNodeBuilder) TextTemplate(template string) *TextToVoiceNodeBuilder {
	b.node.SetTextTemplate(template)
	return b
}

// Voice 设置语音类型
func (b *TextToVoiceNodeBuilder) Voice(voice string) *TextToVoiceNodeBuilder {
	b.node.SetVoice(voice)
	return b
}

// Speed 设置语音速度
func (b *TextToVoiceNodeBuilder) Speed(speed float64) *TextToVoiceNodeBuilder {
	b.node.SetSpeed(speed)
	return b
}

// ValuesFrom 设置值来源
func (b *TextToVoiceNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *TextToVoiceNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// CacheEnabled 设置是否启用缓存
func (b *TextToVoiceNodeBuilder) CacheEnabled(enabled bool) *TextToVoiceNodeBuilder {
	b.node.SetCacheEnabled(enabled)
	return b
}

// Build 构建
func (b *TextToVoiceNodeBuilder) Build() *TextToVoiceNode {
	return b.node
}