# PROMPT: IMPLEMENTAÇÃO DA FASE 07 - IMAP MONITOR (--wait-for-response)

## CONTEXTO

Você é o Cursor IDE atuando como Engenheiro de Software Sênior especialista em Go (Golang), protocolos de e-mail (SMTP/IMAP), e ferramentas CLI com `spf13/cobra` e `spf13/viper`.

Este prompt tem como objetivo implementar o recurso `--wait-for-response` (ou `--wfr`) para o comando `cast send mail`, conforme especificado no documento **10_FASE_07_IMAP_MONITOR_SPECS.md**.

## DOCUMENTO DE REFERÊNCIA

**LEITURA OBRIGATÓRIA:** `specifications/10_FASE_07_IMAP_MONITOR_SPECS.md`

Este documento contém **TODAS** as especificações técnicas, regras de negócio, formato de saída, tratamento de erros e critérios de aceite. Você DEVE seguir cada seção à risca.

## OBJETIVO DA FASE 07

Implementar suporte a **espera por resposta via IMAP** após envio de e-mail, permitindo que o CAST:

1. Envie um e-mail via SMTP com `Message-ID` único.
2. Aguarde por até N minutos uma resposta identificada por cabeçalhos IMAP (`In-Reply-To`, `References`).
3. Exiba o corpo completo da resposta (ou truncado, se configurado).
4. Retorne exit codes específicos: `0` (resposta recebida), `3` (timeout sem resposta), `2`/`4` (erros técnicos).

## REGRAS IMPERATIVAS

### 1. PARIDADE DE RECURSOS (Protocolo de Implementação)

Conforme `specifications/06_PHASE_IMPLEMENTATION_PROTOCOL.md`:

- Esta funcionalidade é **específica do provider `email`**.
- Se `--wfr` for usado com outro provider (`tg`, `zap`, `googlechat`), **IGNORAR** e exibir aviso amarelo (não falhar).
- NÃO implementar para outros providers nesta fase (mesmo que tecnicamente possível).

### 2. ADERÊNCIA À ESPECIFICAÇÃO

**TUDO** está detalhado em `10_FASE_07_IMAP_MONITOR_SPECS.md`. Você DEVE:

- Seguir a estrutura exata de `EmailConfig` (seção 3.1).
- Implementar os defaults e validações conforme seção 3.3.
- Usar a lógica de geração de `Message-ID` da seção 4.1.
- Implementar o algoritmo de busca IMAP **exatamente** como descrito na seção 5.
- Respeitar o formato de saída (seção 6).
- Retornar os exit codes corretos (seção 7 e 9).
- Adicionar suporte a `--verbose` conforme seção 8.

### 3. BIBLIOTECA IMAP

Use a biblioteca Go **`emersion/go-imap`** (ou `emersion/go-imap/v2` se disponível):

```bash
go get github.com/emersion/go-imap
go get github.com/emersion/go-imap/client
```

Para parsing de mensagens, use `emersion/go-message`:

```bash
go get github.com/emersion/go-message/mail
```

Não invente protocolo IMAP manualmente; use essas bibliotecas padrão.

### 4. ASSINATURA DA FUNÇÃO DE ENVIO

A função `Send` em `internal/providers/email.go` atualmente retorna `error`.

Para suportar captura do `Message-ID`, você tem duas opções:

**Opção A (Recomendada):** Criar função auxiliar que retorna `(messageID string, err error)`:

```go
func (p *emailProvider) SendWithMessageID(target string, message string) (string, error) {
    // gera messageID
    // envia e-mail
    // retorna messageID, nil em caso de sucesso
}
```

E manter `Send` chamando `SendWithMessageID` internamente, para manter interface `Provider` intacta.

**Opção B:** Adicionar campo `LastMessageID string` no struct `emailProvider` e atualizá-lo em `Send`. Depois ler via getter.

Escolha a opção que fizer mais sentido arquiteturalmente, mas **NÃO quebre a interface `Provider`** existente.

