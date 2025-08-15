package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/alextavella/agentic-rag/internal/database"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Conecta ao MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	db, err := database.NewMongoDB(ctx, mongoURI)
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer db.Close(ctx)

	// Documentos de exemplo sobre performance em Go
	documents := []database.Document{
		{
			Title:    "Optimizing Go Routines",
			Content:  "Goroutines são leves e eficientes, mas é importante gerenciá-las corretamente. Este guia aborda as melhores práticas para otimização de goroutines, incluindo o uso adequado de channels, wait groups e context.",
			Link:     "/docs/go-optimizing",
			Category: "performance",
		},
		{
			Title:    "Memory Management in Go",
			Content:  "O garbage collector do Go é sofisticado, mas entender como ele funciona é crucial para otimização. Aprenda sobre alocação de memória, escape analysis e dicas para reduzir a pressão no GC.",
			Link:     "/docs/go-memory",
			Category: "performance",
		},
		{
			Title:    "Profiling Go Applications",
			Content:  "Ferramentas de profiling são essenciais para identificar gargalos. Este documento explora o uso de pprof, trace e outras ferramentas built-in do Go para análise de performance.",
			Link:     "/docs/go-profiling",
			Category: "performance",
		},
		{
			Title:    "Database Performance in Go",
			Content:  "Otimize suas consultas de banco de dados em Go. Aprenda sobre connection pooling, prepared statements e como estruturar suas queries para máxima eficiência.",
			Link:     "/docs/go-db-performance",
			Category: "performance",
		},
		{
			Title:    "Network Performance Tuning",
			Content:  "Maximize a performance de rede em aplicações Go. Inclui dicas sobre TCP tuning, HTTP/2, e como implementar client-side caching efetivamente.",
			Link:     "/docs/go-network",
			Category: "performance",
		},
	}

	// Limpa a coleção antes de inserir os novos documentos
	if err := db.ClearCollection(ctx); err != nil {
		log.Fatalf("Erro ao limpar a coleção: %v", err)
	}

	// Insere os documentos
	for _, doc := range documents {
		if err := db.InsertDocument(ctx, doc); err != nil {
			log.Printf("Erro ao inserir documento '%s': %v", doc.Title, err)
			continue
		}
		log.Printf("Documento inserido com sucesso: %s", doc.Title)
	}

	log.Println("Seed concluído com sucesso!")
}
