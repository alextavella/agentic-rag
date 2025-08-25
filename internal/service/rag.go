package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/alextavella/agentic-rag/internal/domain"
	"github.com/alextavella/agentic-rag/internal/infrastructure/llm"
)

// RAGServiceImpl implementa a interface RAGService
type RAGServiceImpl struct {
	docRepo   domain.DocumentRepository
	llmClient domain.LLMClient
	logger    *slog.Logger
	config    RAGConfig
}

// RAGConfig contém configurações para o serviço RAG
type RAGConfig struct {
	MaxSearchResults int
	SearchTimeout    time.Duration
	LLMTimeout       time.Duration
}

// NewRAGService cria uma nova instância do serviço RAG
func NewRAGService(
	docRepo domain.DocumentRepository,
	llmClient domain.LLMClient,
	logger *slog.Logger,
	config RAGConfig,
) *RAGServiceImpl {
	return &RAGServiceImpl{
		docRepo:   docRepo,
		llmClient: llmClient,
		logger:    logger,
		config:    config,
	}
}

// ProcessQuery processa uma query e retorna uma resposta
func (s *RAGServiceImpl) ProcessQuery(ctx context.Context, req *domain.RAGRequest) (*domain.RAGResponse, error) {
	start := time.Now()

	// Validação básica
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	logger := s.logger.With(
		slog.String("operation", "process_query"),
		slog.String("query", req.Query),
		slog.String("user_id", req.UserID),
	)

	logger.Info("processando query RAG")

	// Prepara as mensagens iniciais
	messages := []*domain.ConversationMessage{
		{
			Role:      "user",
			Content:   req.Query,
			Timestamp: time.Now().Unix(),
		},
	}

	// Define as ferramentas disponíveis
	tools := []domain.Tool{llm.CreateSearchTool()}

	// Primeira chamada ao LLM para decidir se precisa buscar
	llmResp, err := s.llmClient.GenerateResponse(ctx, messages, tools)
	if err != nil {
		logger.Error("erro na primeira chamada ao LLM", slog.Any("error", err))
		return nil, fmt.Errorf("erro ao gerar resposta: %w", err)
	}

	var sources []*domain.Document
	searchPerformed := false

	// Processa tool calls se houver
	if len(llmResp.ToolCalls) > 0 {
		// Adiciona a resposta do assistente às mensagens
		messages = append(messages, &domain.ConversationMessage{
			Role:      "assistant",
			Content:   llmResp.Content,
			Timestamp: time.Now().Unix(),
		})

		for _, toolCall := range llmResp.ToolCalls {
			if toolCall.Name == "search_metadata" {
				searchPerformed = true

				// Extrai argumentos da tool call
				args, err := llm.ParseToolArguments(toolCall.Arguments)
				if err != nil {
					logger.Error("erro ao fazer parse dos argumentos", slog.Any("error", err))
					continue
				}

				query, ok := args["query"].(string)
				if !ok {
					logger.Error("query não encontrada nos argumentos")
					continue
				}

				logger.Info("executando busca", slog.String("search_query", query))

				// Executa a busca
				searchResults, err := s.SearchDocuments(ctx, query, req.MaxResults)
				if err != nil {
					logger.Error("erro na busca", slog.Any("error", err))
					searchResults = []*domain.Document{} // Fallback para array vazio
				}

				sources = searchResults

				// Converte resultados para JSON
				resultsJSON, err := json.Marshal(searchResults)
				if err != nil {
					logger.Error("erro ao converter resultados para JSON", slog.Any("error", err))
					resultsJSON = []byte("[]")
				}

				// Adiciona resposta da ferramenta às mensagens
				messages = append(messages, &domain.ConversationMessage{
					Role:      "tool",
					Content:   string(resultsJSON),
					ToolCall:  toolCall.Name,
					ToolID:    toolCall.ID,
					Timestamp: time.Now().Unix(),
				})
			}
		}

		// Segunda chamada ao LLM com o contexto da busca
		finalResp, err := s.llmClient.GenerateResponse(ctx, messages, nil)
		if err != nil {
			logger.Error("erro na segunda chamada ao LLM", slog.Any("error", err))
			return nil, fmt.Errorf("erro ao gerar resposta final: %w", err)
		}

		llmResp = finalResp
	}

	processingTime := time.Since(start).Milliseconds()

	response := &domain.RAGResponse{
		Answer:          llmResp.Content,
		Sources:         sources,
		Query:           req.Query,
		ProcessingTime:  processingTime,
		SearchPerformed: searchPerformed,
		Model:           llmResp.Model,
		TokensUsed:      llmResp.TokensUsed,
	}

	logger.Info("query processada com sucesso",
		slog.Int("sources_count", len(sources)),
		slog.Int64("processing_time_ms", processingTime),
		slog.Bool("search_performed", searchPerformed),
	)

	return response, nil
}

