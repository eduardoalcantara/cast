package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eduardoalcantara/cast/internal/config"
)

// telegramProvider implementa o Provider para Telegram.
type telegramProvider struct {
	config        *config.TelegramConfig
	defaultTarget string
	verbose       bool
}

// NewTelegramProvider cria uma nova instância do TelegramProvider.
func NewTelegramProvider(cfg *config.TelegramConfig, defaultTarget string) Provider {
	return &telegramProvider{
		config:        cfg,
		defaultTarget: defaultTarget,
		verbose:       false,
	}
}

// NewTelegramProviderWithVerbose cria uma nova instância do TelegramProvider com modo verbose.
func NewTelegramProviderWithVerbose(cfg *config.TelegramConfig, defaultTarget string, verbose bool) Provider {
	return &telegramProvider{
		config:        cfg,
		defaultTarget: defaultTarget,
		verbose:       verbose,
	}
}

// Name retorna o nome do provider.
func (p *telegramProvider) Name() string {
	return "telegram"
}

// Send envia uma mensagem via Telegram.
// Suporta múltiplos targets separados por vírgula ou ponto-e-vírgula.
func (p *telegramProvider) Send(target string, message string) error {
	// Parseia múltiplos targets
	targets := config.ParseTargets(target)

	// Se não há targets, tenta usar "me" ou default
	if len(targets) == 0 {
		if p.defaultTarget != "" {
			targets = []string{p.defaultTarget}
		} else if p.config.DefaultChatID != "" {
			targets = []string{p.config.DefaultChatID}
		} else {
			return fmt.Errorf("target 'me' requer default_chat_id configurado ou alias com target")
		}
	}

	// Processa cada target
	for i, t := range targets {
		// Se target for "me" ou vazio, usa DefaultChatID
		chatID := t
		if t == "me" || t == "" {
			if p.defaultTarget != "" {
				chatID = p.defaultTarget
			} else if p.config.DefaultChatID != "" {
				chatID = p.config.DefaultChatID
			} else {
				return fmt.Errorf("target 'me' requer default_chat_id configurado ou alias com target")
			}
		}

		// Envia para este chat ID
		if err := p.sendToChatID(chatID, message); err != nil {
			return fmt.Errorf("erro ao enviar para chat_id %s (target %d/%d): %w", chatID, i+1, len(targets), err)
		}
	}

	return nil
}

