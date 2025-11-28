package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load inicializa e carrega a configuração seguindo a ordem de precedência:
// 1. Variáveis de Ambiente (CAST_*)
// 2. Arquivo Local (cast.*)
func Load() error {
	viper.SetEnvPrefix("CAST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Busca arquivo de configuração no diretório atual
	viper.SetConfigName("cast")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	// Tenta carregar arquivo (não é erro se não existir)
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
		}
	}

	return nil
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
