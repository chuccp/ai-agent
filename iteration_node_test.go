package ai_agent

import (
	"testing"

	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/value"
	"github.com/stretchr/testify/assert"
)

// TestIterationNode 测试迭代节点
func TestIterationNode(t *testing.T) {
	// 创建一个处理单个项目的节点
	processItem := node.NewFunctionNodeBuilder("processItem").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			// 从当前迭代项获取值
			item := int(state.GetRootValue().GetNumber("item"))

			result := value.NewObjectValue()
			result.PutNumber("squared", float64(item*item))
			return result, nil
		}).
		Build()

	// 创建迭代节点，对输入数组的每个元素执行子工作流
	iterationNode := node.NewIterationNodeBuilder("iterate").
		IterationFrom(value.NewValueFrom("", "items", "")).
		Workflow(Of(processItem)).
		Build()

	workflow := Of(iterationNode)
	ag := NewAgentBuilder("test-iteration").Workflow(workflow).Build()

	execConfig := &executor.Config{
		MaxConcurrency: 2,
	}
	exec := NewAgentExecutor(ag, execConfig)

	// 创建输入：[1, 2, 3, 4]
	input := value.NewObjectValue()
	items := value.NewArrayValue()
	for i := 1; i <= 4; i++ {
		itemObj := value.NewObjectValue()
		itemObj.PutNumber("item", float64(i))
		items.Add(itemObj)
	}
	input.Put("items", items)

	response, err := exec.Exec(input)

	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.NodeValue)

	arrayResult, ok := response.NodeValue.(*value.ArrayValue)
	assert.True(t, ok, "response is not ArrayValue, got: %T", response.NodeValue)
	assert.Equal(t, 4, arrayResult.Size(), "array size mismatch, got %d", arrayResult.Size())

	// 检查每个结果
	expectedSquares := []int{1, 4, 9, 16}
	arrayResult.ForEach(func(i int, item value.NodeValue) bool {
		obj, ok := item.(*value.ObjectValue)
		assert.True(t, ok, "item %d is not ObjectValue", i)
		if obj != nil {
			squared := int(obj.GetNumber("squared"))
			assert.Equal(t, expectedSquares[i], squared)
		}
		return true
	})
}
