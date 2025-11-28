# ESPECIFICAÇÃO TÉCNICA: CAST ARCHITECTURE

## 1. STACK TECNOLÓGICA
- **Linguagem:** Go (Golang) 1.22+
- **CLI Framework:** `github.com/spf13/cobra`
- **Config Management:** `github.com/spf13/viper`
- **UI/Color:** `github.com/fatih/color`

## 2. ESTRATÉGIA DE CONFIGURAÇÃO (12-FACTOR APP)
O `cast` deve carregar configurações com a seguinte **ordem de precedência** (topo vence):

1.  **Variáveis de Ambiente (ENV):**
    - Prefixo: `CAST_` (ex: `CAST_TELEGRAM_TOKEN`).
    - *Uso:* Docker, Kubernetes, n8n.
2.  **Arquivo Local (Current Dir):**
    - Formatos: `.yaml`, `.json`, `.properties`.
    - Nome: `cast.*` (ex: `cast.properties`).
    - *Uso:* Desenvolvimento local e servidores Java legados.

## 3. ESTRUTURA DE PASTAS (GO STANDARD LAYOUT)
```text
/cast
  /cmd/cast/
    main.go       (Entrypoint)
    root.go       (Comando Raiz + Banner)
    send.go       (Comando 'send')
  /internal/
    config/       (Lógica do Viper)
    providers/    (Interfaces e Implementações)
  /specifications/
  /documents/
  /tests/
  /prompts/
  /scripts/
  /logs/
  /run/
  go.mod
```

## 4. INTERFACES (CONTRATO)
```go
type Provider interface {
    Name() string
    // Send envia a mensagem. Retorna erro se falhar.
    Send(target string, message string) error
}
```
