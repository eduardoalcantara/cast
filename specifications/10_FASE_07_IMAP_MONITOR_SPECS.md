<img src="https://r2cdn.perplexity.ai/pplx-full-logo-primary-dark%402x.png" style="height:64px;margin-right:32px"/>

# Sim faça uma especificação bem completa para funcionar desde a primeira vez.

O recurso `--wait-for-response` deve ser especificado para IMAP de forma completa, garantindo rastreio preciso de **respostas a um e‑mail específico**, usando `Message-ID` como chave primária e `Subject` apenas como fallback.[^1][^2]

***

## 1. Objetivo do recurso

Permitir que o comando:

```bash
cast send mail destinatario@exemplo.com "Assunto" "Mensagem" --wfr 15
# ou
cast send mail destinatario@exemplo.com "Assunto" "Mensagem" --wait-for-response 15
```

1. Envie um e‑mail via SMTP normalmente.
2. Gere e registre um `Message-ID` único para essa mensagem.
3. Abra uma sessão IMAP na conta configurada e **aguarde até 15 minutos** por uma resposta a esse e‑mail, identificada por cabeçalhos de reply.
4. Se encontrar uma resposta:
    - Exibir metadados (From, Date, Subject).
    - Exibir **corpo completo** do e‑mail de resposta (por padrão).
    - Retornar exit code `0`.
5. Se não encontrar resposta até o limite:
    - Exibir mensagem clara de que o destinatário não respondeu.
    - Retornar exit code `3` (timeout lógico de resposta).

Tudo isso sem tratar outros e‑mails da caixa como “resposta”.[^1]

***

## 2. Parâmetros CLI

### 2.1 Flags

- Curta: `--wfr <minutos>`
- Longa: `--wait-for-response <minutos>`

Regras:

- `<minutos>` é inteiro ≥ 0.
- Se a flag for omitida:
    - Usa `email.wait_for_response_default_minutes` do arquivo de config (se > 0).
- Se `<minutos> = 0` explicitamente:
    - Desabilita a espera para este comando, mesmo que haja default.
- Se `<minutos>` > `email.wait_for_response_max_minutes`:
    - Erro de validação: mensagem clara e exit code `2` (erro de configuração).


### 2.2 Escopo por provider

- Inicialmente **apenas para provider `mail` (email)**.
- Se `cast send` for chamado com outro provider (`tg`, `zap`, `googlechat`) e `--wfr` presente:
    - Ignorar com aviso amarelo:
`⚠ Parâmetro --wait-for-response suportado apenas para provider 'mail'.`

***

## 3. Configuração IMAP em `EmailConfig`

### 3.1 Estrutura de configuração

Em `internal/config/config.go`:[^3]

