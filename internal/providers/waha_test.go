package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eduardoalcantara/cast/internal/config"
)

// TestWAHAProvider_NewProvider testa criação do provider com validações
func TestWAHAProvider_NewProvider(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.WAHAConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "configuração válida completa",
			cfg: config.WAHAConfig{
				APIURL:  "http://localhost:3000",
				Session: "test",
				APIKey:  "secret123",
				Timeout: 30,
			},
			expectError: false,
		},
		{
			name: "URL obrigatória",
			cfg: config.WAHAConfig{
				Session: "test",
			},
			expectError: true,
			errorMsg:    "não configurada",
		},
		{
			name: "URL inválida sem protocolo",
			cfg: config.WAHAConfig{
				APIURL: "localhost:3000",
			},
			expectError: true,
			errorMsg:    "http://",
		},
		{
			name: "session default aplicado",
			cfg: config.WAHAConfig{
				APIURL:  "http://localhost:3000",
				Session: "",
				Timeout: 30,
			},
			expectError: false,
		},
		{
			name: "timeout default aplicado",
			cfg: config.WAHAConfig{
				APIURL:  "http://localhost:3000",
				Timeout: 0,
			},
			expectError: false,
		},
		{
			name: "timeout muito baixo",
			cfg: config.WAHAConfig{
				APIURL:  "http://localhost:3000",
				Timeout: 3,
			},
			expectError: true,
			errorMsg:    "muito baixo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewWAHAProvider(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Errorf("Esperado erro, mas não ocorreu")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Erro não contém '%s': %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Erro inesperado: %v", err)
				}
				if provider == nil {
					t.Error("Provider é nil")
				}
			}
		})
	}
}

// TestWAHAProvider_Send_Success testa envio bem-sucedido
func TestWAHAProvider_Send_Success(t *testing.T) {
	// Mock server que simula WAHA
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Valida endpoint
		if r.URL.Path != "/api/sendText" {
			t.Errorf("Endpoint incorreto: %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}

		// Valida método
		if r.Method != "POST" {
			t.Errorf("Método incorreto: %s", r.Method)
			w.WriteHeader(405)
			return
		}

		// Valida Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type incorreto")
		}

		// Parse payload
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Erro ao parsear payload: %v", err)
			w.WriteHeader(400)
			return
		}

		// Valida campos obrigatórios
		if payload["session"] != "test-session" {
			t.Errorf("Session incorreta: %v", payload["session"])
		}
		if payload["chatId"] != "5511999998888@c.us" {
			t.Errorf("ChatId incorreto: %v", payload["chatId"])
		}
		if payload["text"] != "Mensagem de teste" {
			t.Errorf("Text incorreto: %v", payload["text"])
		}

		// Resposta de sucesso
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{
			"id":     "msg-123",
			"status": "sent",
		})
	}))
	defer server.Close()

	// Criar provider com mock URL
	cfg := config.WAHAConfig{
		APIURL:  server.URL,
		Session: "test-session",
		Timeout: 5,
	}

	provider, err := NewWAHAProvider(cfg)
	if err != nil {
		t.Fatalf("Erro ao criar provider: %v", err)
	}

	// Testar envio
	err = provider.Send("5511999998888@c.us", "Mensagem de teste")
	if err != nil {
		t.Errorf("Send falhou: %v", err)
	}
}

// TestWAHAProvider_Send_InvalidChatID testa validação de Chat ID
func TestWAHAProvider_Send_InvalidChatID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request não deveria ter sido enviada")
	}))
	defer server.Close()

	cfg := config.WAHAConfig{
		APIURL:  server.URL,
		Session: "test",
		Timeout: 5,
	}

	provider, _ := NewWAHAProvider(cfg)

	tests := []struct {
		name   string
		chatID string
	}{
		{"sem arroba", "5511999998888"},
		{"sufixo inválido", "5511999998888@invalid"},
		{"vazio", ""},
		{"só espaços", "   "},
		{"muito curto", "123@c.us"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.Send(tt.chatID, "Teste")
			if err == nil {
				t.Errorf("Esperado erro para chatId '%s', mas não ocorreu", tt.chatID)
			}
		})
	}
}

// TestWAHAProvider_Send_SessionNotConnected testa erro de sessão
func TestWAHAProvider_Send_SessionNotConnected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Session is not connected",
		})
	}))
	defer server.Close()

	cfg := config.WAHAConfig{
		APIURL:  server.URL,
		Session: "disconnected",
		Timeout: 5,
	}

	provider, _ := NewWAHAProvider(cfg)
	err := provider.Send("5511999998888@c.us", "Teste")

	if err == nil {
		t.Error("Esperado erro, mas não ocorreu")
	}

	if !strings.Contains(err.Error(), "não conectada") {
		t.Errorf("Mensagem de erro não é amigável: %v", err)
	}
}

// TestWAHAProvider_Send_SessionNotFound testa sessão inexistente
func TestWAHAProvider_Send_SessionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Session not found",
		})
	}))
	defer server.Close()

	cfg := config.WAHAConfig{
		APIURL:  server.URL,
		Session: "inexistente",
		Timeout: 5,
	}

	provider, _ := NewWAHAProvider(cfg)
	err := provider.Send("5511999998888@c.us", "Teste")

	if err == nil {
		t.Error("Esperado erro, mas não ocorreu")
	}

	if !strings.Contains(err.Error(), "não encontrada") {
		t.Errorf("Mensagem não indica sessão inexistente: %v", err)
	}
}

// TestWAHAProvider_Send_WithAPIKey testa autenticação
func TestWAHAProvider_Send_WithAPIKey(t *testing.T) {
	expectedKey := "secret-api-key-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Valida header de auth
		apiKey := r.Header.Get("X-Api-Key")
		if apiKey != expectedKey {
			t.Errorf("API Key incorreta. Esperado: %s, Recebido: %s", expectedKey, apiKey)
			w.WriteHeader(401)
			return
		}

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"id": "msg-456"})
	}))
	defer server.Close()

	cfg := config.WAHAConfig{
		APIURL:  server.URL,
		Session: "test",
		APIKey:  expectedKey,
		Timeout: 5,
	}

	provider, _ := NewWAHAProvider(cfg)
	err := provider.Send("5511999998888@c.us", "Teste com auth")

	if err != nil {
		t.Errorf("Envio com API Key falhou: %v", err)
	}
}

// TestWAHAProvider_Name testa método Name
func TestWAHAProvider_Name(t *testing.T) {
	cfg := config.WAHAConfig{
		APIURL: "http://localhost:3000",
	}

	provider, _ := NewWAHAProvider(cfg)

	if provider.Name() != "WAHA" {
		t.Errorf("Nome incorreto: %s", provider.Name())
	}
}

// TestWAHAProvider_Send_MultipleTargets testa múltiplos destinatários
func TestWAHAProvider_Send_MultipleTargets(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"id": "msg-123"})
	}))
	defer server.Close()

	cfg := config.WAHAConfig{
		APIURL:  server.URL,
		Session: "test",
		Timeout: 5,
	}

	provider, _ := NewWAHAProvider(cfg)
	err := provider.Send("5511999998888@c.us,5511888777666@c.us", "Mensagem para múltiplos")

	if err != nil {
		t.Errorf("Send falhou: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("Esperado 2 requisições, recebido %d", requestCount)
	}
}
