# RESULTADOS DA FASE 07 - IMAP MONITOR (--wait-for-response)

**Data:** 2025-01-13
**Versão:** 0.7.0
**Status:** ✅ **CONCLUÍDA**

---

## 📋 RESUMO EXECUTIVO

A Fase 07 implementou o recurso **IMAP Monitor** para o provider Email, permitindo que o CAST aguarde e capture respostas por email após o envio. Esta funcionalidade é essencial para permitir controle remoto do Cursor IDE via email.

### Objetivos Alcançados

- ✅ Geração de Message-ID único para cada email enviado
- ✅ Conexão IMAP com SSL/TLS para monitorar caixa de entrada
- ✅ Busca inteligente por resposta (In-Reply-To, References, Subject)
- ✅ Validação robusta para evitar capturar mensagens antigas
- ✅ Polling configurável entre ciclos de busca
- ✅ Exibição do corpo completo da resposta
- ✅ Exit codes específicos para diferentes cenários
- ✅ Logs detalhados para debugging

---

## 🎯 FUNCIONALIDADES IMPLEMENTADAS

### 1. Geração de Message-ID

**Arquivo:** `internal/providers/email.go`

- Função `generateMessageID()` cria Message-ID único no formato:
  - `<cast-{timestamp}-{random}@{domain}>`
- Função `extractDomain()` extrai domínio do email para usar no Message-ID
- Message-ID incluído nos headers do email enviado
- Interface `EmailProviderExtended` com método `GetLastMessageID()`

**Exemplo:**
```go
Message-ID: <cast-1765647182907412000-b275be1448122d58@gmail.com>
```

### 2. Monitoramento IMAP

**Arquivo:** `internal/providers/email_imap.go`

#### 2.1 Função Principal: `WaitForEmailResponse()`

- Orquestra todo o processo de aguardar resposta
- Calcula deadline baseado em `waitMinutes`
- Implementa polling com intervalo configurável
- Gerencia conexões IMAP e tratamento de erros
- Retorna exit codes específicos (0, 2, 3, 4)

#### 2.2 Conexão IMAP: `connectIMAP()`

- Suporte a SSL (porta 993) e TLS (porta 143)
- Autenticação com username/password
- Seleção de pasta (default: INBOX)
- Timeout configurável
- Logs detalhados em modo verbose

#### 2.3 Busca por Resposta: `searchEmailResponse()`

**Ordem de busca:**
1. **Primária:** `SEARCH HEADER In-Reply-To "<messageID>"`
2. **Secundária:** `SEARCH HEADER References "<messageID>"`
3. **Fallback:** `SEARCH HEADER Subject "Re: <subject>"` (apenas após 3 ciclos)

**Validação:**
- Fallback por Subject só é usado após 3 ciclos (dá tempo para resposta)
- Quando usado, valida que `InReplyTo` corresponde ao Message-ID correto
- Evita capturar mensagens antigas com mesmo subject

#### 2.4 Fetch e Parse: `fetchLatestMessage()` e `parseEmailMessage()`

- Busca mensagem completa usando `BODY[]`
- Extrai From, Date, Subject e Body
- Suporte a mensagens multipart (HTML + texto)
- Tratamento robusto de diferentes formatos de email

### 3. Configuração IMAP

**Arquivo:** `internal/config/config.go`

**Novos campos em `EmailConfig`:**
```go
IMAPHost        string  // Host do servidor IMAP
IMAPPort        int     // Porta (993 para SSL, 143 para TLS)
IMAPUsername    string  // Username (geralmente igual ao SMTP)
IMAPPassword    string  // Password (geralmente igual ao SMTP)
IMAPUseTLS      bool    // Usar TLS (porta 143)
IMAPUseSSL      bool    // Usar SSL (porta 993)
IMAPFolder      string  // Pasta a monitorar (default: INBOX)
IMAPTimeout     int     // Timeout em segundos (default: 60)
IMAPPollInterval int    // Intervalo entre ciclos (default: 15s, min: 5s, max: 60s)

WaitForResponseDefault  int  // Minutos padrão (0 = desabilitado)
WaitForResponseMax      int  // Teto de segurança (default: 120)
WaitForResponseMaxLines int  // Limite de linhas do corpo (0 = completo)
```

