package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eduardoalcantara/cast/internal/config"
)

func TestGoogleChatProvider_Name(t *testing.T) {
	cfg := &config.GoogleChatConfig{
		WebhookURL: "https://chat.googleapis.com/v1/spaces/test/messages",
	}
	provider := NewGoogleChatProvider(cfg)
	if provider.Name() != "google_chat" {
		t.Errorf("Esperado 'google_chat', obtido '%s'", provider.Name())
	}
}

func TestGoogleChatProvider_Send_Success(t *testing.T) {
	// Cria servidor HTTP mock
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verifica método
		if r.Method != "POST" {
			t.Errorf("Esperado método POST, obtido %s", r.Method)
		}

		// Verifica Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Esperado Content-Type 'application/json', obtido '%s'", r.Header.Get("Content-Type"))
		}

		// Lê e verifica payload
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Erro ao decodificar payload: %v", err)
		}

		if payload["text"] != "Teste de mensagem" {
			t.Errorf("Esperado text 'Teste de mensagem', obtido '%v'", payload["text"])
		}

		// Retorna sucesso
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"text":"Teste de mensagem"}`))
	}))
	defer server.Close()

	// Cria provider com URL do servidor mock
	cfg := &config.GoogleChatConfig{
		WebhookURL: server.URL,
		Timeout:    30,
	}
	provider := NewGoogleChatProvider(cfg)

	// Envia mensagem
	err := provider.Send(server.URL, "Teste de mensagem")
	if err != nil {
		t.Errorf("Erro inesperado ao enviar mensagem: %v", err)
	}
}

func TestGoogleChatProvider_Send_ErrorResponse(t *testing.T) {
	// Cria servidor HTTP mock que retorna erro
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid webhook URL"}`))
	}))
	defer server.Close()

	cfg := &config.GoogleChatConfig{
		WebhookURL: server.URL,
		Timeout:    30,
	}
	provider := NewGoogleChatProvider(cfg)

	// Tenta enviar mensagem
	err := provider.Send(server.URL, "Teste")
	if err == nil {
		t.Error("Esperado erro, mas não ocorreu")
	}
}

func TestGoogleChatProvider_Send_DefaultWebhook(t *testing.T) {
	// Cria servidor HTTP mock
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"text":"Teste"}`))
	}))
	defer server.Close()

	cfg := &config.GoogleChatConfig{
		WebhookURL: server.URL,
		Timeout:    30,
	}
	provider := NewGoogleChatProvider(cfg)

	// Envia com "default" - deve usar WebhookURL configurado
	err := provider.Send("default", "Teste")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}
}

func TestGoogleChatProvider_Send_MultipleTargets(t *testing.T) {
	callCount := 0
	// Cria servidor HTTP mock
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"text":"Teste"}`))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"text":"Teste"}`))
	}))
	defer server2.Close()

	cfg := &config.GoogleChatConfig{
		WebhookURL: server1.URL,
		Timeout:    30,
	}
	provider := NewGoogleChatProvider(cfg)

	// Envia para múltiplos webhooks
	err := provider.Send(server1.URL+","+server2.URL, "Teste")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	// Verifica se foi chamado 2 vezes
	if callCount != 2 {
		t.Errorf("Esperado 2 chamadas, obtido %d", callCount)
	}
}

func TestGoogleChatProvider_Send_NoWebhookConfigured(t *testing.T) {
	cfg := &config.GoogleChatConfig{
		WebhookURL: "",
		Timeout:    30,
	}
	provider := NewGoogleChatProvider(cfg)

	// Tenta enviar sem webhook configurado e sem target
	err := provider.Send("", "Teste")
	if err == nil {
		t.Error("Esperado erro, mas não ocorreu")
	}

	if err != nil && !strings.Contains(err.Error(), "nenhum webhook especificado") {
		t.Errorf("Esperado erro sobre webhook não especificado, obtido: %v", err)
	}
}
