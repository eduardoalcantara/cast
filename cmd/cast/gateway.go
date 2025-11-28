package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/eduardoalcantara/cast/internal/config"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Gerencia configurações de gateways",
	Long: `Gerencia configurações de gateways (Telegram, WhatsApp, Email, Google Chat).

Exemplos:
  cast gateway add telegram --token "123456:ABC" --default-chat-id "123456789"
  cast gateway add email --interactive
  cast gateway show telegram`,
}

var gatewayAddCmd = &cobra.Command{
	Use:   "add [provider]",
	Short: "Adiciona/Configura um gateway",
	Long: `Adiciona ou configura um gateway.

Use --interactive para modo wizard interativo.
Ou use flags para configurar diretamente.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")

		// Determina provider
		var providerName string
		if len(args) > 0 {
			providerName = args[0]
		} else if !interactive {
			return fmt.Errorf("provider é obrigatório ou use --interactive")
		}

		// Modo interativo
		if interactive {
			return runGatewayWizard(providerName)
		}

		// Modo flags
		return runGatewayAddFlags(cmd, providerName)
	},
}

var gatewayShowCmd = &cobra.Command{
	Use:   "show <provider>",
	Short: "Mostra configuração de um gateway",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]
		mask, _ := cmd.Flags().GetBool("mask")

		cfg, err := config.LoadConfig()
		if err != nil {
			yellow := color.New(color.FgYellow)
			yellow.Printf("Gateway '%s' não configurado\n", providerName)
			return nil
		}

		normalized := normalizeGatewayName(providerName)
		switch normalized {
		case "telegram":
			showTelegramConfig(cfg.Telegram, mask)
		case "email":
			showEmailConfig(cfg.Email, mask)
		case "whatsapp":
			showWhatsAppConfig(cfg.WhatsApp, mask)
		case "google_chat":
			showGoogleChatConfig(cfg.GoogleChat, mask)
		default:
			return fmt.Errorf("provider desconhecido: %s", providerName)
		}

		return nil
	},
}

var gatewayRemoveCmd = &cobra.Command{
	Use:   "remove <provider>",
	Short: "Remove configuração de um gateway",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]
		confirm, _ := cmd.Flags().GetBool("confirm")

		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		normalized := normalizeGatewayName(providerName)
		if normalized == "" {
			return fmt.Errorf("provider desconhecido: %s", providerName)
		}

		// Confirmação
		if !confirm {
			yellow := color.New(color.FgYellow)
			yellow.Printf("Tem certeza que deseja remover a configuração do gateway '%s'? (s/N): ", providerName)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "s" && strings.ToLower(response) != "sim" {
				cyan := color.New(color.FgCyan)
				cyan.Println("Operação cancelada")
				return nil
			}
		}

		// Remove configuração
		switch normalized {
		case "telegram":
			cfg.Telegram = config.TelegramConfig{}
		case "email":
			cfg.Email = config.EmailConfig{}
		case "whatsapp":
			cfg.WhatsApp = config.WhatsAppConfig{}
		case "google_chat":
			cfg.GoogleChat = config.GoogleChatConfig{}
		}

		// Salva
		if err := config.Save(cfg); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao salvar configuração: %v\n", err)
			return err
		}

		green := color.New(color.FgHiGreen, color.Bold)
		green.Printf("✓ Configuração do gateway '%s' removida com sucesso\n", providerName)

		return nil
	},
}

var gatewayUpdateCmd = &cobra.Command{
	Use:   "update <provider>",
	Short: "Atualiza configuração de um gateway",
	Long: `Atualiza configuração de um gateway existente.

Atualiza apenas os campos fornecidos nas flags.
Mantém os outros campos intactos (atualização parcial).

Falha se o gateway não estiver configurado.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]
		normalized := normalizeGatewayName(providerName)
		if normalized == "" {
			return fmt.Errorf("provider desconhecido: %s", providerName)
		}

		// Carrega configuração
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		// Verifica se gateway existe
		var exists bool
		switch normalized {
		case "telegram":
			exists = cfg.Telegram.Token != ""
		case "email":
			exists = cfg.Email.SMTPHost != ""
		case "whatsapp":
			exists = cfg.WhatsApp.PhoneNumberID != ""
		case "google_chat":
			exists = cfg.GoogleChat.WebhookURL != ""
		}

		if !exists {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Gateway '%s' não está configurado\n", providerName)
			red.Println("Use 'cast gateway add' para configurar primeiro")
			return fmt.Errorf("gateway '%s' não está configurado", providerName)
		}

		// Atualiza apenas campos fornecidos
		switch normalized {
		case "telegram":
			if err := updateTelegramViaFlags(cmd, cfg); err != nil {
				return err
			}
		case "email":
			if err := updateEmailViaFlags(cmd, cfg); err != nil {
				return err
			}
		default:
			return fmt.Errorf("update não implementado para: %s", normalized)
		}

		// Valida configuração completa antes de salvar
		if err := cfg.Validate(); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Configuração inválida após update: %v\n", err)
			return fmt.Errorf("configuração inválida: %w", err)
		}

		// Salva
		if err := config.Save(cfg); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao salvar configuração: %v\n", err)
			return err
		}

		green := color.New(color.FgHiGreen, color.Bold)
		green.Printf("✓ Configuração do gateway '%s' atualizada com sucesso\n", providerName)

		return nil
	},
}

