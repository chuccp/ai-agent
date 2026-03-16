package cache

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// Manager 缓存管理器
type Manager struct {
	cachePath string
	enabled   bool
	mu        sync.RWMutex
}

// NewManager 创建缓存管理器
func NewManager(cachePath string, enabled bool) *Manager {
	m := &Manager{
		cachePath: cachePath,
		enabled:   enabled && cachePath != "",
	}
	if m.enabled {
		os.MkdirAll(cachePath, 0755)
	}
	return m
}

// IsEnabled 是否启用
func (m *Manager) IsEnabled() bool {
	return m.enabled
}

// GenerateCacheKey 生成缓存键
func (m *Manager) GenerateCacheKey(key, nodeID string) string {
	return MD5(key) + "_" + nodeID
}

// GenerateCacheKeyWithParent 生成带父ID的缓存键
func (m *Manager) GenerateCacheKeyWithParent(key, nodeID, parentID string) string {
	if util.IsNotBlank(parentID) {
		return MD5(key) + "_" + parentID + "_" + nodeID
	}
	return m.GenerateCacheKey(key, nodeID)
}

// getCacheFile 获取缓存文件路径
func (m *Manager) getCacheFile(cacheKey string) string {
	return filepath.Join(m.cachePath, cacheKey)
}

// SaveCache 保存缓存
func (m *Manager) SaveCache(cacheKey string, nodeValue value.NodeValue) error {
	if !m.enabled || nodeValue == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	data := nodeValue.ToJSON()
	cacheFile := m.getCacheFile(cacheKey)
	err := os.WriteFile(cacheFile, data, 0644)
	if err != nil {
		log.Printf("保存缓存失败: cacheKey=%s, error=%v", cacheKey, err)
		return err
	}
	log.Printf("保存缓存成功: cacheKey=%s, file=%s, size=%d", cacheKey, cacheFile, len(data))
	return nil
}

// GetCache 获取缓存
func (m *Manager) GetCache(cacheKey string) (value.NodeValue, error) {
	if !m.enabled {
		return nil, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	cacheFile := m.getCacheFile(cacheKey)
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		log.Printf("缓存未命中: cacheKey=%s", cacheKey)
		return nil, nil
	}

	log.Printf("缓存命中: cacheKey=%s, file=%s, size=%d", cacheKey, cacheFile, len(data))
	return value.FromJSON(data)
}

// HasCache 检查缓存是否存在
func (m *Manager) HasCache(cacheKey string) bool {
	if !m.enabled {
		return false
	}
	_, err := os.Stat(m.getCacheFile(cacheKey))
	return err == nil
}

// DeleteCache 删除缓存
func (m *Manager) DeleteCache(cacheKey string) bool {
	if !m.enabled {
		return false
	}
	return os.Remove(m.getCacheFile(cacheKey)) == nil
}

// ClearCache 清空缓存
func (m *Manager) ClearCache() {
	if !m.enabled || m.cachePath == "" {
		return
	}
	files, _ := os.ReadDir(m.cachePath)
	for _, file := range files {
		os.Remove(filepath.Join(m.cachePath, file.Name()))
	}
}

// GetCachePath 获取缓存路径
func (m *Manager) GetCachePath() string {
	return m.cachePath
}

// MD5 计算MD5
func MD5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// CacheFunction 缓存函数
type CacheFunction func() (value.NodeValue, error)

// GetOrCompute 获取或计算缓存
func (m *Manager) GetOrCompute(key, nodeID string, fn CacheFunction) (value.NodeValue, error) {
	cached, err := m.GetCache(m.GenerateCacheKey(key, nodeID))
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	result, err := fn()
	if err != nil {
		return nil, err
	}

	if result != nil {
		m.SaveCache(m.GenerateCacheKey(key, nodeID), result)
	}

	return result, nil
}

// CacheFileFunction 文件缓存函数
type CacheFileFunction func(targetPath string) (string, error)

// CacheFile 缓存文件
func (m *Manager) CacheFile(key, suffix string, fn CacheFileFunction) (string, error) {
	if !m.enabled {
		return fn("")
	}

	fileName := MD5(key)
	if suffix != "" {
		if !strings.HasPrefix(suffix, ".") {
			suffix = "." + suffix
		}
		fileName += suffix
	}

	cacheFile := filepath.Join(m.cachePath, fileName)

	if _, err := os.Stat(cacheFile); err == nil {
		return cacheFile, nil
	}

	return fn(cacheFile)
}