**Suporte a ENV:**
- `CAST_EMAIL_IMAP_HOST`
- `CAST_EMAIL_IMAP_PORT`
- `CAST_EMAIL_IMAP_USERNAME`
- `CAST_EMAIL_IMAP_PASSWORD`
- `CAST_EMAIL_IMAP_USE_TLS`
- `CAST_EMAIL_IMAP_USE_SSL`
- `CAST_EMAIL_IMAP_FOLDER`
- `CAST_EMAIL_IMAP_TIMEOUT`
- `CAST_EMAIL_IMAP_POLL_INTERVAL_SECONDS`
- `CAST_EMAIL_WAIT_FOR_RESPONSE_DEFAULT_MINUTES`
- `CAST_EMAIL_WAIT_FOR_RESPONSE_MAX_MINUTES`
- `CAST_EMAIL_WAIT_FOR_RESPONSE_MAX_LINES`

### 4. Integração CLI

**Arquivo:** `cmd/cast/send.go`

**Flags adicionadas:**
- `--wfr` ou `--wait-for-response` (bool): Ativa espera por resposta (usa tempo do config ou 30min)
- `--wfr-minutes N` (int): Especifica tempo de espera em minutos (sobrescreve config)

**Comportamento:**
- Só funciona com provider `email` (ignora com aviso para outros)
- `--wfr` e `--wait-for-response` têm o mesmo comportamento (flags bool)
- Se `--wfr-minutes` for usado sozinho, ativa automaticamente a espera
- Calcula `waitMinutes`: `--wfr-minutes` > config > padrão 30min
- Após envio bem-sucedido, chama `WaitForEmailResponse()`
- Trata exit codes específicos:
  - `0`: Resposta recebida
  - `2`: Erro de configuração
  - `3`: Timeout ou erro de rede
  - `4`: Erro de autenticação IMAP

**Exemplos de uso:**
```bash
# Usar tempo do config ou 30min (padrão)
cast send mail destinatario@exemplo.com "Pergunta" \
  --subject "Confirmação" \
  --wfr --verbose

# Especificar tempo customizado
cast send mail destinatario@exemplo.com "Pergunta" \
  --subject "Confirmação" \
  --wfr --wfr-minutes 5 --verbose

# Apenas --wfr-minutes (ativa automaticamente)
cast send mail destinatario@exemplo.com "Pergunta" \
  --subject "Confirmação" \
  --wfr-minutes 10 --verbose
```

### 5. Exit Codes Específicos

Conforme especificação:
- **0**: Resposta recebida com sucesso
- **2**: Erro de configuração (IMAP não configurado, flags inválidas)
- **3**: Timeout lógico (destinatário não respondeu) ou erro de rede IMAP
- **4**: Erro de autenticação IMAP

---

## 📁 ARQUIVOS CRIADOS/MODIFICADOS

### Arquivos Criados

1. **`internal/providers/email_imap.go`** (673 linhas)
   - `WaitForEmailResponse()` - Função principal
   - `connectIMAP()` - Conexão IMAP
   - `searchEmailResponse()` - Busca por resposta
   - `fetchAndValidateMessage()` - Validação de InReplyTo
   - `fetchLatestMessage()` - Fetch de mensagem completa
   - `parseEmailMessage()` - Parse de email
   - Structs: `EmailResponse`, erros específicos

2. **`internal/providers/email_imap_test.go`** (120 linhas)
   - `TestGenerateMessageID()` - Valida formato e unicidade
   - `TestExtractDomain()` - Testa extração de domínio
   - `TestFormatDuration()` - Testa formatação de duração
   - Testes básicos de parsing

### Arquivos Modificados

1. **`internal/providers/email.go`**
   - Adicionado `generateMessageID()` e `extractDomain()`
   - `SendEmail()` agora gera e retorna Message-ID
   - `buildMultipartMessage()` inclui Message-ID nos headers
   - Adicionado campo `lastMessageID` no struct
   - Interface `EmailProviderExtended` com `GetLastMessageID()`

2. **`internal/config/config.go`**
   - Expandido `EmailConfig` com campos IMAP
   - Adicionado `viper.BindEnv()` para campos IMAP
   - Atualizado `applyEnvOverrides()` para IMAP
   - Atualizado `applyDefaults()` para IMAP
   - Atualizado `Validate()` para validar IMAP quando necessário

3. **`cmd/cast/send.go`**
   - Adicionadas flags `--wfr` e `--wait-for-response`
   - Lógica para calcular `waitMinutes`
   - Chamada a `WaitForEmailResponse()` após envio
   - Tratamento de exit codes específicos
   - Aviso se `--wfr` usado com provider não-email
   - Help atualizado

