package domain

import (
	"context"
	"time"
)

// Document representa um documento no sistema RAG
type Document struct {
	ID        string            `json:"id" bson:"_id,omitempty"`
	Title     string            `json:"title" bson:"title"`
	Content   string            `json:"content" bson:"content"`
	Link      string            `json:"link" bson:"link"`
	Category  string            `json:"category" bson:"category"`
	Metadata  map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" bson:"updated_at"`
}

// NewDocument cria uma nova instância de Document com timestamps
func NewDocument(title, content, link, category string) *Document {
	now := time.Now()
	return &Document{
		Title:     title,
		Content:   content,
		Link:      link,
		Category:  category,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateContent atualiza o conteúdo do documento e o timestamp
func (d *Document) UpdateContent(content string) {
	d.Content = content
	d.UpdatedAt = time.Now()
}

// AddMetadata adiciona metadados ao documento
func (d *Document) AddMetadata(key, value string) {
	if d.Metadata == nil {
		d.Metadata = make(map[string]string)
	}
	d.Metadata[key] = value
	d.UpdatedAt = time.Now()
}

// DocumentRepository define as operações de persistência para documentos
type DocumentRepository interface {
	// Search busca documentos baseado em uma query de texto
	Search(ctx context.Context, query string, limit int) ([]*Document, error)

	// FindByID busca um documento pelo ID
	FindByID(ctx context.Context, id string) (*Document, error)

	// FindByCategory busca documentos por categoria
	FindByCategory(ctx context.Context, category string, limit int) ([]*Document, error)

	// Insert insere um novo documento
	Insert(ctx context.Context, doc *Document) error

	// Update atualiza um documento existente
	Update(ctx context.Context, doc *Document) error

	// Delete remove um documento pelo ID
	Delete(ctx context.Context, id string) error

	// DeleteAll remove todos os documentos (usado para limpeza)
	DeleteAll(ctx context.Context) error

	// SetupIndexes configura os índices necessários
	SetupIndexes(ctx context.Context) error

	// Count retorna o número total de documentos
	Count(ctx context.Context) (int64, error)

	// HealthCheck verifica se o repositório está funcionando
	HealthCheck(ctx context.Context) error
}
