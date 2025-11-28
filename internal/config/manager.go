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
