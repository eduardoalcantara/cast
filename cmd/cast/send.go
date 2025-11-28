package main

import (
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Envia uma mensagem através do provider especificado",
	Long: `Envia uma mensagem através do provider especificado (telegram, whatsapp, email, etc).

A ordem de precedência para configuração é:
  1. Variáveis de Ambiente (CAST_*)
  2. Arquivo Local (cast.*)`,
	Example: `  # Telegram (usando alias 'me' definido no config)
  cast send tg me "Deploy finalizado com sucesso"

  # WhatsApp (formato internacional)
  cast send zap 5511999998888 "Alerta: Disco cheio"

  # Email
  cast send mail admin@empresa.com "Bom dia!" "Relatório Diário" c:\rel.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implementar lógica de envio na Fase 02
		return nil
	},
}
