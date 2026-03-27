package ai_agent

import (
	"testing"

	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/value"
	"github.com/stretchr/testify/assert"
)

// TestIFNodeThen 测试 IFNode then 分支
func TestIFNodeThen(t *testing.T) {
	// 创建 pass 分支
	passNode := node.NewFunctionNodeBuilder("pass").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutString("result", "passed")
			return res, nil
		}).Build()

	// 创建 fail 分支
	failNode := node.NewFunctionNodeBuilder("fail").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutString("result", "failed")
			return res, nil
		}).Build()

	// 创建条件节点
	ifNode, err := node.NewIFNodeBuilder("check").
		Condition(func(ctx node.WorkflowContext) bool {
			rootVal := ctx.GetRootValue()
			score := int(rootVal.GetNumber("score"))
			return score >= 60
		}).
		Then(Of(passNode)).
		Else(Of(failNode)).
		ValuesFrom(value.NewValueFrom("", "score", "score")).
		Build()

	assert.NoError(t, err)

	// 创建工作流
	workflow := Of(ifNode)
	ag := NewAgentBuilder("test-if").Workflow(workflow).Build()

	execConfig := &executor.Config{
		MaxConcurrency: 2,
	}
	exec := NewAgentExecutor(ag, execConfig)

	// 测试条件为 true (pass)
	input := value.NewObjectValue()
	input.PutNumber("score", 80)

	response, err := exec.Exec(input)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	resultObj, ok := response.NodeValue.(*value.ObjectValue)
	assert.True(t, ok)
	assert.Equal(t, "passed", resultObj.GetString("result"))
}

// TestIFNodeElse 测试 IFNode else 分支
func TestIFNodeElse(t *testing.T) {
	// 创建 pass 分支
	passNode := node.NewFunctionNodeBuilder("pass").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutString("result", "passed")
			return res, nil
		}).Build()

	// 创建 fail 分支
	failNode := node.NewFunctionNodeBuilder("fail").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutString("result", "failed")
			return res, nil
		}).Build()

	// 创建条件节点
	ifNode, err := node.NewIFNodeBuilder("check").
		Condition(func(ctx node.WorkflowContext) bool {
			rootVal := ctx.GetRootValue()
			score := int(rootVal.GetNumber("score"))
			return score >= 60
		}).
		Then(Of(passNode)).
		Else(Of(failNode)).
		ValuesFrom(value.NewValueFrom("", "score", "score")).
		Build()

	assert.NoError(t, err)

	// 创建工作流
	workflow := Of(ifNode)
	ag := NewAgentBuilder("test-if-else").Workflow(workflow).Build()

	execConfig := &executor.Config{
		MaxConcurrency: 2,
	}
	exec := NewAgentExecutor(ag, execConfig)

	// 测试条件为 false (fail)
	input := value.NewObjectValue()
	input.PutNumber("score", 50)

	response, err := exec.Exec(input)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	resultObj, ok := response.NodeValue.(*value.ObjectValue)
	assert.True(t, ok)
	assert.Equal(t, "failed", resultObj.GetString("result"))
}

// TestIFNodeMissingCondition 测试缺少条件应该返回错误
func TestIFNodeMissingCondition(t *testing.T) {
	ifNode, err := node.NewIFNodeBuilder("check").
		Then(Of(node.NewFunctionNodeBuilder("test").Build())).
		Build()

	assert.Nil(t, ifNode)
	assert.ErrorIs(t, err, node.ErrIFNodeConditionRequired)
}

// TestIFNodeMissingWorkflow 测试缺少工作流应该返回错误
func TestIFNodeMissingWorkflow(t *testing.T) {
	ifNode, err := node.NewIFNodeBuilder("check").
		Condition(func(ctx node.WorkflowContext) bool { return true }).
		Build()

	assert.Nil(t, ifNode)
	assert.ErrorIs(t, err, node.ErrIFNodeWorkflowRequired)
}

// TestIFNodeNullWhenNoMatch 测试当条件不匹配且没有对应分支时返回 null
func TestIFNodeNullWhenNoMatch(t *testing.T) {
	ifNode, err := node.NewIFNodeBuilder("check").
		Condition(func(ctx node.WorkflowContext) bool {
			return true
		}).
		Then(Of(
			node.NewFunctionNodeBuilder("then").
				ExecFunc(func(state *node.State) (value.NodeValue, error) {
					res := value.NewObjectValue()
					res.PutString("result", "ok")
					return res, nil
				}).Build(),
		)).
		// 没有 else 分支
		Build()

	assert.NoError(t, err)

	// 条件为 false，且没有 else 分支，应该返回 null
	ifNode.SetCondition(func(ctx node.WorkflowContext) bool {
		return false
	})

	workflow := Of(ifNode)
	ag := NewAgentBuilder("test-if-null").Workflow(workflow).Build()

	execConfig := &executor.Config{
		MaxConcurrency: 2,
	}
	exec := NewAgentExecutor(ag, execConfig)

	input := value.NewObjectValue()
	response, err := exec.Exec(input)

	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.True(t, response.NodeValue.IsNull())
}

// TestNestedIFNode 测试嵌套 IFNode
func TestNestedIFNode(t *testing.T) {
	// 等级 A
	gradeA := node.NewFunctionNodeBuilder("gradeA").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutString("grade", "A")
			return res, nil
		}).Build()

	// 等级 B
	gradeB := node.NewFunctionNodeBuilder("gradeB").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutString("grade", "B")
			return res, nil
		}).Build()

	// 等级 F
	gradeF := node.NewFunctionNodeBuilder("gradeF").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutString("grade", "F")
			return res, nil
		}).Build()

	// 内层 IF：区分 A 和 B
	innerIf, err := node.NewIFNodeBuilder("checkTop").
		Condition(func(ctx node.WorkflowContext) bool {
			rootVal := ctx.GetRootValue()
			score := int(rootVal.GetNumber("score"))
			return score >= 90
		}).
		Then(Of(gradeA)).
		Else(Of(gradeB)).
		Build()

	assert.NoError(t, err)

	// 外层 IF：区分及格和不及格
	outerIf, err := node.NewIFNodeBuilder("checkPass").
		Condition(func(ctx node.WorkflowContext) bool {
			rootVal := ctx.GetRootValue()
			score := int(rootVal.GetNumber("score"))
			return score >= 60
		}).
		Then(Of(innerIf)).
		Else(Of(gradeF)).
		Build()

	assert.NoError(t, err)

	workflow := Of(outerIf)
	ag := NewAgentBuilder("test-nested-if").Workflow(workflow).Build()

	execConfig := &executor.Config{
		MaxConcurrency: 2,
	}
	exec := NewAgentExecutor(ag, execConfig)

	testCases := []struct {
		name     string
		score    int
		expected string
	}{
		{"A grade", 95, "A"},
		{"B grade", 80, "B"},
		{"F grade", 50, "F"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := value.NewObjectValue()
			input.PutNumber("score", float64(tc.score))

			response, err := exec.Exec(input)
			assert.NoError(t, err)

			result := response.NodeValue.(*value.ObjectValue).GetString("grade")
			assert.Equal(t, tc.expected, result)
		})
	}
}
