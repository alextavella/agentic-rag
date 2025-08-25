package domain

import (
	"errors"
	"fmt"
)

// Erros específicos do domínio
var (
	// Document errors
	ErrDocumentNotFound = errors.New("documento não encontrado")
	ErrDocumentInvalid  = errors.New("documento inválido")
	ErrDocumentExists   = errors.New("documento já existe")

	// Query errors
	ErrQueryEmpty    = errors.New("query não pode estar vazia")
	ErrQueryTooShort = errors.New("query muito curta")
	ErrQueryTooLong  = errors.New("query muito longa")

	// Repository errors
	ErrRepositoryUnavailable = errors.New("repositório indisponível")
	ErrRepositoryTimeout     = errors.New("timeout no repositório")

	// LLM errors
	ErrLLMUnavailable     = errors.New("modelo de linguagem indisponível")
	ErrLLMTimeout         = errors.New("timeout no modelo de linguagem")
	ErrLLMQuotaExceeded   = errors.New("cota do modelo excedida")
	ErrLLMInvalidResponse = errors.New("resposta inválida do modelo")

	// Configuration errors
	ErrConfigInvalid = errors.New("configuração inválida")
	ErrConfigMissing = errors.New("configuração obrigatória ausente")

	// General errors
	ErrInternalServer     = errors.New("erro interno do servidor")
	ErrServiceUnavailable = errors.New("serviço indisponível")
)

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("campo '%s': %s", e.Field, e.Message)
}

// NewValidationError cria um novo erro de validação
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// BusinessError representa um erro de regra de negócio
type BusinessError struct {
	Code    string
	Message string
	Cause   error
}

func (e BusinessError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e BusinessError) Unwrap() error {
	return e.Cause
}

// NewBusinessError cria um novo erro de negócio
func NewBusinessError(code, message string, cause error) BusinessError {
	return BusinessError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// IsNotFoundError verifica se o erro é de "não encontrado"
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrDocumentNotFound)
}

// IsValidationError verifica se o erro é de validação
func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}

// IsBusinessError verifica se o erro é de negócio
func IsBusinessError(err error) bool {
	var businessErr BusinessError
	return errors.As(err, &businessErr)
}

// IsTemporaryError verifica se o erro é temporário (pode ser tentado novamente)
func IsTemporaryError(err error) bool {
	return errors.Is(err, ErrRepositoryTimeout) ||
		errors.Is(err, ErrLLMTimeout) ||
		errors.Is(err, ErrServiceUnavailable)
}
