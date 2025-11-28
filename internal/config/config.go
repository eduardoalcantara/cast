package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config representa a estrutura de configuração do CAST.
type Config struct {
	Telegram  TelegramConfig              `mapstructure:"telegram"`
	WhatsApp  WhatsAppConfig              `mapstructure:"whatsapp"`
	Email     EmailConfig                 `mapstructure:"email"`
	GoogleChat GoogleChatConfig           `mapstructure:"google_chat"`
	Aliases   map[string]AliasConfig      `mapstructure:"aliases"`
}

// TelegramConfig contém as configurações do Telegram.
type TelegramConfig struct {
	Token        string `mapstructure:"token"`
	DefaultChatID string `mapstructure:"default_chat_id"`
	APIURL       string `mapstructure:"api_url"`
	Timeout      int    `mapstructure:"timeout"`
}

// WhatsAppConfig contém as configurações do WhatsApp (Meta Cloud API).
type WhatsAppConfig struct {
	PhoneNumberID    string `mapstructure:"phone_number_id"`
	AccessToken      string `mapstructure:"access_token"`
	BusinessAccountID string `mapstructure:"business_account_id"`
	APIVersion       string `mapstructure:"api_version"`
	APIURL           string `mapstructure:"api_url"`
	Timeout          int    `mapstructure:"timeout"`
}

// EmailConfig contém as configurações de Email (SMTP).
type EmailConfig struct {
	SMTPHost  string `mapstructure:"smtp_host"`
	SMTPPort  int    `mapstructure:"smtp_port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
	UseTLS    bool   `mapstructure:"use_tls"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Timeout   int    `mapstructure:"timeout"`
}

// GoogleChatConfig contém as configurações do Google Chat.
type GoogleChatConfig struct {
	WebhookURL string `mapstructure:"webhook_url"`
	Timeout    int    `mapstructure:"timeout"`
}

// AliasConfig representa um alias para facilitar o uso do CLI.
type AliasConfig struct {
	Provider string `mapstructure:"provider"`
	Target   string `mapstructure:"target"`
	Name     string `mapstructure:"name"`
}

// Load inicializa e carrega a configuração seguindo a ordem de precedência:
// 1. Variáveis de Ambiente (CAST_*)
// 2. Arquivo Local (cast.*) - suporta .yaml, .json, .properties
func Load() error {
	viper.SetEnvPrefix("CAST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Busca arquivo de configuração no diretório atual
	viper.SetConfigName("cast")
	viper.AddConfigPath(".")

	// Tenta carregar arquivo em múltiplos formatos (ordem: yaml, json, properties)
	configTypes := []string{"yaml", "json", "properties"}
	var lastErr error
	var found bool

	for _, configType := range configTypes {
		viper.SetConfigType(configType)
		if err := viper.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				lastErr = fmt.Errorf("erro ao ler arquivo de configuração (%s): %w", configType, err)
				continue
			}
			// Arquivo não encontrado neste formato, tenta próximo
			continue
		}
		found = true
		break
	}

	// Se nenhum arquivo foi encontrado, não é erro (pode usar apenas ENV)
	if !found && lastErr != nil {
		return lastErr
	}

	return nil
}

// LoadConfig carrega a configuração e retorna a struct Config.
func LoadConfig() (*Config, error) {
	if err := Load(); err != nil {
		return nil, fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("erro ao fazer unmarshal da configuração: %w", err)
	}

	// Aplica valores padrão
	cfg.applyDefaults()

	// Valida configuração
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("erro na validação da configuração: %w", err)
	}

	return &cfg, nil
}

