package types

// NodeType 节点类型
type NodeType int

const (
	NodeTypeSingle NodeType = iota
	NodeTypeMultiple
)

func (nt NodeType) IsSingle() bool {
	return nt == NodeTypeSingle
}

func (nt NodeType) IsMultiple() bool {
	return nt == NodeTypeMultiple
}