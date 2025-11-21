package relay

import (
	"fmt"
	"sync"
	"time"
)

// ChannelCacheLevel 缓存级别
type ChannelCacheLevel int

const (
	// 内存缓存
	ChannelCacheLevelMemory ChannelCacheLevel = iota
	// Redis 缓存
	ChannelCacheLevelRedis
	// 两级缓存
	ChannelCacheLevelHybrid
)

// ChannelCache 渠道缓存
type ChannelCache struct {
	// 缓存级别
	level ChannelCacheLevel

	// 内存缓存
	memoryCache   map[string]*Channel
	memoryCacheMu sync.RWMutex

	// 索引缓存（快速查询）
	// 按类型索引
	indexByType   map[string][]*Channel
	indexByTypeMu sync.RWMutex

	// 按模型索引
	indexByModel   map[string][]*Channel
	indexByModelMu sync.RWMutex

	// 按地区索引
	indexByRegion   map[string][]*Channel
	indexByRegionMu sync.RWMutex

	// 缓存过期时间
	cacheTTL time.Duration

	// 缓存更新时间
	lastUpdateTime time.Time
	updateMu       sync.RWMutex

	// 统计信息
	cacheHits   int64
	cacheMisses int64
	statsLock   sync.RWMutex
}

// NewChannelCache 创建新的渠道缓存
func NewChannelCache(level ChannelCacheLevel) *ChannelCache {
	return &ChannelCache{
		level:          level,
		memoryCache:    make(map[string]*Channel),
		indexByType:    make(map[string][]*Channel),
		indexByModel:   make(map[string][]*Channel),
		indexByRegion:  make(map[string][]*Channel),
		cacheTTL:       5 * time.Minute,
		lastUpdateTime: time.Now(),
	}
}

// AddChannel 添加渠道
func (cc *ChannelCache) AddChannel(ch *Channel) error {
	if ch == nil {
		return fmt.Errorf("channel cannot be nil")
	}
	if ch.ID == "" {
		return fmt.Errorf("channel ID is required")
	}

	cc.memoryCacheMu.Lock()
	defer cc.memoryCacheMu.Unlock()

	cc.memoryCache[ch.ID] = ch

	// 更新索引
	cc.updateIndices()

	return nil
}

// RemoveChannel 移除渠道
func (cc *ChannelCache) RemoveChannel(channelID string) error {
	cc.memoryCacheMu.Lock()
	defer cc.memoryCacheMu.Unlock()

	if _, ok := cc.memoryCache[channelID]; !ok {
		return fmt.Errorf("channel %s not found", channelID)
	}

	delete(cc.memoryCache, channelID)

	// 更新索引
	cc.updateIndices()

	return nil
}

// GetChannel 获取单个渠道
func (cc *ChannelCache) GetChannel(channelID string) (*Channel, error) {
	cc.memoryCacheMu.RLock()
	defer cc.memoryCacheMu.RUnlock()

	ch, ok := cc.memoryCache[channelID]
	if !ok {
		cc.recordMiss()
		return nil, fmt.Errorf("channel %s not found", channelID)
	}

	cc.recordHit()
	return ch, nil
}

// GetAllChannels 获取所有渠道
func (cc *ChannelCache) GetAllChannels() []*Channel {
	cc.memoryCacheMu.RLock()
	defer cc.memoryCacheMu.RUnlock()

	channels := make([]*Channel, 0, len(cc.memoryCache))
	for _, ch := range cc.memoryCache {
		channels = append(channels, ch)
	}

	return channels
}

// GetChannelsByType 按类型获取渠道
func (cc *ChannelCache) GetChannelsByType(channelType string) []*Channel {
	cc.indexByTypeMu.RLock()
	defer cc.indexByTypeMu.RUnlock()

	channels := cc.indexByType[channelType]
	if len(channels) > 0 {
		cc.recordHit()
		return channels
	}

	cc.recordMiss()
	return make([]*Channel, 0)
}

