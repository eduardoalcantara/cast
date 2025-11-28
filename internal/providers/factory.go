package providers

import (
	"fmt"

	"github.com/eduardoalcantara/cast/internal/config"
)

// GetProvider retorna a implementação do provider baseado no nome.
// A resolução de aliases deve ser feita antes de chamar esta função.
func GetProvider(name string, conf *config.Config) (Provider, error) {
	// Normaliza o nome do provider
	providerName := normalizeProviderName(name)

	// Instancia o provider baseado no nome
	switch providerName {
	case "telegram", "tg":
		if conf == nil || conf.Telegram.Token == "" {
			return nil, fmt.Errorf("configuração do Telegram não encontrada: token obrigatório")
		}
		return NewTelegramProvider(&conf.Telegram, ""), nil

	case "email", "mail":
		if conf == nil {
			return nil, fmt.Errorf("configuração do Email não encontrada")
		}
		if conf.Email.SMTPHost == "" || conf.Email.Username == "" || conf.Email.Password == "" {
			return nil, fmt.Errorf("configuração do Email incompleta: smtp_host, username e password são obrigatórios")
		}
		return NewEmailProvider(&conf.Email), nil

	case "whatsapp", "zap":
		return nil, fmt.Errorf("whatsapp ainda não implementado (Fase 03)")

	case "google_chat", "googlechat":
		return nil, fmt.Errorf("google_chat ainda não implementado (Fase 03)")

	default:
		return nil, fmt.Errorf("provider desconhecido: %s (suportados: tg, mail, zap, google_chat)", name)
	}
}

// normalizeProviderName normaliza o nome do provider para comparação.
func normalizeProviderName(name string) string {
	switch name {
	case "tg", "telegram":
		return "telegram"
	case "mail", "email":
		return "email"
	case "zap", "whatsapp":
		return "whatsapp"
	case "google_chat", "googlechat":
		return "google_chat"
	default:
		return name
	}
}
