# Agentic RAG com MongoDB

Este projeto implementa um sistema RAG (Retrieval-Augmented Generation) usando Go, OpenAI GPT-4 e MongoDB. O sistema permite que um agente de IA busque informações relevantes em uma base de dados antes de responder às perguntas do usuário.

## 🚀 Tecnologias

- Go 1.21+
- MongoDB
- OpenAI GPT-4
- Docker & Docker Compose

## 📁 Estrutura do Projeto

O projeto segue o **Standard Go Project Layout** e princípios de **Clean Architecture**:

```
.
├── cmd/                              # Aplicações executáveis
│   ├── app/
│   │   └── main.go                  # Aplicação principal RAG
│   └── seed/
│       └── main.go                  # Script para popular o banco
├── internal/                        # Código privado da aplicação
│   ├── config/                      # Gerenciamento de configurações
│   │   └── config.go
│   ├── domain/                      # Entidades e interfaces de negócio
│   │   ├── document.go              # Entidade Document
│   │   ├── rag.go                   # Interfaces RAG e LLM
│   │   └── errors.go                # Erros específicos do domínio
│   ├── infrastructure/              # Implementações de infraestrutura
│   │   ├── database/
│   │   │   └── mongodb.go           # Repositório MongoDB
│   │   └── llm/
│   │       └── openai.go            # Cliente OpenAI
│   └── service/                     # Lógica de negócio
│       └── rag.go                   # Serviço RAG principal
├── config.example                   # Exemplo de configuração
├── docker-compose.yml              # Configuração do MongoDB
└── go.mod                          # Dependências do Go
```

## 🛠️ Configuração

### 1. Variáveis de Ambiente

Copie o arquivo de exemplo e configure as variáveis:

```bash
cp .env.example .env
```

Configure as seguintes variáveis no arquivo `.env`:

```env
# OpenAI
OPENAI_API_KEY=sua-chave-da-openai
OPENAI_MODEL=gpt-4-turbo-preview

# MongoDB
MONGO_URI=mongodb://admin:password123@localhost:27017
MONGO_DATABASE=rag_docs
MONGO_COLLECTION=documents

# Aplicação
LOG_LEVEL=info
REQUEST_TIMEOUT=30s
SEARCH_LIMIT=5
DEFAULT_QUERY=What are the documents related to Golang performance?
```

### 2. Instalação

1. Clone o repositório:

```bash
git clone https://github.com/alextavella/agentic-rag.git
cd agentic-rag
```

2. Instale as dependências:

```bash
go mod tidy
```

3. Inicie o MongoDB:

```bash
docker-compose up -d
```

4. Verifique se os containers estão rodando:

```bash
docker-compose ps
```

### 3. População do Banco

Execute o script de seed para inserir documentos de exemplo:

```bash
go run cmd/seed/main.go
```

## 💻 Uso

### Executar a aplicação principal

```bash
# Com a query padrão
go run cmd/app/main.go

# Com uma query personalizada
go run cmd/app/main.go "Como otimizar goroutines em Go?"
go run cmd/app/main.go "Explique sobre garbage collector em Go"

# Usando binário compilado
go build -o app cmd/app/main.go
./app "Explique sobre garbage collector em Go"
```

### Como funciona

1. **Configuração**: A aplicação carrega configurações do ambiente
2. **Inicialização**: Conecta ao MongoDB e inicializa cliente OpenAI
3. **Processamento**:
   - Recebe a query do usuário
   - O agente (GPT-4) decide se precisa buscar informações
   - Se necessário, executa busca semântica no MongoDB
   - Gera resposta combinando conhecimento base com dados encontrados
4. **Resposta**: Exibe a resposta final, fontes consultadas e estatísticas

### Exemplo de Saída

```
=== RESPOSTA DO AGENTE ===
Para otimizar goroutines em Go, você deve seguir algumas práticas...

=== FONTES CONSULTADAS ===
1. Optimizing Go Routines
   Link: /docs/go-optimizing
   Categoria: performance
   Conteúdo: Goroutines são leves e eficientes...

=== ESTATÍSTICAS ===
Tempo de processamento: 1250ms
Busca realizada: true
Modelo usado: gpt-4-turbo-preview
Tokens utilizados: 1847
```

## 📊 MongoDB Express

Uma interface web para gerenciar o MongoDB está disponível em:

- URL: http://localhost:8081
- Usuário: admin
- Senha: password123

## 🗃️ Estrutura dos Documentos

Os documentos no MongoDB seguem esta estrutura:

```go
type Document struct {
    Title    string `json:"title"`    // Título do documento
    Content  string `json:"content"`  // Conteúdo principal
    Link     string `json:"link"`     // Link/caminho do documento
    Category string `json:"category"` // Categoria (ex: "performance")
}
```

## 🏗️ Arquitetura

O projeto implementa **Clean Architecture** com separação clara de responsabilidades:

### Camadas

1. **Domain Layer** (`internal/domain/`):

   - Entidades de negócio (`Document`)
   - Interfaces (`DocumentRepository`, `LLMClient`, `RAGService`)
   - Regras de negócio e erros específicos

2. **Infrastructure Layer** (`internal/infrastructure/`):

   - Implementação MongoDB (`database/mongodb.go`)
   - Cliente OpenAI (`llm/openai.go`)
   - Detalhes de infraestrutura

3. **Service Layer** (`internal/service/`):

   - Lógica de aplicação (`rag.go`)
   - Orquestração entre componentes
   - Validações e transformações

4. **Application Layer** (`cmd/`):
   - Pontos de entrada (`main.go`)
   - Configuração de dependências
   - Interface com usuário

### Princípios Aplicados

- **Dependency Inversion**: Camadas superiores dependem de abstrações
- **Single Responsibility**: Cada componente tem uma responsabilidade clara
- **Interface Segregation**: Interfaces pequenas e focadas
- **Separation of Concerns**: Separação entre lógica de negócio e infraestrutura

## 🔍 Funcionalidades

1. **Busca Semântica Inteligente**

   - Índices otimizados no MongoDB
   - Busca full-text em títulos e conteúdo
   - Ranking por relevância
   - Limite configurável de resultados

2. **Integração Avançada com OpenAI**

   - Suporte a múltiplos modelos GPT
   - Sistema de ferramentas (function calling)
   - Controle de tokens e custos
   - Timeout e retry configuráveis

3. **Persistência Robusta**

   - Repositório MongoDB com pool de conexões
   - Índices automáticos para performance
   - Health checks e monitoramento
   - Transações e consistência de dados

4. **Observabilidade**
   - Logging estruturado (JSON)
   - Métricas de performance
   - Rastreamento de requests
   - Error tracking detalhado

## 🤝 Contribuindo

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## ✨ Próximos Passos

- [ ] Adicionar testes automatizados
- [ ] Implementar busca vetorial
- [ ] Adicionar mais opções de configuração
- [ ] Melhorar tratamento de erros
- [ ] Adicionar métricas e monitoramento
- [ ] Implementar cache de resultados
- [ ] Adicionar suporte a mais tipos de documentos
