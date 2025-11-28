package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/eduardoalcantara/cast/internal/config"
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
	Use:   "show",
	Short: "Mostra a configuração completa",
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
	Use:   "validate",
	Short: "Valida a configuração",
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

func init() {
	configShowCmd.Flags().BoolP("mask", "m", true, "Mascara campos sensíveis")
	configShowCmd.Flags().StringP("format", "f", "yaml", "Formato de saída (yaml, json)")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
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