var gatewayTestCmd = &cobra.Command{
	Use:   "test <provider>",
	Short: "Testa conectividade de um gateway",
	Long: `Testa a conectividade e autenticação de um gateway.

Telegram: Chama getMe na API
Email: Conecta ao SMTP, faz autenticação e fecha conexão
WhatsApp: Chama endpoint de metadados
Google Chat: Valida URL do webhook`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]
		target, _ := cmd.Flags().GetString("target")
		normalized := normalizeGatewayName(providerName)
		if normalized == "" {
			return fmt.Errorf("provider desconhecido: %s", providerName)
		}

		// Carrega configuração
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		// Testa gateway
		switch normalized {
		case "telegram":
			return testTelegram(cfg.Telegram)
		case "email":
			return testEmail(cfg.Email, target)
		case "whatsapp":
			return fmt.Errorf("teste de WhatsApp ainda não implementado")
		case "google_chat":
			return testGoogleChat(cfg.GoogleChat, target)
		default:
			return fmt.Errorf("teste não implementado para: %s", normalized)
		}
	},
}

func init() {
	// Flags para gateway add
	gatewayAddCmd.Flags().String("token", "", "Token do Telegram")
	gatewayAddCmd.Flags().String("default-chat-id", "", "Chat ID padrão do Telegram")
	gatewayAddCmd.Flags().String("smtp-host", "", "Servidor SMTP")
	gatewayAddCmd.Flags().Int("smtp-port", 0, "Porta SMTP")
	gatewayAddCmd.Flags().String("username", "", "Usuário SMTP")
	gatewayAddCmd.Flags().String("password", "", "Senha SMTP")
	gatewayAddCmd.Flags().String("from-email", "", "Email remetente")
	gatewayAddCmd.Flags().String("from-name", "", "Nome remetente")
	gatewayAddCmd.Flags().Bool("use-tls", false, "Usar TLS")
	gatewayAddCmd.Flags().Bool("use-ssl", false, "Usar SSL")
	gatewayAddCmd.Flags().Int("timeout", 0, "Timeout em segundos")
	gatewayAddCmd.Flags().BoolP("interactive", "i", false, "Modo wizard interativo")

	// Flags para gateway update (mesmas do add)
	gatewayUpdateCmd.Flags().String("token", "", "Token do Telegram")
	gatewayUpdateCmd.Flags().String("default-chat-id", "", "Chat ID padrão do Telegram")
	gatewayUpdateCmd.Flags().String("smtp-host", "", "Servidor SMTP")
	gatewayUpdateCmd.Flags().Int("smtp-port", 0, "Porta SMTP")
	gatewayUpdateCmd.Flags().String("username", "", "Usuário SMTP")
	gatewayUpdateCmd.Flags().String("password", "", "Senha SMTP")
	gatewayUpdateCmd.Flags().String("from-email", "", "Email remetente")
	gatewayUpdateCmd.Flags().String("from-name", "", "Nome remetente")
	gatewayUpdateCmd.Flags().Bool("use-tls", false, "Usar TLS")
	gatewayUpdateCmd.Flags().Bool("use-ssl", false, "Usar SSL")
	gatewayUpdateCmd.Flags().Int("timeout", 0, "Timeout em segundos")

	gatewayTestCmd.Flags().StringP("target", "t", "", "Target para teste (opcional, para Email e Google Chat)")

	gatewayShowCmd.Flags().BoolP("mask", "m", true, "Mascara campos sensíveis")
	gatewayRemoveCmd.Flags().BoolP("confirm", "y", false, "Confirma sem perguntar")

	gatewayCmd.AddCommand(gatewayAddCmd)
	gatewayCmd.AddCommand(gatewayShowCmd)
	gatewayCmd.AddCommand(gatewayRemoveCmd)
	gatewayCmd.AddCommand(gatewayUpdateCmd)
	gatewayCmd.AddCommand(gatewayTestCmd)
	rootCmd.AddCommand(gatewayCmd)
}

