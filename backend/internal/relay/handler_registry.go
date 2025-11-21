package relay

import (
	"fmt"
	"sync"
)

// HandlerRegistry 处理器注册表
type HandlerRegistry struct {
	// 处理器映射（type -> handler）
	handlers map[RequestType]RelayHandler
	mu       sync.RWMutex

	// 处理器工厂映射（用于动态创建）
	factories map[string]HandlerFactory
	factMu    sync.RWMutex
}

// HandlerFactory 处理器工厂接口
type HandlerFactory interface {
	// 创建处理器
	Create() (RelayHandler, error)
}

// NewHandlerRegistry 创建新的处理器注册表
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers:  make(map[RequestType]RelayHandler),
		factories: make(map[string]HandlerFactory),
	}
}

// RegisterHandler 注册处理器
func (hr *HandlerRegistry) RegisterHandler(handler RelayHandler) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	handlerType := handler.GetType()
	if _, exists := hr.handlers[handlerType]; exists {
		return fmt.Errorf("handler for type %s already registered", handlerType.String())
	}

	hr.handlers[handlerType] = handler
	return nil
}

// RegisterFactory 注册处理器工厂
func (hr *HandlerRegistry) RegisterFactory(name string, factory HandlerFactory) error {
	hr.factMu.Lock()
	defer hr.factMu.Unlock()

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	if _, exists := hr.factories[name]; exists {
		return fmt.Errorf("factory %s already registered", name)
	}

	hr.factories[name] = factory
	return nil
}

// GetHandler 获取处理器
func (hr *HandlerRegistry) GetHandler(handlerType RequestType) (RelayHandler, error) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	handler, ok := hr.handlers[handlerType]
	if !ok {
		return nil, fmt.Errorf("no handler registered for type %s", handlerType.String())
	}

	return handler, nil
}

// GetAllHandlers 获取所有处理器
func (hr *HandlerRegistry) GetAllHandlers() map[RequestType]RelayHandler {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	// 返回副本
	result := make(map[RequestType]RelayHandler)
	for k, v := range hr.handlers {
		result[k] = v
	}
	return result
}

// UnregisterHandler 注销处理器
func (hr *HandlerRegistry) UnregisterHandler(handlerType RequestType) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.handlers[handlerType]; !exists {
		return fmt.Errorf("no handler registered for type %s", handlerType.String())
	}

	delete(hr.handlers, handlerType)
	return nil
}

// CreateFromFactory 从工厂创建处理器
func (hr *HandlerRegistry) CreateFromFactory(name string) (RelayHandler, error) {
	hr.factMu.RLock()
	defer hr.factMu.RUnlock()

	factory, ok := hr.factories[name]
	if !ok {
		return nil, fmt.Errorf("factory %s not registered", name)
	}

	return factory.Create()
}

// GetRegisteredHandlerTypes 获取已注册的处理器类型
func (hr *HandlerRegistry) GetRegisteredHandlerTypes() []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	var types []string
	for handlerType := range hr.handlers {
		types = append(types, handlerType.String())
	}
	return types
}

// GetStatistics 获取所有处理器的统计信息
func (hr *HandlerRegistry) GetStatistics() map[string]map[string]interface{} {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	result := make(map[string]map[string]interface{})
	for handlerType, handler := range hr.handlers {
		result[handlerType.String()] = handler.GetStatistics()
	}
	return result
}

// ResetStatistics 重置所有处理器的统计信息
func (hr *HandlerRegistry) ResetStatistics() {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	for _, handler := range hr.handlers {
		handler.ResetStatistics()
	}
}

// HandlerManager 处理器管理器（高级功能）
type HandlerManager struct {
	registry *HandlerRegistry

	// 请求路由缓存
	routeCache map[RequestType]RelayHandler
	routeMu    sync.RWMutex

	// 统计信息
	totalHandledRequests int64
	mu                   sync.RWMutex
}

// NewHandlerManager 创建新的处理器管理器
func NewHandlerManager() *HandlerManager {
	return &HandlerManager{
		registry:   NewHandlerRegistry(),
		routeCache: make(map[RequestType]RelayHandler),
	}
}

// RegisterHandler 注册处理器
func (hm *HandlerManager) RegisterHandler(handler RelayHandler) error {
	return hm.registry.RegisterHandler(handler)
}

// RegisterFactory 注册处理器工厂
func (hm *HandlerManager) RegisterFactory(name string, factory HandlerFactory) error {
	return hm.registry.RegisterFactory(name, factory)
}

// GetHandler 获取处理器（使用缓存）
func (hm *HandlerManager) GetHandler(handlerType RequestType) (RelayHandler, error) {
	// 先检查缓存
	hm.routeMu.RLock()
	if handler, ok := hm.routeCache[handlerType]; ok {
		hm.routeMu.RUnlock()
		return handler, nil
	}
	hm.routeMu.RUnlock()

	// 从注册表获取
	handler, err := hm.registry.GetHandler(handlerType)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	hm.routeMu.Lock()
	hm.routeCache[handlerType] = handler
	hm.routeMu.Unlock()

	return handler, nil
}

// ClearRouteCache 清除路由缓存
func (hm *HandlerManager) ClearRouteCache() {
	hm.routeMu.Lock()
	defer hm.routeMu.Unlock()

	hm.routeCache = make(map[RequestType]RelayHandler)
}

// GetRegistry 获取处理器注册表
func (hm *HandlerManager) GetRegistry() *HandlerRegistry {
	return hm.registry
}

// GetStatistics 获取统计信息
func (hm *HandlerManager) GetStatistics() map[string]interface{} {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	return map[string]interface{}{
		"total_handled_requests": hm.totalHandledRequests,
		"handlers":               hm.registry.GetStatistics(),
		"registered_types":       hm.registry.GetRegisteredHandlerTypes(),
	}
}

// RecordRequest 记录处理的请求
func (hm *HandlerManager) RecordRequest() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.totalHandledRequests++
}

// ResetStatistics 重置统计信息
func (hm *HandlerManager) ResetStatistics() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.totalHandledRequests = 0
	hm.registry.ResetStatistics()
}

// DefaultHandlerFactory 默认处理器工厂
type DefaultHandlerFactory struct {
	handlerType RequestType
	client      *RequestClient
}

// NewDefaultHandlerFactory 创建默认处理器工厂
func NewDefaultHandlerFactory(handlerType RequestType, client *RequestClient) *DefaultHandlerFactory {
	return &DefaultHandlerFactory{
		handlerType: handlerType,
		client:      client,
	}
}

// Create 创建处理器
func (dhf *DefaultHandlerFactory) Create() (RelayHandler, error) {
	switch dhf.handlerType {
	case RequestTypeChat:
		return NewChatHandler(dhf.client), nil
	case RequestTypeEmbedding:
		return NewEmbeddingHandler(dhf.client), nil
	case RequestTypeImage:
		return NewImageHandler(dhf.client), nil
	case RequestTypeAudio:
		return NewAudioHandler(dhf.client), nil
	default:
		return nil, fmt.Errorf("unsupported handler type: %s", dhf.handlerType.String())
	}
}

// InitializeDefaultHandlers 初始化默认处理器
func InitializeDefaultHandlers(manager *HandlerManager, client *RequestClient) error {
	handlers := []RelayHandler{
		NewChatHandler(client),
		NewEmbeddingHandler(client),
		NewImageHandler(client),
		NewAudioHandler(client),
	}

	for _, handler := range handlers {
		if err := manager.RegisterHandler(handler); err != nil {
			return fmt.Errorf("failed to register handler %s: %v", handler.GetName(), err)
		}
	}

	return nil
}

