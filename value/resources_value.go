package value

import (
	"encoding/json"
	"net/url"
	"sync"

	"github.com/chuccp/ai-agent/util"
	"github.com/spf13/cast"
)

// ResourcesValue 通用资源值，可存放任意资源标识（URL、文件路径、数据库ID等）
type ResourcesValue struct {
	NodeValueBase
	resources []string
	mu        sync.RWMutex
}

// NewResourcesValue 创建资源值
func NewResourcesValue() *ResourcesValue {
	return &ResourcesValue{
		resources: make([]string, 0),
	}
}

// NewResourcesValueFromSlice 从字符串切片创建资源值
func NewResourcesValueFromSlice(resources []string) *ResourcesValue {
	rv := NewResourcesValue()
	rv.resources = append(rv.resources, resources...)
	return rv
}

func (r *ResourcesValue) IsResources() bool {
	return true
}

func (r *ResourcesValue) IsFiles() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.resources {
		if util.IsFilePath(s) {
			return true
		}
	}
	return false
}

func (r *ResourcesValue) IsUrls() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.resources {
		if util.IsURL(s) || util.IsFilePath(s) {
			return true
		}
	}
	return false
}

func (r *ResourcesValue) AsResources() *ResourcesValue {
	return r
}
func (r *ResourcesValue) AsInts() []int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]int, 0, len(r.resources))
	for i, s := range r.resources {
		result[i] = cast.ToInt(s)
	}
	return result
}
func (r *ResourcesValue) AsUints() []uint {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]uint, len(r.resources))
	for index, s := range r.resources {
		result[index] = cast.ToUint(s)
	}
	return result
}

// AsUrls 将资源值转换为UrlsValue，只保留能解析为URL的字符串
func (r *ResourcesValue) AsUrls() *UrlsValue {
	uv := NewUrlsValue()
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.resources {
		if util.IsURL(s) || util.IsFilePath(s) {
			uv.ResourcesValue.resources = append(uv.ResourcesValue.resources, s)
		}
	}
	return uv
}

// AsFiles 将资源值转换为FilesValue，只保留文件路径格式的字符串
func (r *ResourcesValue) AsFiles() *FilesValue {
	fv := NewFilesValue()
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.resources {
		if util.IsFilePath(s) {
			fv.ResourcesValue.resources = append(fv.ResourcesValue.resources, s)
		}
	}
	return fv
}

// Add 添加资源
func (r *ResourcesValue) Add(resource string) *ResourcesValue {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources = append(r.resources, resource)
	return r
}

// Adds 添加多个资源
func (r *ResourcesValue) Adds(resources []string) *ResourcesValue {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources = append(r.resources, resources...)
	return r
}

// AddAll 添加所有资源
func (r *ResourcesValue) AddAll(other *ResourcesValue) {
	if other == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	r.resources = append(r.resources, other.resources...)
}

// Get 获取资源
func (r *ResourcesValue) Get(index int) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if index < 0 || index >= len(r.resources) {
		return ""
	}
	return r.resources[index]
}

// Size 获取大小
func (r *ResourcesValue) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.resources)
}

// IsEmpty 是否为空
func (r *ResourcesValue) IsEmpty() bool {
	return r.Size() == 0
}

// Resources 获取所有资源
func (r *ResourcesValue) Resources() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]string, len(r.resources))
	copy(result, r.resources)
	return result
}

// ForEach 遍历
func (r *ResourcesValue) ForEach(fn func(index int, resource string) bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i, s := range r.resources {
		if !fn(i, s) {
			break
		}
	}
}

// Filter 过滤
func (r *ResourcesValue) Filter(fn func(index int, resource string) bool) *ResourcesValue {
	newRV := NewResourcesValue()
	r.ForEach(func(index int, resource string) bool {
		if fn(index, resource) {
			newRV.Add(resource)
		}
		return true
	})
	return newRV
}

// String 返回字符串表示
func (r *ResourcesValue) String() string {
	return string(r.ToJSON())
}

// ToJSON 返回JSON字符串表示
func (r *ResourcesValue) ToJSON() json.RawMessage {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, _ := json.Marshal(r.resources)
	return data
}

// UrlsValue URL值
type UrlsValue struct {
	*ResourcesValue
}

// NewUrlsValue 创建URL值
func NewUrlsValue() *UrlsValue {
	return &UrlsValue{
		ResourcesValue: NewResourcesValue(),
	}
}

// NewUrlsValueFromSlice 从URL切片创建URL值
func NewUrlsValueFromSlice(urls []url.URL) *UrlsValue {
	uv := NewUrlsValue()
	for _, u := range urls {
		uv.ResourcesValue.resources = append(uv.ResourcesValue.resources, u.String())
	}
	return uv
}

