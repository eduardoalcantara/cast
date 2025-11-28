# ESPECIFICAÇÃO DE CONFIGURAÇÃO: GATEWAYS

**Objetivo:** Definir os parâmetros de configuração necessários para cadastrar e configurar os gateways de mensageria suportados pelo CAST.

## 1. ESTRUTURA GERAL DE CONFIGURAÇÃO

A configuração segue a ordem de precedência definida na `02_TECH_SPEC.md`:
1. Variáveis de Ambiente (`CAST_*`)
2. Arquivo Local (`cast.*`)

### 1.1 Estrutura Base (Struct Go)

```go
type Config struct {
    Telegram  TelegramConfig  `mapstructure:"telegram"`
    WhatsApp  WhatsAppConfig  `mapstructure:"whatsapp"`
    Email     EmailConfig     `mapstructure:"email"`
    GoogleChat GoogleChatConfig `mapstructure:"google_chat"`
    Aliases   map[string]AliasConfig `mapstructure:"aliases"`
}
```

## 2. TELEGRAM (Bot API)

### 2.1 Parâmetros Obrigatórios
- `token`: Token do Bot obtido via [@BotFather](https://t.me/botfather)
- `default_chat_id`: Chat ID padrão (opcional, pode ser sobrescrito por alias)

### 2.2 Parâmetros Opcionais
- `api_url`: URL da API (padrão: `https://api.telegram.org/bot`)
- `timeout`: Timeout em segundos (padrão: `30`)

### 2.3 Estrutura (Struct Go)

```go
type TelegramConfig struct {
    Token        string `mapstructure:"token"`
    DefaultChatID string `mapstructure:"default_chat_id"`
    APIURL       string `mapstructure:"api_url"`
    Timeout      int    `mapstructure:"timeout"`
}
```

### 2.4 Exemplos de Configuração

**YAML (`cast.yaml`):**
```yaml
telegram:
  token: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
  default_chat_id: "123456789"
  api_url: "https://api.telegram.org/bot"
  timeout: 30
```

**JSON (`cast.json`):**
```json
{
  "telegram": {
    "token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
    "default_chat_id": "123456789",
    "api_url": "https://api.telegram.org/bot",
    "timeout": 30
  }
}
```

**Properties (`cast.properties`):**
```properties
telegram.token=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
telegram.default_chat_id=123456789
telegram.api_url=https://api.telegram.org/bot
telegram.timeout=30
```

**Variáveis de Ambiente:**
```bash
CAST_TELEGRAM_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
CAST_TELEGRAM_DEFAULT_CHAT_ID=123456789
CAST_TELEGRAM_API_URL=https://api.telegram.org/bot
CAST_TELEGRAM_TIMEOUT=30
```

## 3. WHATSAPP (Meta Cloud API)

### 3.1 Parâmetros Obrigatórios
- `phone_number_id`: Phone Number ID da Meta Cloud API
- `access_token`: Access Token da Meta Cloud API
- `business_account_id`: Business Account ID (opcional para Sandbox)

### 3.2 Parâmetros Opcionais
- `api_version`: Versão da API (padrão: `v18.0`)
- `api_url`: URL base da API (padrão: `https://graph.facebook.com`)
- `timeout`: Timeout em segundos (padrão: `30`)

### 3.3 Estrutura (Struct Go)

```go
type WhatsAppConfig struct {
    PhoneNumberID    string `mapstructure:"phone_number_id"`
    AccessToken      string `mapstructure:"access_token"`
    BusinessAccountID string `mapstructure:"business_account_id"`
    APIVersion       string `mapstructure:"api_version"`
    APIURL           string `mapstructure:"api_url"`
    Timeout          int    `mapstructure:"timeout"`
}
```

### 3.4 Exemplos de Configuração

**YAML (`cast.yaml`):**
```yaml
whatsapp:
  phone_number_id: "123456789012345"
  access_token: "EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  business_account_id: "987654321098765"
  api_version: "v18.0"
  api_url: "https://graph.facebook.com"
  timeout: 30
```

**Variáveis de Ambiente:**
```bash
CAST_WHATSAPP_PHONE_NUMBER_ID=123456789012345
CAST_WHATSAPP_ACCESS_TOKEN=EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
CAST_WHATSAPP_BUSINESS_ACCOUNT_ID=987654321098765
CAST_WHATSAPP_API_VERSION=v18.0
CAST_WHATSAPP_API_URL=https://graph.facebook.com
CAST_WHATSAPP_TIMEOUT=30
```

## 4. EMAIL (SMTP)

### 4.1 Parâmetros Obrigatórios
- `smtp_host`: Host do servidor SMTP
- `smtp_port`: Porta do servidor SMTP (padrão: `587` para TLS, `465` para SSL)
- `username`: Username/email para autenticação
- `password`: Senha ou App Password

### 4.2 Parâmetros Opcionais
- `from_email`: Email remetente (padrão: usa `username`)
- `from_name`: Nome do remetente
- `use_tls`: Usar TLS (padrão: `true`)
- `use_ssl`: Usar SSL (padrão: `false`, mutuamente exclusivo com TLS)
- `timeout`: Timeout em segundos (padrão: `30`)

### 4.3 Estrutura (Struct Go)

```go
type EmailConfig struct {
    SMTPHost  string `mapstructure:"smtp_host"`
    SMTPPort  int    `mapstructure:"smtp_port"`
    Username  string `mapstructure:"username"`
    Password  string `mapstructure:"password"`
    FromEmail string `mapstructure:"from_email"`
    FromName  string `mapstructure:"from_name"`
    UseTLS    bool   `mapstructure:"use_tls"`
    UseSSL    bool   `mapstructure:"use_ssl"`
    Timeout   int    `mapstructure:"timeout"`
}
```

### 4.4 Exemplos de Configuração

**YAML (`cast.yaml`):**
```yaml
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "seu-email@gmail.com"
  password: "sua-app-password"
  from_email: "seu-email@gmail.com"
  from_name: "CAST Notifications"
  use_tls: true
  use_ssl: false
  timeout: 30
```

**Gmail com App Password:**
```yaml
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "usuario@gmail.com"
  password: "abcd efgh ijkl mnop"  # App Password do Gmail
  use_tls: true
```

**SendGrid:**
```yaml
email:
  smtp_host: "smtp.sendgrid.net"
  smtp_port: 587
  username: "apikey"
  password: "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  from_email: "noreply@empresa.com"
  use_tls: true
```

**Variáveis de Ambiente:**
```bash
CAST_EMAIL_SMTP_HOST=smtp.gmail.com
CAST_EMAIL_SMTP_PORT=587
CAST_EMAIL_USERNAME=seu-email@gmail.com
CAST_EMAIL_PASSWORD=sua-app-password
CAST_EMAIL_FROM_EMAIL=seu-email@gmail.com
CAST_EMAIL_FROM_NAME=CAST Notifications
CAST_EMAIL_USE_TLS=true
CAST_EMAIL_USE_SSL=false
CAST_EMAIL_TIMEOUT=30
```

## 5. GOOGLE CHAT (Incoming Webhook)

### 5.1 Parâmetros Obrigatórios
- `webhook_url`: URL do Incoming Webhook do Google Chat

### 5.2 Parâmetros Opcionais
- `timeout`: Timeout em segundos (padrão: `30`)

### 5.3 Estrutura (Struct Go)

```go
type GoogleChatConfig struct {
    WebhookURL string `mapstructure:"webhook_url"`
    Timeout    int    `mapstructure:"timeout"`
}
```

### 5.4 Exemplos de Configuração

**YAML (`cast.yaml`):**
```yaml
google_chat:
  webhook_url: "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"
  timeout: 30
```

**Variáveis de Ambiente:**
```bash
CAST_GOOGLE_CHAT_WEBHOOK_URL=https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ
CAST_GOOGLE_CHAT_TIMEOUT=30
```

## 6. ALIASES (Atalhos para Targets)

### 6.1 Conceito
Aliases permitem criar atalhos para targets frequentes, facilitando o uso do CLI.

### 6.2 Estrutura (Struct Go)

```go
type AliasConfig struct {
    Provider string `mapstructure:"provider"`
    Target   string `mapstructure:"target"`
    Name     string `mapstructure:"name"`  // Nome descritivo (opcional)
}
```

### 6.3 Exemplos de Configuração

**YAML (`cast.yaml`):**
```yaml
aliases:
  me:
    provider: "tg"
    target: "123456789"
    name: "Meu Telegram Pessoal"

  team:
    provider: "google_chat"
    target: "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"
    name: "Time de Desenvolvimento"

  alerts:
    provider: "zap"
    target: "5511999998888"
    name: "WhatsApp de Alertas"
```

**Uso no CLI:**
```bash
# Ao invés de:
cast send tg 123456789 "Mensagem"

# Pode usar:
cast send tg me "Mensagem"
```

## 7. VALIDAÇÕES E REGRAS

### 7.1 Validações Obrigatórias
- **Telegram:** `token` não pode estar vazio
- **WhatsApp:** `phone_number_id` e `access_token` não podem estar vazios
- **Email:** `smtp_host`, `username` e `password` não podem estar vazios
- **Google Chat:** `webhook_url` não pode estar vazio e deve ser uma URL válida

### 7.2 Regras de Negócio
- `use_tls` e `use_ssl` são mutuamente exclusivos (se ambos `true`, priorizar TLS)
- Se `from_email` não especificado, usar `username`
- Aliases devem ter `provider` e `target` válidos
- Timeout mínimo: 5 segundos, máximo: 300 segundos

### 7.3 Mensagens de Erro
- Configuração faltando: `"configuração obrigatória não encontrada: [gateway].[campo]"`
- URL inválida: `"URL inválida para [gateway]: [url]"`
- Alias não encontrado: `"alias não encontrado: [alias]"`

## 8. EXEMPLO COMPLETO (cast.yaml)

```yaml
telegram:
  token: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
  default_chat_id: "123456789"
  timeout: 30

whatsapp:
  phone_number_id: "123456789012345"
  access_token: "EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  api_version: "v18.0"
  timeout: 30

email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "notificacoes@empresa.com"
  password: "app-password-aqui"
  from_email: "notificacoes@empresa.com"
  from_name: "Sistema CAST"
  use_tls: true
  timeout: 30

google_chat:
  webhook_url: "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"
  timeout: 30

aliases:
  me:
    provider: "tg"
    target: "123456789"
    name: "Meu Telegram"

  dev-team:
    provider: "google_chat"
    target: "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"
    name: "Time de Desenvolvimento"
```

## 9. NOTAS DE IMPLEMENTAÇÃO

- Todos os campos sensíveis (tokens, senhas) devem ser mascarados nos logs
- Configurações devem ser validadas no momento do carregamento (`config.Load()`)
- Erros de configuração devem retornar exit code `2` (conforme `.cursorrules`)
- Suporte a múltiplos gateways do mesmo tipo (futuro: namespaces)
- Um help bem detalhado deve ser implementado no CLI para cada tipo de gateway que o usuário deseje configurar, inlcuindo URLs oficiais de suporte desses gatways por onde o usuário pode conseguir mais detalhes.
- A configuração deverá ter dois modos: via parâmetros e via wizard.
- O usuário deverá escolher como cadastrar suas configurações: em variáveis de ambiente, yaml, properties ou json.
