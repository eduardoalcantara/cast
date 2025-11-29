package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/eduardoalcantara/cast/internal/config"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Gerencia aliases (atalhos para provider + target)",
	Long: `Gerencia aliases que permitem usar nomes curtos no lugar de provider e target.

Exemplos:
  cast alias add me tg "123456789" --name "Meu Telegram"
  cast alias list
  cast alias remove me`,
}

var aliasAddCmd = &cobra.Command{
	Use:          "add <nome> <provider> <target>",
	Short:        "Adiciona um novo alias",
	SilenceUsage: true,
	Long: `Adiciona um novo alias que mapeia um nome para um provider e target.

Argumentos:
  nome     - Nome do alias (ex: me, team, alerts)
  provider - Provider (tg, mail, zap, google_chat)
  target   - Target (chat_id, email, número, webhook_url)`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliasName := args[0]
		provider := args[1]
		target := args[2]
		description, _ := cmd.Flags().GetString("name")

		// Carrega configuração existente
		cfg, err := config.LoadConfig()
		if err != nil {
			// Se não existe, cria nova
			cfg = &config.Config{}
		}

		// Inicializa map se necessário
		if cfg.Aliases == nil {
			cfg.Aliases = make(map[string]config.AliasConfig)
		}

		// Valida se alias já existe
		if existing := cfg.GetAlias(aliasName); existing != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro: Alias '%s' já existe\n", aliasName)
			return fmt.Errorf("alias '%s' já existe", aliasName)
		}

		// Valida provider
		normalizedProvider := normalizeProviderName(provider)
		if normalizedProvider == "" {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro: Provider '%s' inválido\n", provider)
			return fmt.Errorf("provider '%s' inválido (suportados: tg, mail, zap, google_chat)", provider)
		}

		// Valida target
		if target == "" {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro: Target não pode estar vazio\n")
			return fmt.Errorf("target não pode estar vazio")
		}

		// Adiciona alias
		cfg.Aliases[aliasName] = config.AliasConfig{
			Provider: normalizedProvider,
			Target:   target,
			Name:     description,
		}

		// Salva configuração
		if err := config.Save(cfg); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao salvar configuração: %v\n", err)
			return err
		}

		green := color.New(color.FgHiGreen, color.Bold)
		green.Printf("✓ Alias '%s' adicionado com sucesso\n", aliasName)

		return nil
	},
}

var aliasListCmd = &cobra.Command{
	Use:          "list",
	Short:        "Lista todos os aliases",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			// Se não existe, mostra mensagem
			yellow := color.New(color.FgYellow)
			yellow.Println("Nenhum alias configurado")
			return nil
		}

		if cfg.Aliases == nil || len(cfg.Aliases) == 0 {
			yellow := color.New(color.FgYellow)
			yellow.Println("Nenhum alias configurado")
			return nil
		}

		// Imprime cabeçalho
		cyan := color.New(color.FgCyan, color.Bold)
		cyan.Printf("%-15s %-10s %-30s %s\n", "Nome", "Provider", "Target", "Descrição")
		fmt.Println(strings.Repeat("-", 80))

		// Imprime aliases
		for name, alias := range cfg.Aliases {
			desc := alias.Name
			if desc == "" {
				desc = "-"
			}
			fmt.Printf("%-15s %-10s %-30s %s\n", name, alias.Provider, alias.Target, desc)
		}
		return nil
	},
}

var aliasRemoveCmd = &cobra.Command{
	Use:          "remove <nome>",
	Short:        "Remove um alias",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliasName := args[0]
		confirm, _ := cmd.Flags().GetBool("confirm")

		// Carrega configuração
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		if cfg.Aliases == nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro: Alias '%s' não encontrado\n", aliasName)
			return fmt.Errorf("alias '%s' não encontrado", aliasName)
		}

		// Verifica se existe
		if cfg.GetAlias(aliasName) == nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro: Alias '%s' não encontrado\n", aliasName)
			return fmt.Errorf("alias '%s' não encontrado", aliasName)
		}

		// Confirmação
		if !confirm {
			yellow := color.New(color.FgYellow)
			yellow.Printf("Tem certeza que deseja remover o alias '%s'? (s/N): ", aliasName)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "s" && strings.ToLower(response) != "sim" {
				cyan := color.New(color.FgCyan)
				cyan.Println("Operação cancelada")
				return nil
			}
		}

		// Remove alias
		delete(cfg.Aliases, aliasName)

		// Salva configuração
		if err := config.Save(cfg); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao salvar configuração: %v\n", err)
			return err
		}

		green := color.New(color.FgHiGreen, color.Bold)
		green.Printf("✓ Alias '%s' removido com sucesso\n", aliasName)

		return nil
	},
}

