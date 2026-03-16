package types

// OutType 输出类型
type OutType int

const (
	OutTypeText OutType = iota
	OutTypeObject
	OutTypeArrays
	OutTypeJSON
	OutTypeTemplate
	OutTypeFile
	OutTypeFiles
	OutTypeImage
	OutTypeList
)

func (t OutType) String() string {
	switch t {
	case OutTypeText:
		return "TEXT"
	case OutTypeObject:
		return "OBJECT"
	case OutTypeArrays:
		return "ARRAYS"
	case OutTypeJSON:
		return "JSON"
	case OutTypeTemplate:
		return "TEMPLATE"
	case OutTypeFile:
		return "FILE"
	case OutTypeFiles:
		return "FILES"
	case OutTypeImage:
		return "IMAGE"
	case OutTypeList:
		return "LIST"
	default:
		return "UNKNOWN"
	}
}