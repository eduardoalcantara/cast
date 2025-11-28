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

// whatsappProvider implementa o Provider para WhatsApp (Meta Cloud API).
type whatsappProvider struct {
	config *config.WhatsAppConfig
}

// NewWhatsAppProvider cria uma nova instância do WhatsAppProvider.
func NewWhatsAppProvider(cfg *config.WhatsAppConfig) Provider {
	return &whatsappProvider{
		config: cfg,
	}
}

// Name retorna o nome do provider.
func (p *whatsappProvider) Name() string {
	return "whatsapp"
}

// Send envia uma mensagem via WhatsApp (Meta Cloud API).
// Suporta múltiplos targets separados por vírgula ou ponto-e-vírgula.
func (p *whatsappProvider) Send(target string, message string) error {
	// Parseia múltiplos targets
	targets := config.ParseTargets(target)

	if len(targets) == 0 {
		return fmt.Errorf("nenhum destinatário especificado")
	}

	// Processa cada target
	for i, t := range targets {
		if err := p.sendToPhone(t, message); err != nil {
			return fmt.Errorf("erro ao enviar para %s (target %d/%d): %w", t, i+1, len(targets), err)
		}
	}

	return nil
}

// sendToPhone envia mensagem para um número de telefone específico.
func (p *whatsappProvider) sendToPhone(phoneNumber string, message string) error {
	// Monta a URL da API
	apiURL := p.config.APIURL
	if apiURL == "" {
		apiURL = "https://graph.facebook.com"
	}

	apiVersion := p.config.APIVersion
	if apiVersion == "" {
		apiVersion = "v18.0"
	}

	url := fmt.Sprintf("%s/%s/%s/messages", apiURL, apiVersion, p.config.PhoneNumberID)

	// Monta o payload JSON
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                phoneNumber,
		"type":              "text",
		"text": map[string]string{
			"body": message,
		},
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

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.AccessToken))
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
		// Lê corpo da resposta para parsear erro do Facebook
		var responseBody bytes.Buffer
		responseBody.ReadFrom(resp.Body)

		// Tenta parsear erro do Facebook
		var fbError struct {
			Error struct {
				Message   string `json:"message"`
				Type      string `json:"type"`
				Code      int    `json:"code"`
				ErrorData struct {
					MessagingProduct string `json:"messaging_product"`
					Details          string `json:"details"`
				} `json:"error_data"`
			} `json:"error"`
		}

		errorMsg := responseBody.String()
		if err := json.Unmarshal(responseBody.Bytes(), &fbError); err == nil && fbError.Error.Message != "" {
			// Erro parseado do Facebook
			if fbError.Error.Code == 131047 {
				errorMsg = fmt.Sprintf("janela de conversa fechada (24h): %s. Envie uma mensagem de template primeiro ou aguarde o usuário iniciar uma conversa", fbError.Error.Message)
			} else {
				errorMsg = fmt.Sprintf("%s (código: %d, tipo: %s)", fbError.Error.Message, fbError.Error.Code, fbError.Error.Type)
			}
		}

		return fmt.Errorf("erro da API do WhatsApp (status %d): %s", resp.StatusCode, errorMsg)
	}

	return nil
}
