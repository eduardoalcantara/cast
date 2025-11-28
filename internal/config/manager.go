package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/spf13/viper"
)

// Save salva a configuração no disco.
// Se um arquivo de configuração já existe, usa o mesmo formato.
// Se não existe, cria em YAML (padrão).
// Usa permissão 0600 (apenas leitura/escrita para o dono) por segurança.
func Save(cfg *Config) error {
	// Aplica defaults antes de salvar
	cfg.applyDefaults()

	// Determina o arquivo e formato
	configFile := viper.ConfigFileUsed()
	format := "yaml"

	if configFile != "" {
		// Arquivo existe, usa o mesmo formato
		ext := filepath.Ext(configFile)
		switch ext {
		case ".yaml", ".yml":
			format = "yaml"
		case ".json":
			format = "json"
		case ".properties":
			format = "properties"
		default:
			format = "yaml"
		}
	} else {
		// Arquivo não existe, cria em YAML no diretório atual
		configFile = "cast.yaml"
		// Se estiver em teste, usa o diretório atual
		if wd, err := os.Getwd(); err == nil {
			configFile = filepath.Join(wd, "cast.yaml")
		}
	}

	// Garante que mapas vazios sejam inicializados
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]AliasConfig)
	}

	// Salva baseado no formato
	switch format {
	case "yaml":
		return saveYAML(cfg, configFile)
	case "json":
		return saveJSON(cfg, configFile)
	case "properties":
		return saveProperties(cfg, configFile)
	default:
		return fmt.Errorf("formato não suportado: %s", format)
	}
}

// saveYAML salva a configuração em formato YAML.
func saveYAML(cfg *Config, filename string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("erro ao serializar YAML: %w", err)
	}

	// Garante que o diretório existe
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório: %w", err)
		}
	}

	// Escreve em arquivo temporário primeiro (atomicidade)
	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("erro ao escrever arquivo temporário: %w", err)
	}

	// Renomeia para o arquivo final
	if err := os.Rename(tmpFile, filename); err != nil {
		os.Remove(tmpFile) // Limpa em caso de erro
		return fmt.Errorf("erro ao renomear arquivo: %w", err)
	}

	return nil
}

// saveJSON salva a configuração em formato JSON.
func saveJSON(cfg *Config, filename string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar JSON: %w", err)
	}

	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("erro ao escrever arquivo temporário: %w", err)
	}

	if err := os.Rename(tmpFile, filename); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("erro ao renomear arquivo: %w", err)
	}

	return nil
}

// saveProperties salva a configuração em formato Properties.
// Properties é mais limitado, então vamos converter a estrutura para formato chave=valor.
func saveProperties(cfg *Config, filename string) error {
	// Properties é mais complexo para estruturas aninhadas
	// Por enquanto, vamos apenas retornar erro informando que não está implementado
	// ou podemos usar uma biblioteca de properties
	return fmt.Errorf("salvar em formato properties ainda não está implementado (use YAML ou JSON)")
}

// MergeConfig faz merge profundo de duas configurações.
// Campos presentes em source sobrescrevem os de dest.
// Campos ausentes em source são mantidos em dest.
func MergeConfig(source, dest *Config) {
	// Merge Telegram
	if source.Telegram.Token != "" {
		dest.Telegram.Token = source.Telegram.Token
	}
	if source.Telegram.DefaultChatID != "" {
		dest.Telegram.DefaultChatID = source.Telegram.DefaultChatID
	}
	if source.Telegram.APIURL != "" {
		dest.Telegram.APIURL = source.Telegram.APIURL
	}
	if source.Telegram.Timeout > 0 {
		dest.Telegram.Timeout = source.Telegram.Timeout
	}

	// Merge WhatsApp
	if source.WhatsApp.PhoneNumberID != "" {
		dest.WhatsApp.PhoneNumberID = source.WhatsApp.PhoneNumberID
	}
	if source.WhatsApp.AccessToken != "" {
		dest.WhatsApp.AccessToken = source.WhatsApp.AccessToken
	}
	if source.WhatsApp.BusinessAccountID != "" {
		dest.WhatsApp.BusinessAccountID = source.WhatsApp.BusinessAccountID
	}
	if source.WhatsApp.APIVersion != "" {
		dest.WhatsApp.APIVersion = source.WhatsApp.APIVersion
	}
	if source.WhatsApp.APIURL != "" {
		dest.WhatsApp.APIURL = source.WhatsApp.APIURL
	}
	if source.WhatsApp.Timeout > 0 {
		dest.WhatsApp.Timeout = source.WhatsApp.Timeout
	}

	// Merge Email
	if source.Email.SMTPHost != "" {
		dest.Email.SMTPHost = source.Email.SMTPHost
	}
	if source.Email.SMTPPort > 0 {
		dest.Email.SMTPPort = source.Email.SMTPPort
	}
	if source.Email.Username != "" {
		dest.Email.Username = source.Email.Username
	}
	if source.Email.Password != "" {
		dest.Email.Password = source.Email.Password
	}
	if source.Email.FromEmail != "" {
		dest.Email.FromEmail = source.Email.FromEmail
	}
	if source.Email.FromName != "" {
		dest.Email.FromName = source.Email.FromName
	}
	// UseTLS e UseSSL são booleanos, então verificamos se foram explicitamente definidos
	// Como não temos um campo "definido", vamos assumir que se o valor em source é diferente do default, foi definido
	if source.Email.UseTLS {
		dest.Email.UseTLS = source.Email.UseTLS
	}
	if source.Email.UseSSL {
		dest.Email.UseSSL = source.Email.UseSSL
	}
	if source.Email.Timeout > 0 {
		dest.Email.Timeout = source.Email.Timeout
	}

	// Merge Google Chat
	if source.GoogleChat.WebhookURL != "" {
		dest.GoogleChat.WebhookURL = source.GoogleChat.WebhookURL
	}
	if source.GoogleChat.Timeout > 0 {
		dest.GoogleChat.Timeout = source.GoogleChat.Timeout
	}

	// Merge Aliases: novos adicionam, existentes atualizam
	if source.Aliases != nil {
		if dest.Aliases == nil {
			dest.Aliases = make(map[string]AliasConfig)
		}
		for name, alias := range source.Aliases {
			dest.Aliases[name] = alias
		}
	}
}

// BackupConfig cria uma cópia de backup do arquivo de configuração atual.
// Retorna o caminho do arquivo de backup criado.
func BackupConfig() (string, error) {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// Se não há arquivo configurado, tenta encontrar cast.yaml no diretório atual
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("erro ao obter diretório de trabalho: %w", err)
		}
		configFile = filepath.Join(wd, "cast.yaml")
	}

	// Verifica se o arquivo existe
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return "", fmt.Errorf("arquivo de configuração não encontrado: %s", configFile)
	}

	// Cria caminho do backup
	backupFile := configFile + ".bak"

	// Lê conteúdo do arquivo original
	data, err := os.ReadFile(configFile)
	if err != nil {
		return "", fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}

	// Escreve backup
	if err := os.WriteFile(backupFile, data, 0600); err != nil {
		return "", fmt.Errorf("erro ao criar backup: %w", err)
	}

	return backupFile, nil
}
