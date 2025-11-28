# CAST - PROJECT STATUS

**Última atualização:** 2025-01-XX
**Versão:** 0.1.0 (Fase 01 - Bootstrap)
**Status Geral:** 🟡 Em Desenvolvimento

---

## 📊 VISÃO GERAL

O CAST (CAST Automates Sending Tasks) é uma ferramenta CLI standalone para envio agnóstico de mensagens (Fire & Forget) via múltiplos gateways: Telegram, WhatsApp, Email e Google Chat.

**Stack:** Go 1.22+, Cobra, Viper, fatih/color

---

## ✅ FASES CONCLUÍDAS

### ✅ Fase 00: Pesquisa & Naming
- [x] Nome definido: CAST
- [x] Stack definida: Go + Cobra + Viper
- [x] Análise de Gateways (Telegram, WhatsApp, Email, Google Chat)
- [x] Documentação de pesquisa de mercado

### ✅ Fase 01: Bootstrap & CLI Skeleton
- [x] Estrutura de pastas (Standard Go Layout)
- [x] `go.mod` configurado
- [x] Comando raiz (`root.go`) com banner ASCII
- [x] Comando `send` básico
- [x] Configuração do Viper (ENV > File)
- [x] Suporte a múltiplos formatos (YAML, JSON, Properties)
- [x] Help traduzido para português
- [x] Dummy implementation do comando `send`
- [x] Validação de argumentos
- [x] Exit codes configurados
- [x] Script de build (`scripts/build.bat`)
- [x] Configuração VS Code (`.vscode/settings.json`)

---

## 🚧 FASE 01 - DETALHAMENTO

### ✅ Estrutura de Pastas
```
/cast
  /cmd/cast/          ✅ main.go, root.go, send.go
  /internal/
    /config/          ✅ config.go (Viper)
    /providers/       ✅ provider.go (Interface)
  /specifications/    ✅ 00-04
  /documents/         ✅ Tutoriais 01-04
  /tests/             ✅ Criado
  /scripts/           ✅ build.bat
  /logs/              ✅ Criado
  /run/               ✅ Criado
```

### ✅ Funcionalidades Implementadas

#### CLI Core
- [x] Banner ASCII (Verde Claro)
- [x] Help em português
- [x] Comando `send` com validação de argumentos
- [x] Mensagens de erro em português
- [x] Exit codes: 0 (sucesso), 1 (erro), 2 (config)

#### Configuração
- [x] Struct `Config` com `TelegramConfig`
- [x] Função `Load()` - carrega ENV e arquivos
- [x] Função `LoadConfig()` - retorna struct
- [x] Suporte a YAML, JSON, Properties
- [x] Ordem de precedência: ENV > File
- [x] Integração no `main.go`

#### Comando Send (Dummy)
- [x] Aceita 3 argumentos: `[provider] [target] [message]`
- [x] Validação com `cobra.MinimumNArgs(3)`
- [x] Imprime mensagem dummy: `"Sending via [provider] to [target]: [message]"`
- [x] Feedback visual em verde

### ⚠️ Pendências Fase 01
- [ ] Struct completa de Config (WhatsApp, Email, GoogleChat, Aliases)
- [ ] Validação de configuração obrigatória
- [ ] Sistema de aliases funcional
- [ ] Comando para configurar gateways (wizard/interativo)

---

## 📋 FASE 02 - IMPLEMENTAÇÃO DE DRIVERS (PENDENTE)

### 🔴 Driver: Telegram
- [ ] Implementar `TelegramProvider` (interface `Provider`)
- [ ] HTTP POST para API do Telegram
- [ ] Tratamento de erros da API
- [ ] Suporte a Chat ID e aliases
- [ ] Testes unitários

### 🔴 Driver: Email (SMTP)
- [ ] Implementar `EmailProvider` (interface `Provider`)
- [ ] Conexão SMTP com TLS/SSL
- [ ] Suporte a HTML e anexos
- [ ] Compatibilidade com Gmail, SendGrid, Resend
- [ ] Testes unitários

### 🔴 Integração
- [ ] Factory de providers
- [ ] Resolução de aliases
- [ ] Integração com comando `send`
- [ ] Logging estruturado
- [ ] Tratamento de erros de rede

---

## 📋 FASE 03 - INTEGRAÇÃO AVANÇADA (PENDENTE)

### 🔴 Driver: WhatsApp
- [ ] Implementar `WhatsAppProvider`
- [ ] Integração com Meta Cloud API
- [ ] Suporte a Sandbox e Produção
- [ ] Tratamento de templates (Sandbox)

### 🔴 Driver: Google Chat
- [ ] Implementar `GoogleChatProvider`
- [ ] Incoming Webhook
- [ ] Formatação de mensagens

