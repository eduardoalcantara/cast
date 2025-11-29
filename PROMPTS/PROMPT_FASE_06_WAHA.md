# ROLE: Senior Go Engineer (WhatsApp Integration Specialist)
# PROJECT: CAST (Fase 06 - WAHA Provider)

## CONTEXTO

Os 4 providers principais (Telegram, Email, WhatsApp Cloud API, Google Chat) estão implementados e funcionais [file:28][file:29]. Agora precisamos adicionar **WAHA** (WhatsApp HTTP API) como **5º provider alternativo** para casos de uso específicos: notificações controladas para grupos pequenos/privados, contornando limites da Meta Cloud API [web:19][web:23].

**Premissa Arquitetural Fundamental:**
- CAST mantém arquitetura **stateless fire&forget**
- WAHA roda como **microserviço separado** (Docker/servidor)
- CAST é apenas um **HTTP client** que consome API WAHA
- **NÃO embutir browser/engine no CAST** - isso violaria o princípio stateless [file:29]

## INPUTS OBRIGATÓRIOS

Antes de começar, leia:
1. `specifications/06_PHASE_IMPLEMENTATION_PROTOCOL.md` - Protocolo de implementação [file:29]
2. `specifications/04_GATEWAY_CONFIG_SPEC.md` - Estruturas de config [file:5]
3. `specifications/08_FASE_04_ZAP_GOOGLE_IMPLEMENTATION_DEEP_SPECIFICATIONS.md` - Padrão de drivers HTTP [file:28]
4. Documentação oficial WAHA: https://waha.devlike.pro/docs/how-to/send-messages/

## REQUIREMENTS

### 1. Estrutura de Configuração (`internal/config/config.go`)

Adicione ao struct `Config`:

```
type Config struct {
    Telegram    TelegramConfig    `mapstructure:"telegram"`
    WhatsApp    WhatsAppConfig    `mapstructure:"whatsapp"`
    Email       EmailConfig       `mapstructure:"email"`
    GoogleChat  GoogleChatConfig  `mapstructure:"googlechat"`
    WAHA        WAHAConfig        `mapstructure:"waha"`  // NOVO
    Aliases     map[string]AliasConfig `mapstructure:"aliases"`
}

type WAHAConfig struct {
    APIURL     string `mapstructure:"apiurl"`     // Ex: http://localhost:3000
    Session    string `mapstructure:"session"`    // Nome da sessão (default: "default")
    APIKey     string `mapstructure:"apikey"`     // Opcional, se WAHA tiver auth
    Timeout    int    `mapstructure:"timeout"`    // Em segundos
}
```

**Validação:**
- `APIURL` obrigatório e deve começar com `http://` ou `https://`
- `Session` tem default "default" se vazio
- `Timeout` default 30s

### 2. Driver WAHA (`internal/providers/waha.go`)

Implemente o provider seguindo a interface existente:

```
package providers

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/eduardoalcantara/cast/internal/config"
)

type wahaProvider struct {
    apiURL  string
    session string
    apiKey  string
    timeout time.Duration
}

// NewWAHAProvider cria instância do provider WAHA
func NewWAHAProvider(cfg config.WAHAConfig) (Provider, error) {
    if cfg.APIURL == "" {
        return nil, fmt.Errorf("WAHA API URL não configurada")
    }
    
    timeout := time.Duration(cfg.Timeout) * time.Second
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    
    session := cfg.Session
    if session == "" {
        session = "default"
    }
    
    return &wahaProvider{
        apiURL:  cfg.APIURL,
        session: session,
        apiKey:  cfg.APIKey,
        timeout: timeout,
    }, nil
}

func (w *wahaProvider) Send(target, message string) error {
    // Endpoint: POST /api/sendText
    url := fmt.Sprintf("%s/api/sendText", w.apiURL)
    
    // Payload conforme doc WAHA
    payload := map[string]interface{}{
        "session": w.session,
        "chatId":  target,  // Formato: 5511999998888@c.us
        "text":    message,
    }
    
    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("erro ao criar payload: %w", err)
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        return fmt.Errorf("erro ao criar request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    if w.apiKey != "" {
        req.Header.Set("X-Api-Key", w.apiKey)
    }
    
    client := &http.Client{Timeout: w.timeout}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("erro ao enviar: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 && resp.StatusCode != 201 {
        var errResp map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&errResp)
        return fmt.Errorf("WAHA retornou erro %d: %v", resp.StatusCode, errResp)
    }
    
    return nil
}

func (w *wahaProvider) Name() string {
    return "WAHA"
}
```

