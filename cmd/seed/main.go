package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/alextavella/agentic-rag/internal/config"
	"github.com/alextavella/agentic-rag/internal/domain"
	"github.com/alextavella/agentic-rag/internal/infrastructure/database"
	"github.com/alextavella/agentic-rag/internal/service"
)

func main() {
	// Configura logging estruturado
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("iniciando seed do banco de dados")

	// Carrega configurações
	cfg, err := config.Load()
	if err != nil {
		logger.Error("erro ao carregar configurações", slog.Any("error", err))
		os.Exit(1)
	}

	// Cria contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
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

	// Configura índices
	if err := docRepo.SetupIndexes(ctx); err != nil {
		logger.Error("erro ao configurar índices", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("índices configurados com sucesso")

	// Inicializa o serviço RAG para adicionar documentos
	ragConfig := service.RAGConfig{
		MaxSearchResults: cfg.App.SearchLimit,
		SearchTimeout:    10 * time.Second,
		LLMTimeout:       30 * time.Second,
	}

	ragService := service.NewRAGService(docRepo, nil, logger, ragConfig) // LLM não é necessário para seed

	// Documentos de exemplo sobre performance em Go
	documents := []*domain.Document{
		domain.NewDocument(
			"Optimizing Go Routines",
			"Goroutines são leves e eficientes, mas é importante gerenciá-las corretamente. Este guia aborda as melhores práticas para otimização de goroutines, incluindo o uso adequado de channels, wait groups e context. Aprenda a evitar vazamentos de goroutines e como usar pools de workers para controlar a concorrência.",
			"/docs/go-optimizing",
			"performance",
		),
		domain.NewDocument(
			"Memory Management in Go",
			"O garbage collector do Go é sofisticado, mas entender como ele funciona é crucial para otimização. Aprenda sobre alocação de memória, escape analysis e dicas para reduzir a pressão no GC. Descubra técnicas para minimizar alocações desnecessárias e otimizar o uso de memória em aplicações Go.",
			"/docs/go-memory",
			"performance",
		),
		domain.NewDocument(
			"Profiling Go Applications",
			"Ferramentas de profiling são essenciais para identificar gargalos. Este documento explora o uso de pprof, trace e outras ferramentas built-in do Go para análise de performance. Aprenda a identificar hotspots no código e otimizar aplicações baseado em dados reais de profiling.",
			"/docs/go-profiling",
			"performance",
		),
		domain.NewDocument(
			"Database Performance in Go",
			"Otimize suas consultas de banco de dados em Go. Aprenda sobre connection pooling, prepared statements e como estruturar suas queries para máxima eficiência. Descubra padrões para trabalhar com ORMs, drivers nativos e técnicas de cache para melhorar a performance de acesso a dados.",
			"/docs/go-db-performance",
			"performance",
		),
		domain.NewDocument(
			"Network Performance Tuning",
			"Maximize a performance de rede em aplicações Go. Inclui dicas sobre TCP tuning, HTTP/2, e como implementar client-side caching efetivamente. Aprenda sobre timeouts, connection pooling e estratégias para otimizar comunicação entre serviços distribuídos.",
			"/docs/go-network",
			"performance",
		),
		domain.NewDocument(
			"Concurrency Patterns in Go",
			"Explore padrões avançados de concorrência em Go. Desde worker pools até pipelines de processamento, este guia apresenta técnicas para maximizar o paralelismo. Aprenda sobre fan-in, fan-out, rate limiting e como estruturar aplicações altamente concorrentes de forma segura.",
			"/docs/go-concurrency",
			"patterns",
		),
		domain.NewDocument(
			"Error Handling Best Practices",
			"Estratégias eficazes para tratamento de erros em Go. Aprenda sobre error wrapping, custom error types e padrões para propagação de erros em aplicações complexas. Descubra como implementar logging estruturado e monitoramento eficaz de erros em produção.",
			"/docs/go-errors",
			"best-practices",
		),
		domain.NewDocument(
			"Testing Strategies for Go Applications",
			"Guia completo para testes em Go, incluindo testes unitários, de integração e end-to-end. Aprenda sobre table-driven tests, mocking, benchmarking e como estruturar suites de testes para projetos grandes. Inclui estratégias para testes de aplicações concorrentes.",
			"/docs/go-testing",
			"testing",
		),
	}

	// Limpa a coleção antes de inserir os novos documentos
	logger.Info("limpando coleção existente")
	if err := docRepo.DeleteAll(ctx); err != nil {
		logger.Error("erro ao limpar a coleção", slog.Any("error", err))
		os.Exit(1)
	}

	// Insere os documentos
	logger.Info("inserindo documentos", slog.Int("count", len(documents)))

	successCount := 0
	for _, doc := range documents {
		// Adiciona metadados extras
		doc.AddMetadata("source", "seed")
		doc.AddMetadata("version", "1.0")

		if err := ragService.AddDocument(ctx, doc); err != nil {
			logger.Error("erro ao inserir documento",
				slog.String("title", doc.Title),
				slog.Any("error", err),
			)
			continue
		}

		logger.Info("documento inserido com sucesso",
			slog.String("id", doc.ID),
			slog.String("title", doc.Title),
			slog.String("category", doc.Category),
		)
		successCount++
	}

	// Verifica o total de documentos
	totalDocs, err := docRepo.Count(ctx)
	if err != nil {
		logger.Error("erro ao contar documentos", slog.Any("error", err))
	} else {
		logger.Info("contagem final de documentos", slog.Int64("total", totalDocs))
	}

	logger.Info("seed concluído",
		slog.Int("inserted", successCount),
		slog.Int("total_attempted", len(documents)),
		slog.Int64("final_count", totalDocs),
	)

	if successCount == len(documents) {
		logger.Info("todos os documentos foram inseridos com sucesso!")
	} else {
		logger.Warn("alguns documentos falharam ao ser inseridos",
			slog.Int("failed", len(documents)-successCount),
		)
		os.Exit(1)
	}
}
