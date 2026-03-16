package executor

import (
	"sync"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/value"
)

// GroupExecutor 组执行器
type GroupExecutor struct {
	nodes          []node.Node
	parentID       string
	parentContext  *Context
}

// NewGroupExecutor 创建组执行器
func NewGroupExecutor(nodes []node.Node, parentID string, parentContext *Context) *GroupExecutor {
	return &GroupExecutor{
		nodes:         nodes,
		parentID:      parentID,
		parentContext: parentContext,
	}
}

// Exec 执行
func (e *GroupExecutor) Exec() (value.NodeValue, bool, error) {
	// 简化实现
	return value.NullValue, true, nil
}

// ExecBatch 批量执行（使用goroutine并发）
func (e *GroupExecutor) ExecBatch(statusGroup *graph.NodeStatusGroup, inputs []*value.ObjectValue) (value.NodeValue, bool, error) {
	if len(inputs) == 0 {
		return value.NewArrayValue(), true, nil
	}

	results := make([]value.NodeValue, len(inputs))
	for i := range results {
		results[i] = value.NullValue
	}

	statuses := make([][]*graph.NodeStatus, len(inputs))

	// 单输入直接执行
	if len(inputs) == 1 {
		result, statusList, err := e.executeSingle(0, inputs[0])
		if err != nil {
			return nil, false, err
		}
		results[0] = result
		statuses[0] = statusList
		statusGroup.SetChildren(statuses)
		arr := value.NewArrayValue()
		arr.Add(result)
		return arr, true, nil
	}

	// 多输入并发执行
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(inputs))

	for i, input := range inputs {
		wg.Add(1)
		go func(index int, in *value.ObjectValue) {
			defer wg.Done()
			result, statusList, err := e.executeSingle(index, in)
			if err != nil {
				errChan <- err
				return
			}
			mu.Lock()
			results[index] = result
			statuses[index] = statusList
			mu.Unlock()
		}(i, input)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	if err := <-errChan; err != nil {
		return nil, false, err
	}

	statusGroup.SetChildren(statuses)

	arr := value.NewArrayValue()
	for _, result := range results {
		arr.Add(result)
	}

	return arr, true, nil
}

// executeSingle 执行单个
func (e *GroupExecutor) executeSingle(index int, input *value.ObjectValue) (value.NodeValue, []*graph.NodeStatus, error) {
	_ = e.buildChildParentID(index)
	_ = input // 使用input避免编译警告
	// 这里需要创建新的NodeExecutor并执行
	// 简化实现
	return value.NullValue, []*graph.NodeStatus{}, nil
}

// buildChildParentID 构建子父ID
func (e *GroupExecutor) buildChildParentID(index int) string {
	if e.parentID == "" {
		return string(rune(index))
	}
	return e.parentID + "_" + string(rune(index))
}