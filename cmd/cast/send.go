package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/eduardoalcantara/cast/internal/config"
	"github.com/eduardoalcantara/cast/internal/providers"
)

var sendCmd = &cobra.Command{
	Use:           "send [provider|alias] [target] [message]",
	Short:         "Envia uma mensagem através do provider especificado",
	SilenceUsage:  true,  // Não mostra help automaticamente em caso de erro
	SilenceErrors: true,  // Não mostra erro automaticamente (já imprimimos com formatação customizada)
	Long: `Envia uma mensagem através do provider especificado (telegram, whatsapp, email, etc).

Formato:
  - cast send [alias] [message]                    (usando alias configurado)
  - cast send [provider] [target] [message]        (formato tradicional)

A ordem de precedência para configuração é:
  1. Variáveis de Ambiente (CAST_*)
  2. Arquivo Local (cast.*)

Múltiplos Recipientes:
  Você pode enviar para múltiplos recipientes separando-os por vírgula (,) ou ponto-e-vírgula (;):
  - cast send mail "user1@exemplo.com,user2@exemplo.com" "Mensagem"
  - cast send tg "123456789;987654321" "Mensagem"

WAHA (WhatsApp HTTP API):
  Formato do target: 5511999998888@c.us (contato) ou 120363XXX@g.us (grupo)
  - cast send waha 5511999998888@c.us "Notificação controlada"

Email com Assunto e Anexos:
  Para emails, você pode usar flags adicionais:
  - --subject, -s: Define o assunto do email (padrão: "Notificação CAST")
  - --attachment, -a: Adiciona um arquivo anexo (pode ser usado múltiplas vezes)
  - cast send mail admin@empresa.com "Mensagem" --subject "Assunto" --attachment arquivo.pdf`,
	Example: `  # Usando alias 'me' (mais simples)
  cast send me "Deploy finalizado com sucesso"

  # Telegram (formato tradicional)
  cast send tg me "Deploy finalizado com sucesso"

  # Telegram para múltiplos destinatários
  cast send tg "123456789,987654321" "Mensagem para todos"

  # WhatsApp (formato internacional)
  cast send zap 5511999998888 "Alerta: Disco cheio"

  # Email para um destinatário
  cast send mail admin@empresa.com "Bom dia!"

	# Email para múltiplos destinatários
	cast send mail "admin@empresa.com;dev@empresa.com" "Relatório Diário"

	# Email com assunto customizado
	cast send mail admin@empresa.com "Bom dia!" --subject "Assunto Personalizado"

	# Email com anexo
	cast send mail admin@empresa.com "Veja o anexo" --attachment arquivo.pdf

	# Email com assunto e múltiplos anexos
	cast send mail admin@empresa.com "Relatório" --subject "Relatório Mensal" --attachment relatorio.pdf --attachment dados.xlsx

	# WAHA (WhatsApp HTTP API)
	cast send waha 5511999998888@c.us "Notificação via WAHA"
	cast send waha 120363XXXXX@g.us "Mensagem para grupo"`,
	Args: cobra.MinimumNArgs(2), // Aceita 2 args (alias + message) ou 3 args (provider + target + message)
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Carrega configuração primeiro para verificar aliases
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return fmt.Errorf("erro de configuração: %w", err)
		}

		// Verifica se o primeiro argumento é um alias
		var actualProviderName string
		var actualTarget string
		var message string
		var providerName string
		var target string

		// Verifica se o primeiro argumento é um alias
		// Primeiro tenta verificar se é um alias (mesmo que cfg.Aliases seja nil)
		var alias *config.AliasConfig
		if cfg != nil {
			alias = cfg.GetAlias(args[0])
		}

		if alias != nil {
			// É um alias - usa provider e target do alias
			// Formato: cast send me "mensagem" (2 argumentos)
			if len(args) < 2 {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro: Mensagem não fornecida\n")
				return fmt.Errorf("mensagem não fornecida")
			}
			actualProviderName = alias.Provider
			actualTarget = alias.Target
			message = strings.Join(args[1:], " ")
			providerName = args[0] // Para debug
			target = alias.Target  // Para debug
		} else {
			// Não é alias - formato tradicional: cast send provider target "mensagem" (3 argumentos)
			if len(args) < 3 {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro: Formato inválido.\n")
				red.Fprintf(os.Stderr, "  Use: cast send [provider] [target] [message]\n")
				red.Fprintf(os.Stderr, "  Ou: cast send [alias] [message] (se alias estiver configurado)\n")
				if cfg != nil && cfg.Aliases != nil && len(cfg.Aliases) > 0 {
					red.Fprintf(os.Stderr, "\n  Aliases disponíveis: ")
					aliasNames := make([]string, 0, len(cfg.Aliases))
					for name := range cfg.Aliases {
						aliasNames = append(aliasNames, name)
					}
					red.Fprintf(os.Stderr, "%s\n", strings.Join(aliasNames, ", "))
				}
				return fmt.Errorf("formato inválido: requer provider, target e message, ou alias e message")
			}
			providerName = args[0]
			target = args[1]
			message = strings.Join(args[2:], " ")
			actualProviderName = providerName
			actualTarget = target
		}

		// Debug: mostra informações se verbose estiver ativo
		if verbose {
			showDebugInfo(providerName, target, message, cfg)
		}

		// Resolve provider via Factory (com verbose se necessário)
		var provider providers.Provider
		if verbose {
			provider, err = providers.GetProviderWithVerbose(actualProviderName, cfg, true)
		} else {
			provider, err = providers.GetProvider(actualProviderName, cfg)
		}
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao obter provider: %v\n", err)
			return err
		}

		// Envia mensagem
		// Se for email e tiver flags de assunto/anexo, usa método estendido
		if actualProviderName == "email" || actualProviderName == "mail" {
			subject, _ := cmd.Flags().GetString("subject")
			attachments, _ := cmd.Flags().GetStringSlice("attachment")

			// Type assertion para EmailProviderExtended
			if emailProv, ok := provider.(providers.EmailProviderExtended); ok {
				err = emailProv.SendEmail(actualTarget, message, subject, attachments)
			} else {
				// Fallback para método padrão se não conseguir fazer type assertion
				err = provider.Send(actualTarget, message)
			}
		} else {
			err = provider.Send(actualTarget, message)
		}

		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao enviar mensagem: %v\n", err)
			if verbose {
				showErrorDetails(err, actualProviderName, cfg)
			}
			return err
		}

		// Sucesso
		green := color.New(color.FgHiGreen, color.Bold)
		green.Printf("✓ Mensagem enviada com sucesso via %s\n", provider.Name())

		return nil
	},
}

