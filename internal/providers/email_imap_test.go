package providers

import (
	"strings"
	"testing"
	"time"

	"github.com/eduardoalcantara/cast/internal/config"
)

func TestGenerateMessageID(t *testing.T) {
	domain := "exemplo.com"
	messageID := generateMessageID(domain)

	// Verifica formato
	if !strings.HasPrefix(messageID, "<cast-") {
		t.Errorf("Message-ID deve começar com '<cast-', obteve: %s", messageID)
	}
	if !strings.HasSuffix(messageID, "@exemplo.com>") {
		t.Errorf("Message-ID deve terminar com '@exemplo.com>', obteve: %s", messageID)
	}

	// Verifica unicidade (gera dois e compara)
	messageID2 := generateMessageID(domain)
	if messageID == messageID2 {
		t.Errorf("Message-IDs devem ser únicos, mas foram iguais: %s", messageID)
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"user@exemplo.com", "exemplo.com"},
		{"test@domain.org", "domain.org"},
		{"invalid", "cast.local"},
		{"", "cast.local"},
	}

	for _, tt := range tests {
		result := extractDomain(tt.email)
		if result != tt.expected {
			t.Errorf("extractDomain(%q) = %q, esperado %q", tt.email, result, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{2*time.Minute + 34*time.Second, "2m34s"},
		{1*time.Minute + 5*time.Second, "1m5s"},
		{45*time.Second, "45s"},
		{0, "0s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %q, esperado %q", tt.duration, result, tt.expected)
		}
	}
}

func TestWaitForEmailResponse_ConfigValidation(t *testing.T) {
	cfg := config.EmailConfig{
		// IMAP não configurado
		IMAPHost:     "",
		IMAPPort:     0,
		IMAPUsername: "",
		IMAPPassword: "",
	}

	err := WaitForEmailResponse(cfg, "<test@exemplo.com>", "Test", 15, false)
	if err == nil {
		t.Error("Esperado erro quando IMAP não está configurado")
	}
	if err != ErrIMAPConfigMissing && !strings.Contains(err.Error(), "configurar email.imap_*") {
		t.Errorf("Esperado ErrIMAPConfigMissing ou mensagem sobre configuração, obteve: %v", err)
	}
}

func TestWaitForEmailResponse_ZeroMinutes(t *testing.T) {
	cfg := config.EmailConfig{
		IMAPHost:     "imap.exemplo.com",
		IMAPPort:     993,
		IMAPUsername: "user@exemplo.com",
		IMAPPassword: "password",
	}

	// waitMinutes = 0 deve retornar nil imediatamente
	err := WaitForEmailResponse(cfg, "<test@exemplo.com>", "Test", 0, false)
	if err != nil {
		t.Errorf("Esperado nil quando waitMinutes=0, obteve: %v", err)
	}
}

func TestWaitForEmailResponse_MaxMinutesExceeded(t *testing.T) {
	cfg := config.EmailConfig{
		IMAPHost:                "imap.exemplo.com",
		IMAPPort:                993,
		IMAPUsername:            "user@exemplo.com",
		IMAPPassword:            "password",
		WaitForResponseMax:      30,
		WaitForResponseMaxLines: 0,
	}

	// waitMinutes > max deve retornar erro
	err := WaitForEmailResponse(cfg, "<test@exemplo.com>", "Test", 60, false)
	if err == nil {
		t.Error("Esperado erro quando waitMinutes excede o máximo")
	}
	if !strings.Contains(err.Error(), "excede") {
		t.Errorf("Esperado erro sobre exceder máximo, obteve: %v", err)
	}
}