// normalizeGatewayName normaliza o nome do gateway.
func normalizeGatewayName(name string) string {
	switch strings.ToLower(name) {
	case "tg", "telegram":
		return "telegram"
	case "mail", "email":
		return "email"
	case "zap", "whatsapp":
		return "whatsapp"
	case "google_chat", "googlechat":
		return "google_chat"
	default:
		return ""
	}
}

// runGatewayWizard executa o wizard interativo para configurar um gateway.
func runGatewayWizard(providerName string) error {
	// Se provider não foi especificado, pergunta
	if providerName == "" {
		var selected string
		prompt := &survey.Select{
			Message: "Selecione o gateway a configurar:",
			Options: []string{"telegram", "email", "whatsapp", "google_chat"},
		}
		if err := survey.AskOne(prompt, &selected); err != nil {
			return err
		}
		providerName = selected
	}

	normalized := normalizeGatewayName(providerName)
	if normalized == "" {
		return fmt.Errorf("provider desconhecido: %s", providerName)
	}

	// Carrega configuração existente
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.Config{}
	}

	// Executa wizard específico do provider
	switch normalized {
	case "telegram":
		return runTelegramWizard(cfg)
	case "email":
		return runEmailWizard(cfg)
	case "whatsapp":
		return runWhatsAppWizard(cfg)
	case "google_chat":
		return runGoogleChatWizard(cfg)
	default:
		return fmt.Errorf("wizard não implementado para: %s", normalized)
	}
}

// runGatewayAddFlags executa o add via flags.
func runGatewayAddFlags(cmd *cobra.Command, providerName string) error {
	normalized := normalizeGatewayName(providerName)
	if normalized == "" {
		return fmt.Errorf("provider desconhecido: %s", providerName)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.Config{}
	}

	switch normalized {
	case "telegram":
		return addTelegramViaFlags(cmd, cfg)
	case "email":
		return addEmailViaFlags(cmd, cfg)
	default:
		return fmt.Errorf("add via flags não implementado para: %s (use --interactive)", normalized)
	}
}