// GetChannelsByModel 按模型获取渠道
func (cc *ChannelCache) GetChannelsByModel(model string) []*Channel {
	cc.indexByModelMu.RLock()
	defer cc.indexByModelMu.RUnlock()

	channels := cc.indexByModel[model]
	if len(channels) > 0 {
		cc.recordHit()
		return channels
	}

	cc.recordMiss()
	return make([]*Channel, 0)
}

// GetChannelsByRegion 按地区获取渠道
func (cc *ChannelCache) GetChannelsByRegion(region string) []*Channel {
	cc.indexByRegionMu.RLock()
	defer cc.indexByRegionMu.RUnlock()

	channels := cc.indexByRegion[region]
	if len(channels) > 0 {
		cc.recordHit()
		return channels
	}

	cc.recordMiss()
	return make([]*Channel, 0)
}

// FilterChannels 按条件过滤渠道
func (cc *ChannelCache) FilterChannels(filter *ChannelFilter) []*Channel {
	channels := cc.GetAllChannels()
	filtered := make([]*Channel, 0)

	for _, ch := range channels {
		if ch.Matches(filter) {
			filtered = append(filtered, ch)
		}
	}

	return filtered
}

// UpdateChannel 更新渠道
func (cc *ChannelCache) UpdateChannel(ch *Channel) error {
	if ch == nil {
		return fmt.Errorf("channel cannot be nil")
	}

	cc.memoryCacheMu.Lock()
	defer cc.memoryCacheMu.Unlock()

	if _, ok := cc.memoryCache[ch.ID]; !ok {
		return fmt.Errorf("channel %s not found", ch.ID)
	}

	ch.UpdatedAt = time.Now()
	cc.memoryCache[ch.ID] = ch

	// 更新索引
	cc.updateIndices()

	return nil
}

// GetStatistics 获取统计信息
func (cc *ChannelCache) GetStatistics() map[string]interface{} {
	cc.statsLock.RLock()
	defer cc.statsLock.RUnlock()

	total := cc.cacheHits + cc.cacheMisses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(cc.cacheHits) / float64(total) * 100
	}

	cc.memoryCacheMu.RLock()
	channelCount := len(cc.memoryCache)
	cc.memoryCacheMu.RUnlock()

	return map[string]interface{}{
		"cache_level":   cc.level,
		"channel_count": channelCount,
		"cache_hits":    cc.cacheHits,
		"cache_misses":  cc.cacheMisses,
		"hit_rate":      hitRate,
		"last_update":   cc.lastUpdateTime,
		"index_types":   len(cc.indexByType),
		"index_models":  len(cc.indexByModel),
		"index_regions": len(cc.indexByRegion),
	}
}

// ClearCache 清空缓存
func (cc *ChannelCache) ClearCache() {
	cc.memoryCacheMu.Lock()
	defer cc.memoryCacheMu.Unlock()

	cc.memoryCache = make(map[string]*Channel)

	cc.indexByTypeMu.Lock()
	cc.indexByType = make(map[string][]*Channel)
	cc.indexByTypeMu.Unlock()

	cc.indexByModelMu.Lock()
	cc.indexByModel = make(map[string][]*Channel)
	cc.indexByModelMu.Unlock()

	cc.indexByRegionMu.Lock()
	cc.indexByRegion = make(map[string][]*Channel)
	cc.indexByRegionMu.Unlock()
}

// RefreshCache 刷新缓存
func (cc *ChannelCache) RefreshCache(channels []*Channel) error {
	cc.memoryCacheMu.Lock()
	defer cc.memoryCacheMu.Unlock()

	cc.memoryCache = make(map[string]*Channel)
	for _, ch := range channels {
		if ch != nil && ch.ID != "" {
			cc.memoryCache[ch.ID] = ch
		}
	}

	cc.updateMu.Lock()
	cc.lastUpdateTime = time.Now()
	cc.updateMu.Unlock()

	// 重建索引
	cc.updateIndices()

	return nil
}