// SearchDocuments busca documentos relevantes
func (s *RAGServiceImpl) SearchDocuments(ctx context.Context, query string, limit int) ([]*domain.Document, error) {
	if query == "" {
		return nil, domain.ErrQueryEmpty
	}

	if limit <= 0 {
		limit = s.config.MaxSearchResults
	}

	// Aplica timeout específico para busca
	searchCtx, cancel := context.WithTimeout(ctx, s.config.SearchTimeout)
	defer cancel()

	documents, err := s.docRepo.Search(searchCtx, query, limit)
	if err != nil {
		s.logger.Error("erro ao buscar documentos",
			slog.String("query", query),
			slog.Int("limit", limit),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("erro na busca de documentos: %w", err)
	}

	s.logger.Debug("busca realizada com sucesso",
		slog.String("query", query),
		slog.Int("results_count", len(documents)),
	)

	return documents, nil
}

// AddDocument adiciona um novo documento ao sistema
func (s *RAGServiceImpl) AddDocument(ctx context.Context, doc *domain.Document) error {
	if doc == nil {
		return domain.ErrDocumentInvalid
	}

	// Validação básica do documento
	if err := s.validateDocument(doc); err != nil {
		return err
	}

	err := s.docRepo.Insert(ctx, doc)
	if err != nil {
		s.logger.Error("erro ao inserir documento",
			slog.String("title", doc.Title),
			slog.Any("error", err),
		)
		return fmt.Errorf("erro ao adicionar documento: %w", err)
	}

	s.logger.Info("documento adicionado com sucesso",
		slog.String("id", doc.ID),
		slog.String("title", doc.Title),
		slog.String("category", doc.Category),
	)

	return nil
}

// HealthCheck verifica se o serviço está funcionando
func (s *RAGServiceImpl) HealthCheck(ctx context.Context) error {
	// Verifica o repositório de documentos
	if err := s.docRepo.HealthCheck(ctx); err != nil {
		return fmt.Errorf("repositório indisponível: %w", err)
	}

	// Verifica o cliente LLM
	if err := s.llmClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("cliente LLM indisponível: %w", err)
	}

	return nil
}

// validateRequest valida uma requisição RAG
func (s *RAGServiceImpl) validateRequest(req *domain.RAGRequest) error {
	if req == nil {
		return domain.NewValidationError("request", "requisição não pode ser nula")
	}

	if req.Query == "" {
		return domain.ErrQueryEmpty
	}

	if len(req.Query) < 3 {
		return domain.ErrQueryTooShort
	}

	if len(req.Query) > 1000 {
		return domain.ErrQueryTooLong
	}

	if req.MaxResults <= 0 {
		req.MaxResults = s.config.MaxSearchResults
	}

	return nil
}

// validateDocument valida um documento
func (s *RAGServiceImpl) validateDocument(doc *domain.Document) error {
	if doc.Title == "" {
		return domain.NewValidationError("title", "título é obrigatório")
	}

	if doc.Content == "" {
		return domain.NewValidationError("content", "conteúdo é obrigatório")
	}

	if len(doc.Title) > 200 {
		return domain.NewValidationError("title", "título muito longo (máximo 200 caracteres)")
	}

	if len(doc.Content) > 10000 {
		return domain.NewValidationError("content", "conteúdo muito longo (máximo 10000 caracteres)")
	}

	return nil
}
