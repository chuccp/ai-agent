package types

// FieldType 字段类型
type FieldType int

const (
	FieldTypeText FieldType = iota
	FieldTypeFile
	FieldTypeFiles
	FieldTypeFrom
	FieldTypeObject
	FieldTypeDate
	FieldTypeNumber
	FieldTypeBoolean
	FieldTypeArray
)

func (t FieldType) String() string {
	switch t {
	case FieldTypeText:
		return "TEXT"
	case FieldTypeFile:
		return "FILE"
	case FieldTypeFiles:
		return "FILES"
	case FieldTypeFrom:
		return "FROM"
	case FieldTypeObject:
		return "OBJECT"
	case FieldTypeDate:
		return "DATE"
	case FieldTypeNumber:
		return "NUMBER"
	case FieldTypeBoolean:
		return "BOOLEAN"
	case FieldTypeArray:
		return "ARRAY"
	default:
		return "UNKNOWN"
	}
}