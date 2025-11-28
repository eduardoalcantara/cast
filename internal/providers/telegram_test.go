package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eduardoalcantara/cast/internal/config"
)

func TestTelegramProvider_Name(t *testing.T) {
	cfg := &config.TelegramConfig{
		Token: "test-token",
	}
	provider := NewTelegramProvider(cfg, "")
	if provider.Name() != "telegram" {
		t.Errorf("Esperado 'telegram', obtido '%s'", provider.Name())
	}
}

func TestTelegramProvider_Send_Success(t *testing.T) {
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

		if payload["chat_id"] != "123456789" {
			t.Errorf("Esperado chat_id '123456789', obtido '%v'", payload["chat_id"])
		}

		if payload["text"] != "Teste de mensagem" {
			t.Errorf("Esperado text 'Teste de mensagem', obtido '%v'", payload["text"])
		}

		// Retorna sucesso
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	// Cria provider com URL do servidor mock
	cfg := &config.TelegramConfig{
		Token:   "test-token",
		APIURL:  server.URL + "/bot",
		Timeout: 30,
	}
	provider := NewTelegramProvider(cfg, "")

	// Envia mensagem
	err := provider.Send("123456789", "Teste de mensagem")
	if err != nil {
		t.Errorf("Erro inesperado ao enviar mensagem: %v", err)
	}
}

func TestTelegramProvider_Send_ErrorResponse(t *testing.T) {
	// Cria servidor HTTP mock que retorna erro
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok":false,"description":"Bad Request"}`))
	}))
	defer server.Close()

	cfg := &config.TelegramConfig{
		Token:   "test-token",
		APIURL:  server.URL + "/bot",
		Timeout: 30,
	}
	provider := NewTelegramProvider(cfg, "")

	// Tenta enviar mensagem
	err := provider.Send("123456789", "Teste")
	if err == nil {
		t.Error("Esperado erro, mas não ocorreu")
	}
}

func TestTelegramProvider_Send_MultipleTargets(t *testing.T) {
	callCount := 0
	// Cria servidor HTTP mock
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := &config.TelegramConfig{
		Token:   "test-token",
		APIURL:  server.URL + "/bot",
		Timeout: 30,
	}
	provider := NewTelegramProvider(cfg, "")

	// Envia para múltiplos targets
	err := provider.Send("123456789,987654321", "Teste")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	// Verifica se foi chamado 2 vezes
	if callCount != 2 {
		t.Errorf("Esperado 2 chamadas, obtido %d", callCount)
	}
}

func TestTelegramProvider_Send_DefaultChatID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		if payload["chat_id"] != "999888777" {
			t.Errorf("Esperado chat_id '999888777', obtido '%v'", payload["chat_id"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := &config.TelegramConfig{
		Token:        "test-token",
		APIURL:       server.URL + "/bot",
		DefaultChatID: "999888777",
		Timeout:      30,
	}
	provider := NewTelegramProvider(cfg, "")

	// Envia com "me" - deve usar DefaultChatID
	err := provider.Send("me", "Teste")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}
}
