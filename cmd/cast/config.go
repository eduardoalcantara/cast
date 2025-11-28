package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/eduardoalcantara/cast/internal/config"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Comandos gerais de configuração",
	Long: `Gerencia a configuração do CAST.

Exemplos:
  cast config show
  cast config validate`,
}

var configShowCmd = &cobra.Command{
	Use:          "show",
	Short:        "Mostra a configuração completa",
	SilenceUsage: true,
	Long: `Mostra a configuração completa do CAST.

Por padrão, mascara campos sensíveis (tokens, senhas).
Use --mask=false para mostrar valores reais.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mask, _ := cmd.Flags().GetBool("mask")
		format, _ := cmd.Flags().GetString("format")

		// Carrega configuração
		cfg, err := config.LoadConfig()
		if err != nil {
			// Se não existe, mostra mensagem
			yellow := color.New(color.FgYellow)
			yellow.Println("Nenhuma configuração encontrada")
			return nil
		}

		// Cria cópia para mascarar se necessário
		displayCfg := *cfg
		if mask {
			maskSensitiveData(&displayCfg)
		}

		// Formata saída
		switch format {
		case "json":
			data, err := json.MarshalIndent(displayCfg, "", "  ")
			if err != nil {
				return fmt.Errorf("erro ao serializar JSON: %w", err)
			}
			fmt.Println(string(data))
		case "yaml", "":
			data, err := yaml.Marshal(displayCfg)
			if err != nil {
				return fmt.Errorf("erro ao serializar YAML: %w", err)
			}
			fmt.Print(string(data))
		default:
			return fmt.Errorf("formato não suportado: %s (use yaml ou json)", format)
		}

		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "Valida a configuração",
	SilenceUsage: true,
	Long: `Valida a configuração atual do CAST.

Verifica se todos os campos obrigatórios estão preenchidos
e se os valores estão dentro dos limites permitidos.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		// Valida configuração
		if err := cfg.Validate(); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Configuração inválida: %v\n", err)
			return err
		}

		// Mostra resumo
		green := color.New(color.FgHiGreen, color.Bold)
		green.Println("✓ Configuração válida")

		cyan := color.New(color.FgCyan)
		if cfg.Telegram.Token != "" {
			cyan.Println("  - Telegram: configurado")
		}
		if cfg.Email.SMTPHost != "" {
			cyan.Println("  - Email: configurado")
		}
		if cfg.WhatsApp.PhoneNumberID != "" {
			cyan.Println("  - WhatsApp: configurado")
		}
		if cfg.GoogleChat.WebhookURL != "" {
			cyan.Println("  - Google Chat: configurado")
		}
		if cfg.Aliases != nil && len(cfg.Aliases) > 0 {
			cyan.Printf("  - Aliases: %d definidos\n", len(cfg.Aliases))
		}

		return nil
	},
}

