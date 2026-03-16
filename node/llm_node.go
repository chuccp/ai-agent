package node

import (
	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/out"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// LLMFunction LLM函数
type LLMFunction func(nodeState *State, files *value.FilesValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error)

// LLMNode LLM节点
type LLMNode struct {
	*BaseNode
	formatOut      out.OutFormat
	stream         bool
	systemTemplate string
	userTemplate   string
	llmFunction    LLMFunction
	cacheEnabled   bool
	fileValuesFrom []*value.FilesValueFrom
}

// NewLLMNode 创建LLM节点
func NewLLMNode(id string, fileValuesFrom []*value.FilesValueFrom) *LLMNode {
	return &LLMNode{
		BaseNode:       NewBaseNode(id, types.NodeTypeSingle),
		fileValuesFrom: fileValuesFrom,
		cacheEnabled:   true,
	}
}

// SetFormatOut 设置输出格式
func (n *LLMNode) SetFormatOut(format out.OutFormat) {
	n.formatOut = format
}

// SetStream 设置是否流式
func (n *LLMNode) SetStream(stream bool) {
	n.stream = stream
}

// SetSystemTemplate 设置系统模板
func (n *LLMNode) SetSystemTemplate(template string) {
	n.systemTemplate = template
}

// SetUserTemplate 设置用户模板
func (n *LLMNode) SetUserTemplate(template string) {
	n.userTemplate = template
}

// SetLLMFunction 设置LLM函数
func (n *LLMNode) SetLLMFunction(fn LLMFunction) {
	n.llmFunction = fn
}

// SetCacheEnabled 设置是否启用缓存
func (n *LLMNode) SetCacheEnabled(enabled bool) {
	n.cacheEnabled = enabled
}

// IsCacheEnabled 是否启用缓存
func (n *LLMNode) IsCacheEnabled() bool {
	return n.cacheEnabled
}

// ParseFilesValuesFrom 解析文件值来源
func (n *LLMNode) ParseFilesValuesFrom(state *State) (*value.FilesValue, error) {
	filesValue := value.NewFilesValue()
	if n.fileValuesFrom == nil {
		return filesValue, nil
	}
	for _, vf := range n.fileValuesFrom {
		nodeValue := state.GetNodeValueFromNode(vf.NodeID, vf.From)
		if nodeValue != nil && nodeValue.IsFiles() {
			filesValue.AddAllFiles(nodeValue.AsFiles())
		}
	}
	return filesValue, nil
}

// Exec 执行节点
func (n *LLMNode) Exec(state *State) (value.NodeValue, error) {
	nodeValue, err := n.ParseValuesFromWithError(state, n.ValuesFrom)
	if err != nil {
		return nil, err
	}
	filesValue, err := n.ParseFilesValuesFrom(state)
	if err != nil {
		return nil, err
	}

	// 使用模板解析（支持 ${variable} 格式）
	systemPrompt, err := nodeValue.ExecuteTemplateWithDollarFormat(n.systemTemplate)
	if err != nil {
		return nil, err
	}
	userPrompt, err := nodeValue.ExecuteTemplateWithDollarFormat(n.userTemplate)
	if err != nil {
		return nil, err
	}

	// 解析参数（node设置优先级高于Parameter）
	stream := n.resolveStream(state)
	cacheEnabled := n.resolveCacheEnabled(state)

	// 构建缓存键
	cacheKey := systemPrompt + userPrompt
	if !filesValue.IsEmpty() {
		cacheKey += filesValue.String()
	}

	// 检查缓存
	if cacheEnabled && state.IsCacheEnabled() {
		cachedResult, err := state.GetCache(cacheKey)
		if err == nil && cachedResult != nil && !cachedResult.IsNull() {
			return cachedResult, nil
		}
	}

	// 执行LLM函数
	var result value.NodeValue
	if n.llmFunction != nil {
		result, err = n.llmFunction(state, filesValue, systemPrompt, userPrompt, n.formatOut, stream)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(n.ID + " llmFunction is nil")
	}

	// 保存到缓存
	if cacheEnabled && result != nil && state.IsCacheEnabled() {
		state.SaveCache(cacheKey, result)
	}

	return result, nil
}

// resolveStream 解析stream参数，node设置优先级高于Parameter
func (n *LLMNode) resolveStream(state *State) bool {
	// 否则从Parameter获取，如果也没有则使用node的默认值(false)
	return n.stream
}

// resolveCacheEnabled 解析cacheEnabled参数，node设置优先级高于Parameter
func (n *LLMNode) resolveCacheEnabled(state *State) bool {
	return n.cacheEnabled
}

// GetNodeGraph 获取节点图
func (n *LLMNode) GetNodeGraph() *graph.NodeGraph {
	return graph.NewNodeGraph(n.ID, "LLMNode", n.ValuesFrom)
}

// LLMNodeBuilder LLM节点构建器
type LLMNodeBuilder struct {
	node *LLMNode
}

// NewLLMNodeBuilder 创建LLM节点构建器
func NewLLMNodeBuilder(id string) *LLMNodeBuilder {
	return &LLMNodeBuilder{
		node: NewLLMNode(id, nil),
	}
}

// FormatOut 设置输出格式
func (b *LLMNodeBuilder) FormatOut(format out.OutFormat) *LLMNodeBuilder {
	b.node.formatOut = format
	return b
}

// Stream 设置是否流式
func (b *LLMNodeBuilder) Stream(stream bool) *LLMNodeBuilder {
	b.node.stream = stream
	return b
}

// SystemTemplate 设置系统模板
func (b *LLMNodeBuilder) SystemTemplate(template string) *LLMNodeBuilder {
	b.node.systemTemplate = template
	return b
}

// UserTemplate 设置用户模板
func (b *LLMNodeBuilder) UserTemplate(template string) *LLMNodeBuilder {
	b.node.userTemplate = template
	return b
}

// LLMFunction 设置LLM函数
func (b *LLMNodeBuilder) LLMFunction(fn LLMFunction) *LLMNodeBuilder {
	b.node.llmFunction = fn
	return b
}

// ValuesFrom 设置值来源
func (b *LLMNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *LLMNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// FileValuesFrom 设置文件值来源
func (b *LLMNodeBuilder) FileValuesFrom(fileValuesFrom ...*value.FilesValueFrom) *LLMNodeBuilder {
	b.node.fileValuesFrom = append(b.node.fileValuesFrom, fileValuesFrom...)
	return b
}

// CacheEnabled 设置是否启用缓存
func (b *LLMNodeBuilder) CacheEnabled(enabled bool) *LLMNodeBuilder {
	b.node.cacheEnabled = enabled
	return b
}

// Build 构建
func (b *LLMNodeBuilder) Build() *LLMNode {
	return b.node
}

// 错误定义
var (
	ErrLLMNodeFunctionRequired = errors.New("LLMNode llmFunction is required")
	ErrLLMNodeTemplateRequired = errors.New("LLMNode requires at least one template")
)
