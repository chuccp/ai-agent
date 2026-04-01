package value

import (
	"encoding/json"
	"testing"
)

func TestToJSON(t *testing.T) {
	// Test TextValue
	text := NewTextValue("hello")
	assertJSONEqual(t, text.ToJSON(), `"hello"`)

	// Test BoolValue
	boolTrue := NewBoolValue(true)
	assertJSONEqual(t, boolTrue.ToJSON(), `true`)
	boolFalse := NewBoolValue(false)
	assertJSONEqual(t, boolFalse.ToJSON(), `false`)

	// Test NumberValue
	num := NewNumberValue(123.45)
	assertJSONEqual(t, num.ToJSON(), `123.45`)
	intNum := NewIntValue(42)
	assertJSONEqual(t, intNum.ToJSON(), `42`)

	// Test NullValue
	assertJSONEqual(t, NullValue.ToJSON(), `null`)

	// Test ObjectValue
	obj := NewObjectValue()
	obj.Put("name", NewTextValue("test"))
	obj.Put("count", NewIntValue(10))
	obj.Put("active", NewBoolValue(true))
	objJSON := obj.ToJSON()
	var objMap map[string]interface{}
	if err := json.Unmarshal(objJSON, &objMap); err != nil {
		t.Fatalf("ObjectValue ToJSON produced invalid JSON: %v, got: %s", err, objJSON)
	}
	if objMap["name"] != "test" {
		t.Errorf("Expected name=test, got %v", objMap["name"])
	}

	// Test ArrayValue
	arr := NewArrayValue()
	arr.Add(NewTextValue("item1"))
	arr.Add(NewNumberValue(100))
	arr.Add(NewBoolValue(false))
	arrJSON := arr.ToJSON()
	var arrSlice []interface{}
	if err := json.Unmarshal(arrJSON, &arrSlice); err != nil {
		t.Fatalf("ArrayValue ToJSON produced invalid JSON: %v, got: %s", err, arrJSON)
	}
	if len(arrSlice) != 3 {
		t.Errorf("Expected array length 3, got %d", len(arrSlice))
	}

	// Test nested structures
	nestedObj := NewObjectValue()
	nestedObj.Put("items", arr)
	nestedObj.Put("meta", obj)
	nestedJSON := nestedObj.ToJSON()
	var nested map[string]interface{}
	if err := json.Unmarshal(nestedJSON, &nested); err != nil {
		t.Fatalf("Nested ObjectValue ToJSON produced invalid JSON: %v", err)
	}
}

func TestToJSONReturnsString(t *testing.T) {
	// Verify that ToJSON returns a JSON string representation, not map/slice types
	obj := NewObjectValue()
	obj.Put("key", NewTextValue("value"))

	// ToJSON should return json.RawMessage which is a []byte representing JSON string
	jsonBytes := obj.ToJSON()

	// Verify it's valid JSON string format
	if !json.Valid(jsonBytes) {
		t.Errorf("ToJSON returned invalid JSON: %s", jsonBytes)
	}

	// Verify it starts with { (object) not something else
	if jsonBytes[0] != '{' {
		t.Errorf("Object ToJSON should start with '{', got: %s", jsonBytes)
	}

	arr := NewArrayValue()
	arr.Add(NewTextValue("test"))
	arrJSON := arr.ToJSON()

	if !json.Valid(arrJSON) {
		t.Errorf("Array ToJSON returned invalid JSON: %s", arrJSON)
	}

	if arrJSON[0] != '[' {
		t.Errorf("Array ToJSON should start with '[', got: %s", arrJSON)
	}
}

func assertJSONEqual(t *testing.T, got json.RawMessage, want string) {
	if string(got) != want {
		t.Errorf("JSON mismatch: got %s, want %s", got, want)
	}
}

func TestNestedToJSON(t *testing.T) {
	// 测试 ObjectValue 包含 ArrayValue，ArrayValue 中又包含 ObjectValue 的嵌套结构

	// 创建内层 ObjectValue
	innerObj1 := NewObjectValue()
	innerObj1.Put("id", NewIntValue(1))
	innerObj1.Put("name", NewTextValue("item1"))
	innerObj1.Put("price", NewNumberValue(9.99))

	innerObj2 := NewObjectValue()
	innerObj2.Put("id", NewIntValue(2))
	innerObj2.Put("name", NewTextValue("item2"))
	innerObj2.Put("price", NewNumberValue(19.99))

	// 创建 ArrayValue 包含多个 ObjectValue
	itemsArray := NewArrayValue()
	itemsArray.Add(innerObj1)
	itemsArray.Add(innerObj2)

	// 创建外层 ObjectValue 包含 ArrayValue
	outerObj := NewObjectValue()
	outerObj.Put("orderId", NewIntValue(100))
	outerObj.Put("customer", NewTextValue("John"))
	outerObj.Put("items", itemsArray)
	outerObj.Put("total", NewNumberValue(29.98))

	// 获取 JSON
	jsonBytes := outerObj.ToJSON()
	jsonStr := string(jsonBytes)

	t.Logf("Nested JSON output: %s", jsonStr)

	// 验证是有效的 JSON
	if !json.Valid(jsonBytes) {
		t.Fatalf("ToJSON returned invalid JSON: %s", jsonStr)
	}

	// 验证 JSON 格式正确（以 { 开始，以 } 结束）
	if jsonBytes[0] != '{' || jsonBytes[len(jsonBytes)-1] != '}' {
		t.Errorf("JSON should be an object (wrapped in {}), got: %s", jsonStr)
	}

	// 解析 JSON 验证结构
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v, JSON: %s", err, jsonStr)
	}

	// 验证 orderId
	if result["orderId"] != 100.0 {
		t.Errorf("Expected orderId=100, got %v", result["orderId"])
	}

	// 验证 customer
	if result["customer"] != "John" {
		t.Errorf("Expected customer=John, got %v", result["customer"])
	}

	// 验证 items 是数组
	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be array, got %v", result["items"])
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// 验证数组中的第一个对象
	item1, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected item1 to be object, got %v", items[0])
	}
	if item1["name"] != "item1" {
		t.Errorf("Expected item1.name=item1, got %v", item1["name"])
	}
	if item1["price"] != 9.99 {
		t.Errorf("Expected item1.price=9.99, got %v", item1["price"])
	}

	// 验证数组中的第二个对象
	item2, ok := items[1].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected item2 to be object, got %v", items[1])
	}
	if item2["name"] != "item2" {
		t.Errorf("Expected item2.name=item2, got %v", item2["name"])
	}

	t.Logf("Successfully verified nested JSON structure!")
}