### 5. INTEGRAÇÃO EM `cmd/cast/send.go`

Após o envio bem-sucedido com provider `email`, se `waitMinutes > 0`:

```go
if providerName == "mail" && waitMinutes > 0 {
    messageID := ... // capturado do envio
    subject := ... // subject enviado

    err := waitForEmailResponse(cfg.Email, messageID, subject, waitMinutes, verboseFlag)
    if err != nil {
        // tratar erro e exit code conforme seção 9
    }
}
```

### 6. FUNÇÃO `waitForEmailResponse`

Implementar em novo arquivo `internal/providers/email_imap.go` (ou no próprio `email.go`, mas separar logicamente):

```go
func waitForEmailResponse(
    cfg config.EmailConfig,
    messageID string,
    subject string,
    waitMinutes int,
    verbose bool,
) error
```

**Comportamento detalhado na seção 5 da especificação.**

### 7. BUSCA IMAP (CRÍTICO)

A ordem de busca DEVE ser:

1. **Primária:** `SEARCH HEADER In-Reply-To "<messageID>"`
2. **Secundária:** `SEARCH HEADER References "<messageID>"`
3. **Fallback:** `SEARCH HEADER Subject "Re: <subject>"` (apenas se 1 e 2 falharem sistematicamente)

Use a API do `emersion/go-imap` para criar `SearchCriteria` com `Header`:

```go
import "github.com/emersion/go-imap"

criteria := imap.NewSearchCriteria()
criteria.Header.Add("In-Reply-To", messageID)
```

### 8. FORMATO DE SAÍDA

**Seção 6 da especificação** define o formato exato. Use:

- `✓` (check verde) para sucesso.
- `⏳` (relógio) para aguardando.
- `⏰` (despertador) para timeout.
- `✗` (x vermelho) para falha.

Cores via `fatih/color` (já usado no projeto):

```go
color.Green("✓ Mensagem enviada com sucesso via email")
color.Yellow("⏳ Aguardando resposta por até %d minutos...", waitMinutes)
color.Green("✓ Resposta recebida em %s", elapsed)
color.Yellow("⏰ Tempo de espera esgotado (%d minutos).", waitMinutes)
color.Red("✗ O destinatário não respondeu à mensagem.")
```

### 9. EXIT CODES

Conforme especificação e padrão do projeto:

- `0` → Resposta recebida (sucesso total).
- `2` → Erro de configuração (IMAP não configurado, flags inválidas).
- `3` → Timeout lógico (destinatário não respondeu) OU erro de rede IMAP.
- `4` → Erro de autenticação IMAP.

Use `os.Exit(N)` nos lugares apropriados de `cmd/cast/send.go`.

### 10. MODO VERBOSE

Se `--verbose` estiver ativo, imprimir:

- Conexão IMAP (host, porta, SSL/TLS).
- Pasta selecionada.
- Comandos SEARCH executados.
- Quantidade de mensagens encontradas em cada ciclo.
- UID da mensagem quando encontrada.

**NÃO exibir senhas ou tokens completos** (mesmo em verbose).

### 11. CONFIGURAÇÃO EM `cast.yaml`

O wizard de configuração (`cast gateway add email --interactive`) **NÃO precisa** pedir IMAP nesta fase (a configuração IMAP é opcional, só necessária se `--wfr` for usado).

Mas se o usuário editar manualmente o `cast.yaml`, os campos IMAP devem ser lidos e validados conforme seção 3 da especificação.

### 12. TESTES

Criar `internal/providers/email_imap_test.go` com:

- Teste de geração de `Message-ID` (formato, unicidade).
- Mock de servidor IMAP (ou teste de integração real, se possível).
- Teste de parsing de resposta.
- Teste de timeout (simular nenhuma resposta por X segundos).

**Todos os testes devem passar** antes de considerar a fase concluída.

## ESTRUTURA DE ARQUIVOS MODIFICADOS/CRIADOS

### Arquivos a Modificar