var configExportCmd = &cobra.Command{
	Use:          "export",
	Short:        "Exporta a configuração",
	SilenceUsage: true,
	Long: `Exporta a configuração atual do CAST.

Por padrão, imprime YAML no stdout.
Use --output para salvar em arquivo.
Use --force para sobrescrever arquivo existente.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mask, _ := cmd.Flags().GetBool("mask")
		output, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")
		format, _ := cmd.Flags().GetString("format")

		// Carrega configuração
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		// Valida antes de exportar
		if err := cfg.Validate(); err != nil {
			yellow := color.New(color.FgYellow)
			yellow.Printf("⚠ Aviso: Configuração inválida: %v\n", err)
			yellow.Println("Exportando mesmo assim (pode ser útil para debug)...")
		}

		// Cria cópia para mascarar se necessário
		displayCfg := *cfg
		if mask {
			maskSensitiveData(&displayCfg)
		}

		// Determina formato
		if format == "" {
			if output != "" {
				ext := strings.ToLower(filepath.Ext(output))
				switch ext {
				case ".json":
					format = "json"
				case ".yaml", ".yml":
					format = "yaml"
				default:
					format = "yaml"
				}
			} else {
				format = "yaml"
			}
		}

		// Serializa
		var data []byte
		switch format {
		case "json":
			var err error
			data, err = json.MarshalIndent(displayCfg, "", "  ")
			if err != nil {
				return fmt.Errorf("erro ao serializar JSON: %w", err)
			}
		case "yaml", "yml":
			var err error
			data, err = yaml.Marshal(displayCfg)
			if err != nil {
				return fmt.Errorf("erro ao serializar YAML: %w", err)
			}
		default:
			return fmt.Errorf("formato não suportado: %s (use yaml ou json)", format)
		}

		// Escreve saída
		if output != "" {
			// Verifica se arquivo existe
			if _, err := os.Stat(output); err == nil && !force {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Arquivo já existe: %s\n", output)
				red.Println("Use --force para sobrescrever")
				return fmt.Errorf("arquivo já existe: %s", output)
			}

			if err := os.WriteFile(output, data, 0600); err != nil {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro ao salvar arquivo: %v\n", err)
				return err
			}

			green := color.New(color.FgHiGreen, color.Bold)
			green.Printf("✓ Configuração exportada para: %s\n", output)
		} else {
			// Imprime no stdout
			fmt.Print(string(data))
		}

		return nil
	},
}

var configImportCmd = &cobra.Command{
	Use:          "import <arquivo>",
	Short:        "Importa configuração de um arquivo",
	SilenceUsage: true,
	Long: `Importa configuração de um arquivo.

Por padrão, substitui completamente a configuração atual.
Use --merge para mesclar com a configuração existente.
Um backup automático é criado antes da importação.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		importFile := args[0]
		merge, _ := cmd.Flags().GetBool("merge")

		// Verifica se arquivo existe
		if _, err := os.Stat(importFile); os.IsNotExist(err) {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Arquivo não encontrado: %s\n", importFile)
			return fmt.Errorf("arquivo não encontrado: %s", importFile)
		}

		// Detecta formato pela extensão
		ext := strings.ToLower(filepath.Ext(importFile))
		var format string
		switch ext {
		case ".json":
			format = "json"
		case ".yaml", ".yml":
			format = "yaml"
		default:
			format = "yaml" // Assume YAML se não conseguir detectar
		}

		// Lê arquivo
		data, err := os.ReadFile(importFile)
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao ler arquivo: %v\n", err)
			return err
		}

		// Deserializa
		var importedCfg config.Config
		switch format {
		case "json":
			if err := json.Unmarshal(data, &importedCfg); err != nil {
				return fmt.Errorf("erro ao fazer parse JSON: %w", err)
			}
		case "yaml", "yml":
			if err := yaml.Unmarshal(data, &importedCfg); err != nil {
				return fmt.Errorf("erro ao fazer parse YAML: %w", err)
			}
		default:
			return fmt.Errorf("formato não suportado: %s", format)
		}

		// Carrega configuração atual
		currentCfg, err := config.LoadConfig()
		if err != nil {
			// Se não existe, cria nova
			currentCfg = &config.Config{}
		}

		// Faz merge ou substituição
		if merge {
			// Merge profundo
			config.MergeConfig(&importedCfg, currentCfg)
		} else {
			// Substituição total
			currentCfg = &importedCfg
		}

		// Valida antes de salvar
		if err := currentCfg.Validate(); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Configuração inválida após importação: %v\n", err)
			red.Println("Operação abortada. Nenhuma alteração foi salva.")
			return fmt.Errorf("configuração inválida: %w", err)
		}

		// Cria backup antes de salvar
		backupFile, err := config.BackupConfig()
		if err != nil {
			yellow := color.New(color.FgYellow)
			yellow.Printf("⚠ Aviso: Não foi possível criar backup: %v\n", err)
			yellow.Println("Continuando mesmo assim...")
		} else {
			cyan := color.New(color.FgCyan)
			cyan.Printf("✓ Backup criado: %s\n", backupFile)
		}

		// Salva configuração
		if err := config.Save(currentCfg); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao salvar configuração: %v\n", err)
			return err
		}

		green := color.New(color.FgHiGreen, color.Bold)
		if merge {
			green.Println("✓ Configuração importada e mesclada com sucesso")
		} else {
			green.Println("✓ Configuração importada e substituída com sucesso")
		}

		return nil
	},
}

