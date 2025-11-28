package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func TestSave_NewFile(t *testing.T) {
	// Cria diretório temporário
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "cast.yaml")

	// Configura viper para que Save() use o arquivo correto
	viper.Reset()
	viper.SetConfigFile(configFile)
	// Não chama ReadInConfig() pois o arquivo não existe ainda
	// Mas Save() deve criar o arquivo

	// Cria configuração
	cfg := &Config{
		Telegram: TelegramConfig{
			Token:        "test-token",
			DefaultChatID: "123456789",
			Timeout:      30,
		},
		Aliases: map[string]AliasConfig{
			"me": {
				Provider: "tg",
				Target:   "123456789",
				Name:     "Meu Telegram",
			},
		},
	}

	// Salva
	// Como Save() verifica viper.ConfigFileUsed(), precisamos garantir que está configurado
	// Mas se o arquivo não existe, Save() cria "cast.yaml" no diretório atual
	// Vamos mudar para o diretório temporário
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	viper.Reset()
	viper.SetConfigName("cast")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	if err := Save(cfg); err != nil {
		t.Fatalf("Erro ao salvar: %v", err)
	}

	// Verifica se arquivo foi criado
	if _, err := os.Stat("cast.yaml"); os.IsNotExist(err) {
		t.Fatal("Arquivo cast.yaml não foi criado")
	}

	// Carrega e verifica
	viper.Reset()
	viper.SetConfigFile("cast.yaml")
	viper.ReadInConfig()

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("Erro ao carregar: %v", err)
	}

	if loaded.Telegram.Token != "test-token" {
		t.Errorf("Token não foi salvo corretamente: esperado 'test-token', obtido '%s'", loaded.Telegram.Token)
	}

	if loaded.GetAlias("me") == nil {
		t.Error("Alias 'me' não foi salvo")
	}
}

func TestSave_ExistingFile(t *testing.T) {
	// Cria diretório temporário
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Cria arquivo existente
	existingFile := "cast.yaml"
	os.WriteFile(existingFile, []byte("telegram:\n  token: old-token\n"), 0600)

	// Configura viper para usar este arquivo
	viper.Reset()
	viper.SetConfigFile(existingFile)
	viper.ReadInConfig()

	// Cria nova configuração
	cfg := &Config{
		Telegram: TelegramConfig{
			Token:        "new-token",
			DefaultChatID: "987654321",
			Timeout:      30,
		},
	}

	// Salva
	if err := Save(cfg); err != nil {
		t.Fatalf("Erro ao salvar: %v", err)
	}

	// Recarrega viper para ler o arquivo atualizado
	viper.Reset()
	viper.SetConfigFile(existingFile)
	viper.ReadInConfig()

	// Carrega e verifica
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("Erro ao carregar: %v", err)
	}

	if loaded.Telegram.Token != "new-token" {
		t.Errorf("Token não foi atualizado: esperado 'new-token', obtido '%s'", loaded.Telegram.Token)
	}
}

func TestSave_EmptyAliases(t *testing.T) {
	// Cria diretório temporário
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Cria configuração sem aliases
	cfg := &Config{
		Telegram: TelegramConfig{
			Token: "test-token",
		},
		// Aliases é nil
	}

	// Configura viper
	viper.Reset()
	viper.SetConfigName("cast")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	// Salva
	if err := Save(cfg); err != nil {
		t.Fatalf("Erro ao salvar: %v", err)
	}

	// Lê arquivo diretamente para verificar que aliases foi inicializado
	data, err := os.ReadFile("cast.yaml")
	if err != nil {
		t.Fatalf("Erro ao ler arquivo: %v", err)
	}

	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Erro ao fazer unmarshal: %v", err)
	}

	if loaded.Aliases == nil {
		t.Error("Aliases deveria ser inicializado (não nil)")
	}
}
