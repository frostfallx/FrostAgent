package llm

import (
	"sync"
	"time"
)

// SessionContext 管理单个会话的上下文历史
type SessionContext struct {
	ConversationID string        // 每个会话的唯一标识符
	Messages       []ChatMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SessionManager 管理多个会话上下文，支持多用户/多群聊隔离
type SessionManager struct {
	sessions   map[string]*SessionContext
	mu         sync.RWMutex
	MaxHistory int // 单个会话保留的最大历史消息数
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:   make(map[string]*SessionContext),
		MaxHistory: 20,
	}
}

// GetOrCreate 获取或创建会话
func (sm *SessionManager) GetOrCreate(sessionID string) *SessionContext {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()
	if exists {
		return session
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[sessionID]; exists {
		return session
	}

	session = &SessionContext{
		ConversationID: sessionID,
		Messages:       make([]ChatMessage, 0),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	sm.sessions[sessionID] = session
	return session
}

// Delete 删除指定会话
func (sm *SessionManager) Delete(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}