4. **`cmd/cast/help.go`**
   - Atualizado `ShowSendHelp()` com documentação de `--wfr`

---

## 🧪 TESTES

### Testes Unitários

**Arquivo:** `internal/providers/email_imap_test.go`

- ✅ `TestGenerateMessageID()` - 3 casos de teste
  - Valida formato do Message-ID
  - Verifica unicidade (100 iterações)
  - Valida domínio extraído

- ✅ `TestExtractDomain()` - 4 casos de teste
  - Email simples
  - Email com subdomínio
  - Email com nome completo
  - Email inválido

- ✅ `TestFormatDuration()` - 3 casos de teste
  - Segundos
  - Minutos
  - Minutos e segundos

**Total:** 6 testes unitários novos

### Testes Manuais

**Cenário 1: Resposta rápida (ciclo 2)**
```bash
cast send mail destinatario@exemplo.com "Pergunta 5" \
  "Você pode confirmar novamente?" \
  --wfr --wfr-minutes 2 --verbose
```

**Resultado:**
- ✅ Email enviado com Message-ID único
- ✅ Resposta encontrada no ciclo 2 via `In-Reply-To`
- ✅ Corpo completo da resposta exibido
- ✅ Exit code 0 (sucesso)

**Cenário 2: Fallback por Subject (após 3 ciclos)**
- ✅ Fallback desabilitado nos primeiros 3 ciclos
- ✅ Após 3 ciclos, tenta fallback por Subject
- ✅ Valida InReplyTo antes de aceitar resposta
- ✅ Rejeita mensagens antigas com mesmo subject

**Cenário 3: Timeout**
- ✅ Aguarda tempo configurado
- ✅ Retorna exit code 3 (timeout)
- ✅ Mensagem clara de timeout

---

## 📊 MÉTRICAS

### Código

- **Linhas adicionadas:** ~850
- **Arquivos criados:** 2
- **Arquivos modificados:** 4
- **Funções novas:** 6 principais + 3 auxiliares
- **Testes unitários:** 6 novos

### Funcionalidades

- **Providers afetados:** 1 (Email)
- **Flags novas:** 2 (`--wfr`, `--wait-for-response`)
- **Campos de configuração novos:** 11
- **Exit codes específicos:** 4 (0, 2, 3, 4)

### Bibliotecas Adicionadas

- `github.com/emersion/go-imap` - Cliente IMAP
- `github.com/emersion/go-imap/client` - Cliente IMAP
- `github.com/emersion/go-message/mail` - Parser de email

---

## 🐛 PROBLEMAS RESOLVIDOS

### 1. Corpo da Mensagem Não Sendo Capturado

**Problema:** O corpo da resposta não estava sendo exibido.

**Solução:**
- Modificado `fetchLatestMessage()` para buscar `BODY[]` explicitamente
- Melhorado `parseEmailMessage()` para tentar múltiplas seções do corpo
- Adicionado fallback para qualquer seção disponível se a específica falhar

### 2. Fallback por Subject Pegando Mensagens Antigas

**Problema:** O fallback por Subject estava capturando respostas de emails anteriores.

**Solução:**
- Fallback só é usado após 3 ciclos (dá tempo para resposta)
- Quando usado, valida que `InReplyTo` corresponde ao Message-ID correto
- Se não corresponder, tenta mensagens mais antigas na lista
- Se nenhuma corresponder, retorna `false` (continua buscando)

### 3. Intervalo de Polling Não Configurável

**Problema:** Intervalo de polling era fixo (15-30s baseado em timeout).

**Solução:**
- Adicionado campo `imap_poll_interval_seconds` na configuração
- Configurável via YAML ou ENV
- Limites: mínimo 5s, máximo 60s, default 15s

### 4. Message-ID com/sem Angle Brackets

**Problema:** Alguns servidores IMAP retornam Message-ID com `< >`, outros sem.

**Solução:**
- Função `searchEmailResponse()` tenta ambas as variações
- `messageIDClean` remove `< >` para comparação
- Validação flexível em `fetchAndValidateMessage()`

---

## ✅ CRITÉRIOS DE ACEITE

Conforme especificação `10_FASE_07_IMAP_MONITOR_SPECS.md`:

- [x] **CA-01:** Geração de Message-ID único para cada email
- [x] **CA-02:** Conexão IMAP com SSL/TLS configurável
- [x] **CA-03:** Busca por `In-Reply-To` e `References` headers
- [x] **CA-04:** Fallback por Subject (após alguns ciclos, com validação)
- [x] **CA-05:** Exibição do corpo completo da resposta
- [x] **CA-06:** Exit codes específicos (0, 2, 3, 4)
- [x] **CA-07:** Logs detalhados em modo verbose
- [x] **CA-08:** Polling configurável entre ciclos
- [x] **CA-09:** Validação robusta para evitar mensagens antigas
- [x] **CA-10:** Integração completa no comando `send` com flag `--wfr`

**Status:** ✅ **TODOS OS CRITÉRIOS ATENDIDOS**

---

## 📝 EXEMPLOS DE USO

### Exemplo 1: Aguardar Resposta Básica

```bash
cast send mail admin@empresa.com "Confirmação" \
  "Você pode confirmar o recebimento?" \
  --wfr --wfr-minutes 5
```

**Saída esperada:**
```
✓ Mensagem enviada com sucesso via email
⏳ Aguardando resposta por até 5 minutos (IMAP: imap.gmail.com:993, pasta INBOX)...
✓ Resposta recebida em 1m23s

=== EMAIL RESPONSE ===
From: admin@empresa.com
Date: 2025-01-13 14:33:37
Subject: Re: Confirmação

Confirmado! Recebido com sucesso.
```

### Exemplo 2: Com Modo Verbose

```bash
cast send mail destinatario@exemplo.com "Pergunta" \
  "Você pode confirmar?" \
  --wfr --wfr-minutes 2 --verbose
```

**Saída esperada (com logs detalhados):**
```
[DEBUG] Message-ID sendo buscado: <cast-1765647182907412000-b275be1448122d58@gmail.com>
[DEBUG] Subject original: Pergunta
[DEBUG] Intervalo de polling: 15s (entre cada ciclo de busca)
[DEBUG] Ciclo 1: verificando IMAP...
[DEBUG] Conectando ao IMAP imap.gmail.com:993 (SSL)
[DEBUG] Autenticado com sucesso
[DEBUG] Pasta selecionada: INBOX (49576 mensagens)
[DEBUG] SEARCH HEADER In-Reply-To não encontrou mensagens
[DEBUG] Nenhuma mensagem correspondente encontrada, tentando References...
[DEBUG] SEARCH HEADER References não encontrou mensagens
[DEBUG] Fallback por Subject desabilitado (aguardando mais ciclos para dar tempo de resposta)...
[DEBUG] Ciclo 2: verificando IMAP...
[DEBUG] SEARCH HEADER In-Reply-To encontrou 1 mensagem(ns): [49577]
[DEBUG] Mensagem encontrada, UID=49577
✓ Resposta recebida em 45s
```

### Exemplo 3: Timeout

```bash
cast send mail destinatario@exemplo.com "Pergunta" \
  "Você pode responder?" \
  --wfr --wfr-minutes 1
```

**Saída esperada (se não houver resposta):**
```
✓ Mensagem enviada com sucesso via email
⏳ Aguardando resposta por até 1 minuto (IMAP: imap.gmail.com:993, pasta INBOX)...
⏰ Tempo de espera esgotado (1 minuto).
✗ O destinatário não respondeu à mensagem.
```

**Exit code:** `3`

---

## 🔍 LOGS E DEBUGGING

### Modo Verbose

O modo `--verbose` exibe:
- Message-ID sendo buscado
- Subject original
- Intervalo de polling configurado
- Cada ciclo de busca (número e timestamp)
- Conexão IMAP (host, porta, SSL/TLS)
- Autenticação (sucesso/falha)
- Pasta selecionada e quantidade de mensagens
- Resultado de cada busca (In-Reply-To, References, Subject)
- UID da mensagem encontrada
- Seções do corpo disponíveis
- Validação de InReplyTo (quando usar fallback)

### Exemplo de Log Detalhado

