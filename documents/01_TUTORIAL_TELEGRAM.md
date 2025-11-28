# TUTORIAL: Configurando Telegram no CAST

Este tutorial guia você passo a passo para configurar o gateway Telegram no CAST, desde a criação do bot até a configuração no CLI.

## 📋 PRÉ-REQUISITOS

- Conta no Telegram (gratuita)
- Aplicativo Telegram instalado (mobile ou desktop)
- Acesso ao [@BotFather](https://t.me/botfather)

## 🚀 PASSO 1: Criar um Bot no Telegram

### 1.1 Iniciar conversa com BotFather

1. Abra o Telegram (app ou web)
2. Busque por `@BotFather` na barra de pesquisa
3. Clique em "Iniciar" ou envie `/start`

### 1.2 Criar um novo bot

1. Envie o comando `/newbot`
2. BotFather perguntará o **nome do seu bot** (nome público)
   - Exemplo: `Meu Bot de Notificações`
3. BotFather perguntará o **username do bot** (deve terminar em `bot`)
   - Exemplo: `meu_bot_notificacoes_bot`
   - ⚠️ O username deve ser único e terminar com `bot`

### 1.3 Obter o Token

Após criar o bot, BotFather retornará uma mensagem como:

```
Done! Congratulations on your new bot. You will find it at t.me/meu_bot_notificacoes_bot.

Use this token to access the HTTP API:
1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890

Keep your token secure and store it safely, it can be used by anyone to control your bot.
```

**Copie o token** (ex: `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890`)

### 1.4 Comandos úteis do BotFather

- `/token` - Ver o token atual do bot
- `/revoke` - Revogar e gerar novo token
- `/setdescription` - Definir descrição do bot
- `/setabouttext` - Definir texto "Sobre"

## 🔑 PASSO 2: Obter o Chat ID

O Chat ID identifica a conversa onde o bot enviará mensagens. Existem duas formas:

### 2.1 Método 1: Usando @userinfobot

1. Busque por `@userinfobot` no Telegram
2. Inicie conversa e envie `/start`
3. O bot retornará seu Chat ID (número)
   - Exemplo: `123456789`

### 2.2 Método 2: Usando a API do Telegram

1. Envie uma mensagem para seu bot (qualquer mensagem)
2. Acesse no navegador:
   ```
   https://api.telegram.org/bot<SEU_TOKEN>/getUpdates
   ```
   Substitua `<SEU_TOKEN>` pelo token obtido no Passo 1.3

3. Procure por `"chat":{"id":` no JSON retornado
   - O número após `"id":` é seu Chat ID

### 2.3 Método 3: Usando @getidsbot

1. Busque por `@getidsbot` no Telegram
2. Inicie conversa e envie `/start`
3. O bot retornará seu Chat ID

## ⚙️ PASSO 3: Configurar no CAST

Agora vamos configurar o Telegram no CAST. Você pode usar **variáveis de ambiente** ou **arquivo de configuração**.

### 3.1 Opção A: Variáveis de Ambiente (Recomendado para produção)

**Windows (CMD):**
```cmd
set CAST_TELEGRAM_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890
set CAST_TELEGRAM_DEFAULT_CHAT_ID=123456789
```

**Windows (PowerShell):**
```powershell
$env:CAST_TELEGRAM_TOKEN="1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890"
$env:CAST_TELEGRAM_DEFAULT_CHAT_ID="123456789"
```

**Linux/Mac:**
```bash
export CAST_TELEGRAM_TOKEN="1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890"
export CAST_TELEGRAM_DEFAULT_CHAT_ID="123456789"
```

**Docker/Kubernetes:**
```yaml
env:
  - name: CAST_TELEGRAM_TOKEN
    value: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890"
  - name: CAST_TELEGRAM_DEFAULT_CHAT_ID
    value: "123456789"
```

### 3.2 Opção B: Arquivo YAML (`cast.yaml`)

Crie um arquivo `cast.yaml` no diretório onde você executa o CAST:

```yaml
telegram:
  token: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890"
  default_chat_id: "123456789"
  api_url: "https://api.telegram.org/bot"
  timeout: 30
```

### 3.3 Opção C: Arquivo JSON (`cast.json`)

Crie um arquivo `cast.json`:

```json
{
  "telegram": {
    "token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890",
    "default_chat_id": "123456789",
    "api_url": "https://api.telegram.org/bot",
    "timeout": 30
  }
}
```

### 3.4 Opção D: Arquivo Properties (`cast.properties`)

Crie um arquivo `cast.properties`:

```properties
telegram.token=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890
telegram.default_chat_id=123456789
telegram.api_url=https://api.telegram.org/bot
telegram.timeout=30
```

## ✅ PASSO 4: Testar a Configuração

### 4.1 Enviar mensagem de teste

```bash
cast send tg 123456789 "Teste de configuração do Telegram!"
```

Ou usando alias (se configurado):

```bash
cast send tg me "Teste de configuração do Telegram!"
```

### 4.2 Verificar se funcionou

Se a mensagem aparecer no Telegram, a configuração está correta! ✅

## 🔧 CONFIGURAÇÕES AVANÇADAS

### Timeout Customizado

Se você tiver problemas de timeout, aumente o valor:

```yaml
telegram:
  token: "seu-token"
  default_chat_id: "seu-chat-id"
  timeout: 60  # 60 segundos
```

### API URL Customizada

Para usar um proxy ou servidor alternativo:

```yaml
telegram:
  token: "seu-token"
  default_chat_id: "seu-chat-id"
  api_url: "https://api.telegram.org/bot"
```

## 🎯 CONFIGURANDO ALIASES

Para facilitar o uso, configure aliases no `cast.yaml`:

```yaml
telegram:
  token: "seu-token"
  default_chat_id: "123456789"

aliases:
  me:
    provider: "tg"
    target: "123456789"
    name: "Meu Telegram Pessoal"

  trabalho:
    provider: "tg"
    target: "987654321"
    name: "Telegram do Trabalho"
```

Depois use:

```bash
cast send tg me "Mensagem pessoal"
cast send tg trabalho "Mensagem do trabalho"
```

## ⚠️ SEGURANÇA

- **NUNCA** compartilhe seu token publicamente
- **NUNCA** commite tokens em repositórios Git
- Use variáveis de ambiente em produção
- Revogue tokens comprometidos via `/revoke` no BotFather

## 📚 REFERÊNCIAS

- [Documentação Oficial da API do Telegram](https://core.telegram.org/bots/api)
- [BotFather - Guia Completo](https://core.telegram.org/bots/tutorial)
- [Especificação CAST - Telegram](specifications/04_GATEWAY_CONFIG_SPEC.md#2-telegram-bot-api)

## 🆘 SOLUÇÃO DE PROBLEMAS

### Erro: "Unauthorized"
- Verifique se o token está correto
- Use `/token` no BotFather para verificar

### Erro: "Chat not found"
- Certifique-se de ter enviado `/start` para o bot
- Verifique se o Chat ID está correto

### Erro: "Timeout"
- Aumente o valor de `timeout` na configuração
- Verifique sua conexão com a internet

### Mensagem não chega
- Verifique se o bot está ativo
- Confirme que você iniciou conversa com o bot (`/start`)
- Verifique logs do CAST (se habilitado)
