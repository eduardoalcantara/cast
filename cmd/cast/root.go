package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "cast",
	Short:         "CAST - Ferramenta CLI para envio agnóstico de mensagens",
	SilenceErrors: true, // Erros são tratados pelos comandos com formatação customizada
	Long: `Ferramenta CLI standalone para envio agnóstico de mensagens (Fire & Forget).
Suporta múltiplos canais: Telegram, WhatsApp, Email, Google Chat.`,
	Run: func(cmd *cobra.Command, args []string) {
		ShowRootHelp()
	},
}

// Execute executa o comando raiz.
func Execute() error {
	// Configura help customizado para todos os comandos
	// Deve ser chamado depois que todos os comandos foram adicionados
	setupCustomHelp()

	// Aplica templates em português a todos os comandos antes de executar
	applyPortugueseHelpToCommand(rootCmd)

	// Inicializa comandos padrão do Cobra (completion e help)
	// Isso garante que eles existam antes de traduzir
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()

	// Traduz comandos automáticos do Cobra
	translateCobraBuiltinCommands(rootCmd)

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

// setupCustomHelp configura as funções de help customizadas para todos os comandos.
func setupCustomHelp() {
	// Root command
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowRootHelp()
	})

	// Send command
	sendCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowSendHelp()
	})

	// Alias commands
	aliasCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowAliasHelp()
	})
	aliasAddCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowAliasAddHelp()
	})
	aliasListCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowAliasListHelp()
	})
	aliasRemoveCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowAliasRemoveHelp()
	})
	aliasShowCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowAliasShowHelp()
	})
	aliasUpdateCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowAliasUpdateHelp()
	})

	// Gateway commands
	gatewayCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowGatewayHelp()
	})
	gatewayAddCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowGatewayAddHelp()
	})
	gatewayShowCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowGatewayShowHelp()
	})
	gatewayRemoveCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowGatewayRemoveHelp()
	})
	gatewayUpdateCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowGatewayUpdateHelp()
	})
	gatewayTestCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowGatewayTestHelp()
	})

	// Config commands
	configCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowConfigHelp()
	})
	configShowCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowConfigShowHelp()
	})
	configValidateCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowConfigValidateHelp()
	})
	configExportCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowConfigExportHelp()
	})
	configImportCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowConfigImportHelp()
	})
	configReloadCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowConfigReloadHelp()
	})
	// Adiciona help para config sources (se existir)
	if configSourcesCmd := configCmd.Commands(); configSourcesCmd != nil {
		for _, cmd := range configSourcesCmd {
			if cmd.Name() == "sources" {
				cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
					ShowConfigSourcesHelp()
				})
				break
			}
		}
	}
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

	// Aplica templates aos comandos principais
	applyPortugueseHelpToCommand(sendCmd)
	applyPortugueseHelpToCommand(aliasCmd)
	applyPortugueseHelpToCommand(configCmd)
	applyPortugueseHelpToCommand(gatewayCmd)

	// Traduz mensagens de erro comuns
	cobra.MousetrapHelpText = "Este é um comando de linha de comando. Você precisa executá-lo no terminal."
}

// translateCobraBuiltinCommands traduz os comandos automáticos do Cobra.
func translateCobraBuiltinCommands(cmd *cobra.Command) {
	if cmd == nil {
		return
	}

	// Itera pelos comandos para encontrar completion e help
	for _, subCmd := range cmd.Commands() {
		switch subCmd.Name() {
		case "completion":
			subCmd.Short = "Gera script de autocompletar para o shell especificado"
			subCmd.Long = `Gera script de autocompletar para o shell especificado (bash, zsh, fish, powershell).

Permite usar Tab para completar comandos e flags automaticamente.

Exemplos:
  # Bash (Linux/Mac)
  cast completion bash > /etc/bash_completion.d/cast
  source /etc/bash_completion.d/cast

  # PowerShell (Windows)
  cast completion powershell | Out-File -FilePath $PROFILE

  # Zsh (Mac/Linux)
  cast completion zsh > "${fpath[1]}/_cast"
  source "${fpath[1]}/_cast"`

			// Traduz subcomandos do completion
			for _, completionSubCmd := range subCmd.Commands() {
				switch completionSubCmd.Name() {
				case "bash":
					completionSubCmd.Short = "Gera script de autocompletar para bash"
				case "zsh":
					completionSubCmd.Short = "Gera script de autocompletar para zsh"
				case "fish":
					completionSubCmd.Short = "Gera script de autocompletar para fish"
				case "powershell":
					completionSubCmd.Short = "Gera script de autocompletar para PowerShell"
				}
			}

		case "help":
			subCmd.Short = "Ajuda sobre qualquer comando"
			subCmd.Long = `Ajuda sobre qualquer comando do CAST.

Fornece informações detalhadas sobre comandos, subcomandos e flags.

Exemplos:
  cast help send
  cast help gateway add
  cast gateway add --help`
		}

		// Aplica recursivamente aos subcomandos
		translateCobraBuiltinCommands(subCmd)
	}

	// Traduz flag --help (pode estar em Flags ou PersistentFlags)
	if helpFlag := cmd.Flags().Lookup("help"); helpFlag != nil {
		helpFlag.Usage = "Ajuda para cast"
	}
	if helpFlag := cmd.PersistentFlags().Lookup("help"); helpFlag != nil {
		helpFlag.Usage = "Ajuda para cast"
	}
}

// applyPortugueseHelpToCommand aplica templates em português a um comando e seus subcomandos.
func applyPortugueseHelpToCommand(cmd *cobra.Command) {
	if cmd == nil {
		return
	}

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

	cmd.SetUsageTemplate(usageTemplate)
	cmd.SetHelpTemplate(helpTemplate)

	// Traduz flag --help se existir (pode estar em Flags ou PersistentFlags)
	if helpFlag := cmd.Flags().Lookup("help"); helpFlag != nil {
		helpFlag.Usage = fmt.Sprintf("Ajuda para %s", cmd.Name())
	}
	if helpFlag := cmd.PersistentFlags().Lookup("help"); helpFlag != nil {
		helpFlag.Usage = fmt.Sprintf("Ajuda para %s", cmd.Name())
	}

	// Aplica recursivamente aos subcomandos
	for _, subCmd := range cmd.Commands() {
		applyPortugueseHelpToCommand(subCmd)
	}
}
