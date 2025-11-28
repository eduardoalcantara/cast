# ESPECIFICAÇÃO: COMANDOS CRUD DE CONFIGURAÇÃO

**Objetivo:** Definir os comandos CLI para gerenciamento (CRUD) de todas as configurações do CAST, permitindo que o usuário configure gateways e aliases sem editar arquivos manualmente.

**Princípio:** Todas as configurações possíveis nos arquivos de configuração (`cast.yaml`, `cast.json`, `cast.properties`) ou variáveis de ambiente devem ter comandos CRUD correspondentes no CLI.

---

## 1. ESTRUTURA GERAL

### 1.1 Padrão de Comandos

Todos os comandos seguem o padrão:
```
cast <recurso> <ação> [argumentos] [flags]
```

Onde:
- `<recurso>`: Tipo de configuração (gateway, alias, etc)
- `<ação>`: Operação CRUD (add, list, show, update, remove)
- `[argumentos]`: Dados específicos da operação
- `[flags]`: Opções adicionais

### 1.2 Ordem de Precedência para Persistência

Ao salvar configurações, o CLI deve:
1. Verificar se há arquivo de config existente (YAML, JSON ou Properties)
2. Se existir, usar o mesmo formato
3. Se não existir, criar em YAML (padrão)
4. **NUNCA** modificar variáveis de ambiente (apenas ler)

### 1.3 Feedback Visual

- **Verde:** Operação bem-sucedida
- **Vermelho:** Erro
- **Amarelo:** Aviso
- **Ciano:** Informação

---

## 2. COMANDO: `cast gateway`

Gerencia configurações de gateways (Telegram, WhatsApp, Email, Google Chat).

### 2.1 Estrutura

```
cast gateway <ação> [provider] [argumentos] [flags]
```

**Providers suportados:**
- `telegram` ou `tg`
- `whatsapp` ou `zap`
- `email` ou `mail`
- `google_chat` ou `googlechat`

### 2.2 Ações

#### `add` - Adicionar/Configurar Gateway
```bash
cast gateway add [provider] [flags]
```

**Nota:** Se `provider` não for especificado e a flag `--interactive` for usada, o wizard permitirá selecionar o provider interativamente.

**Flags:**
- `--token <token>` (Telegram)
- `--default-chat-id <id>` (Telegram)
- `--phone-number-id <id>` (WhatsApp)
- `--access-token <token>` (WhatsApp)
- `--smtp-host <host>` (Email)
- `--smtp-port <port>` (Email)
- `--username <user>` (Email)
- `--password <password>` (Email)
- `--webhook-url <url>` (Google Chat)
- `--timeout <seconds>` (Todos)
- `--format <yaml|json|properties>` - Formato do arquivo (padrão: yaml)
- `--interactive` ou `-i` - Modo wizard interativo

**Exemplos:**
```bash
# Telegram via flags
cast gateway add telegram --token "123456:ABC" --default-chat-id "123456789"

# Email via wizard interativo
cast gateway add email --interactive

# Wizard interativo (seleciona provider)
cast gateway add --interactive

# WhatsApp
cast gateway add whatsapp --phone-number-id "123" --access-token "EAAxxx"
```

#### `show` - Mostrar Configuração
```bash
cast gateway show <provider> [flags]
```

**Flags:**
- `--mask` - Mascara campos sensíveis (tokens, senhas) - padrão: true

**Exemplos:**
```bash
cast gateway show telegram
cast gateway show email --mask=false  # Mostra senha (cuidado!)
```

#### `update` - Atualizar Configuração
```bash
cast gateway update <provider> [flags]
```

**Flags:** Mesmas do comando `add`

**Exemplos:**
```bash
cast gateway update telegram --default-chat-id "987654321"
cast gateway update email --smtp-port 465 --use-ssl
```

#### `remove` - Remover Configuração
```bash
cast gateway remove <provider> [flags]
```

**Flags:**
- `--confirm` ou `-y` - Confirma sem perguntar

**Exemplos:**
```bash
cast gateway remove telegram
cast gateway remove whatsapp --confirm
```

