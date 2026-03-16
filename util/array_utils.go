package util

// ConcatSlices 连接多个切片
func ConcatSlices[T any](slices ...[]T) []T {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, totalLen)
	var i int
	for _, s := range slices {
		i += copy(result[i:], s)
	}
	return result
}

// AppendSlice 追加元素到切片
func AppendSlice[T any](slice []T, elements ...T) []T {
	return append(slice, elements...)
}

// PrependSlice 在切片前面插入元素
func PrependSlice[T any](slice []T, elements ...T) []T {
	result := make([]T, len(elements)+len(slice))
	copy(result, elements)
	copy(result[len(elements):], slice)
	return result
}

// ContainsSlice 检查切片是否包含元素
func ContainsSlice[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

// IndexOfSlice 查找元素在切片中的索引
func IndexOfSlice[T comparable](slice []T, element T) int {
	for i, v := range slice {
		if v == element {
			return i
		}
	}
	return -1
}

// LastIndexOfSlice 查找元素在切片中的最后索引
func LastIndexOfSlice[T comparable](slice []T, element T) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == element {
			return i
		}
	}
	return -1
}

// RemoveFromSlice 从切片中移除元素（返回新切片）
func RemoveFromSlice[T comparable](slice []T, element T) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if v != element {
			result = append(result, v)
		}
	}
	return result
}

// RemoveAtSlice 从切片中移除指定索引的元素
func RemoveAtSlice[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// ReverseSlice 反转切片
func ReverseSlice[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// ReverseSliceCopy 反转切片（返回新切片）
func ReverseSliceCopy[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// MapSlice 映射切片
func MapSlice[T, R any](slice []T, fn func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// FilterSlice 过滤切片
func FilterSlice[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// ReduceSlice 归约切片
func ReduceSlice[T, R any](slice []T, initial R, fn func(R, T) R) R {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// ForEachSlice 遍历切片
func ForEachSlice[T any](slice []T, fn func(T)) {
	for _, v := range slice {
		fn(v)
	}
}

// AnySlice 检查是否有任意元素满足条件
func AnySlice[T any](slice []T, fn func(T) bool) bool {
	for _, v := range slice {
		if fn(v) {
			return true
		}
	}
	return false
}

// AllSlice 检查是否所有元素都满足条件
func AllSlice[T any](slice []T, fn func(T) bool) bool {
	for _, v := range slice {
		if !fn(v) {
			return false
		}
	}
	return true
}

// CountSlice 统计满足条件的元素数量
func CountSlice[T any](slice []T, fn func(T) bool) int {
	count := 0
	for _, v := range slice {
		if fn(v) {
			count++
		}
	}
	return count
}

// UniqueSlice 去重
func UniqueSlice[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0)
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// FlattenSlice 展平二维切片
func FlattenSlice[T any](slices [][]T) []T {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	result := make([]T, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// ChunkSlice 将切片分成指定大小的块
func ChunkSlice[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}

	chunks := make([][]T, 0, (len(slice)+size-1)/size)
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}