// runTelegramWizard executa o wizard para Telegram.
func runTelegramWizard(cfg *config.Config) error {
	var answers struct {
		Token        string `survey:"token"`
		DefaultChatID string `survey:"defaultChatID"`
		Timeout      string `survey:"timeout"`
	}

	questions := []*survey.Question{
		{
			Name:     "token",
			Prompt:   &survey.Input{Message: "Token do Bot (obtido via @BotFather):"},
			Validate: survey.Required,
		},
		{
			Name:   "defaultChatID",
			Prompt: &survey.Input{Message: "Chat ID padrão (opcional, pode deixar vazio):"},
		},
		{
			Name:   "timeout",
			Prompt: &survey.Input{Message: "Timeout em segundos (padrão: 30):", Default: "30"},
		},
	}

	if err := survey.Ask(questions, &answers); err != nil {
		return err
	}

	// Valida timeout
	timeout := 30
	if answers.Timeout != "" {
		if t, err := strconv.Atoi(answers.Timeout); err == nil && t > 0 {
			timeout = t
		}
	}

	// Atualiza configuração
	cfg.Telegram.Token = answers.Token
	cfg.Telegram.DefaultChatID = answers.DefaultChatID
	cfg.Telegram.Timeout = timeout

	// Mostra resumo
	cyan := color.New(color.FgCyan)
	cyan.Println("\nConfiguração a ser salva:")
	cyan.Printf("  Token: %s\n", maskToken(answers.Token))
	cyan.Printf("  Default Chat ID: %s\n", answers.DefaultChatID)
	cyan.Printf("  Timeout: %d segundos\n", timeout)

	// Confirmação
	var confirm bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Confirmar e salvar?",
		Default: true,
	}, &confirm); err != nil {
		return err
	}

	if !confirm {
		yellow := color.New(color.FgYellow)
		yellow.Println("Operação cancelada")
		return nil
	}

	// Salva
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("erro ao salvar: %w", err)
	}

	green := color.New(color.FgHiGreen, color.Bold)
	green.Println("✓ Configuração do Telegram salva com sucesso")

	return nil
}

