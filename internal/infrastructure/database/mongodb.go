package database

import (
	"context"
	"fmt"
	"time"

	"github.com/alextavella/agentic-rag/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDocumentRepository implementa DocumentRepository usando MongoDB
type MongoDocumentRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// NewMongoDocumentRepository cria uma nova instância do repositório MongoDB
func NewMongoDocumentRepository(ctx context.Context, uri, database, collection string) (*MongoDocumentRepository, error) {
	// Configura as opções de conexão
	clientOptions := options.Client().ApplyURI(uri)

	// Conecta ao MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
	}

	// Verifica a conexão
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("erro ao pingar o MongoDB: %w", err)
	}

	db := client.Database(database)
	coll := db.Collection(collection)

	repo := &MongoDocumentRepository{
		client:     client,
		database:   db,
		collection: coll,
	}

	return repo, nil
}

// Search busca documentos baseado em uma query de texto
func (r *MongoDocumentRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Document, error) {
	if query == "" {
		return nil, domain.ErrQueryEmpty
	}

	// Cria um filtro de busca usando texto
	filter := bson.M{
		"$text": bson.M{
			"$search": query,
		},
	}

	// Configura as opções de busca
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	// Ordena por relevância (score do texto)
	findOptions.SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})
	findOptions.SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}})

	// Executa a busca
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar documentos: %w", err)
	}
	defer cursor.Close(ctx)

	// Decodifica os resultados
	var results []*domain.Document
	for cursor.Next(ctx) {
		var doc domain.Document
		if err := cursor.Decode(&doc); err != nil {
			continue // Skip documentos com erro de decodificação
		}
		results = append(results, &doc)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar resultados: %w", err)
	}

	return results, nil
}

// FindByID busca um documento pelo ID
func (r *MongoDocumentRepository) FindByID(ctx context.Context, id string) (*domain.Document, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrDocumentInvalid
	}

	filter := bson.M{"_id": objectID}

	var doc domain.Document
	err = r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("erro ao buscar documento: %w", err)
	}

	return &doc, nil
}

// FindByCategory busca documentos por categoria
func (r *MongoDocumentRepository) FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Document, error) {
	filter := bson.M{"category": category}

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_at": -1}) // Mais recentes primeiro

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar documentos por categoria: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*domain.Document
	for cursor.Next(ctx) {
		var doc domain.Document
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		results = append(results, &doc)
	}

	return results, nil
}

// Insert insere um novo documento
func (r *MongoDocumentRepository) Insert(ctx context.Context, doc *domain.Document) error {
	if doc == nil {
		return domain.ErrDocumentInvalid
	}

	// Define timestamps se não estiverem definidos
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("erro ao inserir documento: %w", err)
	}

	// Atualiza o ID do documento
	if objectID, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = objectID.Hex()
	}

	return nil
}

// Update atualiza um documento existente
func (r *MongoDocumentRepository) Update(ctx context.Context, doc *domain.Document) error {
	if doc == nil || doc.ID == "" {
		return domain.ErrDocumentInvalid
	}

	objectID, err := primitive.ObjectIDFromHex(doc.ID)
	if err != nil {
		return domain.ErrDocumentInvalid
	}

	// Atualiza o timestamp
	doc.UpdatedAt = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": doc}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar documento: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrDocumentNotFound
	}

	return nil
}

// Delete remove um documento pelo ID
func (r *MongoDocumentRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrDocumentInvalid
	}

	filter := bson.M{"_id": objectID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("erro ao deletar documento: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrDocumentNotFound
	}

	return nil
}

// DeleteAll remove todos os documentos
func (r *MongoDocumentRepository) DeleteAll(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("erro ao limpar coleção: %w", err)
	}
	return nil
}

// SetupIndexes configura os índices necessários
func (r *MongoDocumentRepository) SetupIndexes(ctx context.Context) error {
	// Índice de texto para busca
	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "content", Value: "text"},
		},
	}

	// Índice para categoria
	categoryIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "category", Value: 1},
		},
	}

	// Índice para timestamps
	timestampIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "created_at", Value: -1},
		},
	}

	indexes := []mongo.IndexModel{textIndex, categoryIndex, timestampIndex}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("erro ao criar índices: %w", err)
	}

	return nil
}

// Count retorna o número total de documentos
func (r *MongoDocumentRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("erro ao contar documentos: %w", err)
	}
	return count, nil
}

// HealthCheck verifica se o repositório está funcionando
func (r *MongoDocumentRepository) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}

// Close fecha a conexão com o MongoDB
func (r *MongoDocumentRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
