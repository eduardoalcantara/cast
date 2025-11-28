package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eduardoalcantara/cast/internal/config"
)

func TestWhatsAppProvider_Name(t *testing.T) {
	cfg := &config.WhatsAppConfig{
		PhoneNumberID: "123456789012345",
		AccessToken:   "test-token",
	}
	provider := NewWhatsAppProvider(cfg)
	if provider.Name() != "whatsapp" {
		t.Errorf("Esperado 'whatsapp', obtido '%s'", provider.Name())
	}
}

func TestWhatsAppProvider_Send_Success(t *testing.T) {
	// Cria servidor HTTP mock
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verifica método
		if r.Method != "POST" {
			t.Errorf("Esperado método POST, obtido %s", r.Method)
		}

		// Verifica Authorization header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Esperado Authorization 'Bearer test-token', obtido '%s'", r.Header.Get("Authorization"))
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

		if payload["messaging_product"] != "whatsapp" {
			t.Errorf("Esperado messaging_product 'whatsapp', obtido '%v'", payload["messaging_product"])
		}

		if payload["to"] != "5511999998888" {
			t.Errorf("Esperado to '5511999998888', obtido '%v'", payload["to"])
		}

		if payload["type"] != "text" {
			t.Errorf("Esperado type 'text', obtido '%v'", payload["type"])
		}

		textObj, ok := payload["text"].(map[string]interface{})
		if !ok {
			t.Fatalf("Esperado text como objeto, obtido %T", payload["text"])
		}

		if textObj["body"] != "Teste de mensagem" {
			t.Errorf("Esperado body 'Teste de mensagem', obtido '%v'", textObj["body"])
		}

		// Retorna sucesso
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"messaging_product":"whatsapp"}`))
	}))
	defer server.Close()

	// Cria provider com URL do servidor mock
	cfg := &config.WhatsAppConfig{
		PhoneNumberID: "123456789012345",
		AccessToken:   "test-token",
		APIURL:        server.URL,
		APIVersion:    "v18.0",
		Timeout:       30,
	}
	provider := NewWhatsAppProvider(cfg)

	// Envia mensagem
	err := provider.Send("5511999998888", "Teste de mensagem")
	if err != nil {
		t.Errorf("Erro inesperado ao enviar mensagem: %v", err)
	}
}

func TestWhatsAppProvider_Send_ErrorResponse(t *testing.T) {
	// Cria servidor HTTP mock que retorna erro do Facebook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": {
				"message": "Invalid OAuth access token.",
				"type": "OAuthException",
				"code": 190
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.WhatsAppConfig{
		PhoneNumberID: "123456789012345",
		AccessToken:   "invalid-token",
		APIURL:        server.URL,
		APIVersion:    "v18.0",
		Timeout:       30,
	}
	provider := NewWhatsAppProvider(cfg)

	// Tenta enviar mensagem
	err := provider.Send("5511999998888", "Teste")
	if err == nil {
		t.Error("Esperado erro, mas não ocorreu")
	}

	// Verifica se a mensagem de erro contém informações do Facebook
	if err != nil && !strings.Contains(err.Error(), "Invalid OAuth access token") {
		t.Errorf("Esperado erro com mensagem do Facebook, obtido: %v", err)
	}
}

func TestWhatsAppProvider_Send_WindowClosedError(t *testing.T) {
	// Cria servidor HTTP mock que retorna erro de janela fechada (código 131047)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": {
				"message": "Message failed to send because more than 24 hours have passed since the customer last replied to this number.",
				"type": "OAuthException",
				"code": 131047
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.WhatsAppConfig{
		PhoneNumberID: "123456789012345",
		AccessToken:   "test-token",
		APIURL:        server.URL,
		APIVersion:    "v18.0",
		Timeout:       30,
	}
	provider := NewWhatsAppProvider(cfg)

	// Tenta enviar mensagem
	err := provider.Send("5511999998888", "Teste")
	if err == nil {
		t.Error("Esperado erro, mas não ocorreu")
	}

	// Verifica se a mensagem de erro menciona janela de 24h
	if err != nil && !strings.Contains(err.Error(), "janela de conversa fechada") {
		t.Errorf("Esperado erro sobre janela de 24h, obtido: %v", err)
	}
}

func TestWhatsAppProvider_Send_MultipleTargets(t *testing.T) {
	callCount := 0
	// Cria servidor HTTP mock
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"messaging_product":"whatsapp"}`))
	}))
	defer server.Close()

	cfg := &config.WhatsAppConfig{
		PhoneNumberID: "123456789012345",
		AccessToken:   "test-token",
		APIURL:        server.URL,
		APIVersion:    "v18.0",
		Timeout:       30,
	}
	provider := NewWhatsAppProvider(cfg)

	// Envia para múltiplos targets
	err := provider.Send("5511999998888,5511888777666", "Teste")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	// Verifica se foi chamado 2 vezes
	if callCount != 2 {
		t.Errorf("Esperado 2 chamadas, obtido %d", callCount)
	}
}