var aliasShowCmd = &cobra.Command{
	Use:          "show <nome>",
	Short:        "Mostra detalhes de um alias",
	SilenceUsage: true,
	Long:         `Mostra detalhes de um alias específico em formato "Ficha Técnica".`,
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliasName := args[0]

		// Carrega configuração
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		// Busca alias
		alias := cfg.GetAlias(aliasName)
		if alias == nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Alias '%s' não encontrado\n", aliasName)
			return fmt.Errorf("alias '%s' não encontrado", aliasName)
		}

		// Formata provider name para exibição
		providerDisplay := alias.Provider
		switch alias.Provider {
		case "tg":
			providerDisplay = "tg (Telegram)"
		case "mail":
			providerDisplay = "mail (Email)"
		case "zap":
			providerDisplay = "zap (WhatsApp)"
		case "google_chat":
			providerDisplay = "google_chat (Google Chat)"
		}

		// Exibe em formato "Ficha Técnica"
		cyan := color.New(color.FgCyan)
		cyan.Printf("Alias:      %s\n", aliasName)
		cyan.Printf("Provider:   %s\n", providerDisplay)
		cyan.Printf("Target:     %s\n", alias.Target)
		if alias.Name != "" {
			cyan.Printf("Descrição:  %s\n", alias.Name)
		} else {
			cyan.Println("Descrição:  -")
		}

		return nil
	},
}

var aliasUpdateCmd = &cobra.Command{
	Use:          "update <nome>",
	Short:        "Atualiza um alias existente",
	SilenceUsage: true,
	Long: `Atualiza um alias existente.

Permite atualização parcial: apenas os campos fornecidos nas flags são atualizados.
Mantém os outros campos intactos.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliasName := args[0]

		// Carrega configuração
		cfg, err := config.LoadConfig()
		if err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao carregar configuração: %v\n", err)
			return err
		}

		// Verifica se alias existe
		alias := cfg.GetAlias(aliasName)
		if alias == nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Alias '%s' não encontrado\n", aliasName)
			return fmt.Errorf("alias '%s' não encontrado", aliasName)
		}

		// Atualiza apenas campos fornecidos
		provider, _ := cmd.Flags().GetString("provider")
		target, _ := cmd.Flags().GetString("target")
		description, _ := cmd.Flags().GetString("name")

		if cmd.Flags().Changed("provider") {
			normalizedProvider := normalizeProviderName(provider)
			if normalizedProvider == "" {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro: Provider '%s' inválido\n", provider)
				return fmt.Errorf("provider '%s' inválido", provider)
			}
			alias.Provider = normalizedProvider
		}

		if cmd.Flags().Changed("target") {
			if target == "" {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro: Target não pode estar vazio\n")
				return fmt.Errorf("target não pode estar vazio")
			}
			alias.Target = target
		}

		if cmd.Flags().Changed("name") {
			alias.Name = description
		}

		// Valida provider se foi alterado
		if cmd.Flags().Changed("provider") {
			normalizedProvider := normalizeProviderName(alias.Provider)
			if normalizedProvider == "" {
				red := color.New(color.FgRed, color.Bold)
				red.Fprintf(os.Stderr, "✗ Erro: Provider '%s' inválido\n", alias.Provider)
				return fmt.Errorf("provider '%s' inválido", alias.Provider)
			}
		}

		// Atualiza no map
		cfg.Aliases[aliasName] = *alias

		// Salva configuração
		if err := config.Save(cfg); err != nil {
			red := color.New(color.FgRed, color.Bold)
			red.Fprintf(os.Stderr, "✗ Erro ao salvar configuração: %v\n", err)
			return err
		}

		green := color.New(color.FgHiGreen, color.Bold)
		green.Printf("✓ Alias '%s' atualizado com sucesso\n", aliasName)

		return nil
	},
}

func init() {
	aliasAddCmd.Flags().StringP("name", "n", "", "Nome descritivo do alias")
	aliasRemoveCmd.Flags().BoolP("confirm", "y", false, "Confirma sem perguntar")
	aliasUpdateCmd.Flags().StringP("provider", "p", "", "Provider (tg, mail, zap, google_chat)")
	aliasUpdateCmd.Flags().StringP("target", "t", "", "Target (chat_id, email, número, webhook_url)")
	aliasUpdateCmd.Flags().StringP("name", "n", "", "Nome descritivo do alias")

	aliasCmd.AddCommand(aliasAddCmd)
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasRemoveCmd)
	aliasCmd.AddCommand(aliasShowCmd)
	aliasCmd.AddCommand(aliasUpdateCmd)
	rootCmd.AddCommand(aliasCmd)
}

// normalizeProviderName normaliza o nome do provider.
func normalizeProviderName(name string) string {
	switch strings.ToLower(name) {
	case "tg", "telegram":
		return "tg"
	case "mail", "email":
		return "mail"
	case "zap", "whatsapp":
		return "zap"
	case "google_chat", "googlechat":
		return "google_chat"
	case "waha":
		return "waha"
	default:
		return ""
	}
}
