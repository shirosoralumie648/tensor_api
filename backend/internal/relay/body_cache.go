package relay

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// BodyCacheMode 缓存模式
type BodyCacheMode int

const (
	// 只使用内存缓存
	BodyCacheModeMemory BodyCacheMode = iota
	// 只使用磁盘缓存
	BodyCacheModeDisk
	// 混合模式：小于阈值使用内存，否则使用磁盘
	BodyCacheModeHybrid
)

// BodyCache 请求体缓存
type BodyCache struct {
	// 缓存模式
	mode BodyCacheMode

	// 内存缓存阈值（字节）
	// 超过此大小的请求体会使用磁盘缓存
	memoryThreshold int64

	// 磁盘缓存目录
	diskCachePath string

	// 内存缓存（key=cacheID, value=[]byte）
	memoryCache map[string][]byte
	memoryCacheMu sync.RWMutex

	// 磁盘缓存元数据（key=cacheID, value={filepath, size, hash}）
	diskCacheMetadata map[string]*DiskCacheEntry
	diskCacheMu       sync.RWMutex

	// 清理配置
	maxCacheSize     int64 // 最大缓存大小（字节）
	maxCacheDuration time.Duration // 最大缓存时间

	// 统计信息
	totalCacheHits   int64
	totalCacheMisses int64
	totalCacheSize   int64
	cacheEvictions   int64

	// 清理 goroutine 控制
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// DiskCacheEntry 磁盘缓存条目
type DiskCacheEntry struct {
	FilePath  string    // 文件路径
	Size      int64     // 文件大小
	Hash      string    // MD5 哈希
	CreatedAt time.Time // 创建时间
}

// NewBodyCache 创建新的请求体缓存
func NewBodyCache(diskCachePath string) *BodyCache {
	// 创建磁盘缓存目录
	_ = os.MkdirAll(diskCachePath, 0755)

	return &BodyCache{
		mode:               BodyCacheModeHybrid,
		memoryThreshold:    1024 * 1024,        // 默认 1MB
		diskCachePath:      diskCachePath,
		memoryCache:        make(map[string][]byte),
		diskCacheMetadata:  make(map[string]*DiskCacheEntry),
		maxCacheSize:       10 * 1024 * 1024 * 1024, // 10GB
		maxCacheDuration:   24 * time.Hour,
		stopChan:           make(chan struct{}),
	}
}

// SetMode 设置缓存模式
func (bc *BodyCache) SetMode(mode BodyCacheMode) {
	bc.mode = mode
}

// SetMemoryThreshold 设置内存缓存阈值
func (bc *BodyCache) SetMemoryThreshold(threshold int64) {
	bc.memoryThreshold = threshold
}

// Start 启动缓存清理 goroutine
func (bc *BodyCache) Start() {
	bc.wg.Add(1)
	go bc.cleanupRoutine()
}

// Stop 停止缓存清理
func (bc *BodyCache) Stop() {
	close(bc.stopChan)
	bc.wg.Wait()
}

// CacheRequestBody 缓存请求体
// 返回: (cacheID, 错误)
func (bc *BodyCache) CacheRequestBody(body io.Reader) (string, error) {
	// 读取整个请求体
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	// 生成缓存 ID
	hash := md5.Sum(bodyBytes)
	cacheID := fmt.Sprintf("%x-%d", hash, time.Now().UnixNano())

	// 根据模式缓存
	switch bc.mode {
	case BodyCacheModeMemory:
		bc.cacheInMemory(cacheID, bodyBytes)

	case BodyCacheModeDisk:
		if err := bc.cacheOnDisk(cacheID, bodyBytes); err != nil {
			return "", err
		}

	case BodyCacheModeHybrid:
		if int64(len(bodyBytes)) < bc.memoryThreshold {
			bc.cacheInMemory(cacheID, bodyBytes)
		} else {
			if err := bc.cacheOnDisk(cacheID, bodyBytes); err != nil {
				return "", err
			}
		}
	}

	return cacheID, nil
}

// GetCachedBody 获取缓存的请求体
// 返回: ([]byte, 错误)
func (bc *BodyCache) GetCachedBody(cacheID string) ([]byte, error) {
	// 首先检查内存缓存
	bc.memoryCacheMu.RLock()
	if data, ok := bc.memoryCache[cacheID]; ok {
		bc.memoryCacheMu.RUnlock()
		atomic.AddInt64(&bc.totalCacheHits, 1)
		return data, nil
	}
	bc.memoryCacheMu.RUnlock()

	// 然后检查磁盘缓存
	bc.diskCacheMu.RLock()
	entry, ok := bc.diskCacheMetadata[cacheID]
	bc.diskCacheMu.RUnlock()

	if ok {
		// 从磁盘读取
		data, err := os.ReadFile(entry.FilePath)
		if err == nil {
			atomic.AddInt64(&bc.totalCacheHits, 1)
			return data, nil
		}
		// 文件不存在，删除元数据
		bc.diskCacheMu.Lock()
		delete(bc.diskCacheMetadata, cacheID)
		bc.diskCacheMu.Unlock()
	}

	atomic.AddInt64(&bc.totalCacheMisses, 1)
	return nil, fmt.Errorf("cache not found: %s", cacheID)
}

// GetCachedBodyReader 获取缓存请求体的读取器
func (bc *BodyCache) GetCachedBodyReader(cacheID string) (io.Reader, error) {
	data, err := bc.GetCachedBody(cacheID)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// InvalidateCache 失效缓存
func (bc *BodyCache) InvalidateCache(cacheID string) error {
	// 从内存缓存中移除
	bc.memoryCacheMu.Lock()
	if data, ok := bc.memoryCache[cacheID]; ok {
		delete(bc.memoryCache, cacheID)
		atomic.AddInt64(&bc.totalCacheSize, -int64(len(data)))
	}
	bc.memoryCacheMu.Unlock()

	// 从磁盘缓存中移除
	bc.diskCacheMu.Lock()
	if entry, ok := bc.diskCacheMetadata[cacheID]; ok {
		delete(bc.diskCacheMetadata, cacheID)
		atomic.AddInt64(&bc.totalCacheSize, -entry.Size)
		bc.diskCacheMu.Unlock()

		// 异步删除文件
		go func() {
			_ = os.Remove(entry.FilePath)
		}()
		return nil
	}
	bc.diskCacheMu.Unlock()

	return fmt.Errorf("cache not found: %s", cacheID)
}

// GetStatistics 获取统计信息
func (bc *BodyCache) GetStatistics() map[string]interface{} {
	bc.memoryCacheMu.RLock()
	memoryCount := len(bc.memoryCache)
	bc.memoryCacheMu.RUnlock()

	bc.diskCacheMu.RLock()
	diskCount := len(bc.diskCacheMetadata)
	bc.diskCacheMu.RUnlock()

	totalHits := atomic.LoadInt64(&bc.totalCacheHits)
	totalMisses := atomic.LoadInt64(&bc.totalCacheMisses)
	hitRate := 0.0
	if totalHits+totalMisses > 0 {
		hitRate = float64(totalHits) / float64(totalHits+totalMisses) * 100
	}

	return map[string]interface{}{
		"memory_count":   memoryCount,
		"disk_count":     diskCount,
		"total_size":     atomic.LoadInt64(&bc.totalCacheSize),
		"total_hits":     totalHits,
		"total_misses":   totalMisses,
		"hit_rate":       hitRate,
		"evictions":      atomic.LoadInt64(&bc.cacheEvictions),
	}
}

// cacheInMemory 在内存中缓存
func (bc *BodyCache) cacheInMemory(cacheID string, data []byte) {
	bc.memoryCacheMu.Lock()
	defer bc.memoryCacheMu.Unlock()

	bc.memoryCache[cacheID] = data
	atomic.AddInt64(&bc.totalCacheSize, int64(len(data)))
}

// cacheOnDisk 在磁盘上缓存
func (bc *BodyCache) cacheOnDisk(cacheID string, data []byte) error {
	// 生成文件路径
	filePath := filepath.Join(bc.diskCachePath, cacheID+".dat")

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	// 计算哈希
	hash := fmt.Sprintf("%x", md5.Sum(data))

	// 记录元数据
	bc.diskCacheMu.Lock()
	defer bc.diskCacheMu.Unlock()

	bc.diskCacheMetadata[cacheID] = &DiskCacheEntry{
		FilePath:  filePath,
		Size:      int64(len(data)),
		Hash:      hash,
		CreatedAt: time.Now(),
	}

	atomic.AddInt64(&bc.totalCacheSize, int64(len(data)))

	return nil
}

// cleanupRoutine 清理 goroutine
func (bc *BodyCache) cleanupRoutine() {
	defer bc.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-bc.stopChan:
			return
		case <-ticker.C:
			bc.cleanup()
		}
	}
}

