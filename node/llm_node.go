package node

import (
	"bytes"
	"errors"
	"text/template"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// OutFormat 输出格式接口
type OutFormat interface {
	Format(nodeValue value.NodeValue) (string, error)
}

// LLMFunction LLM函数
type LLMFunction func(files *value.FilesValue, systemPrompt, userPrompt string, format OutFormat, stream bool) (value.NodeValue, error)

// LLMNode LLM节点
type LLMNode struct {
	*BaseNode
	formatOut      OutFormat
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
func (n *LLMNode) SetFormatOut(format OutFormat) {
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
func (n *LLMNode) ParseFilesValuesFrom(state *NodeState) (*value.FilesValue, error) {
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

// executeTemplate 使用 Go text/template 执行模板
func (n *LLMNode) executeTemplate(templateStr string, data *value.ObjectValue) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	// 解析模板
	tmpl, err := template.New("llm").Parse(templateStr)
	if err != nil {
		return "", err
	}

	// 将 ObjectValue 转换为 map 作为模板数据
	templateData := data.ToMap()

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Exec 执行节点
func (n *LLMNode) Exec(state *NodeState) (value.NodeValue, error) {
	nodeValue := n.ParseValuesFrom(state, n.ValuesFrom)
	filesValue, err := n.ParseFilesValuesFrom(state)
	if err != nil {
		return nil, err
	}

	// 使用 Go text/template 执行模板
	systemPrompt, err := n.executeTemplate(n.systemTemplate, nodeValue)
	if err != nil {
		return nil, err
	}
	userPrompt, err := n.executeTemplate(n.userTemplate, nodeValue)
	if err != nil {
		return nil, err
	}

	// 构建缓存键
	cacheKey := systemPrompt + userPrompt
	if !filesValue.IsEmpty() {
		cacheKey += filesValue.String()
	}

	// 检查缓存
	ctx := state.GetWorkflowContext()
	if n.cacheEnabled && ctx != nil {
		// 这里需要实现缓存检查
	}

	// 执行LLM函数
	var result value.NodeValue
	if n.llmFunction != nil {
		result, err = n.llmFunction(filesValue, systemPrompt, userPrompt, n.formatOut, n.stream)
		if err != nil {
			return nil, err
		}
	} else {
		result = nodeValue
	}

	// 保存到缓存
	if n.cacheEnabled && result != nil && ctx != nil {
		// 这里需要实现缓存保存
	}

	return result, nil
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
func (b *LLMNodeBuilder) FormatOut(format OutFormat) *LLMNodeBuilder {
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
func (b *LLMNodeBuilder) Build() (*LLMNode, error) {
	if b.node.llmFunction == nil {
		return nil, ErrLLMNodeFunctionRequired
	}
	if b.node.userTemplate == "" && b.node.systemTemplate == "" {
		return nil, ErrLLMNodeTemplateRequired
	}
	return b.node, nil
}

// 错误定义
var (
	ErrLLMNodeFunctionRequired = errors.New("LLMNode llmFunction is required")
	ErrLLMNodeTemplateRequired = errors.New("LLMNode requires at least one template")
)