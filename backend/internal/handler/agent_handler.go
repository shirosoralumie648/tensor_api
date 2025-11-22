package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shirosoralumie648/Oblivious/backend/internal/service"
	"github.com/shirosoralumie648/Oblivious/backend/internal/utils"
)

// AgentHandler 处理助手相关的 HTTP 请求
type AgentHandler struct {
	agentService *service.AgentService
}

// NewAgentHandler 创建新的 Agent Handler
func NewAgentHandler() *AgentHandler {
	return &AgentHandler{
		agentService: service.NewAgentService(),
	}
}

// CreateAgent 创建新的助手
// POST /api/v1/agents
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req service.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	agent, err := h.agentService.CreateAgent(c.Request.Context(), userID, &req)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, agent, "助手创建成功")
}

// GetAgent 获取助手详情
// GET /api/v1/agents/:id
func (h *AgentHandler) GetAgent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid agent ID")
		return
	}

	agent, err := h.agentService.GetAgentByID(c.Request.Context(), id)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	if agent == nil {
		utils.NotFound(c, "助手不存在")
		return
	}

	utils.Success(c, agent, "")
}

// GetUserAgents 获取用户创建的所有助手
// GET /api/v1/agents/user
func (h *AgentHandler) GetUserAgents(c *gin.Context) {
	userID := c.GetInt("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	agents, total, err := h.agentService.GetUserAgents(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"agents":    agents,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, "")
}

// GetPublicAgents 获取公开的助手列表
// GET /api/v1/agents/public
func (h *AgentHandler) GetPublicAgents(c *gin.Context) {
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	agents, total, err := h.agentService.GetPublicAgents(c.Request.Context(), category, page, pageSize)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"agents":    agents,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, "")
}

// GetFeaturedAgents 获取精选助手
// GET /api/v1/agents/featured
func (h *AgentHandler) GetFeaturedAgents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	agents, err := h.agentService.GetFeaturedAgents(c.Request.Context(), limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, agents, "")
}

// UpdateAgent 更新助手信息
// PUT /api/v1/agents/:id
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid agent ID")
		return
	}

	var req service.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	agent, err := h.agentService.UpdateAgent(c.Request.Context(), userID, id, &req)
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, agent, "助手更新成功")
}

// DeleteAgent 删除助手
// DELETE /api/v1/agents/:id
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid agent ID")
		return
	}

	if err := h.agentService.DeleteAgent(c.Request.Context(), userID, id); err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, nil, "助手删除成功")
}

// SearchAgents 搜索助手
// GET /api/v1/agents/search
func (h *AgentHandler) SearchAgents(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		utils.BadRequest(c, "搜索关键词不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	agents, total, err := h.agentService.SearchAgents(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"agents":    agents,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, "")
}

// LikeAgent 赞助手
// POST /api/v1/agents/:id/like
func (h *AgentHandler) LikeAgent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid agent ID")
		return
	}

	if err := h.agentService.LikeAgent(c.Request.Context(), id); err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, nil, "已赞")
}

// ForkAgent 复制助手
// POST /api/v1/agents/:id/fork
func (h *AgentHandler) ForkAgent(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid agent ID")
		return
	}

	var req struct {
		ForkName string `json:"fork_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	agent, err := h.agentService.ForkAgent(c.Request.Context(), userID, id, req.ForkName)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, agent, "助手复制成功")
}

// GetAgentStats 获取助手统计信息
// GET /api/v1/agents/:id/stats
func (h *AgentHandler) GetAgentStats(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid agent ID")
		return
	}

	stats, err := h.agentService.GetUsageStats(c.Request.Context(), id)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, stats, "")
}

