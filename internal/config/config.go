package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config representa a estrutura de configuração do CAST.
type Config struct {
	Telegram  TelegramConfig              `mapstructure:"telegram" yaml:"telegram" json:"telegram"`
	WhatsApp  WhatsAppConfig              `mapstructure:"whatsapp" yaml:"whatsapp" json:"whatsapp"`
	Email     EmailConfig                 `mapstructure:"email" yaml:"email" json:"email"`
	GoogleChat GoogleChatConfig           `mapstructure:"google_chat" yaml:"google_chat" json:"google_chat"`
	WAHA      WAHAConfig                  `mapstructure:"waha" yaml:"waha" json:"waha"`
	Aliases   map[string]AliasConfig      `mapstructure:"aliases" yaml:"aliases" json:"aliases"`
}

// TelegramConfig contém as configurações do Telegram.
type TelegramConfig struct {
	Token        string `mapstructure:"token" yaml:"token" json:"token"`
	DefaultChatID string `mapstructure:"default_chat_id" yaml:"default_chat_id" json:"default_chat_id"`
	APIURL       string `mapstructure:"api_url" yaml:"api_url" json:"api_url"`
	Timeout      int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
}

// WhatsAppConfig contém as configurações do WhatsApp (Meta Cloud API).
type WhatsAppConfig struct {
	PhoneNumberID    string `mapstructure:"phone_number_id" yaml:"phone_number_id" json:"phone_number_id"`
	AccessToken      string `mapstructure:"access_token" yaml:"access_token" json:"access_token"`
	BusinessAccountID string `mapstructure:"business_account_id" yaml:"business_account_id" json:"business_account_id"`
	APIVersion       string `mapstructure:"api_version" yaml:"api_version" json:"api_version"`
	APIURL           string `mapstructure:"api_url" yaml:"api_url" json:"api_url"`
	Timeout          int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
}

