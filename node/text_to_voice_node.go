package node

import (
	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/util"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// TextToVoiceFunction 文字转语音函数
type TextToVoiceFunction func(state *State, text string, optionsValue *value.OptionsValue) (value.NodeValue, error)

// TextToVoiceNode 文字转语音节点
type TextToVoiceNode struct {
	*BaseNode
	textToVoiceFunction TextToVoiceFunction
	optionsValue        *value.OptionsValue
	optionsFrom         []*value.ValueFrom

	textValueFrom *value.TextValueFrom
}

// NewTextToVoiceNode 创建文字转语音节点
func NewTextToVoiceNode(id string) *TextToVoiceNode {
	return &TextToVoiceNode{
		BaseNode:     NewBaseNode(id, types.NodeTypeSingle),
		optionsValue: value.NewOptionsValue(),
		optionsFrom:  []*value.ValueFrom{},
	}
}

func (n *TextToVoiceNode) ParseTextValuesFromWithError(state *State) (*value.TextValue, error) {
	if n.textValueFrom == nil {
		return nil, errors.New(n.ID + " textValueFrom is nil")
	}
	vf := n.textValueFrom
	nodeValue, err := state.GetNodeValueFromNodeWithError(vf.NodeID, vf.From)
	if err != nil {
		return nil, err
	}
	if nodeValue != nil {
		if nodeValue.IsText() {
			return nodeValue.AsText(), nil
		} else {
			return nil, errors.New(n.ID + " nodeValue is not text")
		}
	}
	return nil, errors.New(n.ID + " textValueFrom is nil")
}

// Exec 执行节点
func (n *TextToVoiceNode) Exec(state *State) (value.NodeValue, error) {
	// 执行文字转语音函数
	if n.textToVoiceFunction == nil {
		return nil, errors.New(n.ID + " textToVoiceFunction is nil")
	}

	nodeValue, err := n.ParseTextValuesFromWithError(state)
	if err != nil {
		return nil, err
	}
	text := nodeValue.Text
	if util.IsBlank(text) {
		return nil, errors.New(n.ID + " text is empty")
	}
	optionsFrom0, err := n.ParseNoRootValuesFromWithError(state, n.optionsFrom)
	if err != nil {
		return nil, err
	}
	options := value.NewOptionsValue()
	options.AddAllIFNULL(n.optionsValue.ObjectValue)
	options.AddAllIFNULL(optionsFrom0)
	state.SetStatusType(types.NodeStatusRunning)
	result, err := n.textToVoiceFunction(state, text, options)
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

func (b *TextToVoiceNodeBuilder) TextFrom(textFrom *value.TextValueFrom) *TextToVoiceNodeBuilder {
	b.node.textValueFrom = textFrom
	b.node.ValuesFrom = append(b.node.ValuesFrom, &value.ValueFrom{
		NodeID: textFrom.NodeID,
		From:   textFrom.From,
		Param:  textFrom.NodeID + "_" + textFrom.From,
	})
	return b
}

func (b *TextToVoiceNodeBuilder) TextToVoiceFunction(textToVoiceFunction TextToVoiceFunction) *TextToVoiceNodeBuilder {
	b.node.textToVoiceFunction = textToVoiceFunction
	return b
}

func (b *TextToVoiceNodeBuilder) Options(key string, value any) *TextToVoiceNodeBuilder {
	b.node.optionsValue.PutAny(key, value)
	return b
}

func (b *TextToVoiceNodeBuilder) OptionsFrom(valuesFrom *value.ValueFrom) *TextToVoiceNodeBuilder {
	b.node.optionsFrom = []*value.ValueFrom{valuesFrom}
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom)
	return b
}

// Build 构建
func (b *TextToVoiceNodeBuilder) Build() *TextToVoiceNode {
	return b.node
}
