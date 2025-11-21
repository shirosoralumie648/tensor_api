package relay

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBodyCacheMemory(t *testing.T) {
	bc := NewBodyCache(t.TempDir())
	defer bc.Stop()

	bc.SetMode(BodyCacheModeMemory)

	// 测试缓存数据
	data := []byte("test data")
	cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to cache: %v", err)
	}

	// 获取缓存数据
	retrieved, err := bc.GetCachedBody(cacheID)
	if err != nil {
		t.Fatalf("Failed to get cached body: %v", err)
	}

	if !bytes.Equal(retrieved, data) {
		t.Errorf("Retrieved data mismatch: got %v, want %v", retrieved, data)
	}
}

func TestBodyCacheDisk(t *testing.T) {
	tmpDir := t.TempDir()
	bc := NewBodyCache(tmpDir)
	defer bc.Stop()

	bc.SetMode(BodyCacheModeDisk)

	// 测试缓存数据
	data := []byte("disk cache test data")
	cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to cache: %v", err)
	}

	// 验证文件存在
	filePath := filepath.Join(tmpDir, cacheID+".dat")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Cache file not created: %s", filePath)
	}

	// 获取缓存数据
	retrieved, err := bc.GetCachedBody(cacheID)
	if err != nil {
		t.Fatalf("Failed to get cached body: %v", err)
	}

	if !bytes.Equal(retrieved, data) {
		t.Errorf("Retrieved data mismatch: got %v, want %v", retrieved, data)
	}
}

func TestBodyCacheHybrid(t *testing.T) {
	tmpDir := t.TempDir()
	bc := NewBodyCache(tmpDir)
	defer bc.Stop()

	bc.SetMode(BodyCacheModeHybrid)
	bc.SetMemoryThreshold(100) // 100 字节阈值

	// 测试小数据（应该使用内存）
	smallData := []byte("small")
	smallCacheID, err := bc.CacheRequestBody(bytes.NewReader(smallData))
	if err != nil {
		t.Fatalf("Failed to cache small data: %v", err)
	}

	// 测试大数据（应该使用磁盘）
	largeData := make([]byte, 200)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	largeCacheID, err := bc.CacheRequestBody(bytes.NewReader(largeData))
	if err != nil {
		t.Fatalf("Failed to cache large data: %v", err)
	}

	// 验证两个都可以被检索
	smallRetrieved, err := bc.GetCachedBody(smallCacheID)
	if err != nil {
		t.Fatalf("Failed to get small cached body: %v", err)
	}
	if !bytes.Equal(smallRetrieved, smallData) {
		t.Errorf("Small data mismatch")
	}

	largeRetrieved, err := bc.GetCachedBody(largeCacheID)
	if err != nil {
		t.Fatalf("Failed to get large cached body: %v", err)
	}
	if !bytes.Equal(largeRetrieved, largeData) {
		t.Errorf("Large data mismatch")
	}
}

func TestBodyCacheInvalidation(t *testing.T) {
	bc := NewBodyCache(t.TempDir())
	defer bc.Stop()

	data := []byte("test data")
	cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to cache: %v", err)
	}

	// 验证缓存存在
	_, err = bc.GetCachedBody(cacheID)
	if err != nil {
		t.Fatalf("Failed to get cached body: %v", err)
	}

	// 失效缓存
	if err := bc.InvalidateCache(cacheID); err != nil {
		t.Fatalf("Failed to invalidate cache: %v", err)
	}

	// 验证缓存不存在
	_, err = bc.GetCachedBody(cacheID)
	if err == nil {
		t.Errorf("Expected error after invalidation, but got none")
	}
}

func TestBodyCacheReader(t *testing.T) {
	bc := NewBodyCache(t.TempDir())
	defer bc.Stop()

	data := []byte("test reader data")
	cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to cache: %v", err)
	}

	// 获取读取器
	reader, err := bc.GetCachedBodyReader(cacheID)
	if err != nil {
		t.Fatalf("Failed to get reader: %v", err)
	}

	// 读取数据
	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if !bytes.Equal(retrieved, data) {
		t.Errorf("Data mismatch: got %v, want %v", retrieved, data)
	}
}

