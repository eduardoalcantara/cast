package providers

import (
	"testing"

	"github.com/eduardoalcantara/cast/internal/config"
)

func TestEmailProvider_Name(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
	}
	provider := NewEmailProvider(cfg)
	if provider.Name() != "email" {
		t.Errorf("Esperado 'email', obtido '%s'", provider.Name())
	}
}

func TestEmailProvider_Send_NoTargets(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
	}
	provider := NewEmailProvider(cfg)

	// Tenta enviar sem targets
	err := provider.Send("", "Teste")
	if err == nil {
		t.Error("Esperado erro para nenhum destinatário")
	}
}

func TestEmailProvider_Send_MultipleTargets(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		FromEmail: "from@example.com",
		FromName:  "Test Sender",
		UseTLS:    true,
	}
	provider := NewEmailProvider(cfg)

	// Tenta enviar para múltiplos targets (vai falhar porque não há servidor SMTP real)
	// Mas pelo menos valida que a função aceita múltiplos targets
	err := provider.Send("user1@example.com,user2@example.com", "Teste")
	// Esperamos erro de conexão, não erro de parsing
	if err != nil {
		// Erro esperado (sem servidor SMTP), mas não deve ser erro de parsing
		if err.Error() == "nenhum destinatário especificado" {
			t.Error("Erro inesperado: nenhum destinatário especificado")
		}
		// Qualquer outro erro (conexão, etc) é esperado neste teste
	}
}

func TestEmailProvider_Send_FromEmailFallback(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		// FromEmail vazio - deve usar Username
		UseTLS: true,
	}
	provider := NewEmailProvider(cfg)

	// Tenta enviar (vai falhar por falta de servidor SMTP)
	err := provider.Send("recipient@example.com", "Teste")
	// Erro esperado (sem servidor SMTP)
	if err != nil {
		// Verifica que não é erro de FromEmail
		if err.Error() == "from_email não pode estar vazio" {
			t.Error("FromEmail deveria usar Username como fallback")
		}
	}
}
