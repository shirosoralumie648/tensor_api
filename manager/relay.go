package manager

import (
	"chat/admin"
	"chat/channel"
	"chat/globals"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ModelAPI(c *gin.Context) {
	c.JSON(http.StatusOK, globals.V1ListModels)
}

func MarketAPI(c *gin.Context) {
	c.JSON(http.StatusOK, admin.MarketInstance.GetModels())
}

func ChargeAPI(c *gin.Context) {
	c.JSON(http.StatusOK, channel.ChargeInstance.ListRules())
}

func PlanAPI(c *gin.Context) {
	c.JSON(http.StatusOK, channel.PlanInstance.GetPlans())
}

func ProvidersAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"OpenAI": gin.H{
			"provider": "OpenAI",
			"docs":     "https://platform.openai.com/docs",
			"features": gin.H{
				"text":           true,
				"vision":         true,
				"tools":          true,
				"images":         true,
				"audio":          true,
				"video":          false,
				"embeddings":     true,
				"context":        128000,
				"json":           true,
				"parallel_tools": true,
				"streaming":      true,
			},
		},
		"Anthropic": gin.H{
			"provider": "Anthropic",
			"docs":     "https://docs.anthropic.com",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     false,
				"audio":      false,
				"video":      false,
				"embeddings": false,
				"context":    200000,
			},
		},
		"Google": gin.H{
			"provider": "Google",
			"docs":     "https://ai.google.dev",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     true,
				"audio":      true,
				"video":      false,
				"embeddings": true,
				"context":    100000,
				"json":       true,
			},
		},
		"DeepSeek": gin.H{
			"provider": "DeepSeek",
			"docs":     "https://platform.deepseek.com",
			"features": gin.H{
				"text":       true,
				"vision":     false,
				"tools":      true,
				"images":     false,
				"audio":      false,
				"video":      false,
				"embeddings": false,
				"context":    32000,
			},
		},
		"Alibaba Qwen": gin.H{
			"provider": "Alibaba Qwen",
			"docs":     "https://help.aliyun.com/zh/model-studio",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     true,
				"audio":      true,
				"video":      false,
				"embeddings": true,
				"context":    128000,
			},
		},
		"Zhipu GLM": gin.H{
			"provider": "Zhipu GLM",
			"docs":     "https://open.bigmodel.cn/dev/api",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     false,
				"audio":      false,
				"video":      false,
				"embeddings": true,
				"context":    128000,
			},
		},
		"Meta Llama": gin.H{
			"provider": "Meta Llama",
			"docs":     "https://ai.meta.com/llama/",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     false,
				"audio":      false,
				"video":      false,
				"embeddings": true,
				"context":    128000,
			},
		},
		"Mistral": gin.H{
			"provider": "Mistral",
			"docs":     "https://docs.mistral.ai",
			"features": gin.H{
				"text":           true,
				"vision":         false,
				"tools":          true,
				"images":         false,
				"audio":          false,
				"video":          false,
				"embeddings":     true,
				"context":        32000,
				"json":           true,
				"parallel_tools": true,
			},
		},
		"Moonshot": gin.H{
			"provider": "Moonshot",
			"docs":     "https://docs.moonshot.cn",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     false,
				"audio":      false,
				"video":      false,
				"embeddings": false,
				"context":    128000,
			},
		},
		"MiniMax": gin.H{
			"provider": "MiniMax",
			"docs":     "https://api.minimax.chat/document",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     false,
				"audio":      false,
				"video":      false,
				"embeddings": false,
				"context":    64000,
			},
		},
		"Baidu ERNIE": gin.H{
			"provider": "Baidu ERNIE",
			"docs":     "https://cloud.baidu.com/doc/WENXINWORKSHOP/index.html",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     true,
				"audio":      true,
				"video":      false,
				"embeddings": true,
				"context":    32000,
			},
		},
		"Tencent Hunyuan": gin.H{
			"provider": "Tencent Hunyuan",
			"docs":     "https://cloud.tencent.com/document/product/1729",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     true,
				"audio":      true,
				"video":      false,
				"embeddings": true,
				"context":    32000,
			},
		},
		"ByteDance Doubao": gin.H{
			"provider": "ByteDance Doubao",
			"docs":     "https://www.volcengine.com/docs/82379/1260579",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      true,
				"images":     true,
				"audio":      true,
				"video":      false,
				"embeddings": true,
				"context":    64000,
			},
		},
		"Ollama": gin.H{
			"provider": "Ollama",
			"docs":     "https://github.com/ollama/ollama",
			"features": gin.H{
				"text":       true,
				"vision":     true,
				"tools":      false,
				"images":     false,
				"audio":      false,
				"video":      false,
				"embeddings": true,
				"context":    8192,
			},
		},
		"Other": gin.H{
			"provider": "Other",
			"features": gin.H{
				"text":       true,
				"vision":     false,
				"tools":      false,
				"images":     false,
				"audio":      false,
				"video":      false,
				"embeddings": false,
			},
		},
	})
}

func sendErrorResponse(c *gin.Context, err error, types ...string) {
	var errType string
	if len(types) > 0 {
		errType = types[0]
	} else {
		errType = "chatnio_api_error"
	}

	c.JSON(http.StatusServiceUnavailable, RelayErrorResponse{
		Error: TranshipmentError{
			Message: err.Error(),
			Type:    errType,
		},
	})
}

func abortWithErrorResponse(c *gin.Context, err error, types ...string) {
	sendErrorResponse(c, err, types...)
	c.Abort()
}