// runEmailWizard executa o wizard para Email.
func runEmailWizard(cfg *config.Config) error {
	var answers struct {
		SMTPHost  string `survey:"smtphost"`
		SMTPPort  string `survey:"smtpport"`
		Username  string `survey:"username"`
		Password  string `survey:"password"`
		FromEmail string `survey:"fromemail"`
		FromName  string `survey:"fromname"`
		UseTLS    bool   `survey:"usetls"`
		UseSSL    bool   `survey:"usessl"`
		Timeout   string `survey:"timeout"`
	}

	questions := []*survey.Question{
		{
			Name:     "smtphost",
			Prompt:   &survey.Input{Message: "Servidor SMTP (ex: smtp.gmail.com):"},
			Validate: survey.Required,
		},
		{
			Name:   "smtpport",
			Prompt: &survey.Input{Message: "Porta SMTP (587 para TLS, 465 para SSL):", Default: "587"},
		},
		{
			Name:     "username",
			Prompt:   &survey.Input{Message: "Usuário (email):"},
			Validate: survey.Required,
		},
		{
			Name:     "password",
			Prompt:   &survey.Password{Message: "Senha:"},
			Validate: survey.Required,
		},
		{
			Name:   "fromemail",
			Prompt: &survey.Input{Message: "Email remetente (opcional, usa usuário se vazio):"},
		},
		{
			Name:   "fromname",
			Prompt: &survey.Input{Message: "Nome remetente (opcional):"},
		},
		{
			Name:   "usetls",
			Prompt: &survey.Confirm{Message: "Usar TLS? (padrão: sim)", Default: true},
		},
		{
			Name:   "usessl",
			Prompt: &survey.Confirm{Message: "Usar SSL? (padrão: não)", Default: false},
		},
		{
			Name:   "timeout",
			Prompt: &survey.Input{Message: "Timeout em segundos (padrão: 30):", Default: "30"},
		},
	}

	if err := survey.Ask(questions, &answers); err != nil {
		return err
	}

	// Valida porta
	port := 587
	if answers.SMTPPort != "" {
		if p, err := strconv.Atoi(answers.SMTPPort); err == nil && p > 0 {
			port = p
		}
	} else {
		if answers.UseSSL {
			port = 465
		} else {
			port = 587
		}
	}

	// Valida TLS/SSL
	useTLS := answers.UseTLS
	useSSL := answers.UseSSL
	if !useTLS && !useSSL {
		useTLS = true // Padrão
	}

	// Valida timeout
	timeout := 30
	if answers.Timeout != "" {
		if t, err := strconv.Atoi(answers.Timeout); err == nil && t > 0 {
			timeout = t
		}
	}

	// Atualiza configuração
	cfg.Email.SMTPHost = answers.SMTPHost
	cfg.Email.SMTPPort = port
	cfg.Email.Username = answers.Username
	cfg.Email.Password = answers.Password
	cfg.Email.FromEmail = answers.FromEmail
	cfg.Email.FromName = answers.FromName
	cfg.Email.UseTLS = useTLS
	cfg.Email.UseSSL = useSSL
	cfg.Email.Timeout = timeout

	// Mostra resumo
	cyan := color.New(color.FgCyan)
	cyan.Println("\nConfiguração a ser salva:")
	cyan.Printf("  SMTP Host: %s\n", answers.SMTPHost)
	cyan.Printf("  SMTP Port: %d\n", port)
	cyan.Printf("  Username: %s\n", answers.Username)
	cyan.Printf("  Password: *****\n")
	cyan.Printf("  From Email: %s\n", answers.FromEmail)
	cyan.Printf("  Use TLS: %v\n", useTLS)
	cyan.Printf("  Use SSL: %v\n", useSSL)
	cyan.Printf("  Timeout: %d segundos\n", timeout)

	// Confirmação
	var confirm bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Confirmar e salvar?",
		Default: true,
	}, &confirm); err != nil {
		return err
	}

	if !confirm {
		yellow := color.New(color.FgYellow)
		yellow.Println("Operação cancelada")
		return nil
	}

	// Salva
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("erro ao salvar: %w", err)
	}

	green := color.New(color.FgHiGreen, color.Bold)
	green.Println("✓ Configuração do Email salva com sucesso")

	return nil
}

// runWhatsAppWizard e runGoogleChatWizard são placeholders para Fase 03.
func runWhatsAppWizard(cfg *config.Config) error {
	return fmt.Errorf("wizard do WhatsApp ainda não implementado (Fase 03)")
}

func runGoogleChatWizard(cfg *config.Config) error {
	return fmt.Errorf("wizard do Google Chat ainda não implementado (Fase 03)")
}

// addTelegramViaFlags adiciona Telegram via flags.
func addTelegramViaFlags(cmd *cobra.Command, cfg *config.Config) error {
	token, _ := cmd.Flags().GetString("token")
	chatID, _ := cmd.Flags().GetString("default-chat-id")
	timeout, _ := cmd.Flags().GetInt("timeout")

	if token == "" {
		return fmt.Errorf("token é obrigatório (use --token)")
	}

	if timeout == 0 {
		timeout = 30
	}

	cfg.Telegram.Token = token
	cfg.Telegram.DefaultChatID = chatID
	cfg.Telegram.Timeout = timeout

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("erro ao salvar: %w", err)
	}

	green := color.New(color.FgHiGreen, color.Bold)
	green.Println("✓ Configuração do Telegram salva com sucesso")

	return nil
}