// updateIndices 重建所有索引
func (cc *ChannelCache) updateIndices() {
	// 清空索引
	newIndexByType := make(map[string][]*Channel)
	newIndexByModel := make(map[string][]*Channel)
	newIndexByRegion := make(map[string][]*Channel)

	// 重建索引
	for _, ch := range cc.memoryCache {
		// 按类型索引
		if ch.Type != "" {
			newIndexByType[ch.Type] = append(newIndexByType[ch.Type], ch)
		}

		// 按模型索引
		if ch.Ability != nil {
			for _, model := range ch.Ability.SupportedModels {
				newIndexByModel[model] = append(newIndexByModel[model], ch)
			}
		}

		// 按地区索引
		if ch.Region != "" {
			newIndexByRegion[ch.Region] = append(newIndexByRegion[ch.Region], ch)
		}
	}

	// 更新索引
	cc.indexByTypeMu.Lock()
	cc.indexByType = newIndexByType
	cc.indexByTypeMu.Unlock()

	cc.indexByModelMu.Lock()
	cc.indexByModel = newIndexByModel
	cc.indexByModelMu.Unlock()

	cc.indexByRegionMu.Lock()
	cc.indexByRegion = newIndexByRegion
	cc.indexByRegionMu.Unlock()
}

// recordHit 记录缓存命中
func (cc *ChannelCache) recordHit() {
	cc.statsLock.Lock()
	defer cc.statsLock.Unlock()
	cc.cacheHits++
}

// recordMiss 记录缓存未命中
func (cc *ChannelCache) recordMiss() {
	cc.statsLock.Lock()
	defer cc.statsLock.Unlock()
	cc.cacheMisses++
}

// ChannelCacheManager 渠道缓存管理器
type ChannelCacheManager struct {
	cache *ChannelCache

	// 定时刷新
	refreshInterval time.Duration
	refreshTicker   *time.Ticker

	// 数据源（从数据库等获取渠道）
	dataSource func() ([]*Channel, error)

	// 控制信号
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewChannelCacheManager 创建渠道缓存管理器
func NewChannelCacheManager(dataSource func() ([]*Channel, error)) *ChannelCacheManager {
	return &ChannelCacheManager{
		cache:           NewChannelCache(ChannelCacheLevelHybrid),
		refreshInterval: 5 * time.Minute,
		dataSource:      dataSource,
		stopCh:          make(chan struct{}),
	}
}

// Start 启动缓存管理器
func (ccm *ChannelCacheManager) Start() error {
	// 初始化缓存
	if err := ccm.refresh(); err != nil {
		return err
	}

	// 启动定时刷新
	ccm.wg.Add(1)
	go ccm.refreshRoutine()

	return nil
}

// Stop 停止缓存管理器
func (ccm *ChannelCacheManager) Stop() {
	close(ccm.stopCh)
	ccm.wg.Wait()

	if ccm.refreshTicker != nil {
		ccm.refreshTicker.Stop()
	}
}

// GetCache 获取缓存
func (ccm *ChannelCacheManager) GetCache() *ChannelCache {
	return ccm.cache
}

// refresh 刷新缓存
func (ccm *ChannelCacheManager) refresh() error {
	if ccm.dataSource == nil {
		return fmt.Errorf("data source not configured")
	}

	channels, err := ccm.dataSource()
	if err != nil {
		return err
	}

	return ccm.cache.RefreshCache(channels)
}

// refreshRoutine 定时刷新
func (ccm *ChannelCacheManager) refreshRoutine() {
	defer ccm.wg.Done()

	ccm.refreshTicker = time.NewTicker(ccm.refreshInterval)
	defer ccm.refreshTicker.Stop()

	for {
		select {
		case <-ccm.stopCh:
			return
		case <-ccm.refreshTicker.C:
			_ = ccm.refresh()
		}
	}
}

// SetRefreshInterval 设置刷新间隔
func (ccm *ChannelCacheManager) SetRefreshInterval(interval time.Duration) {
	ccm.refreshInterval = interval
}
