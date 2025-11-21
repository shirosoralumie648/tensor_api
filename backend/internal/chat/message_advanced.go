package chat

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MessageBranch 消息分支
type MessageBranch struct {
	// 分支 ID
	ID string `json:"id"`

	// 会话 ID
	SessionID string `json:"session_id"`

	// 分支名称
	Name string `json:"name"`

	// 父消息 ID
	ParentMessageID string `json:"parent_message_id"`

	// 分支消息列表
	MessageIDs []string `json:"message_ids"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// MessageEdit 消息编辑记录
type MessageEdit struct {
	// 编辑 ID
	ID string `json:"id"`

	// 消息 ID
	MessageID string `json:"message_id"`

	// 原始内容
	OriginalContent string `json:"original_content"`

	// 新内容
	NewContent string `json:"new_content"`

	// 编辑者
	Editor string `json:"editor"`

	// 编辑时间
	EditedAt time.Time `json:"edited_at"`

	// 编辑原因
	Reason string `json:"reason"`
}

// MessageReference 消息引用
type MessageReference struct {
	// 引用 ID
	ID string `json:"id"`

	// 引用消息 ID
	ReferencedMessageID string `json:"referenced_message_id"`

	// 引用所在消息 ID
	SourceMessageID string `json:"source_message_id"`

	// 引用上下文
	Context string `json:"context"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// ExportFormat 导出格式
type ExportFormat string

const (
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatHTML     ExportFormat = "html"
	ExportFormatCSV      ExportFormat = "csv"
)

// MessageExport 消息导出
type MessageExport struct {
	// 导出 ID
	ID string `json:"id"`

	// 会话 ID
	SessionID string `json:"session_id"`

	// 导出格式
	Format ExportFormat `json:"format"`

	// 导出内容
	Content string `json:"content"`

	// 导出消息数
	MessageCount int `json:"message_count"`

	// 导出时间
	ExportedAt time.Time `json:"exported_at"`
}

// BranchManager 分支管理器
type BranchManager struct {
	// 分支存储
	branches map[string]*MessageBranch
	branchesMu sync.RWMutex

	// 消息分支索引
	messageBranches map[string][]string // messageID -> branchIDs
	messageBranchesMu sync.RWMutex

	// 统计信息
	totalBranches int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewBranchManager 创建分支管理器
func NewBranchManager() *BranchManager {
	return &BranchManager{
		branches:        make(map[string]*MessageBranch),
		messageBranches: make(map[string][]string),
		logFunc:         defaultLogFunc,
	}
}

// CreateBranch 创建分支
func (bm *BranchManager) CreateBranch(sessionID, name, parentMessageID string) (*MessageBranch, error) {
	bm.branchesMu.Lock()
	defer bm.branchesMu.Unlock()

	branchID := fmt.Sprintf("branch-%s-%d", sessionID, time.Now().UnixNano())

	branch := &MessageBranch{
		ID:              branchID,
		SessionID:       sessionID,
		Name:            name,
		ParentMessageID: parentMessageID,
		MessageIDs:      make([]string, 0),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	bm.branches[branchID] = branch

	// 更新消息分支索引
	bm.messageBranchesMu.Lock()
	bm.messageBranches[parentMessageID] = append(bm.messageBranches[parentMessageID], branchID)
	bm.messageBranchesMu.Unlock()

	atomic.AddInt64(&bm.totalBranches, 1)

	bm.logFunc("info", fmt.Sprintf("Created branch %s for message %s", branchID, parentMessageID))

	return branch, nil
}

// GetBranch 获取分支
func (bm *BranchManager) GetBranch(branchID string) (*MessageBranch, error) {
	bm.branchesMu.RLock()
	defer bm.branchesMu.RUnlock()

	branch, exists := bm.branches[branchID]
	if !exists {
		return nil, fmt.Errorf("branch %s not found", branchID)
	}

	return branch, nil
}

// AddMessageToBranch 添加消息到分支
func (bm *BranchManager) AddMessageToBranch(branchID, messageID string) error {
	bm.branchesMu.Lock()
	defer bm.branchesMu.Unlock()

	branch, exists := bm.branches[branchID]
	if !exists {
		return fmt.Errorf("branch %s not found", branchID)
	}

	branch.MessageIDs = append(branch.MessageIDs, messageID)
	branch.UpdatedAt = time.Now()

	return nil
}

// GetBranchesForMessage 获取消息的所有分支
func (bm *BranchManager) GetBranchesForMessage(messageID string) []*MessageBranch {
	bm.messageBranchesMu.RLock()
	branchIDs := bm.messageBranches[messageID]
	bm.messageBranchesMu.RUnlock()

	var branches []*MessageBranch

	bm.branchesMu.RLock()
	defer bm.branchesMu.RUnlock()

	for _, id := range branchIDs {
		if branch, exists := bm.branches[id]; exists {
			branches = append(branches, branch)
		}
	}

	return branches
}

// GetStatistics 获取统计信息
func (bm *BranchManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_branches": atomic.LoadInt64(&bm.totalBranches),
	}
}

// EditManager 编辑管理器
type EditManager struct {
	// 编辑记录
	edits map[string][]*MessageEdit
	editsMu sync.RWMutex

	// 统计信息
	totalEdits int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewEditManager 创建编辑管理器
func NewEditManager() *EditManager {
	return &EditManager{
		edits:   make(map[string][]*MessageEdit),
		logFunc: defaultLogFunc,
	}
}

// RecordEdit 记录编辑
func (em *EditManager) RecordEdit(messageID, originalContent, newContent, editor, reason string) (*MessageEdit, error) {
	em.editsMu.Lock()
	defer em.editsMu.Unlock()

	editID := fmt.Sprintf("edit-%s-%d", messageID, time.Now().UnixNano())

	edit := &MessageEdit{
		ID:              editID,
		MessageID:       messageID,
		OriginalContent: originalContent,
		NewContent:      newContent,
		Editor:          editor,
		EditedAt:        time.Now(),
		Reason:          reason,
	}

	em.edits[messageID] = append(em.edits[messageID], edit)

	atomic.AddInt64(&em.totalEdits, 1)

	em.logFunc("info", fmt.Sprintf("Recorded edit for message %s", messageID))

	return edit, nil
}

// GetEditHistory 获取编辑历史
func (em *EditManager) GetEditHistory(messageID string) []*MessageEdit {
	em.editsMu.RLock()
	defer em.editsMu.RUnlock()

	edits, exists := em.edits[messageID]
	if !exists {
		return make([]*MessageEdit, 0)
	}

	result := make([]*MessageEdit, len(edits))
	copy(result, edits)

	return result
}

// GetLatestEdit 获取最后一次编辑
func (em *EditManager) GetLatestEdit(messageID string) (*MessageEdit, error) {
	em.editsMu.RLock()
	defer em.editsMu.RUnlock()

	edits, exists := em.edits[messageID]
	if !exists || len(edits) == 0 {
		return nil, fmt.Errorf("no edits found for message %s", messageID)
	}

	return edits[len(edits)-1], nil
}

// GetStatistics 获取统计信息
func (em *EditManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_edits": atomic.LoadInt64(&em.totalEdits),
	}
}

// ReferenceManager 引用管理器
type ReferenceManager struct {
	// 引用记录
	references map[string][]*MessageReference
	referencesMu sync.RWMutex

	// 反向引用索引
	reverseIndex map[string][]string // referencedMessageID -> referenceIDs
	reverseIndexMu sync.RWMutex

	// 统计信息
	totalReferences int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewReferenceManager 创建引用管理器
func NewReferenceManager() *ReferenceManager {
	return &ReferenceManager{
		references:   make(map[string][]*MessageReference),
		reverseIndex: make(map[string][]string),
		logFunc:      defaultLogFunc,
	}
}

// CreateReference 创建引用
func (rm *ReferenceManager) CreateReference(referencedMessageID, sourceMessageID, context string) (*MessageReference, error) {
	rm.referencesMu.Lock()
	defer rm.referencesMu.Unlock()

	refID := fmt.Sprintf("ref-%s-%d", sourceMessageID, time.Now().UnixNano())

	reference := &MessageReference{
		ID:                  refID,
		ReferencedMessageID: referencedMessageID,
		SourceMessageID:     sourceMessageID,
		Context:             context,
		CreatedAt:           time.Now(),
	}

	rm.references[sourceMessageID] = append(rm.references[sourceMessageID], reference)

	// 更新反向索引
	rm.reverseIndexMu.Lock()
	rm.reverseIndex[referencedMessageID] = append(rm.reverseIndex[referencedMessageID], refID)
	rm.reverseIndexMu.Unlock()

	atomic.AddInt64(&rm.totalReferences, 1)

	rm.logFunc("info", fmt.Sprintf("Created reference from %s to %s", sourceMessageID, referencedMessageID))

	return reference, nil
}

// GetReferences 获取消息的引用
func (rm *ReferenceManager) GetReferences(sourceMessageID string) []*MessageReference {
	rm.referencesMu.RLock()
	defer rm.referencesMu.RUnlock()

	references, exists := rm.references[sourceMessageID]
	if !exists {
		return make([]*MessageReference, 0)
	}

	result := make([]*MessageReference, len(references))
	copy(result, references)

	return result
}

// GetReferencedBy 获取引用某消息的所有消息
func (rm *ReferenceManager) GetReferencedBy(referencedMessageID string) []*MessageReference {
	rm.reverseIndexMu.RLock()
	refIDs := rm.reverseIndex[referencedMessageID]
	rm.reverseIndexMu.RUnlock()

	var references []*MessageReference

	rm.referencesMu.RLock()
	defer rm.referencesMu.RUnlock()

	for _, refID := range refIDs {
		for _, ref := range rm.references {
			for _, r := range ref {
				if r.ID == refID {
					references = append(references, r)
				}
			}
		}
	}

	return references
}

// GetStatistics 获取统计信息
func (rm *ReferenceManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_references": atomic.LoadInt64(&rm.totalReferences),
	}
}

// ExportManager 导出管理器
type ExportManager struct {
	// 导出记录
	exports map[string]*MessageExport
	exportsMu sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewExportManager 创建导出管理器
func NewExportManager() *ExportManager {
	return &ExportManager{
		exports: make(map[string]*MessageExport),
		logFunc: defaultLogFunc,
	}
}

// ExportMessages 导出消息
func (em *ExportManager) ExportMessages(sessionID string, messages []*Message, format ExportFormat) (*MessageExport, error) {
	content, err := em.formatMessages(messages, format)
	if err != nil {
		return nil, err
	}

	em.exportsMu.Lock()
	defer em.exportsMu.Unlock()

	exportID := fmt.Sprintf("export-%s-%d", sessionID, time.Now().UnixNano())

	export := &MessageExport{
		ID:           exportID,
		SessionID:    sessionID,
		Format:       format,
		Content:      content,
		MessageCount: len(messages),
		ExportedAt:   time.Now(),
	}

	em.exports[exportID] = export

	em.logFunc("info", fmt.Sprintf("Exported %d messages to %s", len(messages), format))

	return export, nil
}

// formatMessages 格式化消息
func (em *ExportManager) formatMessages(messages []*Message, format ExportFormat) (string, error) {
	switch format {
	case ExportFormatJSON:
		data, err := json.MarshalIndent(messages, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil

	case ExportFormatMarkdown:
		var result string
		for _, msg := range messages {
			result += fmt.Sprintf("### %s\n\n%s\n\n", msg.Role, msg.Content)
		}
		return result, nil

	case ExportFormatHTML:
		result := "<html><body>"
		for _, msg := range messages {
			result += fmt.Sprintf("<div class='message %s'><strong>%s:</strong> %s</div>", msg.Role, msg.Role, msg.Content)
		}
		result += "</body></html>"
		return result, nil

	case ExportFormatCSV:
		result := "Role,Content,Tokens,Timestamp\n"
		for _, msg := range messages {
			result += fmt.Sprintf("%s,\"%s\",%d,%s\n", msg.Role, msg.Content, msg.Tokens, msg.Timestamp.Format(time.RFC3339))
		}
		return result, nil

	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// GetExport 获取导出
func (em *ExportManager) GetExport(exportID string) (*MessageExport, error) {
	em.exportsMu.RLock()
	defer em.exportsMu.RUnlock()

	export, exists := em.exports[exportID]
	if !exists {
		return nil, fmt.Errorf("export %s not found", exportID)
	}

	return export, nil
}

// AdvancedMessageManager 高级消息管理器
type AdvancedMessageManager struct {
	// 分支管理
	branchManager *BranchManager

	// 编辑管理
	editManager *EditManager

	// 引用管理
	referenceManager *ReferenceManager

	// 导出管理
	exportManager *ExportManager

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewAdvancedMessageManager 创建高级消息管理器
func NewAdvancedMessageManager() *AdvancedMessageManager {
	return &AdvancedMessageManager{
		branchManager:    NewBranchManager(),
		editManager:      NewEditManager(),
		referenceManager: NewReferenceManager(),
		exportManager:    NewExportManager(),
		logFunc:          defaultLogFunc,
	}
}

// GetStatistics 获取统计信息
func (amm *AdvancedMessageManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"branches":   amm.branchManager.GetStatistics(),
		"edits":      amm.editManager.GetStatistics(),
		"references": amm.referenceManager.GetStatistics(),
	}
}

