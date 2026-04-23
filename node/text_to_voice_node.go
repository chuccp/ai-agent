package node

import (
	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/util"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// TextToVoiceFunction 文字转语音函数
type TextToVoiceFunction func(state *State, text string, parametersValue *value.ParametersValue) (value.NodeValue, error)

// TextToVoiceNode 文字转语音节点
type TextToVoiceNode struct {
	*BaseNode
	textToVoiceFunction TextToVoiceFunction
	parametersValue     *value.ParametersValue
	parametersFrom      []*value.ValueFrom
}

// NewTextToVoiceNode 创建文字转语音节点
func NewTextToVoiceNode(id string) *TextToVoiceNode {
	return &TextToVoiceNode{
		BaseNode:        NewBaseNode(id, types.NodeTypeSingle),
		parametersValue: value.NewParametersValue(),
		parametersFrom:  []*value.ValueFrom{},
	}
}

// Exec 执行节点
func (n *TextToVoiceNode) Exec(state *State) (value.NodeValue, error) {
	// 执行文字转语音函数
	if n.textToVoiceFunction == nil {
		return nil, errors.New(n.ID + " textToVoiceFunction is nil")
	}

	nodeValue, err := n.ParseValuesFromWithError(state, n.ValuesFrom)
	if err != nil {
		return nil, err
	}
	if !nodeValue.IsText() {
		return nil, errors.New(n.ID + " nodeValue is not text")
	}
	text := nodeValue.AsText().Text
	if util.IsBlank(text) {
		return nil, errors.New(n.ID + " text is empty")
	}

	parametersValue0, err := n.ParseValuesFromWithError(state, n.parametersFrom)
	if err != nil {
		return nil, err
	}
	n.parametersValue.AddAllIFNULL(parametersValue0)
	state.SetStatusType(types.NodeStatusRunning)
	result, err := n.textToVoiceFunction(state, text, n.parametersValue)
	if err != nil {
		state.SetStatusType(types.NodeStatusFailed)
		return nil, err
	}

	// 处理结果状态
	if result == nil || result.IsNull() {
		state.SetStatusType(types.NodeStatusRunning)
		return nil, nil
	}

	state.SetStatusType(types.NodeStatusSucceeded)
	return result, nil
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

func (b *TextToVoiceNodeBuilder) TextFrom(valuesFrom *value.ValueFrom) *TextToVoiceNodeBuilder {
	b.node.ValuesFrom = []*value.ValueFrom{valuesFrom}
	return b
}

func (b *TextToVoiceNodeBuilder) TextToVoiceFunction(textToVoiceFunction TextToVoiceFunction) *TextToVoiceNodeBuilder {
	b.node.textToVoiceFunction = textToVoiceFunction
	return b
}

func (b *TextToVoiceNodeBuilder) ParametersValue(key string, value any) *TextToVoiceNodeBuilder {
	b.node.parametersValue.PutAny(key, value)
	return b
}

func (b *TextToVoiceNodeBuilder) ParametersFrom(valuesFrom *value.ValueFrom) *TextToVoiceNodeBuilder {
	b.node.parametersFrom = []*value.ValueFrom{valuesFrom}
	return b
}

// Build 构建
func (b *TextToVoiceNodeBuilder) Build() *TextToVoiceNode {
	return b.node
}
