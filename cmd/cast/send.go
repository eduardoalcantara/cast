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
	Use:   "send [provider] [target] [message]",
	Short: "Envia uma mensagem através do provider especificado",
	Long: `Envia uma mensagem através do provider especificado (telegram, whatsapp, email, etc).

A ordem de precedência para configuração é:
  1. Variáveis de Ambiente (CAST_*)
  2. Arquivo Local (cast.*)

Múltiplos Recipientes:
  Você pode enviar para múltiplos recipientes separando-os por vírgula (,) ou ponto-e-vírgula (;):
  - cast send mail "user1@exemplo.com,user2@exemplo.com" "Mensagem"
  - cast send tg "123456789;987654321" "Mensagem"`,
	Example: `  # Telegram (usando alias 'me' definido no config)
  cast send tg me "Deploy finalizado com sucesso"

  # Telegram para múltiplos destinatários
  cast send tg "123456789,987654321" "Mensagem para todos"

  # WhatsApp (formato internacional)
  cast send zap 5511999998888 "Alerta: Disco cheio"

  # Email para um destinatário
  cast send mail admin@empresa.com "Bom dia!"

  # Email para múltiplos destinatários
  cast send mail "admin@empresa.com;dev@empresa.com" "Relatório Diário"`,
	Args: cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]
		target := args[1]
		message := strings.Join(args[2:], " ")

		// Carrega configuração
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return fmt.Errorf("erro de configuração: %w", err)
		}

		// Verifica se é um alias antes de resolver provider
		var actualProviderName string
		var actualTarget string

		if cfg != nil && cfg.Aliases != nil {
			if alias := cfg.GetAlias(providerName); alias != nil {
				// É um alias - usa provider e target do alias
				actualProviderName = alias.Provider
				actualTarget = alias.Target
			} else {
				// Não é alias - usa valores fornecidos
				actualProviderName = providerName
				actualTarget = target
			}
		} else {
			// Sem aliases configurados - usa valores fornecidos
			actualProviderName = providerName
			actualTarget = target
		}

		// Resolve provider via Factory
		provider, err := providers.GetProvider(actualProviderName, cfg)
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao obter provider: %v\n", err)
			return err
		}

		// Envia mensagem
		err = provider.Send(actualTarget, message)
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao enviar mensagem: %v\n", err)
			return err
		}

		// Sucesso
		green := color.New(color.FgHiGreen, color.Bold)
		green.Printf("✓ Mensagem enviada com sucesso via %s\n", provider.Name())

		return nil
	},
}
