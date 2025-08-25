package domain

import (
	"context"
)

// RAGRequest representa uma solicitação para o sistema RAG
type RAGRequest struct {
	Query      string            `json:"query"`
	UserID     string            `json:"user_id,omitempty"`
	SessionID  string            `json:"session_id,omitempty"`
	MaxResults int               `json:"max_results,omitempty"`
	Category   string            `json:"category,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// RAGResponse representa a resposta do sistema RAG
type RAGResponse struct {
	Answer          string      `json:"answer"`
	Sources         []*Document `json:"sources"`
	Query           string      `json:"query"`
	ProcessingTime  int64       `json:"processing_time_ms"`
	SearchPerformed bool        `json:"search_performed"`
	Model           string      `json:"model"`
	TokensUsed      int         `json:"tokens_used,omitempty"`
}

// SearchResult representa um resultado de busca simplificado
type SearchResult struct {
	Title    string  `json:"title"`
	Link     string  `json:"link"`
	Content  string  `json:"content,omitempty"`
	Category string  `json:"category,omitempty"`
	Score    float64 `json:"score,omitempty"`
}

// ConversationMessage representa uma mensagem na conversa
type ConversationMessage struct {
	Role      string `json:"role"` // user, assistant, tool
	Content   string `json:"content"`
	ToolCall  string `json:"tool_call,omitempty"`
	ToolID    string `json:"tool_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// Conversation representa uma conversa completa
type Conversation struct {
	ID       string                 `json:"id"`
	UserID   string                 `json:"user_id"`
	Messages []*ConversationMessage `json:"messages"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

// LLMClient define a interface para clientes de modelos de linguagem
type LLMClient interface {
	// GenerateResponse gera uma resposta usando o modelo de linguagem
	GenerateResponse(ctx context.Context, messages []*ConversationMessage, tools []Tool) (*LLMResponse, error)

	// GetModel retorna o nome do modelo sendo usado
	GetModel() string

	// HealthCheck verifica se o cliente está funcionando
	HealthCheck(ctx context.Context) error
}

// LLMResponse representa a resposta de um modelo de linguagem
type LLMResponse struct {
	Content      string      `json:"content"`
	ToolCalls    []*ToolCall `json:"tool_calls,omitempty"`
	TokensUsed   int         `json:"tokens_used"`
	Model        string      `json:"model"`
	FinishReason string      `json:"finish_reason"`
}

// Tool representa uma ferramenta disponível para o LLM
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall representa uma chamada de ferramenta feita pelo LLM
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// RAGService define a interface principal do serviço RAG
type RAGService interface {
	// ProcessQuery processa uma query e retorna uma resposta
	ProcessQuery(ctx context.Context, req *RAGRequest) (*RAGResponse, error)

	// SearchDocuments busca documentos relevantes
	SearchDocuments(ctx context.Context, query string, limit int) ([]*Document, error)

	// AddDocument adiciona um novo documento ao sistema
	AddDocument(ctx context.Context, doc *Document) error

	// HealthCheck verifica se o serviço está funcionando
	HealthCheck(ctx context.Context) error
}
