package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
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
		provider := args[0]
		target := args[1]
		message := strings.Join(args[2:], " ")

		// Dummy implementation - apenas imprime a mensagem
		// TODO: Implementar lógica de envio real na Fase 02
		// Suporte a múltiplos targets já implementado via ParseTargets
		fmt.Printf("Sending via [%s] to [%s]: [%s]\n", provider, target, message)

		green := color.New(color.FgHiGreen)
		green.Printf("✓ Mensagem enviada com sucesso (dummy)\n")

		return nil
	},
}
