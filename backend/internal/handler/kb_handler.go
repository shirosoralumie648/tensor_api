package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shirosoralumie648/Oblivious/backend/internal/service"
	"github.com/shirosoralumie648/Oblivious/backend/internal/utils"
)

// KBHandler 处理知识库相关的 HTTP 请求
type KBHandler struct {
	ragService *service.RAGService
}

// NewKBHandler 创建新的 KBHandler
func NewKBHandler(ragService *service.RAGService) *KBHandler {
	return &KBHandler{
		ragService: ragService,
	}
}

// CreateKnowledgeBase 创建知识库
// POST /api/v1/knowledge-bases
func (h *KBHandler) CreateKnowledgeBase(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req service.CreateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	kb, err := h.ragService.CreateKnowledgeBase(c.Request.Context(), userID, &req)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, kb, "知识库创建成功")
}

// GetKnowledgeBase 获取知识库详情
// GET /api/v1/knowledge-bases/:id
func (h *KBHandler) GetKnowledgeBase(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid knowledge base ID")
		return
	}

	kb, err := h.ragService.GetKnowledgeBase(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	if kb == nil {
		utils.NotFound(c, "知识库不存在")
		return
	}

	utils.Success(c, kb, "")
}

// ListKnowledgeBases 获取用户的知识库列表
// GET /api/v1/knowledge-bases
func (h *KBHandler) ListKnowledgeBases(c *gin.Context) {
	userID := c.GetInt("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	kbs, total, err := h.ragService.ListKnowledgeBases(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"knowledge_bases": kbs,
		"total":           total,
		"page":            page,
		"page_size":       pageSize,
	}, "")
}

// DeleteKnowledgeBase 删除知识库
// DELETE /api/v1/knowledge-bases/:id
func (h *KBHandler) DeleteKnowledgeBase(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid knowledge base ID")
		return
	}

	if err := h.ragService.DeleteKnowledgeBase(c.Request.Context(), id, userID); err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, nil, "知识库删除成功")
}

// UploadDocument 上传文档
// POST /api/v1/knowledge-bases/:id/documents
func (h *KBHandler) UploadDocument(c *gin.Context) {
	userID := c.GetInt("user_id")
	kbID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid knowledge base ID")
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		FileContent string `json:"file_content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	doc, err := h.ragService.UploadDocument(c.Request.Context(), userID, kbID, req.Title, req.FileContent)
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, doc, "文档上传成功，正在处理中...")
}

// GetDocumentList 获取文档列表
// GET /api/v1/knowledge-bases/:id/documents
func (h *KBHandler) GetDocumentList(c *gin.Context) {
	userID := c.GetInt("user_id")
	kbID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid knowledge base ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	docs, total, err := h.ragService.GetDocumentList(c.Request.Context(), userID, kbID, page, pageSize)
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"documents":  docs,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	}, "")
}

// DeleteDocument 删除文档
// DELETE /api/v1/knowledge-bases/:id/documents/:doc_id
func (h *KBHandler) DeleteDocument(c *gin.Context) {
	userID := c.GetInt("user_id")
	kbID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid knowledge base ID")
		return
	}

	docIDStr := c.Param("doc_id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		utils.BadRequest(c, "Invalid document ID")
		return
	}

	if err := h.ragService.DeleteDocument(c.Request.Context(), userID, kbID, docID); err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, nil, "文档删除成功")
}

// SearchDocuments 搜索文档
// POST /api/v1/knowledge-bases/:id/search
func (h *KBHandler) SearchDocuments(c *gin.Context) {
	userID := c.GetInt("user_id")
	kbID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid knowledge base ID")
		return
	}

	var req struct {
		Query string `json:"query" binding:"required"`
		Limit int    `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	results, err := h.ragService.SearchDocuments(c.Request.Context(), userID, kbID, req.Query, req.Limit)
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, results, "")
}