// addEmailViaFlags adiciona Email via flags.
func addEmailViaFlags(cmd *cobra.Command, cfg *config.Config) error {
	smtpHost, _ := cmd.Flags().GetString("smtp-host")
	smtpPort, _ := cmd.Flags().GetInt("smtp-port")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	fromEmail, _ := cmd.Flags().GetString("from-email")
	fromName, _ := cmd.Flags().GetString("from-name")
	useTLS, _ := cmd.Flags().GetBool("use-tls")
	useSSL, _ := cmd.Flags().GetBool("use-ssl")
	timeout, _ := cmd.Flags().GetInt("timeout")

	if smtpHost == "" || username == "" || password == "" {
		return fmt.Errorf("smtp-host, username e password são obrigatórios")
	}

	if smtpPort == 0 {
		if useSSL {
			smtpPort = 465
		} else {
			smtpPort = 587
		}
	}

	if timeout == 0 {
		timeout = 30
	}

	if !useTLS && !useSSL {
		useTLS = true
	}

	cfg.Email.SMTPHost = smtpHost
	cfg.Email.SMTPPort = smtpPort
	cfg.Email.Username = username
	cfg.Email.Password = password
	cfg.Email.FromEmail = fromEmail
	cfg.Email.FromName = fromName
	cfg.Email.UseTLS = useTLS
	cfg.Email.UseSSL = useSSL
	cfg.Email.Timeout = timeout

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("erro ao salvar: %w", err)
	}

	green := color.New(color.FgHiGreen, color.Bold)
	green.Println("✓ Configuração do Email salva com sucesso")

	return nil
}

// Funções auxiliares para mostrar configurações
func showTelegramConfig(cfg config.TelegramConfig, mask bool) {
	cyan := color.New(color.FgCyan)
	cyan.Println("Telegram:")
	if cfg.Token != "" {
		if mask {
			cyan.Printf("  Token: %s\n", maskToken(cfg.Token))
		} else {
			cyan.Printf("  Token: %s\n", cfg.Token)
		}
	}
	cyan.Printf("  Default Chat ID: %s\n", cfg.DefaultChatID)
	cyan.Printf("  API URL: %s\n", cfg.APIURL)
	cyan.Printf("  Timeout: %d segundos\n", cfg.Timeout)
}

func showEmailConfig(cfg config.EmailConfig, mask bool) {
	cyan := color.New(color.FgCyan)
	cyan.Println("Email:")
	cyan.Printf("  SMTP Host: %s\n", cfg.SMTPHost)
	cyan.Printf("  SMTP Port: %d\n", cfg.SMTPPort)
	cyan.Printf("  Username: %s\n", cfg.Username)
	if mask {
		cyan.Println("  Password: *****")
	} else {
		cyan.Printf("  Password: %s\n", cfg.Password)
	}
	cyan.Printf("  From Email: %s\n", cfg.FromEmail)
	cyan.Printf("  From Name: %s\n", cfg.FromName)
	cyan.Printf("  Use TLS: %v\n", cfg.UseTLS)
	cyan.Printf("  Use SSL: %v\n", cfg.UseSSL)
	cyan.Printf("  Timeout: %d segundos\n", cfg.Timeout)
}

func showWhatsAppConfig(cfg config.WhatsAppConfig, mask bool) {
	cyan := color.New(color.FgCyan)
	cyan.Println("WhatsApp:")
	cyan.Printf("  Phone Number ID: %s\n", cfg.PhoneNumberID)
	if mask {
		cyan.Println("  Access Token: *****")
	} else {
		cyan.Printf("  Access Token: %s\n", cfg.AccessToken)
	}
	cyan.Printf("  Business Account ID: %s\n", cfg.BusinessAccountID)
	cyan.Printf("  API Version: %s\n", cfg.APIVersion)
	cyan.Printf("  API URL: %s\n", cfg.APIURL)
	cyan.Printf("  Timeout: %d segundos\n", cfg.Timeout)
}

func showGoogleChatConfig(cfg config.GoogleChatConfig, mask bool) {
	cyan := color.New(color.FgCyan)
	cyan.Println("Google Chat:")
	if mask {
		cyan.Println("  Webhook URL: *****")
	} else {
		cyan.Printf("  Webhook URL: %s\n", cfg.WebhookURL)
	}
	cyan.Printf("  Timeout: %d segundos\n", cfg.Timeout)
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "*****"
	}
	return token[:4] + "*****" + token[len(token)-4:]
}

