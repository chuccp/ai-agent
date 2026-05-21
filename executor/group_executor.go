package executor

import (
	"context"
	"strconv"

	"emperror.dev/errors"
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
func (e *GroupExecutor) ExecBatch(statusGroup *graph.NodeStatusGroup, inputs []*value.ObjectValue, isOrder bool) (*value.ArrayValue, bool, error) {

	allArr := value.NewFixArrayValue(len(inputs))
	share := value.NewArrayValue()
	// 单输入直接执行
	if len(inputs) == 1 {
		nodeExecutor := e.executeSingle(0, inputs[0], share)
		statusGroup.AddChildren(nodeExecutor.GetNodeStatus())
		result, err := nodeExecutor.Exec(e.pool2)
		if err != nil {
			return allArr, false, err
		}
		if result == nil || result.IsNull() {
			return allArr, false, nil
		}
		share.Add(result)
		allArr.AddIndex(0, result)
		return allArr, true, nil
	}
	nodeExecutors := make([]*NodeExecutor, len(inputs))
	for i, input := range inputs {
		nodeExecutor := e.executeSingle(i, input, share)
		statusGroup.AddChildren(nodeExecutor.GetNodeStatus())
		nodeExecutors[i] = nodeExecutor
	}
	if isOrder {
		for i := 0; i < len(nodeExecutors); i++ {
			ne := nodeExecutors[i]
			nodeValue, err := ne.Exec(e.pool2)
			if err != nil {
				return allArr, false, err
			}
			if nodeValue == nil || nodeValue.IsNull() {

				return allArr, false, nil
			}
			share.Add(nodeValue)
			allArr.AddIndex(i, nodeValue)
		}
		return allArr, true, nil
	}

	err := e.pool2.WaitGOIndex(len(inputs), func(index int) error {
		nodeValue, err := nodeExecutors[index].Exec(e.pool2)
		if err != nil {
			return errors.Append(err, errors.New("执行失败:"+strconv.Itoa(index)))
		}
		allArr.AddIndex(index, nodeValue)
		return nil
	})
	hasNil := false
	allArr.ForEach(func(index int, value value.NodeValue) bool {
		if !hasNil {
			if value == nil || value.IsNull() {
				hasNil = true
			}
		}
		share.Add(value)
		return true
	})
	if err != nil {
		return allArr, false, err
	}
	return allArr, !hasNil, nil
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