func init() {
	sendCmd.Flags().BoolP("verbose", "v", false, "Mostra informações detalhadas de debug")
	sendCmd.Flags().StringP("subject", "s", "", "Assunto do email (apenas para provider email)")
	sendCmd.Flags().StringSliceP("attachment", "a", []string{}, "Caminho do arquivo anexo (apenas para provider email, pode ser usado múltiplas vezes)")
}

// showDebugInfo exibe informações de debug quando --verbose está ativo.
func showDebugInfo(providerName, target, message string, cfg *config.Config) {
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	cyan.Println("=== DEBUG MODE ===")
	fmt.Println()

	cyan.Printf("Provider: %s\n", providerName)
	cyan.Printf("Target: %s\n", target)
	cyan.Printf("Message: %s\n", message)
	fmt.Println()

	// Mostra informações específicas do provider
	switch providerName {
	case "tg", "telegram":
		if cfg != nil && cfg.Telegram.Token != "" {
			maskedToken := maskToken(cfg.Telegram.Token)
			cyan.Printf("Token: %s\n", maskedToken)
			cyan.Printf("API URL: %s\n", cfg.Telegram.APIURL)
			if cfg.Telegram.DefaultChatID != "" {
				cyan.Printf("Default Chat ID: %s\n", cfg.Telegram.DefaultChatID)
			}
			if target != "me" {
				cyan.Printf("Chat ID (target): %s\n", target)
			}
		} else {
			yellow.Println("⚠ Configuração do Telegram não encontrada")
		}
	case "mail", "email":
		if cfg != nil && cfg.Email.SMTPHost != "" {
			cyan.Printf("SMTP Host: %s\n", cfg.Email.SMTPHost)
			cyan.Printf("SMTP Port: %d\n", cfg.Email.SMTPPort)
			cyan.Printf("Username: %s\n", cfg.Email.Username)
			maskedPassword := maskToken(cfg.Email.Password)
			cyan.Printf("Password: %s\n", maskedPassword)
			cyan.Printf("From: %s <%s>\n", cfg.Email.FromName, cfg.Email.FromEmail)
		}
	case "zap", "whatsapp":
		if cfg != nil && cfg.WhatsApp.PhoneNumberID != "" {
			cyan.Printf("Phone Number ID: %s\n", cfg.WhatsApp.PhoneNumberID)
			maskedToken := maskToken(cfg.WhatsApp.AccessToken)
			cyan.Printf("Access Token: %s\n", maskedToken)
			cyan.Printf("API URL: %s\n", cfg.WhatsApp.APIURL)
			cyan.Printf("API Version: %s\n", cfg.WhatsApp.APIVersion)
		}
	case "google_chat", "googlechat":
		if cfg != nil && cfg.GoogleChat.WebhookURL != "" {
			cyan.Printf("Webhook URL: %s\n", cfg.GoogleChat.WebhookURL)
		}
	}

	fmt.Println()
	cyan.Println("=== FIM DEBUG ===")
	fmt.Println()
}

// showErrorDetails exibe detalhes adicionais do erro em modo verbose.
func showErrorDetails(err error, providerName string, cfg *config.Config) {
	yellow := color.New(color.FgYellow)
	fmt.Println()
	yellow.Println("=== DETALHES DO ERRO (VERBOSE) ===")
	yellow.Printf("Erro completo: %+v\n", err)
	fmt.Println()
}
