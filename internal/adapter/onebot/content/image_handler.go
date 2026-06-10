package content

import (
	"FrostAgent/internal/core"
	"FrostAgent/internal/llm"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func IsContainImage(segments []MessageSegment) bool {
	for _, seg := range segments {
		if seg.Type == "image" {
			return true
		}
	}
	return false
}

// ProcessImage 现在接受 core.LLMProvider 接口
func ProcessImage(segments []MessageSegment, provider core.LLMProvider, baseURL, apiKey, modelName string) string {
	var userTexts []string
	var imageBase64List []string

	for _, seg := range segments {
		if seg.Type == "text" {
			if text, ok := seg.Data["text"].(string); ok {
				userTexts = append(userTexts, text)
			}
		} else if seg.Type == "image" {
			url, ok := seg.Data["url"].(string)
			if !ok || strings.TrimSpace(url) == "" {
				continue
			}
			if b64, err := downloadAndToBase64(url); err == nil {
				imageBase64List = append(imageBase64List, b64)
			}
		}
	}

	combinedText := strings.Join(userTexts, "")
	if len(imageBase64List) > 0 {
		contentBlocks := []llm.ContentBlock{
			{Type: "text", Text: combinedText},
		}

		for _, b64 := range imageBase64List {
			contentBlocks = append(contentBlocks, llm.ContentBlock{
				Type:     "image_url",
				ImageURL: map[string]string{"url": "data:image/jpeg;base64," + b64},
			})
		}
		jsonBytes, _ := json.Marshal(contentBlocks)
		// 调用更新后的视觉模型函数
		return llm.CallVisionModel(provider, baseURL, apiKey, modelName, string(jsonBytes))
	}
	return combinedText
}

func downloadAndToBase64(url string) (string, error) {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(imgBytes), nil
}
