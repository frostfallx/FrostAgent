package main

import (
	"FrostAgent/internal/agent"
	"FrostAgent/internal/llm"
	"FrostAgent/internal/tools"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// 全局引擎实例
var globalEngine *agent.Engine

func init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，将使用默认配置")
	}

	fmt.Println("【初始化】正在初始化智能体引擎...")

	llmClient := llm.NewClient()

	registry := make(map[string]agent.Tool)
	weatherTool := tools.GetWeatherTool()
	registry[weatherTool.Name] = weatherTool

	globalEngine = &agent.Engine{
		MaxIterations: 5,
		ToolRegistry:  registry,
		LLMClient:     llmClient,
		BaseURL:       os.Getenv("UPSTREAM_ENDPOINT"),
		APIKey:        os.Getenv("UPSTREAM_API_KEY"),
		ModelName:     "qwen3-coder-flash",
	}

	fmt.Println("【初始化】✓ 智能体引擎初始化完成")
}

func main() {
	// 创建 Gin 路由
	router := gin.Default()

	// 设置路由
	setupRouter(router)

	// 启动服务器
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}
	fmt.Printf("\n🚀 FrostAgent 智能体服务已启动\n")
	fmt.Printf("📍 监听地址: http://localhost%s\n", listenAddr)
	fmt.Printf("📝 查询接口: POST http://localhost%s/agent/query\n", listenAddr)
	fmt.Printf("✓ 健康检查: GET http://localhost%s/health\n\n", listenAddr)

	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
