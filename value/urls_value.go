package value

import (
	"encoding/json"
	"net/url"
	"sync"
)

// UrlsValue URL值
type UrlsValue struct {
	NodeValueBase
	urls []url.URL
	mu   sync.RWMutex
}

// NewUrlsValue 创建URL值
func NewUrlsValue() *UrlsValue {
	return &UrlsValue{
		urls: make([]url.URL, 0),
	}
}

// NewUrlsValueFromSlice 从切片创建URL值
func NewUrlsValueFromSlice(urls []url.URL) *UrlsValue {
	uv := NewUrlsValue()
	uv.urls = append(uv.urls, urls...)
	return uv
}

func (u *UrlsValue) IsUrls() bool {
	return true
}

func (u *UrlsValue) IsFiles() bool {
	return false
}

func (u *UrlsValue) AsUrls() *UrlsValue {
	return u
}

func (u *UrlsValue) AsFiles() *FilesValue {
	return nil
}

// Add 添加URL
func (u *UrlsValue) Add(url url.URL) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.urls = append(u.urls, url)
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
	u.urls = append(u.urls, other.urls...)
}

// Get 获取URL
func (u *UrlsValue) Get(index int) *url.URL {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if index < 0 || index >= len(u.urls) {
		return nil
	}
	return &u.urls[index]
}

// Size 获取大小
func (u *UrlsValue) Size() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return len(u.urls)
}

// IsEmpty 是否为空
func (u *UrlsValue) IsEmpty() bool {
	return u.Size() == 0
}

// URLs 获取所有URL
func (u *UrlsValue) URLs() []url.URL {
	u.mu.RLock()
	defer u.mu.RUnlock()
	result := make([]url.URL, len(u.urls))
	copy(result, u.urls)
	return result
}

// Strings 获取所有URL字符串
func (u *UrlsValue) Strings() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	result := make([]string, 0, len(u.urls))
	for _, url := range u.urls {
		result = append(result, url.String())
	}
	return result
}

func (u *UrlsValue) String() string {
	return string(u.ToJSON())
}

func (u *UrlsValue) ToJSON() json.RawMessage {
	u.mu.RLock()
	defer u.mu.RUnlock()
	arr := make([]string, 0, len(u.urls))
	for _, url := range u.urls {
		arr = append(arr, url.String())
	}
	data, _ := json.Marshal(arr)
	return data
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
	for _, path := range paths {
		if u, err := url.Parse(path); err == nil {
			fv.urls = append(fv.urls, *u)
		}
	}
	return fv
}

func (f *FilesValue) IsFiles() bool {
	return true
}

func (f *FilesValue) AsFiles() *FilesValue {
	return f
}

// Paths 获取所有文件路径
func (f *FilesValue) Paths() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]string, 0, len(f.urls))
	for _, url := range f.urls {
		result = append(result, url.Path)
	}
	return result
}

// AddAllFiles 添加所有文件值
func (f *FilesValue) AddAllFiles(other *FilesValue) {
	if other == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	f.urls = append(f.urls, other.urls...)
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