```go
type EmailConfig struct {
    SMTPHost  string `mapstructure:"smtp_host" yaml:"smtp_host" json:"smtp_host"`
    SMTPPort  int    `mapstructure:"smtp_port" yaml:"smtp_port" json:"smtp_port"`
    Username  string `mapstructure:"username" yaml:"username" json:"username"`
    Password  string `mapstructure:"password" yaml:"password" json:"password"`
    FromEmail string `mapstructure:"from_email" yaml:"from_email" json:"from_email"`
    FromName  string `mapstructure:"from_name" yaml:"from_name" json:"from_name"`
    UseTLS    bool   `mapstructure:"use_tls" yaml:"use_tls" json:"use_tls"`
    UseSSL    bool   `mapstructure:"use_ssl" yaml:"use_ssl" json:"use_ssl"`
    Timeout   int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`

    // IMAP: usado apenas se wait-for-response estiver ativo
    IMAPHost     string `mapstructure:"imap_host" yaml:"imap_host" json:"imap_host"`
    IMAPPort     int    `mapstructure:"imap_port" yaml:"imap_port" json:"imap_port"`
    IMAPUsername string `mapstructure:"imap_username" yaml:"imap_username" json:"imap_username"`
    IMAPPassword string `mapstructure:"imap_password" yaml:"imap_password" json:"imap_password"`
    IMAPUseTLS   bool   `mapstructure:"imap_use_tls" yaml:"imap_use_tls" json:"imap_use_tls"`
    IMAPUseSSL   bool   `mapstructure:"imap_use_ssl" yaml:"imap_use_ssl" json:"imap_use_ssl"`
    IMAPFolder   string `mapstructure:"imap_folder" yaml:"imap_folder" json:"imap_folder"`
    IMAPTimeout  int    `mapstructure:"imap_timeout" yaml:"imap_timeout" json:"imap_timeout"`

    // Espera por resposta
    WaitForResponseDefault  int `mapstructure:"wait_for_response_default_minutes" yaml:"wait_for_response_default_minutes" json:"wait_for_response_default_minutes"`
    WaitForResponseMax      int `mapstructure:"wait_for_response_max_minutes" yaml:"wait_for_response_max_minutes" json:"wait_for_response_max_minutes"`
    WaitForResponseMaxLines int `mapstructure:"wait_for_response_max_lines" yaml:"wait_for_response_max_lines" json:"wait_for_response_max_lines"`
}
```


### 3.2 YAML típico

```yaml
email:
  smtp_host: "smtp.exemplo.com"
  smtp_port: 587
  username: "notificacoes@exemplo.com"
  password: "********"
  from_email: "notificacoes@exemplo.com"
  from_name: "CAST Notifications"
  use_tls: true
  use_ssl: false
  timeout: 30

  imap_host: "imap.exemplo.com"
  imap_port: 993
  imap_username: "notificacoes@exemplo.com"
  imap_password: "********"
  imap_use_tls: false
  imap_use_ssl: true
  imap_folder: "INBOX"
  imap_timeout: 60

  wait_for_response_default_minutes: 0    # 0 = desabilitado por padrão
  wait_for_response_max_minutes: 120      # teto de segurança
  wait_for_response_max_lines: 0          # 0 = mostrar corpo completo
