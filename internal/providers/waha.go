package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eduardoalcantara/cast/internal/config"
)

// wahaProvider implementa o Provider para WAHA (WhatsApp HTTP API).
type wahaProvider struct {
	apiURL  string        // Base URL do WAHA (ex: http://localhost:3000)
	session string        // Nome da sessão (default: "default")
	apiKey  string        // API Key opcional para auth
	timeout time.Duration // Timeout HTTP
	client  *http.Client // Cliente HTTP reutilizável
}

// NewWAHAProvider cria instância do provider WAHA com validações completas.
func NewWAHAProvider(cfg config.WAHAConfig) (Provider, error) {
	// Validação 1: API URL obrigatória
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("WAHA API URL não configurada. Use: cast gateway add waha --interactive")
	}

	// Validação 2: Formato da URL
	apiURL := strings.TrimRight(cfg.APIURL, "/")
	if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
		return nil, fmt.Errorf("WAHA API URL inválida: deve começar com http:// ou https://")
	}

	// Validação 3: Timeout mínimo e default
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if timeout < 5*time.Second {
		return nil, fmt.Errorf("timeout muito baixo: mínimo 5 segundos")
	}

	// Validação 4: Session default
	session := strings.TrimSpace(cfg.Session)
	if session == "" {
		session = "default"
	}

	// Cliente HTTP reutilizável (performance)
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
	}

	return &wahaProvider{
		apiURL:  apiURL,
		session: session,
		apiKey:  cfg.APIKey,
		timeout: timeout,
		client:  client,
	}, nil
}

// Name retorna o nome do provider.
func (w *wahaProvider) Name() string {
	return "WAHA"
}

// Send envia uma mensagem via WAHA (WhatsApp HTTP API).
func (w *wahaProvider) Send(target string, message string) error {
	// Parseia múltiplos targets
	targets := config.ParseTargets(target)

	if len(targets) == 0 {
		return fmt.Errorf("nenhum destinatário especificado")
	}

	// Processa cada target
	for i, t := range targets {
		if err := w.sendToChatID(t, message); err != nil {
			return fmt.Errorf("erro ao enviar para %s (target %d/%d): %w", t, i+1, len(targets), err)
		}
	}

	return nil
}

// sendToChatID envia mensagem para um chat ID específico.
func (w *wahaProvider) sendToChatID(chatID string, message string) error {
	// Validação 1: Target obrigatório
	if strings.TrimSpace(chatID) == "" {
		return fmt.Errorf("target vazio: forneça chatId no formato 5511999998888@c.us")
	}

	// Validação 2: Formato do chatId
	if err := w.validateChatID(chatID); err != nil {
		return err
	}

	// Validação 3: Mensagem obrigatória
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("mensagem vazia")
	}

	// Construir payload
	payload := map[string]interface{}{
		"session": w.session,
		"chatId":  chatID,
		"text":    message,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	// Construir request
	url := fmt.Sprintf("%s/api/sendText", w.apiURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("erro ao criar request: %w", err)
	}

	// Headers
	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("X-Api-Key", w.apiKey)
	}

	// Executar request
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao conectar com WAHA: %w. Verifique se está rodando em %s", err, w.apiURL)
	}
	defer resp.Body.Close()

	// Ler body da resposta
	respBody, _ := io.ReadAll(resp.Body)

	// Verificar status HTTP
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return w.handleErrorResponse(resp.StatusCode, respBody)
	}

	return nil
}

// validateChatID valida formato do Chat ID do WhatsApp.
func (w *wahaProvider) validateChatID(chatID string) error {
	chatID = strings.TrimSpace(chatID)

	// Validação básica de formato
	if !strings.Contains(chatID, "@") {
		return fmt.Errorf(
			"chatId inválido: '%s'. Formato esperado: 5511999998888@c.us (contato) ou 120363XXX@g.us (grupo)",
			chatID,
		)
	}

	// Valida sufixo
	if !strings.HasSuffix(chatID, "@c.us") && !strings.HasSuffix(chatID, "@g.us") {
		return fmt.Errorf(
			"chatId inválido: deve terminar com @c.us (contato) ou @g.us (grupo). Recebido: %s",
			chatID,
		)
	}

	// Valida prefixo numérico (contatos)
	if strings.HasSuffix(chatID, "@c.us") {
		prefix := strings.Split(chatID, "@")[0]
		if len(prefix) < 10 {
			return fmt.Errorf(
				"chatId muito curto: '%s'. Contatos devem ter código do país + DDD + número",
				chatID,
			)
		}
	}

	return nil
}

// handleErrorResponse processa erros HTTP do WAHA com mensagens amigáveis.
func (w *wahaProvider) handleErrorResponse(statusCode int, body []byte) error {
	// Tentar parsear resposta JSON
	var errorResp map[string]interface{}
	if err := json.Unmarshal(body, &errorResp); err != nil {
		// Se não for JSON, retorna erro genérico
		return fmt.Errorf("WAHA retornou erro %d: %s", statusCode, string(body))
	}

	// Extrair mensagem de erro
	errorMsg := "erro desconhecido"
	if msg, ok := errorResp["error"].(string); ok {
		errorMsg = msg
	} else if msg, ok := errorResp["message"].(string); ok {
		errorMsg = msg
	}

	// Mensagens específicas por código
	switch statusCode {
	case 400:
		return fmt.Errorf("requisição inválida: %s", errorMsg)

	case 401:
		return fmt.Errorf("autenticação falhou: API Key incorreta ou ausente")

	case 404:
		if strings.Contains(strings.ToLower(errorMsg), "session") {
			return fmt.Errorf(
				"sessão '%s' não encontrada. Crie com: curl -X POST %s/api/sessions/start -d '{\"name\":\"%s\"}'",
				w.session,
				w.apiURL,
				w.session,
			)
		}
		return fmt.Errorf("endpoint não encontrado: verifique se WAHA está atualizado")

	case 500:
		if strings.Contains(strings.ToLower(errorMsg), "not connected") ||
			strings.Contains(strings.ToLower(errorMsg), "not authenticated") {
			return fmt.Errorf(
				"sessão '%s' não conectada. Escaneie o QR code em: %s",
				w.session,
				w.apiURL,
			)
		}
		return fmt.Errorf("erro interno do WAHA: %s", errorMsg)

	default:
		return fmt.Errorf("WAHA retornou erro %d: %s", statusCode, errorMsg)
	}
}
