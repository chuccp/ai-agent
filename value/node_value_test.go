package value

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- ObjectValue Tests ---

func TestObjectValuePutChaining(t *testing.T) {
	obj := NewObjectValue()
	result := obj.
		PutString("name", "Alice").
		PutNumber("age", 30).
		PutBool("active", true)

	assert.Same(t, obj, result, "Put* methods should return the same ObjectValue for chaining")
	assert.Equal(t, "Alice", obj.GetString("name"))
	assert.Equal(t, float64(30), obj.GetNumber("age"))
	assert.Equal(t, true, obj.GetBool("active"))
}

func TestObjectValuePutObjectAndArray(t *testing.T) {
	inner := NewObjectValue().PutString("city", "Beijing")
	arr := NewArrayValue()
	arr.AddText("item1")
	arr.AddNumber(42)

	obj := NewObjectValue().
		PutObject("address", inner).
		PutArray("items", arr)

	assert.Equal(t, "Beijing", obj.GetObject("address").GetString("city"))
	assert.Equal(t, 2, obj.GetArray("items").Size())
}

func TestObjectValueAddAll(t *testing.T) {
	obj1 := NewObjectValue().PutString("a", "1").PutNumber("b", 2)
	obj2 := NewObjectValue().PutString("c", "3")

	obj1.AddAll(obj2)

	assert.Equal(t, "1", obj1.GetString("a"))
	assert.Equal(t, float64(2), obj1.GetNumber("b"))
	assert.Equal(t, "3", obj1.GetString("c"))
}

func TestObjectValueAddAllNil(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1")
	result := obj.AddAll(nil)

	assert.Same(t, obj, result)
	assert.Equal(t, "1", obj.GetString("a"))
}

func TestObjectValueAddAllIFNULL(t *testing.T) {
	obj1 := NewObjectValue().PutString("a", "original").PutString("b", "2")
	obj2 := NewObjectValue().PutString("a", "override").PutString("c", "3")

	obj1.AddAllIFNULL(obj2)

	// "a" already exists, should not be overridden
	assert.Equal(t, "original", obj1.GetString("a"))
	// "c" is new, should be added
	assert.Equal(t, "3", obj1.GetString("c"))
}

func TestObjectValueClear(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1").PutNumber("b", 2)
	obj.Clear()

	assert.Equal(t, 0, len(obj.Keys()))
}

func TestObjectValueKeys(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1").PutString("b", "2").PutString("c", "3")
	keys := obj.Keys()

	assert.Len(t, keys, 3)
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	assert.True(t, keySet["a"])
	assert.True(t, keySet["b"])
	assert.True(t, keySet["c"])
}

func TestObjectValueForEach(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1").PutString("b", "2").PutString("c", "3")

	count := 0
	obj.ForEach(func(key string, value NodeValue) bool {
		count++
		return true
	})
	assert.Equal(t, 3, count)
}

func TestObjectValueForEachBreak(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1").PutString("b", "2").PutString("c", "3")

	// Map iteration order is not guaranteed, so we just verify break works
	count := 0
	obj.ForEach(func(key string, value NodeValue) bool {
		count++
		return count < 2
	})
	// Break after first item: count should be 1 or 2 depending on which key is iterated first
	assert.True(t, count >= 1 && count <= 2, "expected 1-2 iterations, got %d", count)
}

func TestObjectValueFromMap(t *testing.T) {
	m := map[string]interface{}{
		"name":   "test",
		"count":  10,
		"active": true,
	}
	obj := NewObjectValueFromMap(m)

	assert.Equal(t, "test", obj.GetString("name"))
	assert.Equal(t, float64(10), obj.GetNumber("count"))
	assert.Equal(t, true, obj.GetBool("active"))
}

func TestObjectValueFromJSON(t *testing.T) {
	obj := NewObjectValue()
	err := obj.FromJSON([]byte(`{"name":"test","count":10,"active":true}`))

	assert.NoError(t, err)
	assert.Equal(t, "test", obj.GetString("name"))
	assert.Equal(t, float64(10), obj.GetNumber("count"))
	assert.Equal(t, true, obj.GetBool("active"))
}