```


### 3.3 Defaults e validação

Em `applyDefaults`:[^3]

- Se `IMAPPort == 0`:
    - Se `IMAPUseSSL=true` → `IMAPPort=993`.
    - Senão, se `IMAPUseTLS=true` → `IMAPPort=143`.
    - Se ambos `IMAPUseSSL` e `IMAPUseTLS` forem `false` → `IMAPUseSSL=true`, `IMAPPort=993`.
- Se `IMAPFolder == ""` → `"INBOX"`.
- Se `IMAPTimeout == 0` → `60`.
- Se `WaitForResponseMax == 0` → `120`.
- Se `WaitForResponseMaxLines < 0` → normalizar para `0`.

Em `Validate`:[^3]

- Se `WaitForResponseDefault > 0`:
    - `IMAPHost`, `IMAPPort`, `IMAPUsername`, `IMAPPassword` obrigatórios.
- Em runtime, se o comando usar `--wfr` com `waitMinutes > 0`:
    - Validar que `WaitForResponseMax > 0` e `waitMinutes <= WaitForResponseMax`.
    - IMAP configurado; caso contrário, erro:
        - `✗ Para usar --wait-for-response é necessário configurar email.imap_* no cast.yaml.`

***

## 4. Geração e rastreamento por Message-ID

### 4.1 Geração de `Message-ID` no envio

- Em `internal/providers/email.go`, antes de montar `headers`:[^3]
    - Gerar um `messageID` único, por exemplo:

```go
func generateMessageID(domain string) string {
    if domain == "" {
        domain = "cast.local"
    }
    // timestamp + random: simples e suficiente
    return fmt.Sprintf("ast-%d-%s@%s>", time.Now().UnixNano(), randomHex(8), domain)
}
```

    - Extrair domínio de `FromEmail` (`notificacoes@exemplo.com` → `exemplo.com`).
    - Incluir header:

```go
messageID := generateMessageID(domain)
headers := []string{
    fmt.Sprintf("Message-ID: %s", messageID),
    // demais headers: From, To, Subject, etc.
}
```

- A função `Send` deve retornar esse `messageID` para quem chamou (possível mudança de assinatura só para o caminho `mail` do comando `send`, via wrapper interno).


### 4.2 Persistência temporária no comando `send`

- Em `cmd/cast/send.go`, após `provider.Send(...)` para `mail`:
    - Se `waitMinutes > 0`, chamar:

```go
err := waitForEmailResponse(cfg.Email, messageID, subject, waitMinutes, verboseFlag)
```


***

## 5. Algoritmo IMAP de espera por resposta

### 5.1 Critérios de match

Para cada ciclo de polling IMAP, o CAST deve procurar **somente** mensagens que:

1. Tenham `In-Reply-To` contendo exatamente o `Message-ID` enviado; ou[^4][^5]
2. Tenham `References` contendo o `Message-ID` enviado;[^4]
3. Como fallback (quando não for possível buscar por header ou o servidor não retornar nada), tenham:
    - `Subject` igual ao original; ou
    - `Subject` começando com `Re: <SubjectOriginal>` (case-insensitive, considerando `Re:`, `RE:`, etc.).[^1][^6]

A lógica deve ser:

- Primeiro tentar busca por `In-Reply-To`/`References`.
- Só se essa busca falhar sistematicamente (erro de servidor ou não suportado) entrar no fallback por `Subject`.


### 5.2 Busca IMAP

Usando IMAP puro, da perspectiva de protocolo:

- Selecionar a pasta:
    - `SELECT <IMAPFolder>` (ex.: `SELECT INBOX`).[^7]
- Buscar por `In-Reply-To` / `References`:
    - `SEARCH HEADER In-Reply-To "<message-id>"`
    - `SEARCH HEADER References "<message-id>"`.[^7][^2]
- Fallback de subject:
    - `SEARCH HEADER Subject "<SubjectOriginal>"`
    - `SEARCH HEADER Subject "Re: <SubjectOriginal>"` (considerar variações de capitalização).[^7][^6]

Na implementação Go, provavelmente via biblioteca IMAP (ex.: `emersion/go-imap`), isso se traduz em filtros equivalentes.

### 5.3 Janela de tempo e polling

Parâmetros:

- `waitMinutes` (derivado de CLI/config).
- `IMAPTimeout` (timeout de comandos).
- Intervalo de polling: `pollInterval`, derivado de `IMAPTimeout` ou constante (ex. 15–30 s).

Loop:

1. `deadline = now + waitMinutes`.
2. Enquanto `now < deadline`:
    - Conectar ao IMAP (`IMAPHost`, `IMAPPort`) com SSL/TLS (conforme flags).
    - Autenticar (`LOGIN IMAPUsername IMAPPassword`).
    - `SELECT IMAPFolder`.
    - Executar `SEARCH` por `In-Reply-To` / `References`:
        - Se IDs encontrados:
            - Buscar a **mensagem mais recente** da lista (`FETCH UID`).
            - Extrair headers e corpo.
            - Imprimir resultado (ver formato abaixo).
            - Retornar sucesso.
    - Se nenhum match:
        - Fallback por `Subject` se necessário.
    - Fechar sessão (`LOGOUT`).
    - Se ainda não passou do `deadline`, `sleep(pollInterval)`.
3. Se o `deadline` for alcançado sem encontrar resposta:
    - Imprimir mensagens de timeout e não resposta.
    - Retornar erro sentinela `ErrNoEmailResponse` mapeado para exit code `3`.

Erros de rede/autenticação IMAP devem ser tratados separadamente como falhas técnicas (exit code `3` ou `4`, conforme design atual de códigos).[^8]

***

## 6. Formato de saída da resposta

### 6.1 Comportamento padrão (sem limite)

- Exibir:

```text
✓ Mensagem enviada com sucesso via email
⏳ Aguardando resposta por até 15 minutos (IMAP: imap.exemplo.com:993, pasta INBOX)...

✓ Resposta recebida em 2m34s

=== EMAIL RESPONSE ===
From: usuario@cliente.com
Date: 2025-12-12 15:40:12
Subject: [translate:Re:] Assunto Original

rpo completo da mensagem de resposta>
=== END EMAIL RESPONSE ===
```

- `wait_for_response_max_lines = 0` → mostrar **corpo completo** (todas as linhas).


### 6.2 Limite opcional de linhas

- Se `WaitForResponseMaxLines > 0`, truncar o corpo após N linhas e acrescentar:

```text
[... corpo truncado em N linhas (ajuste email.wait_for_response_max_lines para 0 se quiser mostrar tudo) ...]
```

- Flag futura opcional em CLI, se desejar override:
    - `--wfr-max-lines N` com precedência sobre configuração.

***

## 7. Comportamento quando não há resposta

Se o tempo expirar sem resposta:

```text
✓ Mensagem enviada com sucesso via email
⏳ Aguardando resposta por até 15 minutos (IMAP: imap.exemplo.com:993, pasta INBOX)...