var configReloadCmd = &cobra.Command{
	Use:          "reload",
	Short:        "Recarrega a configuração do disco",
	SilenceUsage: true,
	Long: `Recarrega a configuração do arquivo do disco.

Útil para verificar se o arquivo é legível após edição manual.
Valida a configuração após recarregar.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Força releitura do arquivo
		// Limpa configuração do Viper
		viper.Reset()

		// Recarrega usando Load()
		if err := config.Load(); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao recarregar configuração: %v\n", err)
			return err
		}

		// Carrega e valida
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		// Valida
		if err := cfg.Validate(); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Configuração inválida: %v\n", err)
			return err
		}

		green := color.New(color.FgHiGreen, color.Bold)
		green.Println("✓ Configuração recarregada e válida")

		return nil
	},
}

func init() {
	configShowCmd.Flags().BoolP("mask", "m", true, "Mascara campos sensíveis")
	configShowCmd.Flags().StringP("format", "f", "yaml", "Formato de saída (yaml, json)")

	configExportCmd.Flags().BoolP("mask", "m", true, "Mascara campos sensíveis")
	configExportCmd.Flags().StringP("output", "o", "", "Arquivo de saída (padrão: stdout)")
	configExportCmd.Flags().BoolP("force", "f", false, "Sobrescreve arquivo existente")
	configExportCmd.Flags().String("format", "", "Formato de saída (yaml, json). Auto-detecta pela extensão se --output for usado")

	configImportCmd.Flags().BoolP("merge", "m", false, "Mescla com configuração existente ao invés de substituir")

	var configSourcesCmd = &cobra.Command{
		Use:          "sources",
		Short:        "Mostra a origem de cada configuração",
		SilenceUsage: true,
		Long: `Mostra de onde vem cada item de configuração (arquivo ou variável de ambiente).

A ordem de precedência é:
  1. Variáveis de Ambiente (CAST_*) - maior prioridade
  2. Arquivo de Configuração (cast.yaml/json/properties)
  3. Valores Padrão

Este comando ajuda a identificar onde cada configuração está sendo definida.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			showConfigSources()
			return nil
		},
	}

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configExportCmd)
	configCmd.AddCommand(configImportCmd)
	configCmd.AddCommand(configReloadCmd)
	configCmd.AddCommand(configSourcesCmd)
	rootCmd.AddCommand(configCmd)
}

// maskSensitiveData mascara campos sensíveis na configuração.
func maskSensitiveData(cfg *config.Config) {
	// Telegram
	if cfg.Telegram.Token != "" {
		cfg.Telegram.Token = "*****"
	}

	// WhatsApp
	if cfg.WhatsApp.AccessToken != "" {
		cfg.WhatsApp.AccessToken = "*****"
	}

	// Email
	if cfg.Email.Password != "" {
		cfg.Email.Password = "*****"
	}

	// Google Chat - Webhook URL pode conter tokens, mas não vamos mascarar por padrão
	// pois é necessário para debug
}