// updateTelegramViaFlags atualiza Telegram via flags (apenas campos fornecidos).
func updateTelegramViaFlags(cmd *cobra.Command, cfg *config.Config) error {
	token, _ := cmd.Flags().GetString("token")
	chatID, _ := cmd.Flags().GetString("default-chat-id")
	timeout, _ := cmd.Flags().GetInt("timeout")

	// Atualiza apenas campos fornecidos
	if cmd.Flags().Changed("token") {
		cfg.Telegram.Token = token
	}
	if cmd.Flags().Changed("default-chat-id") {
		cfg.Telegram.DefaultChatID = chatID
	}
	if cmd.Flags().Changed("timeout") && timeout > 0 {
		cfg.Telegram.Timeout = timeout
	}

	return nil
}

// updateEmailViaFlags atualiza Email via flags (apenas campos fornecidos).
func updateEmailViaFlags(cmd *cobra.Command, cfg *config.Config) error {
	smtpHost, _ := cmd.Flags().GetString("smtp-host")
	smtpPort, _ := cmd.Flags().GetInt("smtp-port")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	fromEmail, _ := cmd.Flags().GetString("from-email")
	fromName, _ := cmd.Flags().GetString("from-name")
	useTLS, _ := cmd.Flags().GetBool("use-tls")
	useSSL, _ := cmd.Flags().GetBool("use-ssl")
	timeout, _ := cmd.Flags().GetInt("timeout")

	// Atualiza apenas campos fornecidos
	if cmd.Flags().Changed("smtp-host") {
		cfg.Email.SMTPHost = smtpHost
	}
	if cmd.Flags().Changed("smtp-port") && smtpPort > 0 {
		cfg.Email.SMTPPort = smtpPort
	}
	if cmd.Flags().Changed("username") {
		cfg.Email.Username = username
	}
	if cmd.Flags().Changed("password") {
		cfg.Email.Password = password
	}
	if cmd.Flags().Changed("from-email") {
		cfg.Email.FromEmail = fromEmail
	}
	if cmd.Flags().Changed("from-name") {
		cfg.Email.FromName = fromName
	}
	if cmd.Flags().Changed("use-tls") {
		cfg.Email.UseTLS = useTLS
	}
	if cmd.Flags().Changed("use-ssl") {
		cfg.Email.UseSSL = useSSL
	}
	if cmd.Flags().Changed("timeout") && timeout > 0 {
		cfg.Email.Timeout = timeout
	}

	return nil
}

