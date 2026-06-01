package models

import "time"

// Message represents a single chat message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatContext stores metadata for a chat conversation context.
type ChatContext struct {
	ConversationID string    `json:"conversation_id"`
	Timestamp      time.Time `json:"timestamp"`
}
