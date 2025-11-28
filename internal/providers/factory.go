package providers

import (
	"fmt"
	"strings"

	"github.com/eduardoalcantara/cast/internal/config"
)

// GetProvider retorna a implementação do provider baseado no nome.
// A resolução de aliases deve ser feita antes de chamar esta função.
func GetProvider(name string, conf *config.Config) (Provider, error) {
	return GetProviderWithVerbose(name, conf, false)
}

// GetProviderWithVerbose retorna a implementação do provider baseado no nome com modo verbose.
func GetProviderWithVerbose(name string, conf *config.Config, verbose bool) (Provider, error) {
	// Normaliza o nome do provider
	providerName := normalizeProviderName(name)

	// Instancia o provider baseado no nome
	switch providerName {
	case "telegram", "tg":
		if conf == nil || conf.Telegram.Token == "" {
			return nil, fmt.Errorf("configuração do Telegram não encontrada: token obrigatório")
		}
		if verbose {
			return NewTelegramProviderWithVerbose(&conf.Telegram, "", true), nil
		}
		return NewTelegramProvider(&conf.Telegram, ""), nil

	case "email", "mail":
		if conf == nil {
			return nil, fmt.Errorf("configuração do Email não encontrada")
		}
		var missing []string
		if conf.Email.SMTPHost == "" {
			missing = append(missing, "smtp_host")
		}
		if conf.Email.SMTPPort == 0 {
			missing = append(missing, "smtp_port")
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("configuração do Email incompleta: %s são obrigatórios", strings.Join(missing, ", "))
		}
		// Username e password são opcionais (servidores como MailHog não requerem autenticação)
		return NewEmailProvider(&conf.Email), nil

	case "whatsapp", "zap":
		if conf == nil {
			return nil, fmt.Errorf("configuração do WhatsApp não encontrada")
		}
		if conf.WhatsApp.PhoneNumberID == "" || conf.WhatsApp.AccessToken == "" {
			return nil, fmt.Errorf("configuração do WhatsApp incompleta: phone_number_id e access_token são obrigatórios")
		}
		return NewWhatsAppProvider(&conf.WhatsApp), nil

	case "google_chat", "googlechat":
		if conf == nil {
			return nil, fmt.Errorf("configuração do Google Chat não encontrada")
		}
		// Webhook URL pode estar vazia se for passada como target no comando send
		return NewGoogleChatProvider(&conf.GoogleChat), nil

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
