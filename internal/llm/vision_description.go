package llm

import (
	"FrostAgent/internal/core"
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
)

// CallVisionModel 现在接受 core.LLMProvider 接口，实现解耦
func CallVisionModel(provider core.LLMProvider, baseURL, apiKey, _, contentBlocks string) string {
	systemPrompt := "请用中文描述图片："

	var parsedContent any = contentBlocks
	var blocks []any
	if err := json.Unmarshal([]byte(contentBlocks), &blocks); err == nil {
		for i, b := range blocks {
			if m, ok := b.(map[string]any); ok {
				if m["type"] == "text" {
					textVal, _ := m["text"].(string)
					if strings.TrimSpace(textVal) == "" {
						m["text"] = "请详细描述这张图片的内容"
					}
					blocks[i] = m
				}
			}
		}
		parsedContent = blocks
	}

	// 构造 core 层的请求
	chatReq := core.ChatRequest{
		Model: os.Getenv("VISUAL_MODEL_NAME"),
		Messages: []core.ChatMessage{
			{Role: core.RoleSystem, Content: systemPrompt},
			{Role: core.RoleUser, Content: parsedContent},
		},
	}

	log.Printf("【视觉模型】即将通过 Provider 发起调用")

	resp, err := provider.Chat(context.Background(), chatReq)
	if err != nil {
		log.Printf("视觉模型调用失败: %v", err)
		return err.Error()
	}

	var contentStr string
	if str, ok := resp.Message.Content.(string); ok {
		contentStr = str
	} else if resp.Message.Content != nil {
		bytes, _ := json.Marshal(resp.Message.Content)
		contentStr = string(bytes)
	} else {
		contentStr = "【视觉模型未返回内容】"
	}

	log.Printf("视觉模型描述完成")
	return contentStr
}
