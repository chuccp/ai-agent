package value

import (
	"strings"
)

// ValueFrom 值来源
type ValueFrom struct {
	NodeID string `json:"nodeId"`
	From   string `json:"from"`
	Param  string `json:"param"`
}

// NewValueFrom 创建值来源
func NewValueFrom(nodeID, from, param string) *ValueFrom {
	return &ValueFrom{
		NodeID: nodeID,
		From:   from,
		Param:  param,
	}
}

// ParseValueFrom 解析值来源字符串
// 格式: "nodeId$.path" 或 "path"
func ParseValueFrom(from, param string) *ValueFrom {
	if !strings.Contains(from, "$") {
		return NewValueFrom(from, "", param)
	}

	parts := strings.SplitN(from, "$", 2)
	if len(parts) > 1 {
		return NewValueFrom(parts[0], "$"+parts[1], param)
	}
	return NewValueFrom("", from, param)
}

// RootValueFrom 创建根值来源
func RootValueFrom(from, param string) *ValueFrom {
	return NewValueFrom("", from, param)
}
func RootALLValueFrom() *ValueFrom {
	return NewValueFrom("", "", "")
}

// NodeValueFrom 创建节点值来源
func NodeValueFrom(nodeID, from, param string) *ValueFrom {
	return NewValueFrom(nodeID, from, param)
}

func (v *ValueFrom) String() string {
	return "ValueFrom{NodeID: " + v.NodeID + ", From: " + v.From + ", Param: " + v.Param + "}"
}
