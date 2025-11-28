package providers

import (
	"testing"

	"github.com/eduardoalcantara/cast/internal/config"
)

func TestGetProvider_Telegram(t *testing.T) {
	cfg := &config.Config{
		Telegram: config.TelegramConfig{
			Token: "test-token",
		},
	}

	provider, err := GetProvider("tg", cfg)
	if err != nil {
		t.Fatalf("Erro ao obter provider Telegram: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider não deveria ser nil")
	}

	if provider.Name() != "telegram" {
		t.Errorf("Esperado nome 'telegram', obtido '%s'", provider.Name())
	}
}

func TestGetProvider_Email(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password",
		},
	}

	provider, err := GetProvider("mail", cfg)
	if err != nil {
		t.Fatalf("Erro ao obter provider Email: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider não deveria ser nil")
	}

	if provider.Name() != "email" {
		t.Errorf("Esperado nome 'email', obtido '%s'", provider.Name())
	}
}

func TestGetProvider_WhatsApp_NotImplemented(t *testing.T) {
	cfg := &config.Config{}

	provider, err := GetProvider("zap", cfg)
	if err == nil {
		t.Error("Esperado erro para WhatsApp não implementado")
	}

	if provider != nil {
		t.Error("Provider não deveria ser retornado para WhatsApp")
	}
}

func TestGetProvider_Unknown(t *testing.T) {
	cfg := &config.Config{}

	provider, err := GetProvider("unknown", cfg)
	if err == nil {
		t.Error("Esperado erro para provider desconhecido")
	}

	if provider != nil {
		t.Error("Provider não deveria ser retornado para provider desconhecido")
	}
}

func TestGetProvider_Telegram_MissingToken(t *testing.T) {
	cfg := &config.Config{
		Telegram: config.TelegramConfig{
			// Token vazio
		},
	}

	provider, err := GetProvider("tg", cfg)
	if err == nil {
		t.Error("Esperado erro para token ausente")
	}

	if provider != nil {
		t.Error("Provider não deveria ser retornado sem token")
	}
}

func TestGetProvider_Email_MissingConfig(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			// Configuração incompleta
		},
	}

	provider, err := GetProvider("mail", cfg)
	if err == nil {
		t.Error("Esperado erro para configuração incompleta")
	}

	if provider != nil {
		t.Error("Provider não deveria ser retornado sem configuração completa")
	}
}

func TestNormalizeProviderName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"tg", "telegram"},
		{"telegram", "telegram"},
		{"mail", "email"},
		{"email", "email"},
		{"zap", "whatsapp"},
		{"whatsapp", "whatsapp"},
		{"google_chat", "google_chat"},
		{"googlechat", "google_chat"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeProviderName(tt.input)
			if result != tt.expected {
				t.Errorf("Esperado '%s', obtido '%s'", tt.expected, result)
			}
		})
	}
}