```
[DEBUG] Message-ID sendo buscado: <cast-1765647182907412000-b275be1448122d58@gmail.com>
[DEBUG] Subject original: Notificação CAST
[DEBUG] Intervalo de polling: 15s (entre cada ciclo de busca)
[DEBUG] Ciclo 1: verificando IMAP...
[DEBUG] Conectando ao IMAP imap.gmail.com:993 (SSL)
[DEBUG] Autenticado com sucesso
[DEBUG] Pasta selecionada: INBOX (49576 mensagens)
[DEBUG] SEARCH HEADER In-Reply-To não encontrou mensagens (Message-ID: <cast-1765647182907412000-b275be1448122d58@gmail.com>)
[DEBUG] Nenhuma mensagem correspondente encontrada, tentando References...
[DEBUG] SEARCH HEADER References não encontrou mensagens
[DEBUG] Fallback por Subject desabilitado (aguardando mais ciclos para dar tempo de resposta)...
[DEBUG] Ciclo 1: 0 respostas encontradas, aguardando 15s antes da próxima verificação...
[DEBUG] Ciclo 2: verificando IMAP...
[DEBUG] Conectando ao IMAP imap.gmail.com:993 (SSL)
[DEBUG] Autenticado com sucesso
[DEBUG] Pasta selecionada: INBOX (49577 mensagens)
[DEBUG] SEARCH HEADER In-Reply-To encontrou 1 mensagem(ns): [49577]
[DEBUG] Mensagem encontrada, UID=49577
[DEBUG] Mensagem recebida, Envelope: &{Date:2025-12-13 14:33:37 ... InReplyTo:<cast-1765647182907412000-b275be1448122d58@gmail.com> ...}
[DEBUG] Body sections disponíveis: 1
[DEBUG] Seção disponível no map: &{... value:BODY[]}
[DEBUG] BodyStructure tipo: multipart
[DEBUG] Seção encontrada! Usando esta seção: &{... value:BODY[]}
✓ Resposta recebida em 45s
```

---

## 🎓 LIÇÕES APRENDIDAS

### 1. IMAP é Complexo

- Diferentes servidores IMAP têm comportamentos diferentes
- Message-ID pode vir com ou sem angle brackets
- Body sections podem estar em formatos diferentes
- É necessário tentar múltiplas abordagens para robustez

### 2. Validação é Essencial

- Fallback por Subject sem validação pode pegar mensagens antigas
- Validação de InReplyTo é crítica para garantir resposta correta
- Dar tempo antes de usar fallback evita falsos positivos

### 3. Polling Configurável

- Intervalo fixo pode ser muito rápido ou muito lento
- Permitir configuração dá flexibilidade ao usuário
- Limites (min/max) previnem configurações problemáticas

### 4. Logs Detalhados

- Modo verbose é essencial para debugging IMAP
- Mostrar cada etapa ajuda a identificar problemas
- Logs devem ser informativos mas não expor senhas

---

## 🚀 PRÓXIMOS PASSOS

### Melhorias Futuras (Não Implementadas)

1. **Suporte a Múltiplas Pastas**
   - Permitir monitorar múltiplas pastas IMAP
   - Configurar pasta por alias

2. **Filtros Avançados**
   - Filtrar por remetente específico
   - Filtrar por palavras-chave no corpo

3. **Notificações**
   - Notificar quando resposta for recebida (além de exibir)
   - Integração com outros providers para notificação

4. **Histórico**
   - Salvar histórico de respostas recebidas
   - Permitir consultar respostas anteriores

5. **Retry Automático**
   - Tentar novamente se conexão IMAP falhar
   - Backoff exponencial para reconexão

---

## 📚 REFERÊNCIAS

- **Especificação:** `specifications/10_FASE_07_IMAP_MONITOR_SPECS.md`
- **Protocolo:** `specifications/06_PHASE_IMPLEMENTATION_PROTOCOL.md`
- **Biblioteca IMAP:** [github.com/emersion/go-imap](https://github.com/emersion/go-imap)
- **Biblioteca Message:** [github.com/emersion/go-message](https://github.com/emersion/go-message)

---

## ✅ CONCLUSÃO

A Fase 07 foi **concluída com sucesso**, implementando todas as funcionalidades especificadas:

- ✅ Geração de Message-ID único
- ✅ Monitoramento IMAP completo
- ✅ Busca inteligente por resposta
- ✅ Validação robusta
- ✅ Polling configurável
- ✅ Exit codes específicos
- ✅ Logs detalhados
- ✅ Integração completa no CLI

O CAST agora pode **aguardar e capturar respostas por email**, permitindo controle remoto via email de forma confiável e robusta.

**Status Final:** ✅ **FASE 07 CONCLUÍDA**

---

*Documento gerado em: 2025-01-13*
*Versão do CAST: 0.7.0*
