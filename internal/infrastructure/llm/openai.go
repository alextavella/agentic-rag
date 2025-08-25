package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alextavella/agentic-rag/internal/domain"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient implementa a interface LLMClient usando OpenAI
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient cria um novo cliente OpenAI
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	client := openai.NewClient(apiKey)
	return &OpenAIClient{
		client: client,
		model:  model,
	}
}

// GenerateResponse gera uma resposta usando o modelo OpenAI
func (c *OpenAIClient) GenerateResponse(ctx context.Context, messages []*domain.ConversationMessage, tools []domain.Tool) (*domain.LLMResponse, error) {
	// Converte mensagens do domínio para formato OpenAI
	openaiMessages := make([]openai.ChatCompletionMessage, 0, len(messages))

	for _, msg := range messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Adiciona informações de tool call se presente
		if msg.ToolCall != "" && msg.ToolID != "" {
			openaiMsg.Name = msg.ToolCall
			openaiMsg.ToolCallID = msg.ToolID
		}

		openaiMessages = append(openaiMessages, openaiMsg)
	}

	// Prepara a requisição
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMessages,
	}

	// Adiciona ferramentas se fornecidas
	if len(tools) > 0 {
		openaiTools := make([]openai.Tool, 0, len(tools))
		for _, tool := range tools {
			openaiTool := openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			}
			openaiTools = append(openaiTools, openaiTool)
		}
		req.Tools = openaiTools
	}

	// Faz a chamada para OpenAI
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("erro na chamada OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, domain.ErrLLMInvalidResponse
	}

	choice := resp.Choices[0]

	// Converte tool calls se presentes
	var toolCalls []*domain.ToolCall
	if len(choice.Message.ToolCalls) > 0 {
		toolCalls = make([]*domain.ToolCall, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			toolCall := &domain.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
			toolCalls = append(toolCalls, toolCall)
		}
	}

	return &domain.LLMResponse{
		Content:      choice.Message.Content,
		ToolCalls:    toolCalls,
		TokensUsed:   resp.Usage.TotalTokens,
		Model:        resp.Model,
		FinishReason: string(choice.FinishReason),
	}, nil
}

// GetModel retorna o modelo sendo usado
func (c *OpenAIClient) GetModel() string {
	return c.model
}

// HealthCheck verifica se o cliente OpenAI está funcionando
func (c *OpenAIClient) HealthCheck(ctx context.Context) error {
	// Tenta fazer uma chamada simples para verificar conectividade
	messages := []*domain.ConversationMessage{
		{
			Role:      "user",
			Content:   "Hello",
			Timestamp: time.Now().Unix(),
		},
	}

	_, err := c.GenerateResponse(ctx, messages, nil)
	if err != nil {
		return fmt.Errorf("health check falhou: %w", err)
	}

	return nil
}

// CreateSearchTool cria a ferramenta de busca para o OpenAI
func CreateSearchTool() domain.Tool {
	return domain.Tool{
		Name:        "search_metadata",
		Description: "Search metadata in database or API from a query",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Text to search in metadata",
				},
			},
			"required": []string{"query"},
		},
	}
}

// ParseToolArguments extrai argumentos de uma tool call
func ParseToolArguments(arguments string) (map[string]interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse dos argumentos: %w", err)
	}
	return args, nil
}