func TestBodyCacheStatistics(t *testing.T) {
	bc := NewBodyCache(t.TempDir())
	defer bc.Stop()

	// 缓存多个数据
	for i := 0; i < 5; i++ {
		data := []byte(fmt.Sprintf("data %d", i))
		_, err := bc.CacheRequestBody(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to cache: %v", err)
		}
	}

	// 获取统计
	stats := bc.GetStatistics()

	if memCount, ok := stats["memory_count"].(int); ok {
		if memCount != 5 {
			t.Errorf("Expected 5 memory items, got %d", memCount)
		}
	}

	if totalSize, ok := stats["total_size"].(int64); ok {
		if totalSize <= 0 {
			t.Errorf("Expected positive total size, got %d", totalSize)
		}
	}
}

func TestBodyCachePurgeAll(t *testing.T) {
	tmpDir := t.TempDir()
	bc := NewBodyCache(tmpDir)
	defer bc.Stop()

	// 缓存数据
	for i := 0; i < 3; i++ {
		data := []byte(fmt.Sprintf("data %d", i))
		_, err := bc.CacheRequestBody(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to cache: %v", err)
		}
	}

	stats := bc.GetStatistics()
	if memCount, ok := stats["memory_count"].(int); ok && memCount == 0 {
		t.Fatalf("Expected cached items before purge")
	}

	// 清空所有缓存
	if err := bc.PurgeAll(); err != nil {
		t.Fatalf("Failed to purge: %v", err)
	}

	stats = bc.GetStatistics()
	if memCount, ok := stats["memory_count"].(int); ok && memCount != 0 {
		t.Errorf("Expected 0 memory items after purge, got %d", memCount)
	}
	if diskCount, ok := stats["disk_count"].(int); ok && diskCount != 0 {
		t.Errorf("Expected 0 disk items after purge, got %d", diskCount)
	}
}

// 性能测试
func BenchmarkBodyCacheMemory(b *testing.B) {
	bc := NewBodyCache(b.TempDir())
	defer bc.Stop()

	bc.SetMode(BodyCacheModeMemory)
	data := make([]byte, 10*1024) // 10KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
		if err != nil {
			b.Fatalf("Cache failed: %v", err)
		}
		_, err = bc.GetCachedBody(cacheID)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func BenchmarkBodyCacheDisk(b *testing.B) {
	bc := NewBodyCache(b.TempDir())
	defer bc.Stop()

	bc.SetMode(BodyCacheModeDisk)
	data := make([]byte, 10*1024) // 10KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
		if err != nil {
			b.Fatalf("Cache failed: %v", err)
		}
		_, err = bc.GetCachedBody(cacheID)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func BenchmarkBodyCacheLargeData(b *testing.B) {
	bc := NewBodyCache(b.TempDir())
	defer bc.Stop()

	bc.SetMode(BodyCacheModeDisk)
	// 100MB 数据
	data := make([]byte, 100*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
		if err != nil {
			b.Fatalf("Cache failed: %v", err)
		}
		_, err = bc.GetCachedBody(cacheID)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
		bc.InvalidateCache(cacheID)
	}
}

func TestBodyCacheExpiration(t *testing.T) {
	bc := NewBodyCache(t.TempDir())
	defer bc.Stop()

	// 设置短的过期时间用于测试
	bc.maxCacheDuration = 1 * time.Second

	// 启动清理线程
	bc.Start()

	// 缓存数据
	data := []byte("test data")
	cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to cache: %v", err)
	}

	// 验证缓存存在
	_, err = bc.GetCachedBody(cacheID)
	if err != nil {
		t.Fatalf("Failed to get cached body: %v", err)
	}

	// 等待过期
	time.Sleep(2 * time.Second)

	// 触发清理
	bc.cleanup()

	// 验证缓存被清除
	_, err = bc.GetCachedBody(cacheID)
	if err == nil {
		t.Errorf("Expected cache to be expired and removed")
	}
}

func TestBodyCacheConcurrency(t *testing.T) {
	bc := NewBodyCache(t.TempDir())
	defer bc.Stop()

	bc.SetMode(BodyCacheModeHybrid)

	// 并发缓存操作
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			data := []byte(fmt.Sprintf("concurrent data %d", id))
			cacheID, err := bc.CacheRequestBody(bytes.NewReader(data))
			if err != nil {
				done <- err
				return
			}

			retrieved, err := bc.GetCachedBody(cacheID)
			if err != nil {
				done <- err
				return
			}

			if !bytes.Equal(retrieved, data) {
				done <- fmt.Errorf("data mismatch for id %d", id)
				return
			}

			done <- nil
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}
}

