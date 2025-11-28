# ROLE: Senior Go Engineer (CLI UX Specialist)
# PROJECT: CAST (Fase 03 - Configuration Management)

# OBJECTIVE
Implementar os comandos de gerenciamento de configuração (CRUD) e o Wizard Interativo, permitindo que o usuário configure o CAST via linha de comando conforme definido em `specifications/05_PARAMETER_SPECS.md`.

# INPUTS
- `specifications/05_PARAMETER_SPECS.md`: Especificação detalhada dos comandos.
- `internal/config/config.go`: Structs existentes.

# TECH STACK ADDITIONS
- `github.com/AlecAivazis/survey/v2`: Para o Wizard Interativo.
- `github.com/olekukonko/tablewriter`: Para listar configurações em tabelas bonitas no terminal.
- `gopkg.in/yaml.v3`: Para marshaling seguro ao salvar o arquivo de configuração (evite usar viper para escrita se ele perder comentários ou formatação, prefira marshal direto da struct).

# REQUIREMENTS

## 1. Gerenciador de Configuração (`internal/config/manager.go`)
Crie funções para **persistir** as alterações:
- `Save(cfg *Config) error`: Salva a struct no disco.
  - **Lógica de Arquivo:** Se `viper.ConfigFileUsed()` retornar um caminho existente, use-o. Se não, crie `cast.yaml` no diretório atual.
  - **Sanitização:** Antes de salvar, garanta que mapas vazios sejam inicializados se necessário.
  - **Segurança:** Use permissão `0600` (apenas leitura/escrita para o dono) pois contém tokens.

## 2. Comando Gateway (`cmd/cast/gateway.go`)
Implemente `cast gateway [provider] [action]`.
- **Estrutura:** Use subcomandos do Cobra (`gatewayCmd`, `gatewayAddCmd`, etc.).
- **Subcomandos:** `add`, `list` (mostra status), `remove`.
- **Flags:** Mapeie todas as flags da Spec 05 (ex: `--token`, `--smtp-host`).
- **Modo Interativo (`--interactive`):**
  - Se a flag for usada, ignore as outras flags e inicie um questionário `survey`.
  - Pergunte os campos obrigatórios do provider escolhido (ex: Telegram -> Token, ChatID).
  - Valide as respostas (ex: ChatID deve ser numérico, Token não vazio).
  - Ao final, mostre o resumo e peça confirmação antes de salvar.

## 3. Comando Alias (`cmd/cast/alias.go`)
Implemente `cast alias [action]`.
- **Action `add`:** `cast alias add <name> <provider> <target>`.
  - Valide se o provider existe e se o target não é vazio.
- **Action `list`:** Use `tablewriter` para mostrar uma tabela ASCII com colunas: Name, Provider, Target.
- **Action `remove`:** `cast alias remove <name>`.

## 4. Comando Config (`cmd/cast/config.go`)
Implemente `cast config [action]`.
- **Action `show`:** Imprime o YAML atual.
  - Implemente flag `--mask` (default true) que substitui tokens por `*****`.
- **Action `validate`:** Roda `cfg.Validate()` e mostra o resultado.

# TEST STRATEGY
- Crie testes unitários para a lógica de persistência (`Save`).
- Como testar Wizard é difícil, foque em testar a validação dos inputs do Wizard.

# DELIVERABLE
Código compilável.
1. `cast gateway telegram add --interactive` -> Deve abrir o wizard, perguntar token, e salvar no arquivo `cast.yaml`.
2. `cast alias add me tg 123456` -> Deve salvar o alias e ele deve aparecer em `cast alias list`.
3. `cast config show` -> Deve mostrar o arquivo gerado.
