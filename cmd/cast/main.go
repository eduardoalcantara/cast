package main

import (
	"fmt"
	"os"

	"github.com/eduardoalcantara/cast/internal/config"
)

func main() {
	// Carrega configuração no bootstrap
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao carregar configuração: %v\n", err)
		os.Exit(2) // Exit code 2 = Config error
	}

	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
