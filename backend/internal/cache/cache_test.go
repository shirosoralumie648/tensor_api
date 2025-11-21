package cache

import (
	"context"
	"testing"
	"time"
)

func TestRedisClient(t *testing.T) {
	cfg := &CacheConfig{
		Addrs:       []string{"localhost:6379"},
		Password:    "",
		DB:          0,
		PoolSize:    10,
		MaxRetries:  3,
		TTL:         1 * time.Hour,
		ClusterMode: false,
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 测试 Set/Get
	err = client.Set(ctx, "test_key", "test_value", 1*time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := client.Get(ctx, "test_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}

	// 测试 Delete
	err = client.Delete(ctx, "test_key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 测试 Exists
	if client.Exists(ctx, "test_key") {
		t.Error("Key should not exist after deletion")
	}
}

func TestRedisClientExpiry(t *testing.T) {
	cfg := &CacheConfig{
		Addrs:       []string{"localhost:6379"},
		Password:    "",
		DB:          0,
		PoolSize:    10,
		MaxRetries:  3,
		TTL:         1 * time.Hour,
		ClusterMode: false,
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 设置短 TTL
	err = client.Set(ctx, "expiring_key", "value", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 立即获取应该成功
	_, err = client.Get(ctx, "expiring_key")
	if err != nil {
		t.Fatalf("Immediate Get failed: %v", err)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 获取应该失败
	_, err = client.Get(ctx, "expiring_key")
	if err == nil {
		t.Error("Get should fail after expiry")
	}
}

func TestRedisClientIncr(t *testing.T) {
	cfg := &CacheConfig{
		Addrs:       []string{"localhost:6379"},
		Password:    "",
		DB:          0,
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 初始化计数器
	err = client.Set(ctx, "counter", int64(0), 1*time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 递增
	val, err := client.Incr(ctx, "counter", 5)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}

	if val != 5 {
		t.Errorf("Expected 5, got %d", val)
	}

	// 再递增
	val, err = client.Incr(ctx, "counter", 3)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}

	if val != 8 {
		t.Errorf("Expected 8, got %d", val)
	}
}

func TestRedisClientMGet(t *testing.T) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 批量设置
	kvs := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	err = client.MSet(ctx, kvs)
	if err != nil {
		t.Fatalf("MSet failed: %v", err)
	}

	// 批量获取
	results := client.MGet(ctx, "key1", "key2", "key3")
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if results["key1"] != "value1" {
		t.Errorf("Expected 'value1', got %v", results["key1"])
	}
}

func TestCacheMiddleware(t *testing.T) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	middleware := NewCacheMiddleware(client, 1*time.Hour)

	// 设置特定路径的 TTL
	middleware.SetPattern("/api/data", 5*time.Minute)

	if !middleware.IsEnabled() {
		t.Error("Middleware should be enabled by default")
	}

	middleware.SetEnabled(false)
	if middleware.IsEnabled() {
		t.Error("Middleware should be disabled")
	}
}

func TestHotDataStrategy(t *testing.T) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	strategy := NewHotDataStrategy(client, 5)

	// 记录访问
	for i := 0; i < 3; i++ {
		strategy.RecordAccess("cold_key")
	}

	for i := 0; i < 6; i++ {
		strategy.RecordAccess("hot_key")
	}

	// 检查热数据
	if strategy.IsHotData("cold_key") {
		t.Error("cold_key should not be hot")
	}

	if !strategy.IsHotData("hot_key") {
		t.Error("hot_key should be hot")
	}
}

func TestCachePrewarmer(t *testing.T) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	prewarmer := NewCachePrewarmer(client)

	// 注册预热数据源
	prewarmer.Register("data1", func(ctx context.Context) (interface{}, error) {
		return "preheated_value_1", nil
	}, 1*time.Hour)

	prewarmer.Register("data2", func(ctx context.Context) (interface{}, error) {
		return "preheated_value_2", nil
	}, 1*time.Hour)

	ctx := context.Background()

	// 执行预热
	err = prewarmer.Preheat(ctx)
	if err != nil {
		t.Fatalf("Preheat failed: %v", err)
	}

	// 验证数据已预热
	val, err := client.Get(ctx, "data1")
	if err != nil {
		t.Fatalf("Get preheated data failed: %v", err)
	}

	if val != "preheated_value_1" {
		t.Errorf("Expected 'preheated_value_1', got %v", val)
	}
}

func TestCacheInvalidator(t *testing.T) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	invalidator := NewCacheInvalidator(client)

	ctx := context.Background()

	// 设置测试数据
	client.Set(ctx, "cache_key_1", "value1", 1*time.Hour)
	client.Set(ctx, "cache_key_2", "value2", 1*time.Hour)

	// 订阅失效事件
	callCount := 0
	invalidator.Subscribe("cache_key_*", func() {
		callCount++
	})

	// 失效缓存
	err = invalidator.Invalidate(ctx, "cache_key_*")
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected callback to be called once, got %d times", callCount)
	}
}

func TestRedisClientPipeline(t *testing.T) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("NewRedisClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 创建管道
	pipe := client.NewPipeline()
	pipe.Set("pipe_key_1", "value1", 1*time.Hour)
	pipe.Set("pipe_key_2", "value2", 1*time.Hour)
	pipe.Get("pipe_key_1")

	// 执行管道
	err = pipe.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 验证数据
	val, err := client.Get(ctx, "pipe_key_1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
}

func BenchmarkRedisSet(b *testing.B) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, _ := NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Set(ctx, "bench_key", "value", 1*time.Hour)
	}
}

func BenchmarkRedisGet(b *testing.B) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, _ := NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()
	client.Set(ctx, "bench_key", "value", 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get(ctx, "bench_key")
	}
}

func BenchmarkCacheHitRate(b *testing.B) {
	cfg := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	client, _ := NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	for i := 0; i < 100; i++ {
		client.Set(ctx, "key_"+string(rune(48+i)), "value", 1*time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		keyIdx := i % 100
		client.Get(ctx, "key_"+string(rune(48+keyIdx)))
	}

	stats := client.GetStats()
	hitRate := stats["hit_rate"].(float64)
	if hitRate < 80 {
		b.Logf("Hit rate too low: %.2f%%", hitRate)
	}
}