func TestParseObjectValue(t *testing.T) {
	obj, err := ParseObjectValue([]byte(`{"key":"value"}`))
	assert.NoError(t, err)
	assert.Equal(t, "value", obj.GetString("key"))
}

func TestParseObjectValueEmptyString(t *testing.T) {
	obj, err := ParseStrObjectValue("")
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Equal(t, 0, len(obj.Keys()))
}

// --- ObjectValue Template Tests ---

func TestObjectValueExecuteTemplateWithDollarFormat(t *testing.T) {
	obj := NewObjectValue().
		PutString("name", "Alice").
		PutNumber("count", 100)

	result, err := obj.ExecuteTemplateWithDollarFormat("Hello, ${name}! Your count is ${count}.")

	assert.NoError(t, err)
	assert.Equal(t, "Hello, Alice! Your count is 100.", result)
}

func TestObjectValueExecuteTemplateWithGoSyntax(t *testing.T) {
	obj := NewObjectValue().PutString("name", "Bob")

	result, err := obj.ExecuteTemplateWithDollarFormat("Hello, {{.name}}!")

	assert.NoError(t, err)
	assert.Equal(t, "Hello, Bob!", result)
}

func TestObjectValueExecuteTemplateUnresolved(t *testing.T) {
	obj := NewObjectValue().PutString("name", "Bob")

	_, err := obj.ExecuteTemplateWithDollarFormat("Hello, ${name}! Missing: ${missing}")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unresolved")
}