**Notas Críticas:**
- Target deve estar no formato WhatsApp ID: `5511999998888@c.us` (grupos: `@g.us`)
- WAHA exige que a sessão esteja conectada (QR code escaneado previamente)
- Erros comuns: sessão desconectada, número inválido

### 3. Integração na Factory (`internal/providers/factory.go`)

Atualize o switch para reconhecer WAHA:

```
func GetProvider(name string, conf config.Config) (Provider, error) {
    // ... código existente ...
    
    switch normalized {
    case "tg", "telegram":
        return NewTelegramProvider(conf.Telegram)
    case "mail", "email":
        return NewEmailProvider(conf.Email)
    case "zap", "whatsapp":
        return NewWhatsAppProvider(conf.WhatsApp)
    case "googlechat":
        return NewGoogleChatProvider(conf.GoogleChat)
    case "waha":  // NOVO
        return NewWAHAProvider(conf.WAHA)
    default:
        return nil, fmt.Errorf("provider desconhecido: %s", name)
    }
}
```

### 4. Wizard Interativo (`cmd/cast/gateway.go`)

Implemente wizard educativo:

```
func runWAHAWizard(cfg *config.Config) error {
    cyan := color.New(color.FgCyan)
    yellow := color.New(color.FgYellow)
    
    // Introdução educativa
    cyan.Println("=== Configuração WAHA (WhatsApp HTTP API) ===")
    yellow.Println("⚠️  WAHA deve estar rodando separadamente (Docker/servidor)")
    yellow.Println("⚠️  Use para notificações controladas - não é API oficial Meta")
    fmt.Println()
    
    var answers struct {
        APIURL  string `survey:"apiurl"`
        Session string `survey:"session"`
        APIKey  string `survey:"apikey"`
        Timeout string `survey:"timeout"`
    }
    
    questions := []*survey.Question{
        {
            Name: "apiurl",
            Prompt: &survey.Input{
                Message: "URL da API WAHA (ex: http://localhost:3000):",
                Default: "http://localhost:3000",
            },
            Validate: func(val interface{}) error {
                url, _ := val.(string)
                if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
                    return fmt.Errorf("URL deve começar com http:// ou https://")
                }
                return nil
            },
        },
        {
            Name: "session",
            Prompt: &survey.Input{
                Message: "Nome da sessão WAHA (default: 'default'):",
                Default: "default",
            },
        },
        {
            Name: "apikey",
            Prompt: &survey.Input{
                Message: "API Key (opcional, deixe vazio se não usar):",
            },
        },
        {
            Name: "timeout",
            Prompt: &survey.Input{
                Message: "Timeout em segundos:",
                Default: "30",
            },
        },
    }
    
    if err := survey.Ask(questions, &answers); err != nil {
        return err
    }
    
    // Validação e conversão
    timeout := 30
    if answers.Timeout != "" {
        if t, err := strconv.Atoi(answers.Timeout); err == nil && t > 0 {
            timeout = t
        }
    }
    
    // Atualiza configuração
    cfg.WAHA.APIURL = answers.APIURL
    cfg.WAHA.Session = answers.Session
    cfg.WAHA.APIKey = answers.APIKey
    cfg.WAHA.Timeout = timeout
    
    // Resumo
    cyan.Println("\n✅ Configuração a ser salva:")
    fmt.Printf("  API URL:  %s\n", answers.APIURL)
    fmt.Printf("  Session:  %s\n", answers.Session)
    fmt.Printf("  Timeout:  %d segundos\n", timeout)
    
    // Confirmação
    var confirm bool
    if err := survey.AskOne(&survey.Confirm{
        Message: "Confirmar e salvar?",
        Default: true,
    }, &confirm); err != nil {
        return err
    }
    
    if !confirm {
        return fmt.Errorf("operação cancelada")
    }
    
    // Salva
    if err := config.Save(cfg); err != nil {
        return fmt.Errorf("erro ao salvar: %w", err)
    }
    
    green := color.New(color.FgHiGreen, color.Bold)
    green.Println("\n✅ Configuração do WAHA salva com sucesso!")
    yellow.Println("\n⚠️  Lembre-se: WAHA deve estar rodando e com sessão conectada")
    
    return nil
}
```

Adicione ao switch em `runGatewayWizard`:

```
case "waha":
    return runWAHAWizard(cfg)
```

### 5. Flags para Configuração Direta

Adicione flags em `gatewayAddCmd` e `gatewayUpdateCmd`:

```
gatewayAddCmd.Flags().String("api-url", "", "URL da API WAHA")
gatewayAddCmd.Flags().String("session", "default", "Nome da sessão WAHA")
gatewayAddCmd.Flags().String("api-key", "", "API Key WAHA (opcional)")
```

Implemente `addWAHAViaFlags`:

```
func addWAHAViaFlags(cmd *cobra.Command, cfg *config.Config) error {
    apiURL, _ := cmd.Flags().GetString("api-url")
    session, _ := cmd.Flags().GetString("session")
    apiKey, _ := cmd.Flags().GetString("api-key")
    timeout, _ := cmd.Flags().GetInt("timeout")
    
    if apiURL == "" {
        return fmt.Errorf("--api-url obrigatório")
    }
    
    if timeout == 0 {
        timeout = 30
    }
    
    if session == "" {
        session = "default"
    }
    
    cfg.WAHA.APIURL = apiURL
    cfg.WAHA.Session = session
    cfg.WAHA.APIKey = apiKey
    cfg.WAHA.Timeout = timeout
    
    if err := config.Save(cfg); err != nil {
        return fmt.Errorf("erro ao salvar: %w", err)
    }
    
    green := color.New(color.FgHiGreen, color.Bold)
    green.Println("✅ Configuração do WAHA salva com sucesso!")
    
    return nil
}
```

### 6. Teste de Conectividade (`cmd/cast/gateway.go`)

Implemente `testWAHA`:

```
func testWAHA(cfg config.WAHAConfig) error {
    cyan := color.New(color.FgCyan)
    green := color.New(color.FgHiGreen, color.Bold)
    red := color.New(color.FgRed, color.Bold)
    
    cyan.Println("🔍 Testando conectividade com WAHA...")
    
    // Endpoint de status: GET /api/sessions/{session}
    url := fmt.Sprintf("%s/api/sessions/%s", cfg.APIURL, cfg.Session)
    
    client := &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Second}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        red.Printf("❌ Erro ao criar request: %v\n", err)
        return err
    }
    
    if cfg.APIKey != "" {
        req.Header.Set("X-Api-Key", cfg.APIKey)
    }
    
    resp, err := client.Do(req)
    if err != nil {
        red.Printf("❌ Falha ao conectar: %v\n", err)
        red.Println("   Verifique se WAHA está rodando e acessível")
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        red.Printf("❌ WAHA retornou status %d\n", resp.StatusCode)
        red.Println("   Verifique se a sessão existe")
        return fmt.Errorf("status %d", resp.StatusCode)
    }
    
    // Parse response
    var sessionInfo map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&sessionInfo); err != nil {
        red.Printf("❌ Erro ao parsear resposta: %v\n", err)
        return err
    }
    
    // Verifica status da sessão
    status, ok := sessionInfo["status"].(string)
    if !ok {
        status = "UNKNOWN"
    }
    
    green.Println("✅ Conectividade OK!")
    fmt.Printf("   URL:     %s\n", cfg.APIURL)
    fmt.Printf("   Session: %s\n", cfg.Session)
    fmt.Printf("   Status:  %s\n", status)
    
    if status != "WORKING" {
        yellow := color.New(color.FgYellow)
        yellow.Println("\n⚠️  Sessão não está ativa!")
        yellow.Println("   Escaneie o QR code no painel WAHA")
    }
    
    return nil
}
```

Adicione ao switch em `gatewayTestCmd`:

```
case "waha":
    return testWAHA(cfg.WAHA)
```

### 7. Comando Send (`cmd/cast/send.go`)

O comando send já deve funcionar automaticamente via factory. Valide que:

```
cast send waha 5511999998888@c.us "Teste via WAHA"
```

**Formato do Target:**
- Contato individual: `5511999998888@c.us`
- Grupo: `120363XXXXX@g.us`

### 8. Help e Documentação

Atualize `cast send --help` para incluir WAHA nos exemplos:

```
sendCmd.Example = `  # Telegram
  cast send tg 123456789 "Mensagem"
  
  # Email
  cast send mail usuario@exemplo.com "Assunto: Teste"
  
  # WhatsApp (Meta Cloud API)
  cast send zap 5511999998888 "Olá"
  
  # Google Chat
  cast send googlechat https://chat.googleapis.com/... "Alerta"
  
  # WAHA (WhatsApp HTTP API)
  cast send waha 5511999998888@c.us "Notificação controlada"`
```

