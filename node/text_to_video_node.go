package node

import (
	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/util"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// TextToVideoFunction 文字转视频函数
type TextToVideoFunction func(state *State, text string, optionsValue *value.OptionsValue) (value.NodeValue, error)

// TextToVideoNode 文字转视频节点
type TextToVideoNode struct {
	*BaseNode
	textToVideoFunction TextToVideoFunction
	optionsValue        *value.OptionsValue
	optionsFrom         []*value.ValueFrom

	textValueFrom *value.TextValueFrom
}

// NewTextToVideoNode 创建文字转视频节点
func NewTextToVideoNode(id string) *TextToVideoNode {
	return &TextToVideoNode{
		BaseNode:     NewBaseNode(id, types.NodeTypeSingle),
		optionsValue: value.NewOptionsValue(),
		optionsFrom:  []*value.ValueFrom{},
	}
}

func (n *TextToVideoNode) ParseTextValuesFromWithError(state *State) (*value.TextValue, error) {
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
func (n *TextToVideoNode) Exec(state *State) (value.NodeValue, error) {
	// 执行文字转视频函数
	if n.textToVideoFunction == nil {
		return nil, errors.New(n.ID + " textToVideoFunction is nil")
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
	result, err := n.textToVideoFunction(state, text, options)
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
func (n *TextToVideoNode) GetNodeGraph() *graph.NodeGraph {
	return graph.NewNodeGraph(n.ID, "TextToVideoNode", n.ValuesFrom)
}

// TextToVideoNodeBuilder 文字转视频节点构建器
type TextToVideoNodeBuilder struct {
	node *TextToVideoNode
}

// NewTextToVideoNodeBuilder 创建文字转视频节点构建器
func NewTextToVideoNodeBuilder(id string) *TextToVideoNodeBuilder {
	return &TextToVideoNodeBuilder{
		node: NewTextToVideoNode(id),
	}
}

func (b *TextToVideoNodeBuilder) TextFrom(textFrom *value.TextValueFrom) *TextToVideoNodeBuilder {
	b.node.textValueFrom = textFrom
	b.node.ValuesFrom = append(b.node.ValuesFrom, &value.ValueFrom{
		NodeID: textFrom.NodeID,
		From:   textFrom.From,
		Param:  textFrom.NodeID + "_" + textFrom.From,
	})
	return b
}

func (b *TextToVideoNodeBuilder) TextToVideoFunction(textToVideoFunction TextToVideoFunction) *TextToVideoNodeBuilder {
	b.node.textToVideoFunction = textToVideoFunction
	return b
}

func (b *TextToVideoNodeBuilder) Options(key string, value any) *TextToVideoNodeBuilder {
	b.node.optionsValue.PutAny(key, value)
	return b
}

func (b *TextToVideoNodeBuilder) OptionsFrom(valuesFrom *value.ValueFrom) *TextToVideoNodeBuilder {
	b.node.optionsFrom = []*value.ValueFrom{valuesFrom}
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom)
	return b
}

// Build 构建
func (b *TextToVideoNodeBuilder) Build() *TextToVideoNode {
	return b.node
}