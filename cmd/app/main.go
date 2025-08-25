package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alextavella/agentic-rag/internal/config"
	"github.com/alextavella/agentic-rag/internal/domain"
	"github.com/alextavella/agentic-rag/internal/infrastructure/database"
	"github.com/alextavella/agentic-rag/internal/infrastructure/llm"
	"github.com/alextavella/agentic-rag/internal/service"
)

func main() {
	// Configura logging estruturado
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Carrega configurações
	cfg, err := config.Load()
	if err != nil {
		logger.Error("erro ao carregar configurações", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("aplicação iniciando",
		slog.String("model", cfg.OpenAI.Model),
		slog.String("database", cfg.Database.Database),
	)

	// Cria contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Inicializa o repositório de documentos
	docRepo, err := database.NewMongoDocumentRepository(
		ctx,
		cfg.Database.URI,
		cfg.Database.Database,
		cfg.Database.Collection,
	)
	if err != nil {
		logger.Error("erro ao conectar ao MongoDB", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := docRepo.Close(ctx); err != nil {
			logger.Error("erro ao fechar conexão com MongoDB", slog.Any("error", err))
		}
	}()

	// Configura índices (necessário apenas uma vez)
	if err := docRepo.SetupIndexes(ctx); err != nil {
		logger.Warn("aviso ao configurar índices", slog.Any("error", err))
	}

	// Inicializa o cliente LLM
	llmClient := llm.NewOpenAIClient(cfg.OpenAI.APIKey, cfg.OpenAI.Model)

	// Configura o serviço RAG
	ragConfig := service.RAGConfig{
		MaxSearchResults: cfg.App.SearchLimit,
		SearchTimeout:    10 * time.Second,
		LLMTimeout:       30 * time.Second,
	}

	ragService := service.NewRAGService(docRepo, llmClient, logger, ragConfig)

	// Verifica se os serviços estão funcionando
	if err := ragService.HealthCheck(ctx); err != nil {
		logger.Error("health check falhou", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("serviços inicializados com sucesso")

	// Processa a query padrão ou uma fornecida via argumento
	query := cfg.App.DefaultQuery
	if len(os.Args) > 1 {
		query = os.Args[1]
	}

	// Cria a requisição RAG
	req := &domain.RAGRequest{
		Query:      query,
		MaxResults: cfg.App.SearchLimit,
		UserID:     "cli-user",
		SessionID:  fmt.Sprintf("session-%d", time.Now().Unix()),
	}

	logger.Info("processando query", slog.String("query", query))

	// Processa a query
	response, err := ragService.ProcessQuery(ctx, req)
	if err != nil {
		logger.Error("erro ao processar query", slog.Any("error", err))
		os.Exit(1)
	}

	// Exibe os resultados
	fmt.Printf("\n=== RESPOSTA DO AGENTE ===\n")
	fmt.Printf("%s\n", response.Answer)

	if response.SearchPerformed && len(response.Sources) > 0 {
		fmt.Printf("\n=== FONTES CONSULTADAS ===\n")
		for i, source := range response.Sources {
			fmt.Printf("%d. %s\n", i+1, source.Title)
			fmt.Printf("   Link: %s\n", source.Link)
			fmt.Printf("   Categoria: %s\n", source.Category)
			if len(source.Content) > 100 {
				fmt.Printf("   Conteúdo: %s...\n", source.Content[:100])
			} else {
				fmt.Printf("   Conteúdo: %s\n", source.Content)
			}
			fmt.Printf("\n")
		}
	}

	fmt.Printf("\n=== ESTATÍSTICAS ===\n")
	fmt.Printf("Tempo de processamento: %dms\n", response.ProcessingTime)
	fmt.Printf("Busca realizada: %v\n", response.SearchPerformed)
	fmt.Printf("Modelo usado: %s\n", response.Model)
	if response.TokensUsed > 0 {
		fmt.Printf("Tokens utilizados: %d\n", response.TokensUsed)
	}

	logger.Info("aplicação finalizada com sucesso")
}
