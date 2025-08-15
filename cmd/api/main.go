package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/alextavella/agentic-rag/internal/database"
	openai "github.com/sashabaranov/go-openai"
)

// SearchResult representa a estrutura de um resultado de busca
// retornado pela função de pesquisa em metadados
type SearchResult struct {
	Title string `json:"title"` // Título do documento encontrado
	Link  string `json:"link"`  // Link/caminho para o documento
}

func main() {
	// Cria um contexto padrão para controlar cancelamento e timeouts
	ctx := context.Background()

	// Inicializa o cliente da OpenAI usando a chave de API do ambiente
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)

	// Inicializa a conexão com o MongoDB
	// mongoURI := "mongodb://admin:password123@localhost:27017"
	mongoURI := os.Getenv("MONGO_URI")
	db, err := database.NewMongoDB(ctx, mongoURI)
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer db.Close(ctx)

	// Configura o índice de texto (necessário apenas uma vez)
	if err := db.SetupTextIndex(ctx); err != nil {
		log.Printf("Aviso ao configurar índice: %v", err)
	}

	// Define a ferramenta de busca que o agente poderá usar
	searchTool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "search_metadata",
			Description: "Search metadata in database or API from a query",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Text to search in metadata",
					},
				},
				"required": []string{"query"},
			},
		},
	}

	// Mensagem inicial do usuário - aqui é onde começa a conversa
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "What are the documents related to Golang performance?",
		},
	}

	// Primeira chamada à API: permite que o agente decida se precisa usar a ferramenta de busca
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4TurboPreview,   // Usando o modelo mais recente da OpenAI
		Messages: messages,                  // Lista de mensagens da conversa
		Tools:    []openai.Tool{searchTool}, // Ferramentas disponíveis para o agente
	})
	if err != nil {
		log.Fatalf("Erro na chamada à OpenAI: %v", err)
	}

	// Processa as chamadas de ferramentas, se houver alguma
	if len(resp.Choices) > 0 && len(resp.Choices[0].Message.ToolCalls) > 0 {
		// Adiciona a resposta do assistente ao histórico de mensagens
		messages = append(messages, resp.Choices[0].Message)

		// Processa cada chamada de ferramenta feita pelo agente
		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			if toolCall.Function.Name == "search_metadata" {
				// Extrai os argumentos da função (a consulta de busca)
				var args struct {
					Query string `json:"query"`
				}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					log.Fatalf("Erro ao processar argumentos: %v", err)
				}

				// Executa a busca real no MongoDB
				results, err := db.SearchDocuments(ctx, args.Query)
				if err != nil {
					log.Printf("Erro na busca: %v", err)
					results = "[]" // Fallback para array vazio em caso de erro
				}

				// Adiciona a resposta da ferramenta ao histórico de mensagens
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    results,
					Name:       toolCall.Function.Name,
					ToolCallID: toolCall.ID,
				})
			}
		}

		// Obtém a resposta final do agente, incluindo o contexto da busca
		finalResp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    openai.GPT4TurboPreview,
			Messages: messages,
		})
		if err != nil {
			log.Fatalf("Erro na resposta final: %v", err)
		}

		fmt.Println("Resposta final do agente:")
		fmt.Println(finalResp.Choices[0].Message.Content)
	} else {
		// Caso o agente decida não usar a ferramenta
		fmt.Println("Resposta do agente (sem busca):")
		fmt.Println(resp.Choices[0].Message.Content)
	}
}