func TestObjectValueExecuteTemplateEmpty(t *testing.T) {
	obj := NewObjectValue()
	result, err := obj.ExecuteTemplateWithDollarFormat("")

	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

// --- ObjectValue GetOrString ---

func TestObjectValueGetOrString(t *testing.T) {
	obj := NewObjectValue().
		PutString("a", "").
		PutString("b", "found")

	result := obj.GetOrString("a", "b", "c")
	assert.Equal(t, "found", result)
}

func TestObjectValueGetOrStringAllEmpty(t *testing.T) {
	obj := NewObjectValue().
		PutString("a", "").
		PutString("b", "")

	result := obj.GetOrString("a", "b")
	assert.Equal(t, "", result)
}

// --- ArrayValue Tests ---

func TestArrayValueAddAndGet(t *testing.T) {
	arr := NewArrayValue()
	arr.Add(NewTextValue("hello"))
	arr.Add(NewNumberValue(42))

	assert.Equal(t, 2, arr.Size())
	assert.Equal(t, "hello", arr.Get(0).AsText().Text)
	assert.Equal(t, float64(42), arr.Get(1).AsNumber().Float64())
}

func TestArrayValueOutOfBounds(t *testing.T) {
	arr := NewArrayValue()
	arr.Add(NewTextValue("hello"))

	assert.True(t, arr.Get(-1).IsNull())
	assert.True(t, arr.Get(10).IsNull())
}

func TestArrayValueIsEmpty(t *testing.T) {
	arr := NewArrayValue()
	assert.True(t, arr.IsEmpty())

	arr.AddText("item")
	assert.False(t, arr.IsEmpty())
}

func TestArrayValueValues(t *testing.T) {
	arr := NewArrayValue()
	arr.AddText("a")
	arr.AddText("b")

	vals := arr.Values()
	assert.Len(t, vals, 2)
}

func TestArrayValueStringValues(t *testing.T) {
	arr := NewArrayValue()
	arr.AddText("a")
	arr.AddNumber(1)
	arr.AddText("b")

	strs := arr.StringValues()
	assert.Len(t, strs, 2)
	assert.Equal(t, []string{"a", "b"}, strs)
}

func TestArrayValueHas(t *testing.T) {
	arr := NewArrayValue()
	arr.AddText("hello")
	arr.AddNumber(42)

	assert.True(t, arr.HasString("hello"))
	assert.True(t, arr.HasNumber(42))
	assert.False(t, arr.HasString("world"))
	assert.False(t, arr.HasNumber(99))
}

func TestArrayValueFind(t *testing.T) {
	arr := NewArrayValue()
	arr.AddText("a")
	arr.AddText("b")
	arr.AddText("c")

	idx := arr.Find(NewTextValue("b"))
	assert.Equal(t, 1, idx)

	idx = arr.Find(NewTextValue("z"))
	assert.Equal(t, -1, idx)
}

func TestArrayValueFilter(t *testing.T) {
	arr := NewArrayValue()
	arr.AddNumber(1)
	arr.AddNumber(2)
	arr.AddNumber(3)
	arr.AddNumber(4)

	filtered := arr.Filter(func(index int, value NodeValue) bool {
		return value.AsNumber().Float64() > 2
	})

	assert.Equal(t, 2, filtered.Size())
	assert.Equal(t, float64(3), filtered.Get(0).AsNumber().Float64())
	assert.Equal(t, float64(4), filtered.Get(1).AsNumber().Float64())
}

func TestArrayValueAddAll(t *testing.T) {
	arr1 := NewArrayValue()
	arr1.AddText("a")
	arr2 := NewArrayValue()
	arr2.AddText("b")
	arr2.AddText("c")

	arr1.AddAll(arr2)
	assert.Equal(t, 3, arr1.Size())
}

func TestArrayValueFromSlice(t *testing.T) {
	arr := NewArrayValueFromSlice([]NodeValue{
		NewTextValue("a"),
		NewTextValue("b"),
	})
	assert.Equal(t, 2, arr.Size())
}

func TestArrayValueFindMaxByScore(t *testing.T) {
	arr := NewArrayValue()
	arr.AddNumber(10)
	arr.AddNumber(50)
	arr.AddNumber(30)

	maxVal, maxScore, found := arr.FindMaxByScore(func(v NodeValue) int {
		return int(v.AsNumber().Float64())
	})

	assert.True(t, found)
	assert.Equal(t, 50, maxScore)
	assert.Equal(t, float64(50), maxVal.AsNumber().Float64())
}

func TestArrayValueFindMaxByScoreEmpty(t *testing.T) {
	arr := NewArrayValue()
	_, _, found := arr.FindMaxByScore(func(v NodeValue) int { return 0 })
	assert.False(t, found)
}

func TestParseArrayValue(t *testing.T) {
	arr, err := ParseArrayValue([]byte(`[1,"two",true,null]`))

	assert.NoError(t, err)
	assert.Equal(t, 4, arr.Size())
	assert.Equal(t, float64(1), arr.Get(0).AsNumber().Float64())
	assert.Equal(t, "two", arr.Get(1).AsText().Text)
	assert.True(t, arr.Get(2).AsBool().Value)
	assert.True(t, arr.Get(3).IsNull())
}

// --- NumberValue Type Conversion Tests ---

func TestNumberValueIntConversions(t *testing.T) {
	n := NewIntValue(42)
	assert.True(t, n.IsInt())
	assert.True(t, n.IsInteger())
	assert.True(t, n.IsSigned())
	assert.False(t, n.IsFloat())
	assert.False(t, n.IsUnsigned())
	assert.Equal(t, 42, n.Int())
	assert.Equal(t, int8(42), n.Int8())
	assert.Equal(t, int64(42), n.Int64())
	assert.Equal(t, uint(42), n.Uint())
	assert.Equal(t, float64(42), n.Float64())
}

func TestNumberValueFloatConversions(t *testing.T) {
	n := NewFloat64Value(3.14)
	assert.True(t, n.IsFloat64())
	assert.True(t, n.IsFloat())
	assert.False(t, n.IsInteger())
	assert.Equal(t, 3, n.Int())
	assert.Equal(t, 3.14, n.Float64())
	assert.Equal(t, float32(3.14), n.Float32())
}

func TestNumberValueCrossTypeConversion(t *testing.T) {
	// int64 -> all others
	n := NewInt64Value(100)
	assert.Equal(t, 100, n.Int())
	assert.Equal(t, int8(100), n.Int8())
	assert.Equal(t, uint16(100), n.Uint16())
	assert.Equal(t, float32(100), n.Float32())

	// uint64 -> all others
	u := NewUint64Value(200)
	assert.Equal(t, 200, u.Int())
	assert.Equal(t, uint64(200), u.Uint64())
	assert.Equal(t, float64(200), u.Float64())

	// float32 -> all others
	f := NewFloat32Value(7.9)
	assert.Equal(t, 7, f.Int())
	assert.InDelta(t, float64(7.9), f.Float64(), 0.01)
}

func TestNumberValueString(t *testing.T) {
	assert.Equal(t, "42", NewIntValue(42).String())
	assert.Equal(t, "3.14", NewFloat64Value(3.14).String())
}

func TestNumberValueEquals(t *testing.T) {
	// Same type, same value
	assert.True(t, NewIntValue(42).Equals(NewIntValue(42)))
	// Same type, different value
	assert.False(t, NewIntValue(42).Equals(NewIntValue(43)))
	// Different type, same numeric value
	assert.True(t, NewIntValue(42).Equals(NewFloat64Value(42)))
	// nil
	assert.False(t, NewIntValue(42).Equals(nil))
}

func TestParseNumber(t *testing.T) {
	n, err := ParseNumber("42")
	assert.NoError(t, err)
	assert.True(t, n.IsInt64())
	assert.Equal(t, int64(42), n.Int64())

	f, err := ParseNumber("3.14")
	assert.NoError(t, err)
	assert.True(t, f.IsFloat64())
	assert.Equal(t, float64(3.14), f.Float64())

	_, err = ParseNumber("not-a-number")
	assert.Error(t, err)
}

func TestMustNumber(t *testing.T) {
	n := MustNumber("123")
	assert.Equal(t, int64(123), n.Int64())

	zero := MustNumber("invalid")
	assert.Equal(t, 0, zero.Int())
}

// --- FindValue / Path Tests ---

func TestFindValueByPath(t *testing.T) {
	obj := NewObjectValue()
	obj.PutString("name", "Alice")
	obj.PutObject("address", NewObjectValue().
		PutString("city", "Beijing"),
	)

	result := obj.FindValue("address.city")
	assert.Equal(t, "Beijing", result.AsText().Text)
}

func TestFindValueByPathWithDollarPrefix(t *testing.T) {
	obj := NewObjectValue().PutString("key", "value")
	result := obj.FindValue("$.key")
	assert.Equal(t, "value", result.AsText().Text)
}

func TestFindValueByPathArrayIndex(t *testing.T) {
	arr := NewArrayValue()
	arr.AddText("a")
	arr.AddText("b")
	arr.AddText("c")

	result := arr.FindValue("[1]")
	assert.Equal(t, "b", result.AsText().Text)
}

func TestFindValueByPathNestedArray(t *testing.T) {
	obj := NewObjectValue()
	obj.PutArray("items", NewArrayValueFromSlice([]NodeValue{
		NewObjectValue().PutString("name", "first"),
		NewObjectValue().PutString("name", "second"),
	}))

	result := obj.FindValue("items[0].name")
	assert.Equal(t, "first", result.AsText().Text)
}

func TestFindValueByPathEmptyPath(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1")
	result := obj.FindValue("")
	assert.True(t, result.IsObject())
}

func TestFindValueByPathInvalidPath(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1")
	result := obj.FindValue("nonexistent")
	assert.Nil(t, result)
}

func TestFindValueByPathNonObjectAtPath(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1")
	result := obj.FindValue("a.b")
	assert.Nil(t, result)
}

func TestFindValueByPathNonArrayAtIndex(t *testing.T) {
	obj := NewObjectValue().PutString("a", "1")
	result := obj.FindValue("a[0]")
	assert.Nil(t, result)
}

// --- FromJSON Tests ---

func TestFromJSON(t *testing.T) {
	tests := []struct {
		input    string
		check    func(t *testing.T, v NodeValue)
	}{
		{`"hello"`, func(t *testing.T, v NodeValue) {
			assert.True(t, v.IsText())
			assert.Equal(t, "hello", v.AsText().Text)
		}},
		{`42`, func(t *testing.T, v NodeValue) {
			assert.True(t, v.IsNumber())
			assert.Equal(t, float64(42), v.AsNumber().Float64())
		}},
		{`true`, func(t *testing.T, v NodeValue) {
			assert.True(t, v.IsBool())
			assert.True(t, v.AsBool().Value)
		}},
		{`null`, func(t *testing.T, v NodeValue) {
			assert.True(t, v.IsNull())
		}},
		{`[1,2,3]`, func(t *testing.T, v NodeValue) {
			assert.True(t, v.IsArray())
			assert.Equal(t, 3, v.AsArray().Size())
		}},
		{`{"a":1}`, func(t *testing.T, v NodeValue) {
			assert.True(t, v.IsObject())
			assert.Equal(t, float64(1), v.AsObject().GetNumber("a"))
		}},
	}

	for _, tt := range tests {
		v, err := FromJSON([]byte(tt.input))
		assert.NoError(t, err)
		tt.check(t, v)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	_, err := FromJSON([]byte(`{invalid}`))
	assert.Error(t, err)
}

// --- Equals Tests ---

func TestEqualsNil(t *testing.T) {
	assert.True(t, Equals(nil, nil))
	assert.False(t, Equals(nil, NewTextValue("a")))
	assert.False(t, Equals(NewTextValue("a"), nil))
}

func TestEqualsText(t *testing.T) {
	assert.True(t, Equals(NewTextValue("a"), NewTextValue("a")))
	assert.False(t, Equals(NewTextValue("a"), NewTextValue("b")))
}

func TestEqualsBool(t *testing.T) {
	assert.True(t, Equals(NewBoolValue(true), NewBoolValue(true)))
	assert.False(t, Equals(NewBoolValue(true), NewBoolValue(false)))
}

func TestEqualsNull(t *testing.T) {
	assert.True(t, Equals(NullValue, NullValue))
}

func TestEqualsDifferentTypes(t *testing.T) {
	assert.False(t, Equals(NewTextValue("42"), NewNumberValue(42)))
	assert.False(t, Equals(NewBoolValue(true), NewNumberValue(1)))
}

// --- Clone Tests ---

func TestCloneNil(t *testing.T) {
	assert.Nil(t, Clone(nil))
}

func TestCloneText(t *testing.T) {
	orig := NewTextValue("hello")
	cloned := Clone(orig)
	assert.Equal(t, "hello", cloned.AsText().Text)
	assert.False(t, orig == cloned)
}

func TestCloneObject(t *testing.T) {
	orig := NewObjectValue().
		PutString("name", "Alice").
		PutObject("nested", NewObjectValue().PutString("key", "value"))

	cloned := Clone(orig).(*ObjectValue)
	assert.Equal(t, "Alice", cloned.GetString("name"))
	assert.Equal(t, "value", cloned.GetObject("nested").GetString("key"))

	// Modifying clone should not affect original
	cloned.PutString("name", "Bob")
	assert.Equal(t, "Alice", orig.GetString("name"))
}

func TestCloneArray(t *testing.T) {
	orig := NewArrayValue()
	orig.AddText("a")
	orig.AddNumber(1)

	cloned := Clone(orig).(*ArrayValue)
	assert.Equal(t, 2, cloned.Size())
	assert.Equal(t, "a", cloned.Get(0).AsText().Text)

	cloned.AddText("c")
	assert.Equal(t, 2, orig.Size())
}

func TestCloneNumber(t *testing.T) {
	n := NewIntValue(42)
	cloned := Clone(n).(*NumberValue)
	assert.True(t, cloned.IsInt())
	assert.Equal(t, 42, cloned.Int())
}

func TestCloneBool(t *testing.T) {
	b := NewBoolValue(true)
	cloned := Clone(b).(*BoolValue)
	assert.True(t, cloned.Value)
}

// --- UrlsValue Tests ---

func TestUrlsValueIsEmpty(t *testing.T) {
	urls := NewUrlsValue()
	assert.True(t, urls.IsEmpty())
}

// --- StreamNodeValue Tests ---

func TestStreamNodeValue(t *testing.T) {
	s := NewStreamNodeValue()
	assert.True(t, s.IsStream())
	assert.False(t, s.IsDone())

	s.Send(NewTextValue("chunk1"))
	s.Send(NewTextValue("chunk2"))
	s.Close()
	assert.True(t, s.IsDone())

	// Collect should have 2 items
	arr := s.Collect()
	assert.Equal(t, 2, arr.Size())
}