1. `internal/config/config.go`
   - Adicionar campos IMAP em `EmailConfig` (seção 3.1).
   - Atualizar `applyDefaults` (seção 3.3).
   - Atualizar `Validate` (seção 3.3).

2. `internal/providers/email.go`
   - Adicionar geração de `Message-ID` (seção 4.1).
   - Incluir `Message-ID` nos headers.
   - Expor `messageID` para quem chama (conforme estratégia escolhida na regra 4).

3. `cmd/cast/send.go`
   - Adicionar flags `--wfr` e `--wait-for-response`.
   - Após envio via `email`, chamar `waitForEmailResponse` se aplicável (seção 5.3).
   - Tratar exit codes (seção 9).

### Arquivos a Criar

4. `internal/providers/email_imap.go`
   - Função `waitForEmailResponse` (seção 5).
   - Função auxiliar `connectIMAP`.
   - Função auxiliar `searchEmailByMessageID`.
   - Função auxiliar `fetchAndPrintResponse`.

5. `internal/providers/email_imap_test.go`
   - Testes unitários (regra 12).

6. `results/06_RESULTS.md`
   - Documento de resultados da Fase 07 (conforme protocolo de implementação).
   - Incluir:
     - Resumo do que foi implementado.
     - Lista de arquivos criados/modificados.
     - Exemplos de comandos testados.
     - Log de testes (`go test ./...`).
     - Métricas (linhas de código, testes, cobertura).

## CHECKLIST DE ENTREGA

Antes de dar a fase como concluída, VERIFIQUE:

- [ ] Todos os 10 critérios de aceite da seção 10 da especificação foram cumpridos.
- [ ] `go build ./cmd/cast` compila sem erros.
- [ ] `go test ./...` passa 100%.
- [ ] Help atualizado: `cast send --help` mostra `--wfr` e `--wait-for-response`.
- [ ] Exemplo de uso funcional:
  ```bash
  cast send mail teste@exemplo.com "Assunto" "Mensagem" --wfr 5 --verbose
  ```
- [ ] Arquivo `results/06_RESULTS.md` criado e completo.
- [ ] `PROJECT_CONTEXT.md` atualizado com status da Fase 07 (marcada como concluída).
- [ ] Commits organizados com mensagens claras.
- [ ] **ZERO regressões**: comandos anteriores (`cast send tg`, `cast alias list`, etc.) continuam funcionando.

## RESTRIÇÕES

1. **NÃO implemente nada além do especificado.** Se tiver dúvidas, pergunte antes de inventar comportamento.
2. **NÃO quebre interfaces existentes** (`Provider`, `Config`) sem justificativa arquitetural clara.
3. **NÃO faça otimizações prematuras.** Priorize clareza e aderência à especificação.
4. **NÃO ignore tratamento de erros.** Cada erro IMAP/SMTP deve ter mensagem clara e exit code correto.

## COMUNICAÇÃO

Ao terminar, responda com:

1. **Lista de arquivos modificados/criados** (paths completos).
2. **Resultado dos testes** (output de `go test ./...`).
3. **Exemplo de execução real** (screenshot ou texto do terminal).
4. **Confirmação do checklist** (todos os itens marcados).

Se encontrar **qualquer ambiguidade** na especificação, **PERGUNTE** antes de implementar. Não assuma comportamento não documentado.

---

## LEMBRE-SE

Este é um recurso crítico para permitir controle remoto do Cursor IDE via e-mail. A precisão na identificação da resposta (via `Message-ID`) é ESSENCIAL. Se implementar errado, o sistema vai capturar e-mails aleatórios ao invés da resposta esperada.

**Confiamos em você para seguir a especificação à risca.**

---

**Boa implementação!**

*Documento de referência obrigatória:* `specifications/10_FASE_07_IMAP_MONITOR_SPECS.md`
*Protocolo de entrega:* `specifications/06_PHASE_IMPLEMENTATION_PROTOCOL.md`
*Contexto do projeto:* `PROJECT_CONTEXT.md`