⏰ Tempo de espera esgotado (15 minutos).
✗ O destinatário não respondeu à mensagem.
```

- Exit code: `3` (timeout lógico de resposta).
- Scripts ou agentes de IA podem tratar esse código como “sem resposta humana” e seguir outro fluxo.

***

## 8. Modo verbose / debug

Com `--verbose`, o comando deve mostrar detalhes do IMAP:

- Conexão:

```text
[DEBUG] Conectando ao IMAP imap.exemplo.com:993 (SSL)
[DEBUG] Pasta selecionada: INBOX
```

- Busca:

```text
[DEBUG] SEARCH HEADER In-Reply-To "ast-...@exemplo.com>"
[DEBUG] Nenhuma mensagem correspondente encontrada, tentando References...
[DEBUG] SEARCH HEADER References "ast-...@exemplo.com>"
[DEBUG] Nenhuma mensagem correspondente, fallback por Subject...
[DEBUG] SEARCH HEADER Subject "Assunto Original"
```

- Ciclos:

```text
[DEBUG] Ciclo 1: 0 respostas encontradas, aguardando 30s antes da próxima verificação...
```

- No sucesso:

```text
[DEBUG] Mensagem encontrada, UID=1234
```


***

## 9. Tratamento de erros

- **Erro SMTP no envio** → comportamento atual, não entra no fluxo IMAP.
- **Erro de configuração IMAP** (host/credenciais ausentes):
    - Mensagem clara.
    - Exit code `2`.
- **Erro de autenticação IMAP**:
    - Mensagem clara (`falha na autenticação IMAP`).
    - Exit code `4`.
- **Erro de rede/timeouts IMAP**:
    - Mensagem clara (`erro de conexão IMAP`, `timeout IMAP`).
    - Exit code `3`.

***

## 10. Critérios de aceite

Para considerar o `--wait-for-response` com IMAP concluído:

1. `EmailConfig` expandido com campos IMAP e wait‑for‑response, com defaults e validação funcionando.[^3]
2. Comando `send` aceita `--wfr` / `--wait-for-response` e calcula `waitMinutes` segundo regras de precedência (flag > config > 0/disable).[^8]
3. E‑mail enviado com `Message-ID` único incluso nos headers.
4. Função `waitForEmailResponse` implementada com:
    - Conexão IMAP segura.
    - Polling até `deadline`.
    - Busca primária por `In-Reply-To` e `References` contendo o `Message-ID` enviado.
    - Fallback por `Subject`/`Re: Subject`.
    - Impressão do corpo completo ou truncado conforme configuração.
5. Caso com resposta dentro do prazo:
    - Mensagem impressa no formato definido.
    - Exit code `0`.
6. Caso sem resposta:
    - Mensagens de timeout e “destinatário não respondeu”.
    - Exit code `3`.
7. `--verbose` exibe todos os detalhes de IMAP sem vazar credenciais.
8. `go test ./...` passando e novo `results/06_RESULTS.md` documentando:
    - exemplos de uso,
    - limitações (ex.: dependência de servidor IMAP padrão),
    - testes reais feitos (ex.: Gmail/Outlook/IMAP local).

Com essa especificação, o desenvolvedor do

<div align="center">⁂</div>

[^1]: https://blog.nodemailer.com/2020/04/27/tracking-email-replies-with-imap-api/

[^2]: https://www.chilkatsoft.com/imap_search_criteria.asp

[^3]: cast.md

[^4]: https://github.com/mscdex/node-imap/issues/796

[^5]: https://forum.rebex.net/1463/how-to-find-a-sent-message-using-mailmessage-inreplyto

[^6]: https://forum.rebex.net/22116/how-can-search-first-10-character-of-subject-in-imap

[^7]: https://stackoverflow.com/questions/21362001/search-criteria-of-imap-protocol-search-command

[^8]: 05_RESULTS.md

