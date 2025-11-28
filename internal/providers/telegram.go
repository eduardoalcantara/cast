package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eduardoalcantara/cast/internal/config"
)

// telegramProvider implementa o Provider para Telegram.
type telegramProvider struct {
	config      *config.TelegramConfig
	defaultTarget string
}

// NewTelegramProvider cria uma nova instância do TelegramProvider.
func NewTelegramProvider(cfg *config.TelegramConfig, defaultTarget string) Provider {
	return &telegramProvider{
		config:        cfg,
		defaultTarget: defaultTarget,
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
	apiURL := p.config.APIURL
	if apiURL == "" {
		apiURL = "https://api.telegram.org/bot"
	}
	url := fmt.Sprintf("%s%s/sendMessage", apiURL, p.config.Token)

	// Monta o payload JSON
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
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
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	// Valida status code
	if resp.StatusCode != http.StatusOK {
		// Lê corpo da resposta para debug
		var responseBody bytes.Buffer
		responseBody.ReadFrom(resp.Body)
		return fmt.Errorf("erro da API do Telegram (status %d): %s", resp.StatusCode, responseBody.String())
	}

	return nil
}
