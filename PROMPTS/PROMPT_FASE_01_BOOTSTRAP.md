# ROLE: Senior Go Engineer (CLI Specialist)
# PROJECT: CAST (CAST Automates Sending Tasks)

# OBJECTIVE
Inicializar o projeto `cast` em Go, criando a estrutura de pastas, configurando Cobra+Viper e implementando a UX definida.

# INPUTS
Consulte as specs definidas em `/documents/02_TECH_SPEC.md` e `/documents/03_CLI_UX.md`.

# REQUIREMENTS

1. **Setup:**
   - Inicie `go.mod` (module `cast`).
   - Crie `cmd/cast/main.go`, `cmd/cast/root.go`, `cmd/cast/send.go`.
   - Crie `internal/config/config.go`.

2. **Configuração (Viper):**
   - Habilite leitura de ENV com prefixo `CAST_`.
   - Habilite leitura de arquivo `cast.properties`, `cast.yaml` ou `cast.json` na pasta local.
   - Struct de config deve prever: `Telegram.Token`, `Telegram.DefaultChatID`.

3. **UX & Commands:**
   - Implemente o Banner ASCII (Light Green) no `rootCmd`.
   - Implemente o comando `send` aceitando 3 argumentos obrigatórios.
   - Preencha o campo `Example` do Cobra com os exemplos da spec UX.

4. **Dummy Implementation:**
   - O comando `send` NÃO deve enviar nada real ainda. Apenas imprima:
     "Sending via [provider] to [target]: [message]"
   - Isso serve para validarmos se os argumentos e configs estão sendo lidos corretamente.

# DELIVERABLE
Código compilável. `go run ./cmd/cast` deve mostrar o banner.