// sendToChatID envia mensagem para um chat ID específico.
func (p *telegramProvider) sendToChatID(chatID string, message string) error {
	// Monta a URL da API
	// Formato correto: https://api.telegram.org/bot<TOKEN>/sendMessage
	apiURL := p.config.APIURL
	if apiURL == "" {
		apiURL = "https://api.telegram.org/bot"
	}
	// Garante que apiURL termina com "bot" (sem barra)
	if !strings.HasSuffix(apiURL, "bot") {
		apiURL = strings.TrimSuffix(apiURL, "/")
		if !strings.HasSuffix(apiURL, "bot") {
			apiURL = apiURL + "/bot"
		}
	}
	// Constrói URL: apiURL + token + /sendMessage
	url := fmt.Sprintf("%s%s/sendMessage", apiURL, p.config.Token)

	// Monta o payload JSON
	// Telegram API aceita chat_id como string ou número
	// Mas vamos tentar converter para número se possível para evitar problemas
	var chatIDValue interface{} = chatID
	if chatIDNum, err := strconv.ParseInt(chatID, 10, 64); err == nil {
		chatIDValue = chatIDNum
	}

	payload := map[string]interface{}{
		"chat_id": chatIDValue,
		"text":    message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	// Debug: mostra informações detalhadas se verbose estiver ativo
	if p.verbose {
		p.showDebugInfo(url, chatID, chatIDValue, string(jsonData))
	}

	// Cria requisição HTTP
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Configura timeout
	timeout := time.Duration(p.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	// Executa requisição
	if p.verbose {
		// Mascara token na URL para exibição
		maskedURL := strings.Replace(url, p.config.Token, maskToken(p.config.Token), 1)
		fmt.Fprintf(os.Stderr, "[DEBUG] Enviando requisição HTTP POST para: %s\n", maskedURL)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	// Debug: mostra resposta
	if p.verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Status Code: %d\n", resp.StatusCode)
	}

	// Valida status code
	if resp.StatusCode != http.StatusOK {
		// Lê corpo da resposta para parsear erro
		var responseBody bytes.Buffer
		responseBody.ReadFrom(resp.Body)
		bodyStr := responseBody.String()

		// Tenta parsear resposta JSON para obter descrição do erro
		var apiResponse struct {
			OK          bool   `json:"ok"`
			ErrorCode   int    `json:"error_code"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal([]byte(bodyStr), &apiResponse); err == nil && apiResponse.Description != "" {
			// Mensagem de erro mais amigável baseada no código de erro
			var userFriendlyMsg string
			switch apiResponse.ErrorCode {
			case 400:
				userFriendlyMsg = "Requisição inválida. Verifique o formato do chat_id."
			case 403:
				userFriendlyMsg = "Bot bloqueado ou sem permissão. O usuário precisa iniciar conversa com o bot primeiro."
			case 404:
				userFriendlyMsg = "Chat não encontrado. Possíveis causas:\n" +
					"  - O chat_id está incorreto\n" +
					"  - O bot não tem permissão para enviar mensagens para este chat\n" +
					"  - O usuário precisa enviar uma mensagem para o bot primeiro (não apenas iniciar a conversa)\n" +
					"  - O token do bot pode estar incorreto ou expirado"
			case 429:
				userFriendlyMsg = "Muitas requisições. Aguarde alguns segundos antes de tentar novamente."
			default:
				userFriendlyMsg = apiResponse.Description
			}

			// Sugestões baseadas no erro
			var suggestions string
			if apiResponse.ErrorCode == 404 {
				suggestions = "\n\nDica: No Telegram Bot API, você não pode usar números de telefone como chat_id.\n" +
					"Para obter o chat_id correto:\n" +
					"  1. O usuário deve iniciar uma conversa com o bot primeiro\n" +
					"  2. Use um bot como @userinfobot para descobrir seu chat_id\n" +
					"  3. Ou use 'me' se você configurou default_chat_id na configuração\n" +
					"  4. Exemplo: cast send tg me \"Mensagem\""
			}

			return fmt.Errorf("erro da API do Telegram (status %d, código %d): %s%s", resp.StatusCode, apiResponse.ErrorCode, userFriendlyMsg, suggestions)
		}

		// Se não conseguiu parsear, retorna erro genérico
		return fmt.Errorf("erro da API do Telegram (status %d): %s", resp.StatusCode, bodyStr)
	}

	return nil
}

// showDebugInfo exibe informações detalhadas de debug.
func (p *telegramProvider) showDebugInfo(url, chatID string, chatIDValue interface{}, jsonPayload string) {
	fmt.Fprintf(os.Stderr, "\n[DEBUG] === Telegram Provider Debug ===\n")

	// Mascara token na URL para exibição
	maskedURL := strings.Replace(url, p.config.Token, maskToken(p.config.Token), 1)
	fmt.Fprintf(os.Stderr, "[DEBUG] URL completa: %s\n", maskedURL)

	// Mostra componentes da URL separadamente
	apiURL := p.config.APIURL
	if apiURL == "" {
		apiURL = "https://api.telegram.org/bot"
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] API URL base: %s\n", apiURL)
	fmt.Fprintf(os.Stderr, "[DEBUG] Token: %s\n", maskToken(p.config.Token))
	fmt.Fprintf(os.Stderr, "[DEBUG] Endpoint: /sendMessage\n")
	fmt.Fprintf(os.Stderr, "[DEBUG] Chat ID (string): %s\n", chatID)
	fmt.Fprintf(os.Stderr, "[DEBUG] Chat ID (valor no JSON): %v (tipo: %T)\n", chatIDValue, chatIDValue)
	fmt.Fprintf(os.Stderr, "[DEBUG] Payload JSON:\n%s\n", jsonPayload)
	fmt.Fprintf(os.Stderr, "[DEBUG] Timeout: %d segundos\n", p.config.Timeout)
	fmt.Fprintf(os.Stderr, "[DEBUG] ================================\n\n")
}

// maskToken mascara um token mostrando apenas os primeiros e últimos caracteres.
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
