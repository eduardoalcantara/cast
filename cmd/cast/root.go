package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cast",
	Short: "CAST - Ferramenta CLI para envio agnóstico de mensagens",
	Long: `Ferramenta CLI standalone para envio agnóstico de mensagens (Fire & Forget).
Suporta múltiplos canais: Telegram, WhatsApp, Email, Google Chat.`,
	Run: func(cmd *cobra.Command, args []string) {
		printBanner()
		cmd.Help()
	},
}

// Execute executa o comando raiz.
func Execute() error {
	return rootCmd.Execute()
}

// printBanner exibe o banner ASCII do CAST.
func printBanner() {
	green := color.New(color.FgHiGreen, color.Bold)
	green.Println("┏┓┏┓┏┓┏┳┓")
	green.Println("┃ ┣┫┗┓ ┃")
	green.Println("┗┛┛┗┗┛ ┻")
	green.Println("CAST Automates Sending Tasks")
	fmt.Println("2025 Ⓒ Eduardo Alcântara")
}

func init() {
	rootCmd.AddCommand(sendCmd)
	setupPortugueseHelp()
}

// setupPortugueseHelp configura as mensagens de help do Cobra para português.
func setupPortugueseHelp() {
	usageTemplate := `Uso:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [comando]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Exemplos:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Comandos Disponíveis:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Flags Globais:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Comandos Adicionais:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [comando] --help" para mais informações sobre um comando.{{end}}
`

	helpTemplate := `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

	// Aplica templates em português para todos os comandos
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetHelpTemplate(helpTemplate)
	sendCmd.SetUsageTemplate(usageTemplate)
	sendCmd.SetHelpTemplate(helpTemplate)

	// Traduz mensagens de erro comuns
	cobra.MousetrapHelpText = "Este é um comando de linha de comando. Você precisa executá-lo no terminal."
}