#### `test` - Testar Configuração
```bash
cast gateway test <provider> [flags]
```

**Flags:**
- `--target <target>` - Target para teste (opcional)

**Exemplos:**
```bash
cast gateway test telegram
cast gateway test email --target "teste@exemplo.com"
```

### 2.3 Exemplos Completos

```bash
# Configurar Telegram
cast gateway add telegram \
  --token "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz" \
  --default-chat-id "123456789" \
  --timeout 30

# Configurar Email via wizard
cast gateway add email --interactive

# Wizard interativo (seleciona provider)
cast gateway add --interactive

# Ver configuração do Telegram
cast gateway show telegram

# Atualizar timeout do Email
cast gateway update email --timeout 60

# Testar conexão SMTP
cast gateway test email --target "admin@empresa.com"

# Remover configuração do WhatsApp
cast gateway remove whatsapp
```

---

## 3. COMANDO: `cast alias`

Gerencia aliases (atalhos para provider + target).

### 3.1 Estrutura

```
cast alias <ação> [argumentos] [flags]
```

### 3.2 Ações

#### `add` - Adicionar Alias
```bash
cast alias add <nome> <provider> <target> [flags]
```

**Argumentos:**
- `<nome>`: Nome do alias (ex: `me`, `team`, `alerts`)
- `<provider>`: Provider (tg, mail, zap, google_chat)
- `<target>`: Target (chat_id, email, número, webhook_url)

**Flags:**
- `--name <descrição>` - Nome descritivo (opcional)
- `--format <yaml|json|properties>` - Formato do arquivo

**Exemplos:**
```bash
cast alias add me tg "123456789" --name "Meu Telegram Pessoal"
cast alias add team mail "sdc@tre-pa.jus.br" --name "Time de Desenvolvimento"
cast alias add alerts zap "5511999998888"
```

#### `list` - Listar Aliases
```bash
cast alias list [flags]
```

**Flags:**
- `--format <table|json|yaml>` - Formato de saída (padrão: table)

**Exemplos:**
```bash
cast alias list
cast alias list --format json
```

**Saída esperada (table):**
```
Nome    Provider    Target                    Descrição
----    --------    ------                    -----------
me      tg          123456789                Meu Telegram Pessoal
team    mail        sdc@tre-pa.jus.br         Time de Desenvolvimento
alerts  zap         5511999998888             WhatsApp de Alertas
```

#### `show` - Mostrar Detalhes de um Alias
```bash
cast alias show <nome> [flags]
```

**Exemplos:**
```bash
cast alias show me
```

**Saída esperada:**
```
Alias: me
Provider: tg
Target: 123456789
Descrição: Meu Telegram Pessoal
```

#### `update` - Atualizar Alias
```bash
cast alias update <nome> [flags]
```

**Flags:**
- `--provider <provider>` - Novo provider
- `--target <target>` - Novo target
- `--name <descrição>` - Nova descrição

**Exemplos:**
```bash
cast alias update me --target "987654321"
cast alias update team --provider "google_chat" --target "https://..."
```

#### `remove` - Remover Alias
```bash
cast alias remove <nome> [flags]
```

**Flags:**
- `--confirm` ou `-y` - Confirma sem perguntar

**Exemplos:**
```bash
cast alias remove me
cast alias remove alerts --confirm
```

### 3.3 Exemplos Completos

```bash
# Adicionar aliases
cast alias add me-tg tg "9198805000" --name "Meu Telegram"
cast alias add team mail "sdc@tre-pa.jus.br" --name "Time TRE-PA"

# Listar todos
cast alias list

# Ver detalhes
cast alias show me-tg

# Atualizar target
cast alias update me-tg --target "9198805001"

# Remover
cast alias remove team
```

---

## 4. COMANDO: `cast config`

Comandos gerais de configuração.

### 4.1 Estrutura

```
cast config <ação> [argumentos] [flags]
```

### 4.2 Ações

#### `show` - Mostrar Configuração Completa
```bash
cast config show [flags]
```

