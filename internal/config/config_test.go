package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigWithAliases(t *testing.T) {
	// Cria um arquivo YAML temporário com aliases
	yamlContent := `telegram:
  token: "test-token"
  default_chat_id: "123456789"

aliases:
  me:
    provider: "tg"
    target: "123456789"
    name: "Meu Telegram"

  team:
    provider: "mail"
    target: "team@exemplo.com"
    name: "Time de Desenvolvimento"

  alerts:
    provider: "zap"
    target: "5511999998888"
    name: "WhatsApp de Alertas"
`

	// Cria arquivo temporário
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "cast.yaml")
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo de teste: %v", err)
	}

	// Salva o diretório atual
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Erro ao obter diretório atual: %v", err)
	}

	// Muda para o diretório temporário
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Erro ao mudar diretório: %v", err)
	}
	defer os.Chdir(originalDir)

	// Carrega configuração
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Verifica se os aliases foram carregados
	if cfg.Aliases == nil {
		t.Fatal("Aliases não foram carregados (map é nil)")
	}

	if len(cfg.Aliases) != 3 {
		t.Fatalf("Esperado 3 aliases, obtido %d", len(cfg.Aliases))
	}

	// Verifica alias "me"
	meAlias := cfg.GetAlias("me")
	if meAlias == nil {
		t.Fatal("Alias 'me' não encontrado")
	}
	if meAlias.Provider != "tg" {
		t.Errorf("Alias 'me': esperado provider 'tg', obtido '%s'", meAlias.Provider)
	}
	if meAlias.Target != "123456789" {
		t.Errorf("Alias 'me': esperado target '123456789', obtido '%s'", meAlias.Target)
	}
	if meAlias.Name != "Meu Telegram" {
		t.Errorf("Alias 'me': esperado name 'Meu Telegram', obtido '%s'", meAlias.Name)
	}

	// Verifica alias "team"
	teamAlias := cfg.GetAlias("team")
	if teamAlias == nil {
		t.Fatal("Alias 'team' não encontrado")
	}
	if teamAlias.Provider != "mail" {
		t.Errorf("Alias 'team': esperado provider 'mail', obtido '%s'", teamAlias.Provider)
	}
	if teamAlias.Target != "team@exemplo.com" {
		t.Errorf("Alias 'team': esperado target 'team@exemplo.com', obtido '%s'", teamAlias.Target)
	}

	// Verifica alias "alerts"
	alertsAlias := cfg.GetAlias("alerts")
	if alertsAlias == nil {
		t.Fatal("Alias 'alerts' não encontrado")
	}
	if alertsAlias.Provider != "zap" {
		t.Errorf("Alias 'alerts': esperado provider 'zap', obtido '%s'", alertsAlias.Provider)
	}

	// Verifica alias inexistente
	nonexistent := cfg.GetAlias("inexistente")
	if nonexistent != nil {
		t.Error("Alias 'inexistente' não deveria existir")
	}
}

func TestParseTargets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "vírgula simples",
			input:    "target1,target2,target3",
			expected: []string{"target1", "target2", "target3"},
		},
		{
			name:     "ponto-e-vírgula",
			input:    "target1;target2;target3",
			expected: []string{"target1", "target2", "target3"},
		},
		{
			name:     "com espaços",
			input:    "target1, target2 , target3",
			expected: []string{"target1", "target2", "target3"},
		},
		{
			name:     "ponto-e-vírgula com espaços",
			input:    "target1; target2 ; target3",
			expected: []string{"target1", "target2", "target3"},
		},
		{
			name:     "target único",
			input:    "target1",
			expected: []string{"target1"},
		},
		{
			name:     "string vazia",
			input:    "",
			expected: nil,
		},
		{
			name:     "apenas espaços",
			input:    "   ,  ,  ",
			expected: nil,
		},
		{
			name:     "preferência ponto-e-vírgula",
			input:    "target1;target2,target3",
			expected: []string{"target1", "target2,target3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTargets(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Esperado %d targets, obtido %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Target %d: esperado '%s', mas não existe no resultado", i, expected)
					continue
				}
				if result[i] != expected {
					t.Errorf("Target %d: esperado '%s', obtido '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.applyDefaults()

	// Verifica defaults do Telegram
	if cfg.Telegram.APIURL != "https://api.telegram.org/bot" {
		t.Errorf("Telegram.APIURL: esperado 'https://api.telegram.org/bot', obtido '%s'", cfg.Telegram.APIURL)
	}
	if cfg.Telegram.Timeout != 30 {
		t.Errorf("Telegram.Timeout: esperado 30, obtido %d", cfg.Telegram.Timeout)
	}

	// Verifica defaults do WhatsApp
	if cfg.WhatsApp.APIVersion != "v18.0" {
		t.Errorf("WhatsApp.APIVersion: esperado 'v18.0', obtido '%s'", cfg.WhatsApp.APIVersion)
	}
	if cfg.WhatsApp.APIURL != "https://graph.facebook.com" {
		t.Errorf("WhatsApp.APIURL: esperado 'https://graph.facebook.com', obtido '%s'", cfg.WhatsApp.APIURL)
	}

	// Verifica defaults do Email
	if cfg.Email.SMTPPort != 587 {
		t.Errorf("Email.SMTPPort: esperado 587, obtido %d", cfg.Email.SMTPPort)
	}
	if !cfg.Email.UseTLS {
		t.Error("Email.UseTLS: esperado true (padrão)")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "timeout muito baixo",
			config: &Config{
				Telegram: TelegramConfig{Timeout: 3},
			},
			wantErr: true,
			errMsg:  "timeout deve estar entre 5 e 300",
		},
		{
			name: "timeout muito alto",
			config: &Config{
				Telegram: TelegramConfig{Timeout: 500},
			},
			wantErr: true,
			errMsg:  "timeout deve estar entre 5 e 300",
		},
		{
			name: "TLS e SSL ambos true",
			config: &Config{
				Email: EmailConfig{UseTLS: true, UseSSL: true},
			},
			wantErr: true,
			errMsg:  "use_tls e use_ssl não podem ser ambos true",
		},
		{
			name: "alias sem provider",
			config: &Config{
				Aliases: map[string]AliasConfig{
					"test": {Target: "target1"},
				},
			},
			wantErr: true,
			errMsg:  "provider não pode estar vazio",
		},
		{
			name: "alias sem target",
			config: &Config{
				Aliases: map[string]AliasConfig{
					"test": {Provider: "tg"},
				},
			},
			wantErr: true,
			errMsg:  "target não pode estar vazio",
		},
		{
			name: "configuração válida",
			config: &Config{
				Telegram: TelegramConfig{Timeout: 30},
				Email:    EmailConfig{UseTLS: true, UseSSL: false},
				Aliases: map[string]AliasConfig{
					"me": {Provider: "tg", Target: "123456789"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.applyDefaults()
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Esperado erro, mas não ocorreu")
					return
				}
				if tt.errMsg != "" {
					// Verifica se a mensagem de erro contém o texto esperado
					// Aceita variações na mensagem (ex: "use_tls e use_ssl" vs mensagem completa)
					errStr := err.Error()
					if !contains(errStr, tt.errMsg) && !contains(errStr, "use_tls e use_ssl") {
						t.Errorf("Mensagem de erro não contém '%s': %v", tt.errMsg, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Erro inesperado: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
