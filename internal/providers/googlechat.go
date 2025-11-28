package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eduardoalcantara/cast/internal/config"
)

// googleChatProvider implementa o Provider para Google Chat (Incoming Webhooks).
type googleChatProvider struct {
	config *config.GoogleChatConfig
}

// NewGoogleChatProvider cria uma nova instância do GoogleChatProvider.
func NewGoogleChatProvider(cfg *config.GoogleChatConfig) Provider {
	return &googleChatProvider{
		config: cfg,
	}
}

// Name retorna o nome do provider.
func (p *googleChatProvider) Name() string {
	return "google_chat"
}

// Send envia uma mensagem via Google Chat (Incoming Webhook).
// Lógica de Target:
// - Se target for uma URL completa (começa com https://), usa essa URL
// - Se target for "default" ou vazio, usa a URL configurada no cast.yaml
// - Suporta múltiplos webhooks separados por vírgula ou ponto-e-vírgula
func (p *googleChatProvider) Send(target string, message string) error {
	// Parseia múltiplos targets
	targets := config.ParseTargets(target)

	// Se não há targets, tenta usar a URL configurada
	if len(targets) == 0 {
		if p.config.WebhookURL == "" {
			return fmt.Errorf("nenhum webhook especificado e nenhum webhook_url configurado")
		}
		targets = []string{p.config.WebhookURL}
	}

	// Processa cada target
	for i, t := range targets {
		webhookURL := t

		// Se target for "default" ou vazio, usa a URL configurada
		if t == "default" || t == "" {
			if p.config.WebhookURL == "" {
				return fmt.Errorf("target 'default' requer webhook_url configurado")
			}
			webhookURL = p.config.WebhookURL
		} else if !strings.HasPrefix(t, "https://") {
			// Se não começa com https://, assume que é a URL configurada
			if p.config.WebhookURL != "" {
				webhookURL = p.config.WebhookURL
			} else {
				return fmt.Errorf("webhook inválido: %s (deve começar com https:// ou usar 'default')", t)
			}
		}

		// Valida se é uma URL do Google Chat
		if !strings.Contains(webhookURL, "chat.googleapis.com") && !strings.Contains(webhookURL, "googleapis.com") {
			// Permite URLs customizadas, mas avisa se não parece ser do Google
			// Não bloqueia, pois pode ser um webhook proxy
		}

		if err := p.sendToWebhook(webhookURL, message); err != nil {
			return fmt.Errorf("erro ao enviar para webhook %s (target %d/%d): %w", webhookURL, i+1, len(targets), err)
		}
	}

	return nil
}

// sendToWebhook envia mensagem para um webhook específico.
func (p *googleChatProvider) sendToWebhook(webhookURL string, message string) error {
	// Monta o payload JSON
	payload := map[string]string{
		"text": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	// Cria requisição HTTP
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		webhookURL,
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
		var responseBody bytes.Buffer
		responseBody.ReadFrom(resp.Body)
		return fmt.Errorf("erro do webhook do Google Chat (status %d): %s", resp.StatusCode, responseBody.String())
	}

	return nil
}
