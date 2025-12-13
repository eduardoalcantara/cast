package main

import (
	"fmt"
	"os"
	"path/filepath"
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

Quebras de Linha:
  Use \n para quebra de linha simples e \n\n para linha em branco:
  - cast send tg me "Linha 1\nLinha 2"
  - cast send tg me "Parágrafo 1\n\nParágrafo 2"

Email com Assunto e Anexos:
  Para emails, você pode usar flags adicionais:
  - --subject, -s: Define o assunto do email (padrão: "Notificação CAST")
  - --attachment, -a: Adiciona um arquivo anexo (pode ser usado múltiplas vezes)
  - cast send mail admin@empresa.com "Mensagem" --subject "Assunto" --attachment arquivo.pdf

Aguardar Resposta (IMAP):
  Para emails, você pode aguardar uma resposta via IMAP:
  - --wfr, --wait-for-response: Aguarda resposta via IMAP (usa tempo do config ou 30min)
  - --wfr-minutes N: Especifica tempo de espera em minutos (sobrescreve config)
  - --full, --full-layout: Inclui HTML no corpo da resposta (padrão: apenas texto)
  - cast send mail destinatario@exemplo.com "Assunto" "Mensagem" --wfr
  - cast send mail destinatario@exemplo.com "Assunto" "Mensagem" --wfr --wfr-minutes 15
  - cast send mail destinatario@exemplo.com "Assunto" "Mensagem" --wfr-minutes 10
  - Se uma resposta for encontrada, exibe o corpo completo da resposta
  - Exit codes: 0 (resposta recebida), 3 (timeout sem resposta), 2 (config), 4 (auth)`,
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

	# Email aguardando resposta (IMAP)
	cast send mail destinatario@exemplo.com "Pergunta" "Você pode confirmar?" --wfr
	cast send mail destinatario@exemplo.com "Assunto" "Mensagem" --wfr --wfr-minutes 15
	cast send mail cliente@empresa.com "Solicitação" "Por favor, responda" --wait-for-response --wfr-minutes 30 --verbose

	# WAHA (WhatsApp HTTP API)
	cast send waha 5511999998888@c.us "Notificação via WAHA"
	cast send waha 120363XXXXX@g.us "Mensagem para grupo"

	# Mensagens com quebras de linha
	cast send tg me "Linha 1\nLinha 2"
	cast send tg me "Parágrafo 1\n\nParágrafo 2"`,
	Args: cobra.MinimumNArgs(2), // Aceita 2 args (alias + message) ou 3 args (provider + target + message)
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Carrega configuração primeiro para verificar aliases
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			yellow := color.New(color.FgYellow)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)

			// Informa onde o CAST está procurando o arquivo
			execPath, _ := os.Executable()
			if execPath != "" {
				execDir := filepath.Dir(execPath)
				yellow.Fprintf(os.Stderr, "\n💡 Dica: O CAST procura cast.yaml em:\n")
				yellow.Fprintf(os.Stderr, "   1. Diretório do executável: %s\n", execDir)
				wd, _ := os.Getwd()
				if wd != "" {
					yellow.Fprintf(os.Stderr, "   2. Diretório atual: %s\n", wd)
				}
				yellow.Fprintf(os.Stderr, "   3. Variáveis de ambiente (CAST_*)\n\n")
			}

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
			message = processNewlines(message) // Processa quebras de linha
			providerName = args[0] // Para debug
			target = alias.Target  // Para debug
		} else {
			// Não é alias - formato tradicional: cast send provider target "mensagem" (3 argumentos)
			// OU: cast send mail target "assunto" "mensagem" (4 argumentos para email)
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

			// Para email/mail: se houver 4 argumentos e --subject não foi usado,
			// o terceiro argumento é o assunto e o quarto é a mensagem
			normalizedProvider := strings.ToLower(providerName)
			if (normalizedProvider == "mail" || normalizedProvider == "email") && len(args) == 4 {
				// Verifica se --subject não foi fornecido via flag
				subjectFlag, _ := cmd.Flags().GetString("subject")
				if subjectFlag == "" {
					// Terceiro argumento é o assunto, quarto é a mensagem
					// Armazena o assunto temporariamente (será usado depois)
					cmd.Flags().Set("subject", args[2])
					message = args[3]
				} else {
					// --subject foi fornecido, então todos os args após target são mensagem
					message = strings.Join(args[2:], " ")
				}
			} else {
				// Formato padrão: todos os args após target são mensagem
				message = strings.Join(args[2:], " ")
			}

			message = processNewlines(message) // Processa quebras de linha
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

		// Determinar se deve aguardar resposta e por quanto tempo
		// NOVA ARQUITETURA: Flag Bool para presença + Flag Int opcional para valor customizado
		waitMinutes := 0
		wfrEnabled := false

		// Verifica se flag bool foi usada (qualquer uma delas)
		if cmd.Flags().Changed("wfr") || cmd.Flags().Changed("wait-for-response") {
			wfrBool, _ := cmd.Flags().GetBool("wfr")
			wfrLongBool, _ := cmd.Flags().GetBool("wait-for-response")

			wfrEnabled = wfrBool || wfrLongBool

			if verbose {
				cyan := color.New(color.FgCyan)
				cyan.Printf("[DEBUG] Flag --wfr detectada: %v\n", wfrEnabled)
			}
		}

		// Se flag habilitada, determinar tempo
		if wfrEnabled {
			// Primeiro, verificar se --wfr-minutes foi especificado
			if cmd.Flags().Changed("wfr-minutes") {
				waitMinutes, _ = cmd.Flags().GetInt("wfr-minutes")
				if verbose {
					cyan := color.New(color.FgCyan)
					cyan.Printf("[DEBUG] Usando --wfr-minutes: %d\n", waitMinutes)
				}
			}

			// Se --wfr-minutes não foi usado ou é 0, usar config ou padrão
			if waitMinutes == 0 {
				if cfg == nil {
					red := color.New(color.FgRed, color.Bold)
					red.Fprintf(os.Stderr, "✗ Erro: --wait-for-response requer arquivo de configuração (cast.yaml) com dados de conexão IMAP\n")
					return fmt.Errorf("--wait-for-response requer arquivo de configuração com dados de conexão IMAP")
				}
				if cfg.Email.WaitForResponseDefault > 0 {
					waitMinutes = cfg.Email.WaitForResponseDefault
					if verbose {
						cyan := color.New(color.FgCyan)
						cyan.Printf("[DEBUG] Usando wait_for_response_default_minutes do config: %d\n", waitMinutes)
					}
				} else {
					waitMinutes = 30 // Padrão hard-coded
					if verbose {
						cyan := color.New(color.FgCyan)
						cyan.Printf("[DEBUG] Usando padrão de 30 minutos\n")
					}
				}
			}

			// Validar contra máximo configurado
			if cfg != nil && cfg.Email.WaitForResponseMax > 0 && waitMinutes > cfg.Email.WaitForResponseMax {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro: tempo de espera (%d min) excede o máximo configurado (%d min)\n", waitMinutes, cfg.Email.WaitForResponseMax)
				return fmt.Errorf("tempo de espera (%d min) excede o máximo configurado (%d min)", waitMinutes, cfg.Email.WaitForResponseMax)
			}
		}

		// CORREÇÃO: Se --wfr-minutes foi usado sozinho (sem --wfr), ativar automaticamente
		if !wfrEnabled && cmd.Flags().Changed("wfr-minutes") {
			waitMinutes, _ = cmd.Flags().GetInt("wfr-minutes")
			if waitMinutes > 0 {
				wfrEnabled = true
				if verbose {
					cyan := color.New(color.FgCyan)
					cyan.Printf("[DEBUG] --wfr-minutes usado sozinho, ativando espera: %d min\n", waitMinutes)
				}
			}
		}

		// Se --wfr foi usado com provider diferente de email, avisa e ignora
		if wfrEnabled && actualProviderName != "email" && actualProviderName != "mail" {
			yellow := color.New(color.FgYellow)
			yellow.Printf("⚠ Parâmetro --wait-for-response suportado apenas para provider 'mail'.\n")
			wfrEnabled = false
			waitMinutes = 0
		}

		if verbose {
			cyan := color.New(color.FgCyan)
			cyan.Printf("[DEBUG] waitMinutes calculado: %d\n", waitMinutes)
			cyan.Printf("[DEBUG] wfrEnabled: %v\n", wfrEnabled)
		}

		// Validação de waitMinutes
		if waitMinutes > 0 {
			if cfg == nil {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro: configuração não carregada\n")
				return fmt.Errorf("configuração não carregada")
			}
		}

		// Envia mensagem
		var messageID string
		// Se for email e tiver flags de assunto/anexo, usa método estendido
		if actualProviderName == "email" || actualProviderName == "mail" {
			subject, _ := cmd.Flags().GetString("subject")
			attachments, _ := cmd.Flags().GetStringSlice("attachment")

			// Type assertion para EmailProviderExtended
			if emailProv, ok := provider.(providers.EmailProviderExtended); ok {
				messageID, err = emailProv.SendEmail(actualTarget, message, subject, attachments)
			} else {
				// Fallback para método padrão se não conseguir fazer type assertion
				err = provider.Send(actualTarget, message)
				if err == nil {
					// Tenta obter Message-ID via getter
					if emailProv, ok := provider.(providers.EmailProviderExtended); ok {
						messageID = emailProv.GetLastMessageID()
					}
				}
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

		// Se waitMinutes > 0 e provider é email, aguarda resposta
		if waitMinutes > 0 && wfrEnabled && (actualProviderName == "email" || actualProviderName == "mail") {
			// Usa subject do flag ou padrão
			subject, _ := cmd.Flags().GetString("subject")
			if subject == "" {
				subject = "Notificação CAST"
			}

			// Chama waitForEmailResponse
			// Verifica flag --full ou --full-layout
			fullLayout, _ := cmd.Flags().GetBool("full")
			if !fullLayout {
				fullLayout, _ = cmd.Flags().GetBool("full-layout")
			}
			// Se flag não foi especificada, usa config (default: false = sem HTML)
			if !fullLayout && cfg != nil {
				fullLayout = cfg.Email.WaitForResponseFullLayout
			}

			err = providers.WaitForEmailResponse(cfg.Email, messageID, subject, waitMinutes, fullLayout, verbose)
			if err != nil {
				// Trata exit codes específicos
				if err == providers.ErrNoEmailResponse {
					// Timeout sem resposta: exit code 3
					os.Exit(3)
				}
				if err == providers.ErrIMAPConfigMissing {
					// Configuração faltando: exit code 2
					red := color.New(color.FgRed, color.Bold)
					red.Fprintf(os.Stderr, "✗ %v\n", err)
					os.Exit(2)
				}
				if err == providers.ErrIMAPAuth {
					// Erro de autenticação: exit code 4
					red := color.New(color.FgRed, color.Bold)
					red.Fprintf(os.Stderr, "✗ %v\n", err)
					os.Exit(4)
				}
				// Outros erros de rede/timeout: exit code 3
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro ao aguardar resposta: %v\n", err)
				os.Exit(3)
			}
		}

		return nil
	},
}

func init() {
	sendCmd.Flags().BoolP("verbose", "v", false, "Mostra informações detalhadas de debug")
	sendCmd.Flags().StringP("subject", "s", "", "Assunto do email (apenas para provider email)")
	sendCmd.Flags().StringSliceP("attachment", "a", []string{}, "Caminho do arquivo anexo (apenas para provider email, pode ser usado múltiplas vezes)")
	// Flags para aguardar resposta via IMAP (apenas para provider email)
	sendCmd.Flags().Bool("wfr", false, "Aguarda resposta do destinatário via IMAP (usa tempo do config ou 30min)")
	sendCmd.Flags().Bool("wait-for-response", false, "Aguarda resposta do destinatário via IMAP (forma longa)")
	sendCmd.Flags().Int("wfr-minutes", 0, "Tempo de espera em minutos (0 = usar config/padrão, apenas para provider email)")
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

// processNewlines processa sequências de quebra de linha na mensagem.
// Converte \n em quebra de linha real e \n\n em linha em branco.
func processNewlines(message string) string {
	// Primeiro processa \n\n (duas quebras de linha = linha em branco)
	// Usa um placeholder temporário para evitar processamento duplo
	message = strings.ReplaceAll(message, "\\n\\n", "\x00DOUBLE_NEWLINE\x00")

	// Depois processa \n (quebra de linha única)
	message = strings.ReplaceAll(message, "\\n", "\n")

	// Restaura \n\n (duas quebras de linha)
	message = strings.ReplaceAll(message, "\x00DOUBLE_NEWLINE\x00", "\n\n")

	return message
}

// showErrorDetails exibe detalhes adicionais do erro em modo verbose.
func showErrorDetails(err error, providerName string, cfg *config.Config) {
	yellow := color.New(color.FgYellow)
	fmt.Println()
	yellow.Println("=== DETALHES DO ERRO (VERBOSE) ===")
	yellow.Printf("Erro completo: %+v\n", err)
	fmt.Println()
}
