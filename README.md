# Agentic RAG com MongoDB

Este projeto implementa um sistema RAG (Retrieval-Augmented Generation) usando Go, OpenAI GPT-4 e MongoDB. O sistema permite que um agente de IA busque informações relevantes em uma base de dados antes de responder às perguntas do usuário.

## 🚀 Tecnologias

- Go 1.21+
- MongoDB
- OpenAI GPT-4
- Docker & Docker Compose

## 📁 Estrutura do Projeto

```
.
├── cmd/
│   ├── api/
        |__ main.go    # Aplicação principal
│   └── seed/
│       └── main.go    # Script para popular o banco
├── internal/
│   └── database/
│       └── mongodb.go # Pacote de acesso ao MongoDB
├── docker-compose.yml # Configuração do MongoDB
└── go.mod            # Dependências do Go
```

## 🛠️ Configuração

### 1. Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
OPENAI_API_KEY=sua-chave-da-openai
MONGO_URI=mongodb://admin:password123@localhost:27017
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

1. Execute a aplicação principal:

```bash
go run cmd/api/main.go
```

2. A aplicação irá:
   - Receber uma pergunta do usuário
   - O agente (GPT-4) decidirá se precisa buscar informações
   - Se necessário, consultará o MongoDB
   - Gerará uma resposta combinando seu conhecimento com os dados encontrados

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

## 🔍 Funcionalidades

1. **Busca Semântica**

   - Índice de texto no MongoDB para busca eficiente
   - Busca em títulos e conteúdo dos documentos
   - Limite configurável de resultados

2. **Integração com OpenAI**

   - Uso do modelo GPT-4 Turbo
   - Sistema de ferramentas (tools) para busca
   - Histórico de conversação mantido

3. **Persistência**
   - Armazenamento em MongoDB
   - Conexão segura com autenticação
   - Volume Docker para persistência dos dados

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
