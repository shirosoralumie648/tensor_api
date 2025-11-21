package cache

import (
	"math"
	"sync"
)

// BloomFilter 布隆过滤器实现
type BloomFilter struct {
	// 位数组
	bits []byte

	// 哈希函数数量
	numHashFuncs int

	// 大小（位数）
	size uint32

	// 互斥锁（并发安全）
	mu sync.RWMutex
}

// NewBloomFilter 创建新的布隆过滤器
// capacity: 期望容量
// falsePositiveRate: 可接受的误判率（0.01 表示 1%）
func NewBloomFilter(capacity int, falsePositiveRate float64) BloomFilter {
	// 计算位数组大小
	size := calculateFilterSize(capacity, falsePositiveRate)

	// 计算哈希函数数量
	numHashFuncs := calculateNumHashFuncs(size, capacity)

	return BloomFilter{
		bits:         make([]byte, (size+7)/8), // 转换为字节数组
		size:         size,
		numHashFuncs: numHashFuncs,
	}
}

// Add 向布隆过滤器添加元素
func (bf *BloomFilter) Add(data []byte) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	for i := 0; i < bf.numHashFuncs; i++ {
		hash := bf.hash(data, uint32(i))
		index := hash % bf.size
		byteIndex := index / 8
		bitIndex := index % 8
		bf.bits[byteIndex] |= 1 << bitIndex
	}
}

// Contains 检查布隆过滤器中是否包含元素
func (bf *BloomFilter) Contains(data []byte) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	for i := 0; i < bf.numHashFuncs; i++ {
		hash := bf.hash(data, uint32(i))
		index := hash % bf.size
		byteIndex := index / 8
		bitIndex := index % 8

		// 如果任何一位是 0，则元素肯定不存在
		if bf.bits[byteIndex]&(1<<bitIndex) == 0 {
			return false
		}
	}

	return true
}

// Reset 重置布隆过滤器
func (bf *BloomFilter) Reset() {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	for i := range bf.bits {
		bf.bits[i] = 0
	}
}

// GetStats 获取布隆过滤器统计信息
func (bf *BloomFilter) GetStats() map[string]interface{} {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	// 计算已设置的位数
	setBits := 0
	for _, b := range bf.bits {
		for i := 0; i < 8; i++ {
			if b&(1<<uint(i)) != 0 {
				setBits++
			}
		}
	}

	return map[string]interface{}{
		"size":           bf.size,
		"num_hash_funcs": bf.numHashFuncs,
		"total_bits":     len(bf.bits) * 8,
		"set_bits":       setBits,
		"utilization":    float64(setBits) / float64(len(bf.bits)*8),
	}
}

// 私有方法

// hash 计算哈希值
func (bf *BloomFilter) hash(data []byte, seed uint32) uint32 {
	// 使用 MurmurHash2 风格的哈希算法
	const (
		c1 uint32 = 0xcc9e2d51
		c2 uint32 = 0x1b873593
	)

	hash := seed ^ uint32(len(data))

	// 处理 4 字节的块
	chunks := len(data) / 4
	for i := 0; i < chunks; i++ {
		k := uint32(data[i*4]) |
			(uint32(data[i*4+1]) << 8) |
			(uint32(data[i*4+2]) << 16) |
			(uint32(data[i*4+3]) << 24)

		k *= c1
		k = rotl32(k, 15)
		k *= c2

		hash ^= k
		hash = rotl32(hash, 13)
		hash = (hash * 5) + 0xe6546b64
	}

	// 处理剩余字节
	tail := data[chunks*4:]
	switch len(tail) {
	case 3:
		hash ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		hash ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		hash ^= uint32(tail[0])
		hash *= c1
		hash = rotl32(hash, 15)
		hash *= c2
	}

	// 最终混合
	hash ^= uint32(len(data))
	return fmix32(hash)
}

// rotl32 32位循环左移
func rotl32(x, r uint32) uint32 {
	return (x << r) | (x >> (32 - r))
}

// fmix32 最终混合函数
func fmix32(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

// calculateFilterSize 计算布隆过滤器的大小（位数）
func calculateFilterSize(capacity int, falsePositiveRate float64) uint32 {
	// 公式: m = -1 / ln(2)^2 * n * ln(p)
	// m: 位数
	// n: 期望的元素数量
	// p: 可接受的误判率
	ln2Squared := math.Ln2 * math.Ln2
	size := -1.0 / ln2Squared * float64(capacity) * math.Log(falsePositiveRate)
	return uint32(math.Ceil(size))
}

// calculateNumHashFuncs 计算哈希函数的数量
func calculateNumHashFuncs(size uint32, capacity int) int {
	// 公式: k = m / n * ln(2)
	// k: 哈希函数数量
	// m: 位数
	// n: 期望的元素数量
	numHashFuncs := float64(size) / float64(capacity) * math.Ln2
	return int(math.Ceil(numHashFuncs))
}

// GetExpectedFalsePositiveRate 获取期望的误判率
func (bf *BloomFilter) GetExpectedFalsePositiveRate(elementsAdded int) float64 {
	if elementsAdded == 0 {
		return 0
	}
	// 公式: (1 - e^(-k*n/m))^k
	exponent := float64(-bf.numHashFuncs*elementsAdded) / float64(bf.size)
	return math.Pow(1-math.Exp(exponent), float64(bf.numHashFuncs))
}