// showConfigSources mostra a origem de cada configuração.
func showConfigSources() {
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgHiGreen)
	yellow := color.New(color.FgYellow)
	white := color.New(color.FgWhite)

	// Carrega configuração
	cfg, err := config.LoadConfig()
	if err != nil {
		yellow.Println("Nenhuma configuração encontrada")
		return
	}

	// Mostra arquivo de configuração
	configFile := config.GetConfigFile()
	fmt.Println()
	cyan.Println("=== ORIGEM DAS CONFIGURAÇÕES ===")
	fmt.Println()
	if configFile != "" {
		absPath, _ := filepath.Abs(configFile)
		green.Printf("Arquivo de Configuração: %s\n", absPath)
	} else {
		yellow.Println("Arquivo de Configuração: Não encontrado")
	}
	fmt.Println()

	// Função auxiliar para verificar origem
	getSource := func(key string) string {
		// Verifica se está em ENV (prioridade)
		envKey := "CAST_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if envVal := os.Getenv(envKey); envVal != "" {
			return "ENV"
		}
		// Verifica se está no arquivo usando viper.IsSet
		// viper.IsSet retorna true se a chave foi definida no arquivo ou ENV
		// Como já verificamos ENV acima, se IsSet retorna true, está no arquivo
		if viper.IsSet(key) {
			return "FILE"
		}
		// Se não está definido em nenhum lugar, pode ser default ou não definido
		// Para valores booleanos false, precisamos verificar se foi explicitamente definido
		// Mas como não temos como distinguir, vamos assumir que se não está setado, é DEFAULT ou N/A
		return "DEFAULT"
	}

	// Telegram
	cyan.Println("Telegram:")
	showSource("  token", cfg.Telegram.Token, getSource("telegram.token"))
	showSource("  default_chat_id", cfg.Telegram.DefaultChatID, getSource("telegram.default_chat_id"))
	showSource("  api_url", cfg.Telegram.APIURL, getSource("telegram.api_url"))
	showSource("  timeout", fmt.Sprintf("%d", cfg.Telegram.Timeout), getSource("telegram.timeout"))
	fmt.Println()

	// WhatsApp
	cyan.Println("WhatsApp:")
	showSource("  phone_number_id", cfg.WhatsApp.PhoneNumberID, getSource("whatsapp.phone_number_id"))
	// Usa função maskToken de gateway.go (mesmo pacote)
	maskedToken := cfg.WhatsApp.AccessToken
	if len(maskedToken) > 8 {
		maskedToken = maskedToken[:4] + "*****" + maskedToken[len(maskedToken)-4:]
	} else if maskedToken != "" {
		maskedToken = "*****"
	}
	showSource("  access_token", maskedToken, getSource("whatsapp.access_token"))
	showSource("  business_account_id", cfg.WhatsApp.BusinessAccountID, getSource("whatsapp.business_account_id"))
	showSource("  api_version", cfg.WhatsApp.APIVersion, getSource("whatsapp.api_version"))
	showSource("  api_url", cfg.WhatsApp.APIURL, getSource("whatsapp.api_url"))
	showSource("  timeout", fmt.Sprintf("%d", cfg.WhatsApp.Timeout), getSource("whatsapp.timeout"))
	fmt.Println()

	// Email
	cyan.Println("Email:")
	showSource("  smtp_host", cfg.Email.SMTPHost, getSource("email.smtp_host"))
	showSource("  smtp_port", fmt.Sprintf("%d", cfg.Email.SMTPPort), getSource("email.smtp_port"))
	showSource("  username", cfg.Email.Username, getSource("email.username"))
	// Usa função maskToken de gateway.go (mesmo pacote)
	maskedPassword := cfg.Email.Password
	if len(maskedPassword) > 8 {
		maskedPassword = maskedPassword[:4] + "*****" + maskedPassword[len(maskedPassword)-4:]
	} else if maskedPassword != "" {
		maskedPassword = "*****"
	}
	showSource("  password", maskedPassword, getSource("email.password"))
	showSource("  from_email", cfg.Email.FromEmail, getSource("email.from_email"))
	showSource("  from_name", cfg.Email.FromName, getSource("email.from_name"))
	showSource("  use_tls", fmt.Sprintf("%v", cfg.Email.UseTLS), getSource("email.use_tls"))
	showSource("  use_ssl", fmt.Sprintf("%v", cfg.Email.UseSSL), getSource("email.use_ssl"))
	showSource("  timeout", fmt.Sprintf("%d", cfg.Email.Timeout), getSource("email.timeout"))
	fmt.Println()

	// Google Chat
	cyan.Println("Google Chat:")
	showSource("  webhook_url", cfg.GoogleChat.WebhookURL, getSource("google_chat.webhook_url"))
	showSource("  timeout", fmt.Sprintf("%d", cfg.GoogleChat.Timeout), getSource("google_chat.timeout"))
	fmt.Println()

	// Aliases
	if cfg.Aliases != nil && len(cfg.Aliases) > 0 {
		cyan.Println("Aliases:")
		for name, alias := range cfg.Aliases {
			white.Printf("  %s:\n", name)
			white.Printf("    provider: %s (FILE)\n", alias.Provider)
			white.Printf("    target: %s (FILE)\n", alias.Target)
			if alias.Name != "" {
				white.Printf("    name: %s (FILE)\n", alias.Name)
			}
		}
		fmt.Println()
	}

	// Legenda
	cyan.Println("Legenda:")
	green.Println("  ENV     - Variável de Ambiente (CAST_*)")
	white.Println("  FILE    - Arquivo de Configuração")
	yellow.Println("  DEFAULT - Valor Padrão")
	fmt.Println()
}

// showSource mostra uma linha de origem de configuração.
func showSource(key, value, source string) {
	white := color.New(color.FgWhite)
	green := color.New(color.FgHiGreen)
	yellow := color.New(color.FgYellow)

	var sourceColor *color.Color
	switch source {
	case "ENV":
		sourceColor = green
	case "FILE":
		sourceColor = white
	case "DEFAULT":
		sourceColor = yellow
	default:
		sourceColor = white
	}

	// Para valores booleanos, não mascarar "false" como "não definido"
	// false é um valor válido e explícito
	if value == "" || value == "0" {
		value = "(não definido)"
	}
	// Para booleanos, manter "false" e "true" como estão

	white.Printf("%-25s = %-30s [", key, value)
	sourceColor.Print(source)
	white.Println("]")
}
