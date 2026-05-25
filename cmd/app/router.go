package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AgentRequest 接收的请求体
type AgentRequest struct {
	Input string `json:"input" binding:"required"`
}

// AgentResponse 返回的响应体
type AgentResponse struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

// setupRouter 设置和配置路由
func setupRouter(engine *gin.Engine) {
	// 添加 CORS 中间件
	engine.Use(corsMiddleware())

	// 注册路由
	engine.GET("/health", handleHealth)
	engine.POST("/agent/query", handleAgentQuery)
}

// corsMiddleware CORS 跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// handleAgentQuery 处理智能体查询的接口
func handleAgentQuery(c *gin.Context) {
	var req AgentRequest

	// 绑定 JSON 请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求绑定失败: %v\n", err)
		c.JSON(http.StatusBadRequest, AgentResponse{
			Error: "无效的请求: " + err.Error(),
		})
		return
	}

	// 检查输入是否为空
	if req.Input == "" {
		c.JSON(http.StatusBadRequest, AgentResponse{
			Error: "输入不能为空",
		})
		return
	}

	log.Printf("【收到用户输入】%s\n", req.Input)

	// 执行智能体
	result := globalEngine.Run(req.Input)

	c.JSON(http.StatusOK, AgentResponse{
		Result: result,
	})
}

// handleHealth 处理健康检查接口
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "FrostAgent 智能体服务运行正常",
	})
}
