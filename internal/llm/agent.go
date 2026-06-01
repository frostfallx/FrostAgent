package llm

import (
	"fmt"
	"os"
	"time"

	"FrostAgent/internal/tools"
)

// Engine 结构体，用于管理智能体的执行
type Engine struct {
	MaxIterations  int
	ToolRegistry   map[string]tools.Tool
	LLMClient      *Client          // API 客户端
	BaseURL        string           // API 地址
	APIKey         string           // API 密钥
	ModelName      string           // 模型名称
	SessionManager *SessionManager  // 会话上下文管理器
}

// Run 执行智能体的主循环（单次无状态调用）
func (e *Engine) Run(prompt string) string {
	systemPrompt := os.Getenv("SYSTEM_PROMPT")
	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}
	result := e.runLoop(messages)
	return result
}

// RunWithSession 执行智能体的主循环（带会话上下文）
// sessionID 用于区分不同的会话（如不同用户或群聊）
func (e *Engine) RunWithSession(sessionID string, prompt string) string {
	session := e.SessionManager.GetOrCreate(sessionID)

	// 获取历史消息
	messages := session.Messages

	// 如果是新会话，添加系统提示词
	if len(messages) == 0 {
		systemPrompt := os.Getenv("SYSTEM_PROMPT")
		messages = append(messages, ChatMessage{Role: "system", Content: systemPrompt})
	}

	// 添加用户输入
	messages = append(messages, ChatMessage{Role: "user", Content: prompt})

	// 运行核心循环，获得最终消息列表和答案
	result := e.runLoop(messages)

	// 将最终的消息列表写回会话
	// 这里先把系统 prompt 保留，然后截取后续的所有消息
	// 注意：runLoop 内部对 messages 做了 append，我们需要拿到修改后的 messages
	// 因此在 runLoop 中我们会直接修改传入的切片
	session.Messages = e.trimMessagesForSession(messages)
	session.UpdatedAt = time.Now()

	return result
}

// runLoop 核心循环逻辑，封装工具调用和多轮推理
func (e *Engine) runLoop(messages []ChatMessage) string {
	// 转换工具注册表为大模型看得懂的形式
	var modelTools []any
	for _, t := range e.ToolRegistry {
		modelTools = append(modelTools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.Parameters,
			},
		})
	}

	// 主循环
	for i := 0; i < e.MaxIterations; i++ {
		fmt.Printf("【第%d轮思考开始】\n", i+1)

		// 调用 internal/llm 包向大模型发送 HTTP 请求
		responseMsg, err := e.LLMClient.CallAPI(e.BaseURL, e.APIKey, e.ModelName, messages, modelTools)
		if err != nil {
			return fmt.Sprintf("LLM掉线了: %v", err)
		}

		// 记下回复
		messages = append(messages, *responseMsg)

		// 是否给出最终答案
		if len(responseMsg.ToolCalls) == 0 {
			fmt.Println("【智能体给出最终答案】")
			contentStr, _ := responseMsg.Content.(string)
			return contentStr
		}

		// 发现有指令，开始干活
		for _, tc := range responseMsg.ToolCalls {
			fmt.Printf("【智能体调用工具】%s，参数: %s\n", tc.Function.Name, tc.Function.Arguments)

			var toolResult string

			// 从 map 中找到工具执行
			if tool, exists := e.ToolRegistry[tc.Function.Name]; exists {
				res, err := tool.Execute(tc.Function.Arguments)
				if err != nil {
					toolResult = fmt.Sprintf("工具执行失败: %v", err)
				} else {
					toolResult = res
				}
			} else {
				toolResult = "工具未找到"
			}

			fmt.Println("【工具执行结果】", toolResult)

			// 把结果包装成 role=tool 的消息，记录
			toolMsg := ChatMessage{
				Role:       "tool",
				Content:    toolResult,
				ToolCallID: tc.ID,
			}
			messages = append(messages, toolMsg)
		}
		// 循环进入下一轮
	}
	return "达到最大迭代次数，未能得出最终答案"
}

// trimMessagesForSession 保留系统提示词和最近的对话历史，防止无限膨胀
func (e *Engine) trimMessagesForSession(messages []ChatMessage) []ChatMessage {
	if len(messages) <= e.SessionManager.MaxHistory+1 {
		return messages
	}
	// 保留第一条（system prompt）和最近的 MaxHistory 条
	trimmed := make([]ChatMessage, 0, e.SessionManager.MaxHistory+1)
	trimmed = append(trimmed, messages[0])
	trimmed = append(trimmed, messages[len(messages)-e.SessionManager.MaxHistory:]...)
	return trimmed
}
