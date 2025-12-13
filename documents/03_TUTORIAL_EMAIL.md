# TUTORIAL: Configurando Email (SMTP) no CAST

Este tutorial guia você passo a passo para configurar o gateway Email no CAST usando SMTP, com exemplos para Gmail, SendGrid e outros provedores.

## 📋 PRÉ-REQUISITOS

- Conta de email (Gmail, Outlook, SendGrid, etc.)
- Senha ou App Password (dependendo do provedor)

## 🚀 PASSO 1: Escolher Provedor SMTP

O CAST suporta qualquer servidor SMTP. Exemplos comuns:

- **Gmail** (gratuito, fácil de configurar)
- **Outlook/Hotmail** (gratuito)
- **SendGrid** (gratuito até 100 emails/dia)
- **Resend** (gratuito até 3.000 emails/mês)
- **Servidor SMTP próprio**

## 📧 PASSO 2: Configurar Gmail

### 2.1 Habilitar App Password

Gmail requer uma "App Password" para aplicativos de terceiros:

1. Acesse [https://myaccount.google.com/](https://myaccount.google.com/)
2. Vá em **"Segurança"**
3. Ative **"Verificação em duas etapas"** (obrigatório)
4. Role até **"Senhas de app"**
5. Selecione **"App"** → **"Outro (nome personalizado)"**
6. Digite: `CAST Notifications`
7. Clique em **"Gerar"**
8. **Copie a senha gerada** (16 caracteres, sem espaços)
   - Exemplo: `abcd efgh ijkl mnop` → use `abcdefghijklmnop`

### 2.2 Configurações SMTP do Gmail

- **Host:** `smtp.gmail.com`
- **Porta:** `587` (TLS) ou `465` (SSL)
- **Username:** Seu email completo
- **Password:** App Password gerada
- **TLS:** Habilitado

## 📧 PASSO 3: Configurar Outlook/Hotmail

### 3.1 Obter Senha de App

1. Acesse [https://account.microsoft.com/security](https://account.microsoft.com/security)
2. Vá em **"Segurança"** → **"Senhas de app"**
3. Clique em **"Criar uma nova senha de app"**
4. Digite um nome: `CAST Notifications`
5. **Copie a senha gerada**

### 3.2 Configurações SMTP do Outlook

- **Host:** `smtp-mail.outlook.com`
- **Porta:** `587` (TLS)
- **Username:** Seu email completo
- **Password:** Senha de app gerada
- **TLS:** Habilitado

## 📧 PASSO 4: Configurar SendGrid

### 4.1 Criar Conta SendGrid

1. Acesse [https://sendgrid.com/](https://sendgrid.com/)
2. Clique em **"Start for free"**
3. Crie uma conta (gratuita até 100 emails/dia)

### 4.2 Criar API Key

1. No painel do SendGrid, vá em **"Settings"** → **"API Keys"**
2. Clique em **"Create API Key"**
3. Dê um nome: `CAST Notifications`
4. Selecione **"Full Access"** ou **"Mail Send"**
5. **Copie a API Key gerada**
   - ⚠️ Você só verá esta chave uma vez!

### 4.3 Configurações SMTP do SendGrid

- **Host:** `smtp.sendgrid.net`
- **Porta:** `587` (TLS)
- **Username:** `apikey` (literalmente)
- **Password:** Sua API Key
- **TLS:** Habilitado

## 📧 PASSO 5: Configurar Resend

### 5.1 Criar Conta Resend

1. Acesse [https://resend.com/](https://resend.com/)
2. Clique em **"Get Started"**
3. Crie uma conta (gratuita até 3.000 emails/mês)

### 5.2 Obter API Key

1. No painel do Resend, vá em **"API Keys"**
2. Clique em **"Create API Key"**
3. Dê um nome: `CAST Notifications`
4. **Copie a API Key**

### 5.3 Configurações SMTP do Resend

- **Host:** `smtp.resend.com`
- **Porta:** `587` (TLS)
- **Username:** `resend` (literalmente)
- **Password:** Sua API Key
- **TLS:** Habilitado

## ⚙️ PASSO 6: Configurar no CAST

### 6.1 Opção A: Variáveis de Ambiente

**Windows (CMD):**
```cmd
set CAST_EMAIL_SMTP_HOST=smtp.gmail.com
set CAST_EMAIL_SMTP_PORT=587
set CAST_EMAIL_USERNAME=seu-email@gmail.com
set CAST_EMAIL_PASSWORD=abcdefghijklmnop
set CAST_EMAIL_FROM_EMAIL=seu-email@gmail.com
set CAST_EMAIL_FROM_NAME=CAST Notifications
set CAST_EMAIL_USE_TLS=true
```

**Windows (PowerShell):**
```powershell
$env:CAST_EMAIL_SMTP_HOST="smtp.gmail.com"
$env:CAST_EMAIL_SMTP_PORT="587"
$env:CAST_EMAIL_USERNAME="seu-email@gmail.com"
$env:CAST_EMAIL_PASSWORD="abcdefghijklmnop"
$env:CAST_EMAIL_FROM_EMAIL="seu-email@gmail.com"
$env:CAST_EMAIL_FROM_NAME="CAST Notifications"
$env:CAST_EMAIL_USE_TLS="true"
```

**Linux/Mac:**
```bash
export CAST_EMAIL_SMTP_HOST="smtp.gmail.com"
export CAST_EMAIL_SMTP_PORT="587"
export CAST_EMAIL_USERNAME="seu-email@gmail.com"
export CAST_EMAIL_PASSWORD="abcdefghijklmnop"
export CAST_EMAIL_FROM_EMAIL="seu-email@gmail.com"
export CAST_EMAIL_FROM_NAME="CAST Notifications"
export CAST_EMAIL_USE_TLS="true"
```

### 6.2 Opção B: Arquivo YAML (`cast.yaml`)

**Gmail:**
```yaml
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "seu-email@gmail.com"
  password: "abcdefghijklmnop"  # App Password
  from_email: "seu-email@gmail.com"
  from_name: "CAST Notifications"
  use_tls: true
  use_ssl: false
  timeout: 30
```

**SendGrid:**
```yaml
email:
  smtp_host: "smtp.sendgrid.net"
  smtp_port: 587
  username: "apikey"
  password: "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  from_email: "noreply@empresa.com"
  from_name: "Sistema CAST"
  use_tls: true
  timeout: 30
```

**Resend:**
```yaml
email:
  smtp_host: "smtp.resend.com"
  smtp_port: 587
  username: "resend"
  password: "re_xxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  from_email: "noreply@empresa.com"
  from_name: "CAST Notifications"
  use_tls: true
  timeout: 30
```

**Outlook:**
```yaml
email:
  smtp_host: "smtp-mail.outlook.com"
  smtp_port: 587
  username: "seu-email@outlook.com"
  password: "senha-de-app"
  from_email: "seu-email@outlook.com"
  from_name: "CAST Notifications"
  use_tls: true
  timeout: 30
```

### 6.3 Opção C: Arquivo JSON (`cast.json`)

```json
{
  "email": {
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "username": "seu-email@gmail.com",
    "password": "abcdefghijklmnop",
    "from_email": "seu-email@gmail.com",
    "from_name": "CAST Notifications",
    "use_tls": true,
    "use_ssl": false,
    "timeout": 30
  }
}
```

### 6.4 Opção D: Arquivo Properties (`cast.properties`)

```properties
email.smtp_host=smtp.gmail.com
email.smtp_port=587
email.username=seu-email@gmail.com
email.password=abcdefghijklmnop
email.from_email=seu-email@gmail.com
email.from_name=CAST Notifications
email.use_tls=true
email.use_ssl=false
email.timeout=30
```

## ✅ PASSO 7: Testar a Configuração

### 7.1 Enviar email de teste

```bash
cast send mail destinatario@exemplo.com "Assunto do Email" "Corpo da mensagem"
```

### 7.2 Enviar email com anexo

```bash
cast send mail destinatario@exemplo.com "Relatório" "Segue o relatório em anexo" --attachment caminho/para/arquivo.pdf
```

### 7.3 Aguardar resposta (IMAP Monitor)

```bash
# Aguarda resposta usando tempo do config ou 30min (padrão)
cast send mail destinatario@exemplo.com "Pergunta importante" \
  --subject "Sua opinião" \
  --wfr

# Aguarda 5 minutos específicos
cast send mail destinatario@exemplo.com "Pergunta importante" \
  --subject "Sua opinião" \
  --wfr --wfr-minutes 5

# Apenas --wfr-minutes (ativa automaticamente)
cast send mail destinatario@exemplo.com "Confirmação" \
  --subject "Confirme recebimento" \
  --wfr-minutes 2 --verbose

# Forma longa --wait-for-response
cast send mail destinatario@exemplo.com "Solicitação" \
  --subject "Por favor, responda" \
  --wait-for-response --wfr-minutes 10
```

### 7.4 Verificar se funcionou

Verifique a caixa de entrada (e spam) do destinatário. Se o email chegou, a configuração está correta! ✅

## 🔧 CONFIGURAÇÕES AVANÇADAS

### Usar SSL ao invés de TLS

Alguns servidores usam SSL na porta 465:

```yaml
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 465
  use_tls: false
  use_ssl: true
```

### Timeout Customizado

```yaml
email:
  timeout: 60  # 60 segundos
```

### From Email Diferente do Username

```yaml
email:
  username: "sistema@empresa.com"
  from_email: "notificacoes@empresa.com"
  from_name: "Sistema de Notificações"
```

## 🎯 CONFIGURANDO ALIASES

```yaml
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "sistema@empresa.com"
  password: "senha"

aliases:
  admin:
    provider: "mail"
    target: "admin@empresa.com"
    name: "Administrador"

  dev-team:
    provider: "mail"
    target: "dev@empresa.com"
    name: "Time de Desenvolvimento"
```

Uso:

```bash
cast send mail admin "Alerta" "Mensagem importante"
cast send mail dev-team "Deploy" "Deploy realizado com sucesso"
```

## ⚠️ SEGURANÇA

- **NUNCA** use sua senha normal do Gmail (use App Password)
- **NUNCA** commite senhas em repositórios Git
- Use variáveis de ambiente em produção
- Revogue App Passwords comprometidas
- Para Gmail, sempre use App Password, não senha normal

## 📚 REFERÊNCIAS

- [Gmail - App Passwords](https://support.google.com/accounts/answer/185833)
- [SendGrid - SMTP Settings](https://docs.sendgrid.com/for-developers/sending-email/getting-started-smtp)
- [Resend - SMTP](https://resend.com/docs/send-with-smtp)
- [Outlook - App Passwords](https://support.microsoft.com/en-us/account-billing/using-app-passwords-with-apps-that-don-t-support-two-step-verification-5896ed9b-4263-681f-128a-12b3910f1b2f)
- [Especificação CAST - Email](specifications/04_GATEWAY_CONFIG_SPEC.md#4-email-smtp)

## 🆘 SOLUÇÃO DE PROBLEMAS

### Erro: "Authentication failed"
- Verifique username e password
- Para Gmail, certifique-se de usar App Password, não senha normal
- Verifique se a verificação em duas etapas está ativada (Gmail)

### Erro: "Connection refused"
- Verifique se o host e porta estão corretos
- Verifique firewall/proxy
- Alguns provedores bloqueiam conexões de IPs não autorizados

### Erro: "TLS/SSL handshake failed"
- Verifique se `use_tls` ou `use_ssl` está correto
- Tente trocar a porta (587 para TLS, 465 para SSL)
- Verifique se o servidor suporta TLS/SSL

### Email vai para spam
- Configure SPF, DKIM e DMARC no domínio
- Use um provedor confiável (SendGrid, Resend)
- Evite palavras suspeitas no assunto/corpo

### Erro: "Timeout"
- Aumente o valor de `timeout`
- Verifique sua conexão com a internet
- Alguns servidores são mais lentos

### Mensagem não chega
- Verifique logs do CAST (se habilitado)
- Confirme que o email do destinatário está correto
- Verifique a pasta de spam
- Teste com outro provedor SMTP

### Erro ao aguardar resposta (--wfr)
- Verifique se a configuração IMAP está completa
- Confirme que `imap_host`, `imap_port`, `imap_username` e `imap_password` estão corretos
- Para Gmail, use a mesma App Password do SMTP
- Verifique se a pasta IMAP está correta (geralmente "INBOX")
- Use `--verbose` para ver logs detalhados da conexão IMAP
- Certifique-se de que o servidor IMAP está acessível (porta 993 para SSL, 143 para TLS)

### Resposta não é detectada
- O CAST busca por `In-Reply-To` e `References` headers primeiro
- Se não encontrar, usa fallback por Subject após 3 ciclos
- Certifique-se de que o cliente de email do destinatário está configurando corretamente os headers de resposta
- Use `--verbose` para ver qual método de busca está sendo usado
- Verifique se o Message-ID do email enviado está sendo referenciado na resposta