// testTelegram testa conectividade do Telegram chamando getMe.
func testTelegram(cfg config.TelegramConfig) error {
	if cfg.Token == "" {
		red := color.New(color.FgRed, color.Bold)
		red.Println("✗ Telegram não está configurado")
		return fmt.Errorf("telegram não está configurado")
	}

	apiURL := cfg.APIURL
	if apiURL == "" {
		apiURL = "https://api.telegram.org"
	}

	url := fmt.Sprintf("%s/bot%s/getMe", apiURL, cfg.Token)

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		red := color.New(color.FgRed, color.Bold)
		red.Printf("✗ Erro ao criar requisição: %v\n", err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		red := color.New(color.FgRed, color.Bold)
		red.Printf("✗ Erro de conectividade: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		red := color.New(color.FgRed, color.Bold)
		red.Printf("✗ Erro na API: Status %d\n", resp.StatusCode)
		return fmt.Errorf("erro na API: status %d", resp.StatusCode)
	}

	green := color.New(color.FgHiGreen, color.Bold)
	green.Printf("✓ Conectividade OK (%dms)\n", latency.Milliseconds())

	return nil
}

// testEmail testa conectividade SMTP sem enviar email.
func testEmail(cfg config.EmailConfig, target string) error {
	if cfg.SMTPHost == "" || cfg.Username == "" || cfg.Password == "" {
		red := color.New(color.FgRed, color.Bold)
		red.Println("✗ Email não está configurado")
		return fmt.Errorf("email não está configurado")
	}

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	if cfg.SMTPPort == 0 {
		if cfg.UseSSL {
			addr = fmt.Sprintf("%s:465", cfg.SMTPHost)
		} else {
			addr = fmt.Sprintf("%s:587", cfg.SMTPHost)
		}
	}

	start := time.Now()

	// Conecta ao SMTP
	var conn *smtp.Client
	var err error

	if cfg.UseSSL {
		// SSL direto (porta 465)
		tlsConfig := &tls.Config{
			ServerName: cfg.SMTPHost,
		}
		tlsConn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Printf("✗ Erro ao conectar (SSL): %v\n", err)
			return err
		}
		defer tlsConn.Close()

		conn, err = smtp.NewClient(tlsConn, cfg.SMTPHost)
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Printf("✗ Erro ao criar cliente SMTP: %v\n", err)
			return err
		}
	} else {
		// TLS (porta 587)
		conn, err = smtp.Dial(addr)
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Printf("✗ Erro ao conectar: %v\n", err)
			return err
		}
	}
	defer conn.Close()

	// EHLO
	if err := conn.Hello("localhost"); err != nil {
		red := color.New(color.FgRed, color.Bold)
		red.Printf("✗ Erro no EHLO: %v\n", err)
		return err
	}

	// StartTLS se necessário
	if cfg.UseTLS && !cfg.UseSSL {
		if ok, _ := conn.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: cfg.SMTPHost,
			}
			if err := conn.StartTLS(tlsConfig); err != nil {
				red := color.New(color.FgRed, color.Bold)
				red.Printf("✗ Erro no StartTLS: %v\n", err)
				return err
			}
		}
	}

	// Autenticação
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
	if err := conn.Auth(auth); err != nil {
		red := color.New(color.FgRed, color.Bold)
		red.Printf("✗ Erro na autenticação: %v\n", err)
		return err
	}

	// QUIT
	if err := conn.Quit(); err != nil {
		red := color.New(color.FgRed, color.Bold)
		red.Printf("✗ Erro ao fechar conexão: %v\n", err)
		return err
	}

	latency := time.Since(start)

	green := color.New(color.FgHiGreen, color.Bold)
	green.Printf("✓ Conectividade OK (%dms)\n", latency.Milliseconds())

	// Se target foi fornecido, envia email de teste
	if target != "" {
		yellow := color.New(color.FgYellow)
		yellow.Println("⚠ Envio de email de teste não implementado ainda")
		// TODO: Implementar envio de email de teste
	}

	return nil
}

// testGoogleChat testa webhook do Google Chat.
func testGoogleChat(cfg config.GoogleChatConfig, target string) error {
	if cfg.WebhookURL == "" {
		red := color.New(color.FgRed, color.Bold)
		red.Println("✗ Google Chat não está configurado")
		return fmt.Errorf("google chat não está configurado")
	}

	// Valida formato da URL
	if !strings.HasPrefix(cfg.WebhookURL, "https://chat.googleapis.com") {
		red := color.New(color.FgRed, color.Bold)
		red.Println("✗ URL do webhook inválida (deve começar com https://chat.googleapis.com)")
		return fmt.Errorf("url do webhook inválida")
	}

	// Se target foi fornecido, envia mensagem de teste
	if target != "" {
		yellow := color.New(color.FgYellow)
		yellow.Println("⚠ Envio de mensagem de teste não implementado ainda")
		// TODO: Implementar envio de mensagem de teste
	} else {
		green := color.New(color.FgHiGreen, color.Bold)
		green.Println("✓ URL do webhook válida")
	}

	return nil
}
