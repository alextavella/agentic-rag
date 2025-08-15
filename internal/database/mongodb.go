package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Document representa um documento armazenado no MongoDB
type Document struct {
	Title    string `bson:"title" json:"title"`
	Content  string `bson:"content" json:"content"`
	Link     string `bson:"link" json:"link"`
	Category string `bson:"category" json:"category"`
}

// MongoDB encapsula a conexão e operações com o MongoDB
type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// NewMongoDB cria uma nova instância de conexão com o MongoDB
func NewMongoDB(ctx context.Context, uri string) (*MongoDB, error) {
	// Configura as opções de conexão
	clientOptions := options.Client().ApplyURI(uri)

	// Conecta ao MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %v", err)
	}

	// Verifica a conexão
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao pingar o MongoDB: %v", err)
	}

	// Seleciona o banco de dados e coleção
	database := client.Database("rag_docs")
	collection := database.Collection("documents")

	return &MongoDB{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

// Close fecha a conexão com o MongoDB
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// ClearCollection limpa a coleção de documentos
func (m *MongoDB) ClearCollection(ctx context.Context) error {
	_, err := m.collection.DeleteMany(ctx, bson.M{})
	return err
}

// SearchDocuments busca documentos baseado em uma query
func (m *MongoDB) SearchDocuments(ctx context.Context, query string) (string, error) {
	// Cria um filtro de busca usando texto
	filter := bson.M{
		"$text": bson.M{
			"$search": query,
		},
	}

	// Configura as opções de busca
	findOptions := options.Find()
	findOptions.SetLimit(5) // Limita a 5 resultados

	// Executa a busca
	cursor, err := m.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return "[]", fmt.Errorf("erro ao buscar documentos: %v", err)
	}
	defer cursor.Close(ctx)

	// Decodifica os resultados
	var results []Document
	if err = cursor.All(ctx, &results); err != nil {
		return "[]", fmt.Errorf("erro ao decodificar resultados: %v", err)
	}

	// Converte os resultados para JSON
	jsonResults, err := json.Marshal(results)
	if err != nil {
		return "[]", fmt.Errorf("erro ao converter para JSON: %v", err)
	}

	return string(jsonResults), nil
}

// InsertDocument insere um novo documento no MongoDB
func (m *MongoDB) InsertDocument(ctx context.Context, doc Document) error {
	_, err := m.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("erro ao inserir documento: %v", err)
	}
	return nil
}

// SetupTextIndex configura um índice de texto para busca
func (m *MongoDB) SetupTextIndex(ctx context.Context) error {
	// Cria um índice de texto nos campos title e content
	model := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "content", Value: "text"},
		},
	}

	_, err := m.collection.Indexes().CreateOne(ctx, model)
	if err != nil {
		return fmt.Errorf("erro ao criar índice de texto: %v", err)
	}

	log.Println("Índice de texto criado com sucesso")
	return nil
}
