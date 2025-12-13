# CAST - CAST Automates Sending Tasks

**Versão:** 0.7.0 | **Status:** 🟡 Em Desenvolvimento

Ferramenta CLI standalone para envio agnóstico de mensagens (Fire & Forget) via múltiplos gateways de comunicação.

---

## 📋 Índice

- [Visão Geral](#visão-geral)
- [Características](#características)
- [Instalação](#instalação)
- [Uso Rápido](#uso-rápido)
- [Providers Suportados](#providers-suportados)
- [Comandos](#comandos)
- [Configuração](#configuração)
- [Exemplos Práticos](#exemplos-práticos)
- [Implementações Pendentes](#implementações-pendentes)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Contribuição](#contribuição)

---

## 🎯 Visão Geral

O **CAST** é uma ferramenta de linha de comando escrita em Go que permite enviar mensagens através de múltiplos canais de comunicação de forma unificada e simples. Ideal para automações, notificações, scripts e integrações.

### Principais Vantagens

- ✅ **Agnóstico de Provider**: Interface única para múltiplos gateways
- ✅ **Fire & Forget**: Envio assíncrono sem bloqueio
- ✅ **Standalone**: Binário único, sem dependências externas
- ✅ **Multi-formato**: Suporte a YAML, JSON e Properties
- ✅ **Multi-ambiente**: Variáveis de ambiente + arquivos de configuração
- ✅ **CLI Intuitivo**: Wizards interativos e comandos simples
- ✅ **100% em Português**: Interface e documentação totalmente traduzidas

---

## ✨ Características

### Funcionalidades Principais

- **Envio de Mensagens**: Suporte a múltiplos destinatários em um único comando
- **Aliases**: Atalhos configuráveis para provider + target
- **Wizards Interativos**: Configuração guiada para todos os providers
- **Testes de Conectividade**: Validação de configuração antes de usar
- **Processamento de Quebras de Linha**: Suporte a `\n` e `\n\n` em mensagens
- **Email Avançado**: Suporte a assunto customizado e anexos
- **Debug Mode**: Flag `--verbose` para troubleshooting
- **Busca Inteligente de Config**: Procura no diretório atual e no diretório do executável

### Stack Tecnológica

- **Linguagem**: Go 1.22+
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Configuração**: [Viper](https://github.com/spf13/viper)
- **UI**: [fatih/color](https://github.com/fatih/color)
- **Wizards**: [survey](https://github.com/AlecAivazis/survey/v2)

---

## 📦 Instalação

### Pré-requisitos

- Go 1.22 ou superior (apenas para compilação)
- Binário compilado para sua plataforma (Windows/Linux)

### Compilação

```bash
# Clone o repositório
git clone https://github.com/eduardoalcantara/cast.git
cd cast

# Compile o projeto
go build -o run/cast.exe ./cmd/cast

# O executável estará em ./run/cast.exe
```

### Instalação Global (Opcional)

```bash
# Windows
copy run\cast.exe C:\Windows\System32\

# Linux
sudo cp run/cast /usr/local/bin/
```

---

## 🚀 Uso Rápido

### 1. Configurar um Provider

```bash
# Wizard interativo (recomendado)
cast gateway add --interactive

# Ou via flags (Telegram)
cast gateway add telegram --token "SEU_TOKEN" --default-chat-id 123456789
```

### 2. Enviar uma Mensagem

```bash
# Formato básico
cast send telegram 123456789 "Olá, mundo!"

# Usando alias (mais simples)
cast send me "Mensagem de teste"

# Múltiplos destinatários
cast send email "user1@exemplo.com,user2@exemplo.com" "Notificação"
```

### 3. Criar um Alias

```bash
cast alias add me --provider telegram --target 123456789 --name "Meu Telegram"
```

---

## 📡 Providers Suportados

### ✅ Telegram

- **API Oficial**: Bot API do Telegram
- **Formato**: `cast send telegram <chat_id> <mensagem>`
- **Configuração**: Token do bot + Chat ID padrão (opcional)
- **Recursos**: Suporte a múltiplos destinatários, validação de chat_id

### ✅ WhatsApp (Meta Cloud API)

- **API Oficial**: Meta Cloud API (WhatsApp Business)
- **Formato**: `cast send whatsapp <phone_number> <mensagem>`
- **Configuração**: Phone Number ID, Access Token, Business Account ID
- **Recursos**: Tratamento de janela de 24h, validação de números

### ✅ Email (SMTP)

- **Protocolo**: SMTP com TLS/SSL
- **Formato**: `cast send email <destinatário> <mensagem>`
- **Configuração**: Host, porta, credenciais, TLS/SSL, IMAP (opcional)
- **Recursos**: Assunto customizado, anexos, múltiplos destinatários, **aguardar resposta via IMAP** (`--wfr`)

### ✅ Google Chat

- **API**: Incoming Webhooks
- **Formato**: `cast send googlechat <webhook_url> <mensagem>`
- **Configuração**: URL do webhook
- **Recursos**: Suporte a múltiplos webhooks

### ✅ WAHA (WhatsApp HTTP API)

- **API**: WAHA (WhatsApp HTTP API) - Self-hosted
- **Formato**: `cast send waha <chat_id> <mensagem>`
- **Configuração**: URL da API, sessão, API Key (opcional)
- **Recursos**: Suporte a contatos (`@c.us`) e grupos (`@g.us`), validação robusta

---

## 📖 Comandos

### `cast send`

Envia uma mensagem através do provider especificado.

```bash
# Sintaxe básica
cast send [provider|alias] [target] [message]

# Exemplos
cast send telegram 123456789 "Mensagem"
cast send me "Usando alias"
cast send email admin@empresa.com "Notificação" --subject "Alerta" --attachment arquivo.pdf
cast send waha 5511999998888@c.us "Notificação WAHA"
```

**Flags:**
- `--verbose, -v`: Modo debug (mostra detalhes da requisição)
- `--subject, -s`: Assunto do email (apenas para email)
- `--attachment, -a`: Arquivo anexo (apenas para email, pode ser usado múltiplas vezes)
- `--wfr, --wait-for-response`: Aguarda resposta do destinatário via IMAP (usa tempo do config ou 30min, apenas para email)
- `--wfr-minutes N`: Especifica tempo de espera em minutos (sobrescreve config, apenas para email)

### `cast gateway`

Gerencia configurações de gateways (providers).

```bash
# Adicionar gateway (wizard interativo)
cast gateway add --interactive

# Adicionar gateway via flags
cast gateway add telegram --token "TOKEN" --default-chat-id 123456789

# Listar gateways configurados
cast gateway show

# Mostrar configuração de um provider específico
cast gateway show telegram

# Atualizar configuração
cast gateway update telegram --token "NOVO_TOKEN"

# Testar conectividade
cast gateway test telegram

# Remover gateway
cast gateway remove telegram
```

**Providers suportados:**
- `telegram` ou `tg`
- `whatsapp` ou `zap`
- `email` ou `mail`
- `googlechat` ou `google_chat`
- `waha`

### `cast alias`

Gerencia aliases (atalhos para provider + target).

```bash
# Adicionar alias
cast alias add me --provider telegram --target 123456789 --name "Meu Telegram"

# Listar aliases
cast alias list

# Mostrar detalhes de um alias
cast alias show me

# Atualizar alias
cast alias update me --target 987654321

# Remover alias
cast alias remove me
```

### `cast config`

Comandos gerais de configuração.

```bash
# Mostrar configuração completa (com mascaramento de senhas)
cast config show

# Validar configuração
cast config validate

# Mostrar origem de cada configuração (ENV, FILE, DEFAULT)
cast config sources

# Exportar configuração
cast config export
cast config export --output config-backup.yaml

# Importar configuração
cast config import config-backup.yaml
cast config import config-backup.yaml --replace

# Recarregar configuração
cast config reload
```

---

## ⚙️ Configuração

### Ordem de Precedência

1. **Variáveis de Ambiente** (`CAST_*`) - Maior prioridade
2. **Arquivo Local** (`cast.yaml`, `cast.json`, `cast.properties`) - Diretório atual ou do executável
3. **Valores Padrão** - Menor prioridade

### Busca de Arquivo de Configuração

O CAST procura o arquivo de configuração na seguinte ordem:

1. **Diretório atual** (onde você está executando o comando)
2. **Diretório do executável** (fallback automático)

Isso permite executar o CAST de qualquer diretório sem se preocupar com a localização do arquivo de configuração.

### Formato de Configuração

#### YAML (Recomendado)

```yaml
telegram:
  token: "123456:ABC-DEF..."
  default_chat_id: 123456789
  timeout: 30

whatsapp:
  phone_number_id: "123456789"
  access_token: "EAAG..."
  business_account_id: "987654321"
  timeout: 30

email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "seu-email@gmail.com"
  password: "sua-senha"
  from_email: "seu-email@gmail.com"
  from_name: "Seu Nome"
  use_tls: true
  use_ssl: false
  timeout: 30
  # IMAP: usado apenas se --wait-for-response estiver ativo
  imap_host: "imap.gmail.com"
  imap_port: 993
  imap_username: "seu-email@gmail.com"
  imap_password: "sua-senha"
  imap_use_tls: false
  imap_use_ssl: true
  imap_folder: "INBOX"
  imap_timeout: 60
  imap_poll_interval_seconds: 15  # Intervalo entre ciclos de busca (5-60s)
  # Espera por resposta
  wait_for_response_default_minutes: 0  # 0 = desabilitado por padrão
  wait_for_response_max_minutes: 120     # Teto de segurança
  wait_for_response_max_lines: 0        # 0 = mostrar corpo completo

google_chat:
  webhook_url: "https://chat.googleapis.com/v1/spaces/..."
  timeout: 30

waha:
  api_url: "http://localhost:3000"
  session: "default"
  api_key: "sua-api-key"
  timeout: 30

aliases:
  me:
    provider: telegram
    target: "123456789"
    name: "Meu Telegram"
```

#### Variáveis de Ambiente

```bash
# Telegram
export CAST_TELEGRAM_TOKEN="123456:ABC-DEF..."
export CAST_TELEGRAM_DEFAULT_CHAT_ID=123456789

# WhatsApp
export CAST_WHATSAPP_PHONE_NUMBER_ID="123456789"
export CAST_WHATSAPP_ACCESS_TOKEN="EAAG..."

# Email
export CAST_EMAIL_SMTP_HOST="smtp.gmail.com"
export CAST_EMAIL_SMTP_PORT=587
export CAST_EMAIL_USERNAME="seu-email@gmail.com"
export CAST_EMAIL_PASSWORD="sua-senha"

# Google Chat
export CAST_GOOGLE_CHAT_WEBHOOK_URL="https://chat.googleapis.com/..."

# WAHA
export CAST_WAHA_API_URL="http://localhost:3000"
export CAST_WAHA_SESSION="default"
export CAST_WAHA_API_KEY="sua-api-key"
```

---

## 💡 Exemplos Práticos

### Notificação Simples

```bash
cast send me "Sistema iniciado com sucesso"
```

### Notificação com Quebras de Linha

```bash
cast send tg me "Status do Sistema:\n\n✅ Servidor: Online\n✅ Banco: Conectado\n✅ API: Respondendo"
```

### Email com Anexo

```bash
cast send email admin@empresa.com "Relatório diário em anexo" \
  --subject "Relatório Diário - $(date +%Y-%m-%d)" \
  --attachment relatorio.pdf
```

### Email Aguardando Resposta (IMAP Monitor)

```bash
# Aguarda resposta usando tempo do config ou 30min (padrão)
cast send email destinatario@exemplo.com "Pergunta importante" \
  --subject "Sua opinião" \
  --wfr

# Aguarda 5 minutos específicos
cast send email destinatario@exemplo.com "Pergunta importante" \
  --subject "Sua opinião" \
  --wfr --wfr-minutes 5

# Apenas --wfr-minutes (ativa automaticamente)
cast send email destinatario@exemplo.com "Confirmação" \
  --subject "Confirme recebimento" \
  --wfr-minutes 2 --verbose

# Forma longa --wait-for-response
cast send email destinatario@exemplo.com "Solicitação" \
  --subject "Por favor, responda" \
  --wait-for-response --wfr-minutes 10
```

### Múltiplos Destinatários

```bash
cast send email "user1@empresa.com,user2@empresa.com,user3@empresa.com" \
  "Notificação importante para toda a equipe"
```

### Integração em Scripts

```bash
#!/bin/bash
# Script de monitoramento

if system_is_down; then
  cast send telegram 123456789 "⚠️ Sistema fora do ar!"
  cast send email admin@empresa.com "Alerta: Sistema fora do ar" --subject "ALERTA CRÍTICO"
fi
```

### WAHA (WhatsApp Self-hosted)

```bash
# Enviar para contato
cast send waha 5511999998888@c.us "Notificação via WAHA"

# Enviar para grupo
cast send waha 120363XXXXX@g.us "Mensagem para o grupo"
```

---

## 🚧 Implementações Pendentes

### ✅ Fase 07 - IMAP Monitor (--wait-for-response)

- [x] **Monitoramento IMAP**: Aguarda resposta por email após envio
- [x] **Busca por Message-ID**: Identifica resposta via `In-Reply-To` e `References`
- [x] **Fallback por Subject**: Busca alternativa após alguns ciclos
- [x] **Validação de InReplyTo**: Garante que a resposta corresponde ao email correto
- [x] **Polling Configurável**: Intervalo entre ciclos de busca (5-60 segundos)
- [x] **Exit Codes Específicos**: 0 (resposta recebida), 3 (timeout), 2/4 (erros)
- [x] **Corpo Completo**: Exibe corpo da mensagem de resposta
- [x] **Logs Detalhados**: Modo verbose para debugging IMAP

### 🔴 Fase 08 - Build & Release (Pendente)

- [ ] **Cross-compilation**: Scripts para Windows e Linux
- [ ] **Versionamento Automático**: Integração com Git tags
- [ ] **Releases no GitHub**: Automação de releases
- [ ] **CI/CD**: GitHub Actions para testes e builds
- [ ] **Distribuição**: Binários pré-compilados para download

### Melhorias Futuras

- [ ] **Templates de Mensagem**: Suporte a templates com variáveis
- [ ] **Agendamento**: Envio agendado de mensagens
- [ ] **Retry Automático**: Tentativas automáticas em caso de falha
- [ ] **Rate Limiting**: Controle de taxa de envio
- [ ] **Logging Estruturado**: Logs em formato JSON
- [ ] **Métricas**: Estatísticas de envio
- [ ] **Webhook Receiver**: Receber notificações via webhook
- [ ] **Provider Discord**: Suporte ao Discord
- [ ] **Provider Slack**: Suporte ao Slack

### Documentação

- [ ] **Guia de Instalação Detalhado**: Passo a passo para cada plataforma
- [ ] **FAQ**: Perguntas frequentes
- [ ] **Troubleshooting Guide**: Guia de resolução de problemas
- [ ] **Video Tutorials**: Tutoriais em vídeo

---

## 📁 Estrutura do Projeto

```
cast/
├── cmd/cast/              # Comandos CLI
│   ├── main.go           # Entrypoint
│   ├── root.go           # Comando raiz
│   ├── send.go           # Comando send
│   ├── gateway.go        # Comando gateway
│   ├── alias.go          # Comando alias
│   ├── config.go         # Comando config
│   └── help.go           # Sistema de help customizado
│
├── internal/
│   ├── config/           # Gerenciamento de configuração
│   │   ├── config.go     # Structs e carregamento
│   │   └── manager.go    # Persistência
│   │
│   └── providers/        # Implementação dos providers
│       ├── provider.go   # Interface Provider
│       ├── factory.go    # Factory de providers
│       ├── telegram.go   # Driver Telegram
│       ├── email.go      # Driver Email
│       ├── whatsapp.go   # Driver WhatsApp
│       ├── googlechat.go # Driver Google Chat
│       └── waha.go       # Driver WAHA
│
├── specifications/       # Especificações técnicas
├── documents/            # Tutoriais e documentação
├── results/              # Resultados das fases
├── scripts/              # Scripts de build
├── tests/                # Testes de integração
└── run/                  # Binários compilados
```

### Arquitetura

O CAST segue o **Standard Go Project Layout**:

- **`cmd/`**: Aplicações principais (CLI)
- **`internal/`**: Código privado da aplicação
- **`pkg/`**: Código que pode ser usado por outras aplicações (não utilizado ainda)

### Princípios de Design

- **SOLID**: Separação de responsabilidades, interfaces bem definidas
- **DRY**: Evita duplicação de código
- **Dependency Inversion**: Uso de interfaces para desacoplamento
- **Error Handling**: Tratamento explícito de erros com contexto

---

## 🧪 Testes

### Executar Testes

```bash
# Todos os testes
go test ./...

# Testes de um pacote específico
go test ./internal/providers/...

# Testes com cobertura
go test -cover ./...
```

### Cobertura Atual

- **Testes Unitários**: 39 testes implementados
- **Providers Testados**: Todos os 5 providers (Telegram, Email, WhatsApp, Google Chat, WAHA)
- **Status**: ✅ Todos os testes passando

---

## 🤝 Contribuição

Contribuições são bem-vindas! Por favor:

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/MinhaFeature`)
3. Commit suas mudanças (`git commit -m 'Adiciona MinhaFeature'`)
4. Push para a branch (`git push origin feature/MinhaFeature`)
5. Abra um Pull Request

### Padrões de Código

- Siga o padrão **Effective Go**
- Use `gofmt` e `goimports` para formatação
- Documente funções exportadas com GoDoc
- Adicione testes para novas funcionalidades
- Mantenha a interface em português

---

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

---

## 📞 Suporte

- **Issues**: [GitHub Issues](https://github.com/eduardoalcantara/cast/issues)
- **Documentação**: Veja a pasta `/documents` para tutoriais detalhados
- **Especificações**: Veja a pasta `/specifications` para detalhes técnicos

---

## 🙏 Agradecimentos

- [Cobra](https://github.com/spf13/cobra) - Framework CLI
- [Viper](https://github.com/spf13/viper) - Gerenciamento de configuração
- [fatih/color](https://github.com/fatih/color) - Cores no terminal
- [survey](https://github.com/AlecAivazis/survey/v2) - Wizards interativos

---

**Desenvolvido com ❤️ por Eduardo Alcântara**

*Resposta Nº 41*

*Modelo: claude-3-5-sonnet-20241022*
