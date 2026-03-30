package executor

import (
	"context"
	"strconv"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/node"
	pool2 "github.com/chuccp/ai-agent/pool"
	"github.com/chuccp/ai-agent/value"
)

// GroupExecutor 组执行器
type GroupExecutor struct {
	nodes         []node.Node
	parentID      string
	parentContext *Context
	ctxContext    context.Context
	pool2         *pool2.GOPool
}

// NewGroupExecutor 创建组执行器
func NewGroupExecutor(nodes []node.Node, parentID string, parentContext *Context, ctxContext context.Context) *GroupExecutor {
	return &GroupExecutor{
		nodes:         nodes,
		parentID:      parentID,
		parentContext: parentContext,
		ctxContext:    ctxContext,
		pool2:         parentContext.pool2,
	}
}

// ExecBatch 批量执行（使用goroutine并发，带panic恢复）
func (e *GroupExecutor) ExecBatch(statusGroup *graph.NodeStatusGroup, inputs []*value.ObjectValue, isOrder bool) (value.NodeValue, bool, error) {

	arr := value.NewArrayValue()
	// 单输入直接执行
	if len(inputs) == 1 {
		nodeExecutor := e.executeSingle(0, inputs[0], arr)
		statusGroup.AddChildren(nodeExecutor.GetNodeStatus())
		result, err := nodeExecutor.Exec(e.pool2)
		if err != nil {
			return nil, false, err
		}
		if result == nil || result.IsNull() {
			return value.NullValue, false, nil
		}
		arr.Add(result)
		return arr, true, nil
	}
	nodeExecutors := make([]*NodeExecutor, len(inputs))
	for i, input := range inputs {
		nodeExecutor := e.executeSingle(i, input, arr)
		statusGroup.AddChildren(nodeExecutor.GetNodeStatus())
		nodeExecutors[i] = nodeExecutor
	}
	if isOrder {
		for i := 0; i < len(nodeExecutors); i++ {
			nodeValue, err := nodeExecutors[i].Exec(e.pool2)
			if err != nil {
				return nil, false, err
			}
			if nodeValue == nil || nodeValue.IsNull() {
				return value.NullValue, false, nil
			}
			arr.Add(nodeValue)
		}
		return arr, true, nil
	}
	results := make([]value.NodeValue, len(inputs))
	for i := range results {
		results[i] = value.NullValue
	}
	err := e.pool2.WaitGOIndex(len(inputs), func(index int) error {
		nodeValue, err := nodeExecutors[index].Exec(e.pool2)
		if err != nil {
			return err
		}
		results[index] = nodeValue
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	for _, result := range results {
		if result == nil || result.IsNull() {
			return value.NullValue, false, nil
		}
		arr.Add(result)
	}
	return arr, true, nil
}

func (e *GroupExecutor) executeSingle(index int, input *value.ObjectValue, shareValue *value.ArrayValue) *NodeExecutor {
	childParentId := e.buildChildParentID(index)
	childContext := e.parentContext.CreateChildContext(e.nodes, input, shareValue, childParentId).(*Context)
	nodeExecutor := NewNodeExecutor(index, e.nodes, childContext)
	return nodeExecutor
}

// buildChildParentID 构建子父ID
func (e *GroupExecutor) buildChildParentID(index int) string {
	if e.parentID == "" {
		return strconv.Itoa(index)
	}
	return e.parentID + "_" + strconv.Itoa(index)
}