---

## 📋 FASE 04 - BUILD & RELEASE (PENDENTE)

### 🔴 Build
- [ ] Cross-compilation (Windows/Linux)
- [ ] Scripts de build para múltiplas plataformas
- [ ] Versionamento automático
- [ ] Releases no GitHub

### 🔴 Documentação
- [ ] README completo
- [ ] Guia de instalação
- [ ] Exemplos de uso
- [ ] Changelog

---

## 📚 DOCUMENTAÇÃO

### ✅ Especificações
- [x] `00_MASTER_PLAN.md` - Visão geral do projeto
- [x] `01_MARKET_RESEARCH.md` - Pesquisa de gateways
- [x] `02_TECH_SPEC.md` - Especificação técnica
- [x] `03_CLI_UX.md` - Especificação de UX
- [x] `04_GATEWAY_CONFIG_SPEC.md` - Configuração de gateways

### ✅ Tutoriais
- [x] `01_TUTORIAL_TELEGRAM.md` - Configurar Telegram
- [x] `02_TUTORIAL_WHATSAPP.md` - Configurar WhatsApp
- [x] `03_TUTORIAL_EMAIL.md` - Configurar Email
- [x] `04_TUTORIAL_GOOGLE_CHAT.md` - Configurar Google Chat
- [x] `README.md` - Índice dos tutoriais

### ⚠️ Pendente
- [ ] README principal do projeto
- [ ] Guia de instalação
- [ ] Exemplos práticos
- [ ] FAQ

---

## 🏗️ ARQUITETURA ATUAL

### Estrutura de Código

```
cmd/cast/
  main.go      ✅ Entrypoint com config.Load()
  root.go      ✅ Comando raiz + banner + help PT
  send.go      ✅ Comando send (dummy)

internal/
  config/
    config.go  ✅ Viper + Struct Config (parcial)
  providers/
    provider.go ✅ Interface Provider
```

### Interfaces Definidas

```go
type Provider interface {
    Name() string
    Send(target string, message string) error
}
```

### Config Atual

```go
type Config struct {
    Telegram TelegramConfig  ✅ Implementado
    // WhatsApp, Email, GoogleChat, Aliases - ⚠️ Pendente
}
```

---

## 🧪 TESTES

### ✅ Estrutura
- [x] Pasta `tests/` criada
- [x] `.gitignore` configurado

### 🔴 Pendente
- [ ] Testes unitários para `config.Load()`
- [ ] Testes unitários para providers
- [ ] Testes de integração
- [ ] Mocks para APIs externas

---

## 🛠️ FERRAMENTAS E SCRIPTS

### ✅ Implementado
- [x] `scripts/build.bat` - Script de build com logs
- [x] `.vscode/settings.json` - Configuração Go + Terminal
- [x] `.gitignore` - Configurado (run/, logs/, tests/)
- [x] `.cursorrules` - Regras do projeto

### ⚠️ Pendente
- [ ] Scripts de build para Linux
- [ ] Scripts de release
- [ ] CI/CD (GitHub Actions)

---

## 📈 MÉTRICAS

### Código
- **Linhas de código:** ~300
- **Arquivos Go:** 5
- **Comandos CLI:** 2 (root, send)
- **Providers:** 0 implementados (4 pendentes)

### Documentação
- **Especificações:** 5 arquivos
- **Tutoriais:** 4 arquivos
- **Cobertura:** ~80% da Fase 01

---

## 🎯 PRÓXIMOS PASSOS

### Curto Prazo (Fase 01 - Finalização)
1. Completar struct `Config` (WhatsApp, Email, GoogleChat, Aliases)
2. Implementar validação de configuração
3. Sistema de aliases funcional
4. Comando `config` para wizard de configuração

### Médio Prazo (Fase 02)
1. Implementar `TelegramProvider`
2. Implementar `EmailProvider`
3. Testes unitários
4. Integração completa

### Longo Prazo (Fase 03-04)
1. WhatsApp e Google Chat providers
2. Cross-compilation
3. Releases e documentação final

---

## 🔗 LINKS ÚTEIS

- **Especificações:** `/specifications/`
- **Tutoriais:** `/documents/`
- **Código:** `/cmd/cast/`, `/internal/`
- **Scripts:** `/scripts/`

---

## 📝 NOTAS

- O projeto está na **Fase 01** (Bootstrap)
- A estrutura base está completa e funcional
- O comando `send` atualmente apenas imprime mensagens (dummy)
- Próximo foco: Implementar drivers reais (Fase 02)

---

**Mantido por:** Equipe CAST
**Última revisão:** 2025-01-XX