// NewUrlsValueFromStrings 从字符串切片创建URL值
func NewUrlsValueFromStrings(strs []string) *UrlsValue {
	uv := NewUrlsValue()
	uv.ResourcesValue.resources = append(uv.ResourcesValue.resources, strs...)
	return uv
}

func (u *UrlsValue) IsUrls() bool {
	return true
}

func (u *UrlsValue) AsUrls() *UrlsValue {
	return u
}

// Add 添加URL
func (u *UrlsValue) Add(v url.URL) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.ResourcesValue.resources = append(u.ResourcesValue.resources, v.String())
}

// Adds 添加多个URL
func (u *UrlsValue) Adds(items []url.URL) {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, v := range items {
		u.ResourcesValue.resources = append(u.ResourcesValue.resources, v.String())
	}
}

func StringToURL(value string) (*url.URL, error) {
	return url.Parse(value)
}

// AddAll 添加所有URL
func (u *UrlsValue) AddAll(other *UrlsValue) {
	if other == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	u.ResourcesValue.resources = append(u.ResourcesValue.resources, other.ResourcesValue.resources...)
}

// Get 获取URL
func (u *UrlsValue) Get(index int) *url.URL {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if index < 0 || index >= len(u.ResourcesValue.resources) {
		return nil
	}
	parsed, err := url.Parse(u.ResourcesValue.resources[index])
	if err != nil {
		return nil
	}
	return parsed
}

// Size 获取大小
func (u *UrlsValue) Size() int {
	return u.ResourcesValue.Size()
}

// IsEmpty 是否为空
func (u *UrlsValue) IsEmpty() bool {
	return u.ResourcesValue.IsEmpty()
}

// URLs 获取所有URL
func (u *UrlsValue) URLs() []url.URL {
	strs := u.ResourcesValue.Resources()
	result := make([]url.URL, 0, len(strs))
	for _, s := range strs {
		if parsed, err := url.Parse(s); err == nil {
			result = append(result, *parsed)
		}
	}
	return result
}

// Strings 获取所有URL字符串
func (u *UrlsValue) Strings() []string {
	return u.ResourcesValue.Resources()
}

// ToJSON 返回JSON字符串表示
func (u *UrlsValue) ToJSON() json.RawMessage {
	return u.ResourcesValue.ToJSON()
}

// FilesValue 文件值
type FilesValue struct {
	UrlsValue
}

// NewFilesValue 创建文件值
func NewFilesValue() *FilesValue {
	return &FilesValue{
		UrlsValue: *NewUrlsValue(),
	}
}

// NewFilesValueFromPaths 从路径创建文件值
func NewFilesValueFromPaths(paths []string) *FilesValue {
	fv := NewFilesValue()
	fv.ResourcesValue.resources = append(fv.ResourcesValue.resources, paths...)
	return fv
}

func (f *FilesValue) IsFiles() bool {
	return true
}

func (f *FilesValue) AsFiles() *FilesValue {
	return f
}

func (f *FilesValue) AsUrls() *UrlsValue {
	uv := NewUrlsValue()
	f.mu.RLock()
	defer f.mu.RUnlock()
	uv.ResourcesValue.resources = append(uv.ResourcesValue.resources, f.ResourcesValue.resources...)
	return uv
}

// Paths 获取所有文件路径
func (f *FilesValue) Paths() []string {
	return f.ResourcesValue.Resources()
}

// ToJSON 返回JSON字符串表示，返回文件路径数组
func (f *FilesValue) ToJSON() json.RawMessage {
	return f.ResourcesValue.ToJSON()
}

// AddAllFiles 添加所有文件值
func (f *FilesValue) AddAllFiles(other *FilesValue) {
	f.ResourcesValue.AddAll(other.ResourcesValue)
}

// UrlsValueFrom URL值来源
type UrlsValueFrom struct {
	NodeID string
	From   string
}

// NewUrlsValueFrom 创建URL值来源
func NewUrlsValueFrom(nodeID, from string) *UrlsValueFrom {
	return &UrlsValueFrom{
		NodeID: nodeID,
		From:   from,
	}
}

// ResourcesValueFrom 资源值来源
type ResourcesValueFrom struct {
	NodeID string
	From   string
}

// NewResourcesValueFrom 创建资源值来源
func NewResourcesValueFrom(nodeID, from string) *ResourcesValueFrom {
	return &ResourcesValueFrom{
		NodeID: nodeID,
		From:   from,
	}
}

// FilesValueFrom 文件值来源
type FilesValueFrom struct {
	UrlsValueFrom
}

// NewFilesValueFrom 创建文件值来源
func NewFilesValueFrom(nodeID, from string) *FilesValueFrom {
	return &FilesValueFrom{
		UrlsValueFrom: UrlsValueFrom{
			NodeID: nodeID,
			From:   from,
		},
	}
}