**Flags:**
- `--format <yaml|json|properties>` - Formato de saída
- `--mask` - Mascara campos sensíveis (padrão: true)
- `--source` - Mostra origem (ENV ou File)

**Exemplos:**
```bash
cast config show
cast config show --format json --mask=false
cast config show --source
```

#### `validate` - Validar Configuração
```bash
cast config validate [flags]
```

**Exemplos:**
```bash
cast config validate
```

**Saída esperada:**
```
✓ Configuração válida
  - Telegram: configurado
  - Email: configurado
  - Aliases: 3 definidos
```

#### `reload` - Recarregar Configuração
```bash
cast config reload
```

Útil após editar arquivo manualmente.

#### `export` - Exportar Configuração
```bash
cast config export [flags]
```

**Flags:**
- `--format <yaml|json|properties>` - Formato de exportação
- `--output <arquivo>` - Arquivo de saída (padrão: stdout)
- `--mask` - Mascara campos sensíveis

**Exemplos:**
```bash
cast config export --format yaml --output cast.backup.yaml
cast config export --format json --mask
```

#### `import` - Importar Configuração
```bash
cast config import <arquivo> [flags]
```

**Flags:**
- `--merge` - Mescla com configuração existente (padrão: false, sobrescreve)
- `--format <yaml|json|properties>` - Formato do arquivo (auto-detect se não especificado)

**Exemplos:**
```bash
cast config import cast.backup.yaml
cast config import config.json --merge
```

---

## 5. MODO WIZARD (INTERATIVO)

### 5.1 Conceito

O modo wizard (`--interactive` ou `-i`) permite configurar gateways através de perguntas interativas, facilitando o uso para usuários menos técnicos.

### 5.2 Fluxo do Wizard

1. **Seleção do Gateway:**
   ```
   Selecione o gateway a configurar:
   1) Telegram
   2) WhatsApp
   3) Email
   4) Google Chat
   ```

2. **Perguntas Contextuais:**
   - Para Telegram: Token, Chat ID, Timeout
   - Para Email: SMTP Host, Porta, Credenciais, TLS/SSL
   - etc.

3. **Confirmação:**
   ```
   Configuração a ser salva:
   [mostra resumo]

   Confirmar? (s/n):
   ```

4. **Escolha de Formato:**
   ```
   Em qual formato salvar?
   1) YAML (cast.yaml)
   2) JSON (cast.json)
   3) Properties (cast.properties)
   ```

### 5.3 Exemplos

```bash
# Wizard completo (seleciona provider)
cast gateway add --interactive

# Wizard para Email específico
cast gateway add email --interactive
```

---

## 6. FORMATO DE SAÍDA

### 6.1 Tabelas

Para comandos `list`, usar tabelas formatadas:

```
┌─────────┬──────────┬──────────────────────┬──────────────────────┐
│ Nome    │ Provider │ Target               │ Descrição             │
├─────────┼──────────┼──────────────────────┼──────────────────────┤
│ me      │ tg       │ 123456789            │ Meu Telegram Pessoal  │
│ team    │ mail     │ sdc@tre-pa.jus.br    │ Time de Desenvolvimento│
└─────────┴──────────┴──────────────────────┴──────────────────────┘
```

### 6.2 JSON/YAML

Para flags `--format json` ou `--format yaml`, saída estruturada.

### 6.3 Mensagens de Sucesso/Erro

**Sucesso:**
```
✓ Alias 'me' adicionado com sucesso
✓ Configuração do Telegram atualizada
```

**Erro:**
```
✗ Erro: Alias 'me' já existe
✗ Erro: Configuração do Email incompleta (smtp_host obrigatório)
```

---

## 7. VALIDAÇÕES

### 7.1 Validações ao Adicionar/Atualizar

- **Telegram:**
  - Token não pode estar vazio
  - Token deve ter formato válido (número:hash)
  - Chat ID deve ser numérico (se fornecido)

- **Email:**
  - SMTP Host não pode estar vazio
  - Porta deve estar entre 1-65535
  - Username e Password obrigatórios
  - TLS e SSL mutuamente exclusivos