// cleanup 执行清理
func (bc *BodyCache) cleanup() {
	now := time.Now()

	// 清理过期的磁盘缓存
	bc.diskCacheMu.Lock()
	var toDelete []string
	for cacheID, entry := range bc.diskCacheMetadata {
		if now.Sub(entry.CreatedAt) > bc.maxCacheDuration {
			toDelete = append(toDelete, cacheID)
		}
	}

	for _, cacheID := range toDelete {
		entry := bc.diskCacheMetadata[cacheID]
		delete(bc.diskCacheMetadata, cacheID)
		atomic.AddInt64(&bc.totalCacheSize, -entry.Size)
		atomic.AddInt64(&bc.cacheEvictions, 1)
		bc.diskCacheMu.Unlock()

		// 异步删除文件
		go func(filePath string) {
			_ = os.Remove(filePath)
		}(entry.FilePath)

		bc.diskCacheMu.Lock()
	}
	bc.diskCacheMu.Unlock()

	// 如果总大小超过最大值，进行 LRU 清理
	totalSize := atomic.LoadInt64(&bc.totalCacheSize)
	if totalSize > bc.maxCacheSize {
		bc.evictOldestCaches()
	}
}

// evictOldestCaches 清理最旧的缓存
func (bc *BodyCache) evictOldestCaches() {
	// 收集所有缓存条目及其时间
	type cacheItem struct {
		cacheID   string
		createdAt time.Time
		size      int64
	}

	items := make([]cacheItem, 0)

	// 收集磁盘缓存
	bc.diskCacheMu.RLock()
	for cacheID, entry := range bc.diskCacheMetadata {
		items = append(items, cacheItem{
			cacheID:   cacheID,
			createdAt: entry.CreatedAt,
			size:      entry.Size,
		})
	}
	bc.diskCacheMu.RUnlock()

	// 收集内存缓存（估计时间为现在）
	bc.memoryCacheMu.RLock()
	for cacheID, data := range bc.memoryCache {
		items = append(items, cacheItem{
			cacheID:   cacheID,
			createdAt: time.Now(),
			size:      int64(len(data)),
		})
	}
	bc.memoryCacheMu.RUnlock()

	// 按创建时间排序（最旧的在前）
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].createdAt.Before(items[i].createdAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// 删除最旧的缓存直到大小降到目标
	targetSize := bc.maxCacheSize * 80 / 100 // 目标为最大值的 80%
	currentSize := atomic.LoadInt64(&bc.totalCacheSize)

	for _, item := range items {
		if currentSize <= targetSize {
			break
		}

		_ = bc.InvalidateCache(item.cacheID)
		currentSize -= item.size
	}
}

// PurgeAll 清空所有缓存
func (bc *BodyCache) PurgeAll() error {
	// 清空内存缓存
	bc.memoryCacheMu.Lock()
	memCount := len(bc.memoryCache)
	bc.memoryCache = make(map[string][]byte)
	bc.memoryCacheMu.Unlock()

	// 清空磁盘缓存
	bc.diskCacheMu.Lock()
	diskEntries := make([]*DiskCacheEntry, 0)
	for _, entry := range bc.diskCacheMetadata {
		diskEntries = append(diskEntries, entry)
	}
	bc.diskCacheMetadata = make(map[string]*DiskCacheEntry)
	bc.diskCacheMu.Unlock()

	// 删除磁盘文件
	var lastErr error
	for _, entry := range diskEntries {
		if err := os.Remove(entry.FilePath); err != nil {
			lastErr = err
		}
	}

	atomic.StoreInt64(&bc.totalCacheSize, 0)

	if lastErr != nil {
		return fmt.Errorf("purged %d memory + %d disk caches with errors: %v",
			memCount, len(diskEntries), lastErr)
	}

	return nil
}