// EmailConfig contém as configurações de Email (SMTP).
type EmailConfig struct {
	SMTPHost  string `mapstructure:"smtp_host" yaml:"smtp_host" json:"smtp_host"`
	SMTPPort  int    `mapstructure:"smtp_port" yaml:"smtp_port" json:"smtp_port"`
	Username  string `mapstructure:"username" yaml:"username" json:"username"`
	Password  string `mapstructure:"password" yaml:"password" json:"password"`
	FromEmail string `mapstructure:"from_email" yaml:"from_email" json:"from_email"`
	FromName  string `mapstructure:"from_name" yaml:"from_name" json:"from_name"`
	UseTLS    bool   `mapstructure:"use_tls" yaml:"use_tls" json:"use_tls"`
	UseSSL    bool   `mapstructure:"use_ssl" yaml:"use_ssl" json:"use_ssl"`
	Timeout   int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`

	// IMAP: usado apenas se wait-for-response estiver ativo
	IMAPHost     string `mapstructure:"imap_host" yaml:"imap_host" json:"imap_host"`
	IMAPPort     int    `mapstructure:"imap_port" yaml:"imap_port" json:"imap_port"`
	IMAPUsername string `mapstructure:"imap_username" yaml:"imap_username" json:"imap_username"`
	IMAPPassword string `mapstructure:"imap_password" yaml:"imap_password" json:"imap_password"`
	IMAPUseTLS   bool   `mapstructure:"imap_use_tls" yaml:"imap_use_tls" json:"imap_use_tls"`
	IMAPUseSSL   bool   `mapstructure:"imap_use_ssl" yaml:"imap_use_ssl" json:"imap_use_ssl"`
	IMAPFolder        string `mapstructure:"imap_folder" yaml:"imap_folder" json:"imap_folder"`
	IMAPTimeout       int    `mapstructure:"imap_timeout" yaml:"imap_timeout" json:"imap_timeout"`
	IMAPPollInterval  int    `mapstructure:"imap_poll_interval_seconds" yaml:"imap_poll_interval_seconds" json:"imap_poll_interval_seconds"`

	// Espera por resposta
	WaitForResponseDefault  int  `mapstructure:"wait_for_response_default_minutes" yaml:"wait_for_response_default_minutes" json:"wait_for_response_default_minutes"`
	WaitForResponseMax      int  `mapstructure:"wait_for_response_max_minutes" yaml:"wait_for_response_max_minutes" json:"wait_for_response_max_minutes"`
	WaitForResponseMaxLines int  `mapstructure:"wait_for_response_max_lines" yaml:"wait_for_response_max_lines" json:"wait_for_response_max_lines"`
	WaitForResponseFullLayout bool `mapstructure:"wait_for_response_full_layout" yaml:"wait_for_response_full_layout" json:"wait_for_response_full_layout"`
}

// GoogleChatConfig contém as configurações do Google Chat.
type GoogleChatConfig struct {
	WebhookURL string `mapstructure:"webhook_url" yaml:"webhook_url" json:"webhook_url"`
	Timeout    int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
}

// WAHAConfig contém as configurações do WAHA (WhatsApp HTTP API).
type WAHAConfig struct {
	APIURL  string `mapstructure:"api_url" yaml:"api_url" json:"api_url"`
	Session string `mapstructure:"session" yaml:"session" json:"session"`
	APIKey  string `mapstructure:"api_key" yaml:"api_key" json:"api_key"`
	Timeout int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
}

// AliasConfig representa um alias para facilitar o uso do CLI.
type AliasConfig struct {
	Provider string `mapstructure:"provider" yaml:"provider" json:"provider"`
	Target   string `mapstructure:"target" yaml:"target" json:"target"`
	Name     string `mapstructure:"name" yaml:"name" json:"name"`
}

// Load inicializa e carrega a configuração seguindo a ordem de precedência:
// 1. Variáveis de Ambiente (CAST_*)
// 2. Arquivo Local (cast.*) - suporta .yaml, .json, .properties
func Load() error {
	viper.SetEnvPrefix("CAST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// BindEnv explícito para estruturas aninhadas (necessário para AutomaticEnv funcionar corretamente)
	// Com SetEnvKeyReplacer(".", "_"), o Viper mapeia automaticamente:
	// telegram.token -> CAST_TELEGRAM_TOKEN
	// telegram.default_chat_id -> CAST_TELEGRAM_DEFAULT_CHAT_ID
	// etc.

	// Telegram
	viper.BindEnv("telegram.token")
	viper.BindEnv("telegram.default_chat_id")
	viper.BindEnv("telegram.api_url")
	viper.BindEnv("telegram.timeout")

	// WhatsApp
	viper.BindEnv("whatsapp.phone_number_id")
	viper.BindEnv("whatsapp.access_token")
	viper.BindEnv("whatsapp.business_account_id")
	viper.BindEnv("whatsapp.api_version")
	viper.BindEnv("whatsapp.api_url")
	viper.BindEnv("whatsapp.timeout")

	// Email
	viper.BindEnv("email.smtp_host")
	viper.BindEnv("email.smtp_port")
	viper.BindEnv("email.username")
	viper.BindEnv("email.password")
	viper.BindEnv("email.from_email")
	viper.BindEnv("email.from_name")
	viper.BindEnv("email.use_tls")
	viper.BindEnv("email.use_ssl")
	viper.BindEnv("email.timeout")
	// IMAP
	viper.BindEnv("email.imap_host")
	viper.BindEnv("email.imap_port")
	viper.BindEnv("email.imap_username")
	viper.BindEnv("email.imap_password")
	viper.BindEnv("email.imap_use_tls")
	viper.BindEnv("email.imap_use_ssl")
	viper.BindEnv("email.imap_folder")
	viper.BindEnv("email.imap_timeout")
	viper.BindEnv("email.imap_poll_interval_seconds")
	// Wait for response
	viper.BindEnv("email.wait_for_response_default_minutes")
	viper.BindEnv("email.wait_for_response_max_minutes")
	viper.BindEnv("email.wait_for_response_max_lines")
	viper.BindEnv("email.wait_for_response_full_layout")

	// Google Chat
	viper.BindEnv("google_chat.webhook_url")
	viper.BindEnv("google_chat.timeout")

	// WAHA
	viper.BindEnv("waha.api_url")
	viper.BindEnv("waha.session")
	viper.BindEnv("waha.api_key")
	viper.BindEnv("waha.timeout")

	// Busca arquivo de configuração de forma transparente:
	// 1. Primeiro procura no diretório atual (onde o usuário está executando)
	// 2. Se não encontrar, procura no diretório do executável (fallback)

	viper.SetConfigName("cast")
	// Adiciona diretório atual como primeiro caminho (prioridade)
	viper.AddConfigPath(".")

	// Adiciona diretório do executável como fallback
	execPath, err := os.Executable()
	if err == nil {
		// Obtém o diretório do executável
		execDir := filepath.Dir(execPath)
		// Normaliza o caminho (resolve symlinks no Linux/Mac)
		execDir, _ = filepath.EvalSymlinks(execDir)
		// Adiciona como fallback (segunda opção)
		viper.AddConfigPath(execDir)
	}

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
// Respeita a ordem de precedência: ENV > Arquivo
func LoadConfig() (*Config, error) {
	if err := Load(); err != nil {
		return nil, fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("erro ao fazer unmarshal da configuração: %w", err)
	}

	// Aplica valores de ENV sobre o arquivo (ENV tem prioridade)
	// viper.Get() respeita a ordem de precedência: ENV > Arquivo
	applyEnvOverrides(&cfg)

	// Aplica valores padrão
	cfg.applyDefaults()

	// Valida configuração
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("erro na validação da configuração: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides aplica valores de variáveis de ambiente sobre os valores do arquivo.
// Isso garante que ENV sempre tenha prioridade sobre o arquivo.
// viper.Get() respeita a ordem de precedência: ENV > Arquivo
func applyEnvOverrides(cfg *Config) {
	// Telegram
	// Se ENV está definido, usa ENV (viper.Get() retorna ENV se existir)
	if envVal := viper.GetString("telegram.token"); envVal != "" {
		cfg.Telegram.Token = envVal
	}
	if envVal := viper.GetString("telegram.default_chat_id"); envVal != "" {
		cfg.Telegram.DefaultChatID = envVal
	}
	if envVal := viper.GetString("telegram.api_url"); envVal != "" {
		cfg.Telegram.APIURL = envVal
	}
	if envVal := viper.GetInt("telegram.timeout"); envVal > 0 {
		cfg.Telegram.Timeout = envVal
	}

	// WhatsApp
	if envVal := viper.GetString("whatsapp.phone_number_id"); envVal != "" {
		cfg.WhatsApp.PhoneNumberID = envVal
	}
	if envVal := viper.GetString("whatsapp.access_token"); envVal != "" {
		cfg.WhatsApp.AccessToken = envVal
	}
	if envVal := viper.GetString("whatsapp.business_account_id"); envVal != "" {
		cfg.WhatsApp.BusinessAccountID = envVal
	}
	if envVal := viper.GetString("whatsapp.api_version"); envVal != "" {
		cfg.WhatsApp.APIVersion = envVal
	}
	if envVal := viper.GetString("whatsapp.api_url"); envVal != "" {
		cfg.WhatsApp.APIURL = envVal
	}
	if envVal := viper.GetInt("whatsapp.timeout"); envVal > 0 {
		cfg.WhatsApp.Timeout = envVal
	}

	// Email
	if envVal := viper.GetString("email.smtp_host"); envVal != "" {
		cfg.Email.SMTPHost = envVal
	}
	if envVal := viper.GetInt("email.smtp_port"); envVal > 0 {
		cfg.Email.SMTPPort = envVal
	}
	if envVal := viper.GetString("email.username"); envVal != "" {
		cfg.Email.Username = envVal
	}
	if envVal := viper.GetString("email.password"); envVal != "" {
		cfg.Email.Password = envVal
	}
	if envVal := viper.GetString("email.from_email"); envVal != "" {
		cfg.Email.FromEmail = envVal
	}
	if envVal := viper.GetString("email.from_name"); envVal != "" {
		cfg.Email.FromName = envVal
	}
	// Booleanos: se ENV está definido, usa ENV
	if viper.IsSet("email.use_tls") {
		cfg.Email.UseTLS = viper.GetBool("email.use_tls")
	}
	if viper.IsSet("email.use_ssl") {
		cfg.Email.UseSSL = viper.GetBool("email.use_ssl")
	}
	if envVal := viper.GetInt("email.timeout"); envVal > 0 {
		cfg.Email.Timeout = envVal
	}
	// IMAP
	if envVal := viper.GetString("email.imap_host"); envVal != "" {
		cfg.Email.IMAPHost = envVal
	}
	if envVal := viper.GetInt("email.imap_port"); envVal > 0 {
		cfg.Email.IMAPPort = envVal
	}
	if envVal := viper.GetString("email.imap_username"); envVal != "" {
		cfg.Email.IMAPUsername = envVal
	}
	if envVal := viper.GetString("email.imap_password"); envVal != "" {
		cfg.Email.IMAPPassword = envVal
	}
	if viper.IsSet("email.imap_use_tls") {
		cfg.Email.IMAPUseTLS = viper.GetBool("email.imap_use_tls")
	}
	if viper.IsSet("email.imap_use_ssl") {
		cfg.Email.IMAPUseSSL = viper.GetBool("email.imap_use_ssl")
	}
	if envVal := viper.GetString("email.imap_folder"); envVal != "" {
		cfg.Email.IMAPFolder = envVal
	}
	if envVal := viper.GetInt("email.imap_timeout"); envVal > 0 {
		cfg.Email.IMAPTimeout = envVal
	}
	if envVal := viper.GetInt("email.imap_poll_interval_seconds"); envVal > 0 {
		cfg.Email.IMAPPollInterval = envVal
	}
	// Wait for response
	if envVal := viper.GetInt("email.wait_for_response_default_minutes"); envVal >= 0 {
		cfg.Email.WaitForResponseDefault = envVal
	}
	if envVal := viper.GetInt("email.wait_for_response_max_minutes"); envVal > 0 {
		cfg.Email.WaitForResponseMax = envVal
	}
	if envVal := viper.GetInt("email.wait_for_response_max_lines"); envVal >= 0 {
		cfg.Email.WaitForResponseMaxLines = envVal
	}
	if envVal := viper.GetBool("email.wait_for_response_full_layout"); envVal {
		cfg.Email.WaitForResponseFullLayout = envVal
	}

	// Google Chat
	if envVal := viper.GetString("google_chat.webhook_url"); envVal != "" {
		cfg.GoogleChat.WebhookURL = envVal
	}
	if envVal := viper.GetInt("google_chat.timeout"); envVal > 0 {
		cfg.GoogleChat.Timeout = envVal
	}

	// WAHA
	if envVal := viper.GetString("waha.api_url"); envVal != "" {
		cfg.WAHA.APIURL = envVal
	}
	if envVal := viper.GetString("waha.session"); envVal != "" {
		cfg.WAHA.Session = envVal
	}
	if envVal := viper.GetString("waha.api_key"); envVal != "" {
		cfg.WAHA.APIKey = envVal
	}
	if envVal := viper.GetInt("waha.timeout"); envVal > 0 {
		cfg.WAHA.Timeout = envVal
	}
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
	// Aplica padrão TLS apenas se NENHUM dos dois foi explicitamente definido
	// Verifica se foram definidos no arquivo ou ENV usando viper.IsSet()
	if !viper.IsSet("email.use_tls") && !viper.IsSet("email.use_ssl") {
		// Nenhum foi definido, aplica padrão TLS
		if !c.Email.UseTLS && !c.Email.UseSSL {
			c.Email.UseTLS = true // Padrão TLS
		}
	}
	if c.Email.Timeout == 0 {
		c.Email.Timeout = 30
	}

	// Email IMAP defaults
	if c.Email.IMAPPort == 0 {
		if c.Email.IMAPUseSSL {
			c.Email.IMAPPort = 993
		} else if c.Email.IMAPUseTLS {
			c.Email.IMAPPort = 143
		} else if c.Email.IMAPHost != "" {
			// Se IMAPHost está configurado mas não há flags SSL/TLS, assume SSL por padrão
			c.Email.IMAPUseSSL = true
			c.Email.IMAPPort = 993
		}
	}
	if c.Email.IMAPFolder == "" {
		c.Email.IMAPFolder = "INBOX"
	}
	if c.Email.IMAPTimeout == 0 {
		c.Email.IMAPTimeout = 60
	}
	if c.Email.WaitForResponseMax == 0 {
		c.Email.WaitForResponseMax = 120
	}
	if c.Email.WaitForResponseMaxLines < 0 {
		c.Email.WaitForResponseMaxLines = 0
	}

	// Google Chat defaults
	if c.GoogleChat.Timeout == 0 {
		c.GoogleChat.Timeout = 30
	}

	// WAHA defaults
	if c.WAHA.Session == "" {
		c.WAHA.Session = "default"
	}
	if c.WAHA.Timeout == 0 {
		c.WAHA.Timeout = 30
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
	if c.WAHA.Timeout < 5 || c.WAHA.Timeout > 300 {
		return fmt.Errorf("waha.timeout deve estar entre 5 e 300 segundos")
	}

	// Validação de Email: TLS e SSL são mutuamente exclusivos
	if c.Email.UseTLS && c.Email.UseSSL {
		return fmt.Errorf("email.use_tls e email.use_ssl não podem ser ambos true (priorizando TLS)")
	}

	// Validação de Email IMAP: se WaitForResponseDefault > 0, IMAP deve estar configurado
	if c.Email.WaitForResponseDefault > 0 {
		if c.Email.IMAPHost == "" {
			return fmt.Errorf("email.wait_for_response_default_minutes > 0 requer email.imap_host configurado")
		}
		if c.Email.IMAPPort == 0 {
			return fmt.Errorf("email.wait_for_response_default_minutes > 0 requer email.imap_port configurado")
		}
		if c.Email.IMAPUsername == "" {
			return fmt.Errorf("email.wait_for_response_default_minutes > 0 requer email.imap_username configurado")
		}
		if c.Email.IMAPPassword == "" {
			return fmt.Errorf("email.wait_for_response_default_minutes > 0 requer email.imap_password configurado")
		}
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
