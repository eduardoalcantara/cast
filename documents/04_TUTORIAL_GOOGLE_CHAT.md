# TUTORIAL: Configurando Google Chat no CAST

Este tutorial guia você passo a passo para configurar o gateway Google Chat no CAST usando Incoming Webhooks.

## 📋 PRÉ-REQUISITOS

- Conta Google Workspace (antigo G Suite) ou Google Account pessoal
- Acesso ao Google Chat
- Permissão para criar webhooks no espaço/chat

## 🚀 PASSO 1: Acessar Google Chat

### 1.1 Abrir Google Chat

1. Acesse [https://chat.google.com/](https://chat.google.com/)
2. Faça login com sua conta Google
3. Você verá seus espaços e conversas

## 🔗 PASSO 2: Criar Incoming Webhook

### 2.1 Criar ou Acessar um Espaço

1. No Google Chat, clique em **"Espaços"** no menu lateral
2. Escolha um espaço existente ou crie um novo:
   - Clique em **"+"** ao lado de "Espaços"
   - Digite um nome: `CAST Notifications`
   - Clique em **"Criar"**

### 2.2 Adicionar Webhook ao Espaço

1. No espaço criado, clique no **ícone de configurações** (⚙️) no topo
2. Vá em **"Aplicativos e integrações"** ou **"Apps and integrations"**
3. Procure por **"Incoming Webhooks"** ou **"Webhooks de entrada"**
4. Clique em **"Configurar"** ou **"Add"**

### 2.3 Configurar o Webhook

1. Dê um nome para o webhook: `CAST Notifications`
2. (Opcional) Adicione um avatar/ícone
3. Clique em **"Salvar"** ou **"Save"**

### 2.4 Obter URL do Webhook

Após salvar, você verá uma URL como:

```
https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ
```

**Copie esta URL completa** - você precisará dela para configurar no CAST.

⚠️ **IMPORTANTE:** Esta URL é sensível e permite enviar mensagens ao espaço. Mantenha-a segura!

## ⚙️ PASSO 3: Configurar no CAST

### 3.1 Opção A: Variáveis de Ambiente

**Windows (CMD):**
```cmd
set CAST_GOOGLE_CHAT_WEBHOOK_URL=https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ
set CAST_GOOGLE_CHAT_TIMEOUT=30
```

**Windows (PowerShell):**
```powershell
$env:CAST_GOOGLE_CHAT_WEBHOOK_URL="https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"
$env:CAST_GOOGLE_CHAT_TIMEOUT="30"
```

**Linux/Mac:**
```bash
export CAST_GOOGLE_CHAT_WEBHOOK_URL="https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"
export CAST_GOOGLE_CHAT_TIMEOUT="30"
```

### 3.2 Opção B: Arquivo YAML (`cast.yaml`)

```yaml
google_chat:
  webhook_url: "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"
  timeout: 30
```

### 3.3 Opção C: Arquivo JSON (`cast.json`)

```json
{
  "google_chat": {
    "webhook_url": "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ",
    "timeout": 30
  }
}
```

### 3.4 Opção D: Arquivo Properties (`cast.properties`)

```properties
google_chat.webhook_url=https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ
google_chat.timeout=30
```

## ✅ PASSO 4: Testar a Configuração

### 4.1 Enviar mensagem de teste

```bash
cast send google_chat "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ" "Teste de configuração do Google Chat!"
```

### 4.2 Verificar se funcionou

Verifique o espaço no Google Chat. Se a mensagem aparecer, a configuração está correta! ✅

## 🔧 CONFIGURAÇÕES AVANÇADAS

### Timeout Customizado

```yaml
google_chat:
  webhook_url: "sua-url-webhook"
  timeout: 60  # 60 segundos
```

## 🎯 CONFIGURANDO ALIASES

Para facilitar o uso, configure aliases no `cast.yaml`:

```yaml
google_chat:
  webhook_url: "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"

aliases:
  team:
    provider: "google_chat"
    target: "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ"
    name: "Time de Desenvolvimento"

  alerts:
    provider: "google_chat"
    target: "https://chat.googleapis.com/v1/spaces/YYYYY/messages?key=AAAAA&token=BBBBB"
    name: "Canal de Alertas"
```

Depois use:

```bash
cast send google_chat team "Mensagem para o time"
cast send google_chat alerts "Alerta crítico!"
```

## 🔄 MÚLTIPLOS WEBHOOKS

Você pode configurar múltiplos webhooks para diferentes espaços:

```yaml
aliases:
  dev-team:
    provider: "google_chat"
    target: "https://chat.googleapis.com/v1/spaces/DEV/messages?key=KEY1&token=TOKEN1"
    name: "Time de Desenvolvimento"

  prod-alerts:
    provider: "google_chat"
    target: "https://chat.googleapis.com/v1/spaces/PROD/messages?key=KEY2&token=TOKEN2"
    name: "Alertas de Produção"
```

## ⚠️ SEGURANÇA

- **NUNCA** compartilhe sua URL de webhook publicamente
- **NUNCA** commite URLs de webhook em repositórios Git
- Use variáveis de ambiente em produção
- Revogue webhooks comprometidos no Google Chat
- Cada webhook pode ser revogado individualmente nas configurações do espaço

## 🔒 REVOGAR WEBHOOK

Se sua URL de webhook for comprometida:

1. Acesse o espaço no Google Chat
2. Vá em **Configurações** → **Aplicativos e integrações**
3. Encontre o webhook "CAST Notifications"
4. Clique em **"Remover"** ou **"Delete"**
5. Crie um novo webhook se necessário

## 📚 REFERÊNCIAS

- [Google Chat API - Incoming Webhooks](https://developers.google.com/chat/api/guides/messages/formats)
- [Google Chat - Documentação Oficial](https://developers.google.com/chat)
- [Criar e gerenciar webhooks](https://support.google.com/chat/answer/7650837)
- [Especificação CAST - Google Chat](specifications/04_GATEWAY_CONFIG_SPEC.md#5-google-chat-incoming-webhook)

## 🆘 SOLUÇÃO DE PROBLEMAS

### Erro: "Invalid webhook URL"
- Verifique se a URL está completa e correta
- Certifique-se de que copiou toda a URL (incluindo `?key=...&token=...`)
- Tente criar um novo webhook

### Erro: "Webhook not found"
- O webhook pode ter sido removido
- Verifique se você ainda tem acesso ao espaço
- Crie um novo webhook

### Erro: "Permission denied"
- Verifique se você tem permissão para enviar mensagens no espaço
- Certifique-se de que o webhook está ativo
- Verifique se você não foi removido do espaço

### Erro: "Timeout"
- Aumente o valor de `timeout` na configuração
- Verifique sua conexão com a internet
- O Google Chat pode estar temporariamente indisponível

### Mensagem não aparece no chat
- Verifique se a URL do webhook está correta
- Confirme que o webhook ainda está ativo
- Verifique logs do CAST (se habilitado)
- Teste a URL diretamente com `curl`:

```bash
curl -X POST "https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=YYYYY&token=ZZZZZ" \
  -H "Content-Type: application/json" \
  -d '{"text": "Teste"}'
```

### Webhook parou de funcionar
- Webhooks podem expirar ou ser revogados
- Crie um novo webhook e atualize a configuração
- Verifique se o espaço ainda existe