- **WhatsApp:**
  - Phone Number ID e Access Token obrigatórios

- **Google Chat:**
  - Webhook URL deve ser uma URL válida

- **Aliases:**
  - Nome único (não pode duplicar)
  - Provider deve ser válido
  - Target não pode estar vazio

### 7.2 Mensagens de Erro

Todas as validações devem retornar mensagens claras em português:
- `"Alias 'me' já existe"`
- `"Token do Telegram inválido (formato esperado: número:hash)"`
- `"Porta SMTP deve estar entre 1 e 65535"`

---

## 8. IMPLEMENTAÇÃO POR FASES

### ✅ Fase 01 - Bootstrap
- [x] Estrutura base de comandos
- [x] Leitura de configuração

### 🚧 Fase 02 - Core Drivers
- [ ] Comando `cast alias` (CRUD completo)
- [ ] Comando `cast gateway` (parcial - Telegram e Email)

### 📋 Fase 03 - Integração Avançada
- [ ] Comando `cast gateway` (completo - WhatsApp e Google Chat)
- [ ] Comando `cast config` (show, validate, export, import)

### 📋 Fase 04 - Build & Release
- [ ] Modo wizard interativo
- [ ] Comando `cast config reload`
- [ ] Testes de integração dos comandos CRUD

---

## 9. EXEMPLOS DE USO COMPLETO

### 9.1 Configuração Inicial (Wizard)

```bash
# Configurar tudo via wizard
cast gateway add telegram --interactive
cast gateway add email --interactive

# Wizard interativo (seleciona provider)
cast gateway add --interactive

# Adicionar aliases
cast alias add me tg "123456789" --name "Meu Telegram"
cast alias add team mail "sdc@tre-pa.jus.br" --name "Time TRE-PA"
```

### 9.2 Uso Diário

```bash
# Enviar usando alias
cast send me "Deploy finalizado!"

# Ver configurações
cast gateway show telegram
cast alias list

# Atualizar configuração
cast gateway update email --smtp-port 465 --use-ssl
```

### 9.3 Backup e Restore

```bash
# Exportar configuração
cast config export --output backup.yaml

# Importar configuração
cast config import backup.yaml

# Validar antes de usar
cast config validate
```

---

## 10. NOTAS DE IMPLEMENTAÇÃO

### 10.1 Persistência

- **Arquivos:** Sempre manter formatação e comentários ao atualizar
- **Backup:** Criar backup automático antes de modificar arquivo existente
- **Atomicidade:** Escrever em arquivo temporário e renomear (evitar corrupção)

### 10.2 Segurança

- **Mascaramento:** Sempre mascarar tokens/senhas na saída (exceto com `--mask=false`)
- **Permissões:** Verificar permissões de escrita antes de modificar arquivos
- **Validação:** Validar todas as entradas antes de salvar

### 10.3 UX

- **Confirmação:** Pedir confirmação para operações destrutivas (`remove`)
- **Feedback:** Sempre mostrar o que foi feito
- **Help:** Help contextual rico para cada comando

### 10.4 Compatibilidade

- **Formatos:** Suportar YAML, JSON e Properties
- **Migração:** Permitir converter entre formatos
- **ENV:** Ler de ENV, mas não escrever (apenas arquivos)

---

## 11. COMANDOS FUTUROS (A Adicionar)

Esta seção será atualizada conforme novas funcionalidades forem implementadas:

### 11.1 Fase 03+
- [ ] `cast gateway whatsapp` (CRUD completo)
- [ ] `cast gateway google_chat` (CRUD completo)
- [ ] `cast config migrate` - Migrar entre formatos

### 11.2 Melhorias Futuras
- [ ] `cast alias test <nome>` - Testar alias
- [ ] `cast gateway <provider> test --dry-run` - Teste sem enviar
- [ ] `cast config diff` - Comparar configurações
- [ ] `cast config history` - Histórico de mudanças (se implementar versionamento)

---

**Última atualização:** 2025-01-XX
**Versão do documento:** 1.0
**Status:** 🟡 Em desenvolvimento (Fase 02)
