# TUTORIAL: Configurando WhatsApp (Meta Cloud API) no CAST

Este tutorial guia você passo a passo para configurar o gateway WhatsApp no CAST usando a Meta Cloud API (oficial).

## 📋 PRÉ-REQUISITOS

- Conta no Facebook/Meta
- Acesso ao [Meta for Developers](https://developers.facebook.com/)
- Número de telefone para testes (Sandbox) ou Business Account (Produção)

## 🚀 PASSO 1: Criar App no Meta for Developers

### 1.1 Acessar Meta for Developers

1. Acesse [https://developers.facebook.com/](https://developers.facebook.com/)
2. Faça login com sua conta Facebook/Meta
3. Clique em **"Meus Apps"** no canto superior direito
4. Clique em **"Criar App"**

### 1.2 Escolher Tipo de App

1. Selecione **"Business"** como tipo de app
2. Clique em **"Próximo"**
3. Preencha:
   - **Nome do App**: Ex: `CAST Notifications`
   - **Email de contato**: Seu email
   - **Finalidade do App**: Selecione conforme sua necessidade
4. Clique em **"Criar App"**

### 1.3 Adicionar Produto WhatsApp

1. No painel do app, procure por **"WhatsApp"**
2. Clique em **"Configurar"** no card do WhatsApp
3. Você será redirecionado para o painel do WhatsApp

## 🔑 PASSO 2: Configurar WhatsApp Business API

### 2.1 Acessar Sandbox (Gratuito para testes)

1. No painel do WhatsApp, você verá a opção **"Sandbox"**
2. Clique em **"Iniciar"** ou **"Começar"**
3. Você verá um número de telefone temporário para testes

### 2.2 Obter Phone Number ID

1. No painel do WhatsApp, vá em **"Configurações"** → **"Números de telefone"**
2. Você verá seu **Phone Number ID** (número longo)
   - Exemplo: `123456789012345`
3. **Copie este ID**

### 2.3 Obter Access Token

1. No painel do WhatsApp, vá em **"Configurações"** → **"Tokens de acesso"**
2. Você verá um token temporário (válido por 24 horas)
3. Para token permanente:
   - Clique em **"Gerar token"**
   - Selecione as permissões necessárias
   - Copie o token gerado
   - ⚠️ **Guarde este token com segurança!**

### 2.4 Obter Business Account ID (Opcional)

1. No painel do WhatsApp, vá em **"Configurações"** → **"Contas comerciais"**
2. Copie o **Business Account ID**
   - Exemplo: `987654321098765`

## 📱 PASSO 3: Adicionar Número de Teste (Sandbox)

### 3.1 Adicionar número ao Sandbox

1. No painel do WhatsApp, vá em **"Sandbox"**
2. Clique em **"Adicionar número de telefone"**
3. Digite seu número (formato internacional, sem +)
   - Exemplo: `5511999998888` (Brasil)
4. Clique em **"Enviar código"**
5. Digite o código recebido via WhatsApp
6. Seu número será adicionado ao Sandbox

### 3.2 Limitações do Sandbox

- ⚠️ Apenas números adicionados ao Sandbox podem receber mensagens
- ⚠️ Mensagens devem ser iniciadas pelo sistema (não pode responder)
- ⚠️ Ideal apenas para testes

## 🚀 PASSO 4: Upgrade para Produção (Opcional)

Para usar em produção, você precisa:

1. **Verificar seu app** no Meta for Developers
2. **Configurar Business Account** completa
3. **Adicionar método de pagamento** (mensagens são cobradas)
4. **Solicitar número de telefone** dedicado

**Custos:** Consulte [preços da Meta Cloud API](https://developers.facebook.com/docs/whatsapp/pricing)

## ⚙️ PASSO 5: Configurar no CAST

### 5.1 Opção A: Variáveis de Ambiente

**Windows (CMD):**
```cmd
set CAST_WHATSAPP_PHONE_NUMBER_ID=123456789012345
set CAST_WHATSAPP_ACCESS_TOKEN=EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
set CAST_WHATSAPP_BUSINESS_ACCOUNT_ID=987654321098765
set CAST_WHATSAPP_API_VERSION=v18.0
```

**Windows (PowerShell):**
```powershell
$env:CAST_WHATSAPP_PHONE_NUMBER_ID="123456789012345"
$env:CAST_WHATSAPP_ACCESS_TOKEN="EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
$env:CAST_WHATSAPP_BUSINESS_ACCOUNT_ID="987654321098765"
$env:CAST_WHATSAPP_API_VERSION="v18.0"
```

**Linux/Mac:**
```bash
export CAST_WHATSAPP_PHONE_NUMBER_ID="123456789012345"
export CAST_WHATSAPP_ACCESS_TOKEN="EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export CAST_WHATSAPP_BUSINESS_ACCOUNT_ID="987654321098765"
export CAST_WHATSAPP_API_VERSION="v18.0"
```

### 5.2 Opção B: Arquivo YAML (`cast.yaml`)

```yaml
whatsapp:
  phone_number_id: "123456789012345"
  access_token: "EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  business_account_id: "987654321098765"
  api_version: "v18.0"
  api_url: "https://graph.facebook.com"
  timeout: 30
```

### 5.3 Opção C: Arquivo JSON (`cast.json`)

```json
{
  "whatsapp": {
    "phone_number_id": "123456789012345",
    "access_token": "EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "business_account_id": "987654321098765",
    "api_version": "v18.0",
    "api_url": "https://graph.facebook.com",
    "timeout": 30
  }
}
```

### 5.4 Opção D: Arquivo Properties (`cast.properties`)

```properties
whatsapp.phone_number_id=123456789012345
whatsapp.access_token=EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
whatsapp.business_account_id=987654321098765
whatsapp.api_version=v18.0
whatsapp.api_url=https://graph.facebook.com
whatsapp.timeout=30
```

## ✅ PASSO 6: Testar a Configuração

### 6.1 Enviar mensagem de teste

```bash
cast send zap 5511999998888 "Teste de configuração do WhatsApp!"
```

**Formato do número:** Use formato internacional sem `+`
- Brasil: `5511999998888` (55 = código do país, 11 = DDD, resto = número)

### 6.2 Verificar se funcionou

Se a mensagem aparecer no WhatsApp, a configuração está correta! ✅

## 🔧 CONFIGURAÇÕES AVANÇADAS

### Versão da API

A versão padrão é `v18.0`. Para usar outra versão:

```yaml
whatsapp:
  api_version: "v19.0"  # Versão mais recente
```

### Timeout Customizado

```yaml
whatsapp:
  timeout: 60  # 60 segundos
```

## 🎯 CONFIGURANDO ALIASES

```yaml
whatsapp:
  phone_number_id: "seu-phone-number-id"
  access_token: "seu-access-token"

aliases:
  alerts:
    provider: "zap"
    target: "5511999998888"
    name: "WhatsApp de Alertas"

  pessoal:
    provider: "zap"
    target: "5511888887777"
    name: "WhatsApp Pessoal"
```

Uso:

```bash
cast send zap alerts "Alerta crítico!"
cast send zap pessoal "Mensagem pessoal"
```

## ⚠️ SEGURANÇA

- **NUNCA** compartilhe seu Access Token
- **NUNCA** commite tokens em repositórios Git
- Tokens temporários expiram em 24 horas
- Use tokens permanentes apenas em produção
- Revogue tokens comprometidos no painel do Meta

## 📚 REFERÊNCIAS

- [Meta for Developers - WhatsApp](https://developers.facebook.com/docs/whatsapp)
- [WhatsApp Cloud API - Documentação](https://developers.facebook.com/docs/whatsapp/cloud-api)
- [Guia de Início Rápido](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started)
- [Preços e Cobrança](https://developers.facebook.com/docs/whatsapp/pricing)
- [Especificação CAST - WhatsApp](specifications/04_GATEWAY_CONFIG_SPEC.md#3-whatsapp-meta-cloud-api)

## 🆘 SOLUÇÃO DE PROBLEMAS

### Erro: "Invalid OAuth access token"
- Verifique se o token está correto
- Tokens temporários expiram em 24h - gere um novo
- Verifique se o token tem as permissões corretas

### Erro: "Phone number not found"
- Certifique-se de que o número está no Sandbox
- Verifique o formato do número (sem +, formato internacional)

### Erro: "Message template required"
- No Sandbox, você pode enviar apenas mensagens de template
- Para mensagens livres, precisa de conta de produção

### Erro: "Rate limit exceeded"
- Você atingiu o limite de mensagens
- Aguarde alguns minutos ou faça upgrade para produção

### Mensagem não chega
- Verifique se o número está no Sandbox
- Confirme que o número está no formato correto
- Verifique logs do CAST (se habilitado)