// applyDefaults aplica valores padrão para campos opcionais.
func (c *Config) applyDefaults() {
	// Telegram defaults
	if c.Telegram.APIURL == "" {
		c.Telegram.APIURL = "https://api.telegram.org/bot"
	}
	if c.Telegram.Timeout == 0 {
		c.Telegram.Timeout = 30
	}

	// WhatsApp defaults
	if c.WhatsApp.APIVersion == "" {
		c.WhatsApp.APIVersion = "v18.0"
	}
	if c.WhatsApp.APIURL == "" {
		c.WhatsApp.APIURL = "https://graph.facebook.com"
	}
	if c.WhatsApp.Timeout == 0 {
		c.WhatsApp.Timeout = 30
	}

	// Email defaults
	if c.Email.SMTPPort == 0 {
		if c.Email.UseSSL {
			c.Email.SMTPPort = 465
		} else {
			c.Email.SMTPPort = 587
		}
	}
	if c.Email.FromEmail == "" {
		c.Email.FromEmail = c.Email.Username
	}
	if !c.Email.UseTLS && !c.Email.UseSSL {
		c.Email.UseTLS = true // Padrão TLS
	}
	if c.Email.Timeout == 0 {
		c.Email.Timeout = 30
	}

	// Google Chat defaults
	if c.GoogleChat.Timeout == 0 {
		c.GoogleChat.Timeout = 30
	}
}

// Validate valida a configuração obrigatória.
func (c *Config) Validate() error {
	// Validação de timeout (mínimo 5, máximo 300)
	if c.Telegram.Timeout < 5 || c.Telegram.Timeout > 300 {
		return fmt.Errorf("telegram.timeout deve estar entre 5 e 300 segundos")
	}
	if c.WhatsApp.Timeout < 5 || c.WhatsApp.Timeout > 300 {
		return fmt.Errorf("whatsapp.timeout deve estar entre 5 e 300 segundos")
	}
	if c.Email.Timeout < 5 || c.Email.Timeout > 300 {
		return fmt.Errorf("email.timeout deve estar entre 5 e 300 segundos")
	}
	if c.GoogleChat.Timeout < 5 || c.GoogleChat.Timeout > 300 {
		return fmt.Errorf("google_chat.timeout deve estar entre 5 e 300 segundos")
	}

	// Validação de Email: TLS e SSL são mutuamente exclusivos
	if c.Email.UseTLS && c.Email.UseSSL {
		return fmt.Errorf("email.use_tls e email.use_ssl não podem ser ambos true (priorizando TLS)")
	}

	// Validação de aliases
	for aliasName, alias := range c.Aliases {
		if alias.Provider == "" {
			return fmt.Errorf("alias '%s': provider não pode estar vazio", aliasName)
		}
		if alias.Target == "" {
			return fmt.Errorf("alias '%s': target não pode estar vazio", aliasName)
		}
	}

	return nil
}

// GetAlias retorna um alias pelo nome, ou nil se não existir.
func (c *Config) GetAlias(name string) *AliasConfig {
	if c.Aliases == nil {
		return nil
	}
	alias, exists := c.Aliases[name]
	if !exists {
		return nil
	}
	return &alias
}

// Get retorna o valor da configuração (com fallback para default).
func Get(key string, defaultValue interface{}) interface{} {
	if viper.IsSet(key) {
		return viper.Get(key)
	}
	return defaultValue
}

// GetString retorna o valor da configuração como string.
func GetString(key string, defaultValue string) string {
	return viper.GetString(key)
}

// MustGetString retorna o valor da configuração como string ou sai com erro.
func MustGetString(key string) (string, error) {
	value := viper.GetString(key)
	if value == "" {
		return "", fmt.Errorf("configuração obrigatória não encontrada: %s", key)
	}
	return value, nil
}

// GetConfigFile retorna o caminho do arquivo de configuração carregado (se existir).
func GetConfigFile() string {
	return viper.ConfigFileUsed()
}

// ParseTargets parseia uma string de targets separados por vírgula ou ponto-e-vírgula.
// Retorna um slice de targets limpos (sem espaços).
func ParseTargets(targets string) []string {
	if targets == "" {
		return nil
	}

	// Tenta separar por ponto-e-vírgula primeiro (mais específico)
	separator := ";"
	if !strings.Contains(targets, ";") {
		separator = ","
	}

	parts := strings.Split(targets, separator)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