### 9. Testes Unitários (`internal/providers/waha_test.go`)

```
package providers

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/eduardoalcantara/cast/internal/config"
)

func TestWAHAProvider_Send_Success(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/api/sendText" {
            t.Errorf("Esperado /api/sendText, recebido %s", r.URL.Path)
        }
        
        if r.Method != "POST" {
            t.Errorf("Esperado POST, recebido %s", r.Method)
        }
        
        var payload map[string]interface{}
        json.NewDecoder(r.Body).Decode(&payload)
        
        if payload["session"] != "test-session" {
            t.Errorf("Session incorreta: %v", payload["session"])
        }
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"id": "msg-123"})
    }))
    defer server.Close()
    
    // Config
    cfg := config.WAHAConfig{
        APIURL:  server.URL,
        Session: "test-session",
        Timeout: 5,
    }
    
    // Provider
    provider, err := NewWAHAProvider(cfg)
    if err != nil {
        t.Fatalf("Erro ao criar provider: %v", err)
    }
    
    // Test
    err = provider.Send("5511999998888@c.us", "Teste")
    if err != nil {
        t.Errorf("Send falhou: %v", err)
    }
}

func TestWAHAProvider_Send_SessionDisconnected(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Session is not connected",
        })
    }))
    defer server.Close()
    
    cfg := config.WAHAConfig{
        APIURL:  server.URL,
        Session: "disconnected",
        Timeout: 5,
    }
    
    provider, _ := NewWAHAProvider(cfg)
    err := provider.Send("5511999998888@c.us", "Teste")
    
    if err == nil {
        t.Error("Esperado erro, mas não ocorreu")
    }
}
```

### 10. Normalização do Provider

Atualize as funções de normalização em `cmd/cast/alias.go` e `gateway.go`:

```
func normalizeProviderName(name string) string {
    switch strings.ToLower(name) {
    case "tg", "telegram":
        return "tg"
    case "mail", "email":
        return "mail"
    case "zap", "whatsapp":
        return "zap"
    case "googlechat", "gchat":
        return "googlechat"
    case "waha":  // NOVO
        return "waha"
    default:
        return ""
    }
}
```

## CHECKLIST DE IMPLEMENTAÇÃO (DEFINITION OF DONE)

Marque conforme for completando:

- [ ] `config.go`: Struct `WAHAConfig` adicionada
- [ ] `waha.go`: Provider implementado com `Send()` e `Name()`
- [ ] `factory.go`: Case "waha" no switch
- [ ] `gateway.go`: Função `runWAHAWizard()` implementada
- [ ] `gateway.go`: Função `addWAHAViaFlags()` implementada
- [ ] `gateway.go`: Função `testWAHA()` implementada
- [ ] `alias.go`: Provider "waha" na normalização
- [ ] `send.go`: Help atualizado com exemplo WAHA
- [ ] `waha_test.go`: Testes unitários criados
- [ ] `go test ./...` passa 100%
- [ ] Criar `05_TUTORIAL_WAHA.md` (próxima seção)
- [ ] Atualizar `PROJECT_CONTEXT.md`
- [ ] Criar `06_RESULTS.md` com evidências

## DELIVERABLE

Código compilável onde:

1. ✅ `cast gateway add waha --interactive` executa wizard
2. ✅ `cast gateway add waha --api-url http://localhost:3000` configura via flags
3. ✅ `cast gateway show waha` exibe configuração
4. ✅ `cast gateway test waha` valida conectividade
5. ✅ `cast send waha 5511999998888@c.us "Teste"` envia mensagem
6. ✅ `cast alias add meu-zap waha 5511999998888@c.us` cria alias
7. ✅ Testes unitários passam com cobertura >80%

## AVISOS IMPORTANTES

🚨 **NÃO FAÇA:**
- Embutir browser/engine no CAST
- Gerenciar sessão WhatsApp Web dentro do CAST
- Implementar QR code scan no CAST

✅ **FAÇA:**
- CAST = HTTP client stateless
- WAHA = Microserviço externo
- Documentação clara sobre dependência externa

---

**Última atualização:** 2025-11-29  
**Versão do documento:** 1.0  
**Autor:** Arquiteto CAST + Eduardo (PO)
