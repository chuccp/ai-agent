package node

import (
	"log"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// ImageGenerationFunction 图片生成函数
type ImageGenerationFunction func(state *State, resourcesValue *value.ResourcesValue, userPrompt string, maxNumber int, scale string) (value.NodeValue, error)

// ImageGenerationNode 图片生成节点
type ImageGenerationNode struct {
	*BaseNode
	imageGenerationFunction ImageGenerationFunction
	userTemplate            string
	resourcesValueFrom      []*value.ResourcesValueFrom
	maxNumber               int
	scale                   string
	cacheEnabled            bool
}

// NewImageGenerationNode 创建图片生成节点
func NewImageGenerationNode(id string) *ImageGenerationNode {
	return &ImageGenerationNode{
		BaseNode:     NewBaseNode(id, types.NodeTypeSingle),
		maxNumber:    1,
		scale:        "",
		cacheEnabled: true,
	}
}

// SetImageGenerationFunction 设置图片生成函数
func (n *ImageGenerationNode) SetImageGenerationFunction(fn ImageGenerationFunction) *ImageGenerationNode {
	n.imageGenerationFunction = fn
	return n
}

// SetUserTemplate 设置用户提示词模板
func (n *ImageGenerationNode) SetUserTemplate(template string) *ImageGenerationNode {
	n.userTemplate = template
	return n
}

// SetResourcesValueFrom 设置资源值来源
func (n *ImageGenerationNode) SetResourcesValueFrom(froms ...*value.ResourcesValueFrom) *ImageGenerationNode {
	n.resourcesValueFrom = append(n.resourcesValueFrom, froms...)
	return n
}

// SetMaxNumber 设置最大生成数量
func (n *ImageGenerationNode) SetMaxNumber(max int) *ImageGenerationNode {
	n.maxNumber = max
	return n
}

// SetScale 设置图片比例
func (n *ImageGenerationNode) SetScale(scale string) *ImageGenerationNode {
	n.scale = scale
	return n
}

// SetCacheEnabled 设置是否启用缓存
func (n *ImageGenerationNode) SetCacheEnabled(enabled bool) *ImageGenerationNode {
	n.cacheEnabled = enabled
	return n
}

// ParseResourcesValuesFrom 解析资源值来源
func (n *ImageGenerationNode) ParseResourcesValuesFromWithError(state *State) (*value.ResourcesValue, error) {
	newResourcesValue := value.NewResourcesValue()
	if n.resourcesValueFrom == nil {
		return newResourcesValue, nil
	}

	for _, vf := range n.resourcesValueFrom {
		nodeValue, err := state.GetNodeValueFromNodeWithError(vf.NodeID, vf.From)
		if err != nil {
			return nil, err
		}
		if nodeValue != nil {

			if nodeValue.IsResources() {
				newResourcesValue.AddAll(nodeValue.AsResources())
			}
			if nodeValue.IsArray() {
				nodeValue.AsArray().ForEach(func(index int, v value.NodeValue) bool {
					if v.IsText() {
						newResourcesValue.Add(v.AsText().Text)
					}
					return true
				})
			}
			if nodeValue.IsNumber() {
				newResourcesValue.Add(nodeValue.AsNumber().String())
			}

		}
	}
	return newResourcesValue, nil
}

// Exec 执行节点
func (n *ImageGenerationNode) Exec(state *State) (value.NodeValue, error) {
	// 解析输入
	nodeValue, err := n.ParseValuesFromWithError(state, n.ValuesFrom)
	if err != nil {
		return nil, err
	}
	resourcesValue, err := n.ParseResourcesValuesFromWithError(state)
	if err != nil {
		return nil, err
	}
	// 执行模板
	userPrompt, err := nodeValue.ExecuteTemplateWithDollarFormat(n.userTemplate)
	if err != nil {
		return nil, err
	}
	if userPrompt == "" {
		return nil, errors.New(" userPrompt is empty ")
	}

	// 解析参数（node设置优先级高于Parameter）
	maxNumber := n.resolveMaxNumber(state)
	scale := n.resolveScale(state)

	// 执行图片生成
	if n.imageGenerationFunction == nil {
		return nil, errors.Errorf(" nodeID %s imageGenerationFunction is nil ", n.ID)
	}

	state.SetStatusType(types.NodeStatusRunning)
	result, err := n.imageGenerationFunction(state, resourcesValue, userPrompt, maxNumber, scale)
	log.Println("ImageGenerationNode", "Exec", "result", result, "err", err)
	if err != nil {
		state.SetStatusType(types.NodeStatusFailed)
		return nil, err
	}

	// 处理结果状态
	if result == nil || result.IsNull() {
		state.SetStatusType(types.NodeStatusRunning)
		return nil, nil
	}

	if result.IsResources() && result.AsResources().IsEmpty() {
		state.SetStatusType(types.NodeStatusRunning)
		return nil, nil
	}

	state.SetStatusType(types.NodeStatusSucceeded)
	return result, nil
}

func (n *ImageGenerationNode) resolveMaxNumber(state *State) int {
	if n.maxNumber != 0 {
		return n.maxNumber
	}
	return state.GetParameterInt("maxNumber", n.maxNumber)
}

// resolveScale 解析scale参数，node设置优先级高于Parameter
func (n *ImageGenerationNode) resolveScale(state *State) string {
	// 如果node设置的值不是默认值("1:1")，说明是显式设置，优先使用
	if n.scale != "" {
		return n.scale
	}
	// 否则从Parameter获取，如果也没有则使用node的默认值
	scale := state.GetParameterString("scale", n.scale)
	if scale == "" {
		return n.scale
	}
	return scale
}

// GetNodeGraph 获取节点图
func (n *ImageGenerationNode) GetNodeGraph() *graph.NodeGraph {
	var valuesFrom []*value.ValueFrom
	for _, vf := range n.resourcesValueFrom {
		valuesFrom = append(valuesFrom, &value.ValueFrom{
			NodeID: vf.NodeID,
			From:   vf.From,
		})
	}
	return graph.NewNodeGraph(n.ID, "ImageGenerationNode", valuesFrom)
}

// ImageGenerationNodeBuilder 图片生成节点构建器
type ImageGenerationNodeBuilder struct {
	node *ImageGenerationNode
}

// NewImageGenerationNodeBuilder 创建图片生成节点构建器
func NewImageGenerationNodeBuilder(id string) *ImageGenerationNodeBuilder {
	return &ImageGenerationNodeBuilder{
		node: NewImageGenerationNode(id),
	}
}

// ImageGenerationFunction 设置图片生成函数
func (b *ImageGenerationNodeBuilder) ImageGenerationFunction(fn ImageGenerationFunction) *ImageGenerationNodeBuilder {
	b.node.SetImageGenerationFunction(fn)
	return b
}

// UserTemplate 设置用户提示词模板
func (b *ImageGenerationNodeBuilder) UserTemplate(template string) *ImageGenerationNodeBuilder {
	b.node.SetUserTemplate(template)
	return b
}

// ValuesFrom 设置值来源
func (b *ImageGenerationNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *ImageGenerationNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// ResourcesValueFrom 设置资源值来源
func (b *ImageGenerationNodeBuilder) ResourcesValueFrom(froms ...*value.ResourcesValueFrom) *ImageGenerationNodeBuilder {
	b.node.SetResourcesValueFrom(froms...)
	return b
}

// MaxNumber 设置最大生成数量
func (b *ImageGenerationNodeBuilder) MaxNumber(max int) *ImageGenerationNodeBuilder {
	b.node.SetMaxNumber(max)
	return b
}

// Scale 设置图片比例
func (b *ImageGenerationNodeBuilder) Scale(scale string) *ImageGenerationNodeBuilder {
	b.node.SetScale(scale)
	return b
}

// CacheEnabled 设置是否启用缓存
func (b *ImageGenerationNodeBuilder) CacheEnabled(enabled bool) *ImageGenerationNodeBuilder {
	b.node.SetCacheEnabled(enabled)
	return b
}

// Build 构建
func (b *ImageGenerationNodeBuilder) Build() *ImageGenerationNode {
	return b.node
}
