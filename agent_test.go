package ai_agent

import (
	"testing"

	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/value"
	"github.com/stretchr/testify/assert"
)

// TestSimpleWorkflow 测试简单工作流
func TestSimpleWorkflow(t *testing.T) {
	// 创建函数节点
	processNode := node.NewFunctionNodeBuilder("process").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			rootValue := state.GetRootValue()
			name := rootValue.GetString("name")

			result := value.NewObjectValue()
			result.PutString("greeting", "Hello, "+name+"!")
			return result, nil
		}).
		Build()

	// 创建输出节点
	outputNode := node.NewOutputNodeBuilder("output").
		ValuesFrom(value.NewValueFrom("process", "", "")).
		Build()

	// 创建工作流和代理
	workflow := Of(processNode, outputNode)
	ag := NewAgentBuilder("test-agent").
		Workflow(workflow).
		Build()

	// 执行
	execConfig := &executor.Config{
		MaxConcurrency: 2,
	}
	exec := NewAgentExecutor(ag, execConfig)
	input := value.NewObjectValue()
	input.PutString("name", "World")

	response, err := exec.Exec(input)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Success)
}

// TestEmptyWorkflow 测试空工作流应该返回错误
func TestEmptyWorkflow(t *testing.T) {
	builder := NewWorkflowBuilder()
	workflow, err := builder.Build()

	assert.Nil(t, workflow)
	assert.ErrorIs(t, err, ErrWorkflowRequiresNode)
}

// TestGetRootParams 测试获取根参数
func TestGetRootParams(t *testing.T) {
	node1 := node.NewOutputNodeBuilder("output").
		ValuesFrom(value.NewValueFrom("", "name", "name")).
		Build()

	node2 := node.NewOutputNodeBuilder("output2").
		ValuesFrom(value.NewValueFrom("", "age", "age")).
		Build()

	workflow := Of(node1, node2)

	params := workflow.GetRootParams()

	assert.Len(t, params, 2)
	paramSet := make(map[string]bool)
	for _, p := range params {
		paramSet[p] = true
	}
	assert.True(t, paramSet["name"])
	assert.True(t, paramSet["age"])
}

// TestAsyncExecution 测试异步执行
func TestAsyncExecution(t *testing.T) {
	processNode := node.NewFunctionNodeBuilder("process").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			result := value.NewObjectValue()
			result.PutString("result", "done")
			return result, nil
		}).
		Build()

	workflow := Of(processNode)
	ag := NewAgentBuilder("test-async").Workflow(workflow).Build()

	execConfig := &executor.Config{
		MaxConcurrency: 2,
	}
	exec := NewAgentExecutor(ag, execConfig)
	input := value.NewObjectValue()

	result := exec.ExecAsync(input)

	assert.NoError(t, result.Error)
	assert.NotNil(t, result.Response)
	assert.True(t, result.Response.Success)
}

// TestConcurrentExecution 测试并发执行
func TestConcurrentExecution(t *testing.T) {
	// nodeA 和 nodeB 并行执行，nodeC 依赖两者
	nodeA := node.NewFunctionNodeBuilder("nodeA").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutNumber("value", 1)
			return res, nil
		}).
		Build()

	nodeB := node.NewFunctionNodeBuilder("nodeB").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutNumber("value", 2)
			return res, nil
		}).
		Build()

	nodeC := node.NewFunctionNodeBuilder("nodeC").
		ValuesFrom(
			value.NewValueFrom("nodeA", "", ""),
			value.NewValueFrom("nodeB", "", ""),
		).
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			ctx := state.GetWorkflowContext()
			valA := int(ctx.GetNodeValue("nodeA").(*value.ObjectValue).GetNumber("value"))
			valB := int(ctx.GetNodeValue("nodeB").(*value.ObjectValue).GetNumber("value"))

			res := value.NewObjectValue()
			res.PutNumber("sum", float64(valA+valB))
			return res, nil
		}).
		Build()

	workflow := Of(nodeA, nodeB, nodeC)
	ag := NewAgentBuilder("test-concurrent").Workflow(workflow).Build()

	execConfig := &executor.Config{
		MaxConcurrency: 2,
	}
	exec := NewAgentExecutor(ag, execConfig)
	input := value.NewObjectValue()

	response, err := exec.Exec(input)

	assert.NoError(t, err)
	assert.True(t, response.Success)

	result := int(response.NodeValue.(*value.ObjectValue).GetNumber("sum"))
	assert.Equal(t, 3, result)
}
