# ESPECIFICAÇÃO TÉCNICA PROFUNDA: IMPLEMENTAÇÃO WAHA PROVIDER

**Objetivo:** Fornecer o caminho das pedras técnico para implementação do driver WAHA (WhatsApp HTTP API) na Fase 06, garantindo paridade de recursos (Wizard, Test, Send) com os demais providers e mantendo arquitetura stateless do CAST.

---

## 1. WAHA (WhatsApp HTTP API)

### 1.1 Arquitetura do Driver (`internal/providers/waha.go`)

O WAHA é fundamentalmente **diferente** dos outros providers por ser uma API de **terceiros** que precisa estar rodando como microserviço separado. O driver CAST é apenas um **HTTP client stateless** que consome endpoints REST.

**⚠️ PRINCÍPIO ARQUITETURAL CRÍTICO:**
- **CAST**: Stateless fire&forget (executa → envia → encerra)
- **WAHA**: Stateful persistent service (mantém sessão WhatsApp Web ativa)
- **NUNCA** tentar embutir browser/engine no CAST
- **SEMPRE** assumir WAHA como dependência externa

#### 1.1.1 Estrutura do Provider

```
package providers

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
    
    "github.com/eduardoalcantara/cast/internal/config"
)

type wahaProvider struct {
    apiURL  string        // Base URL do WAHA (ex: http://localhost:3000)
    session string        // Nome da sessão (default: "default")
    apiKey  string        // API Key opcional para auth
    timeout time.Duration // Timeout HTTP
    client  *http.Client  // Cliente HTTP reutilizável
}
```

#### 1.1.2 Construtor com Validações Robustas

```
// NewWAHAProvider cria instância do provider WAHA com validações completas
func NewWAHAProvider(cfg config.WAHAConfig) (Provider, error) {
    // Validação 1: API URL obrigatória
    if cfg.APIURL == "" {
        return nil, fmt.Errorf("WAHA API URL não configurada. Use: cast gateway add waha --interactive")
    }
    
    // Validação 2: Formato da URL
    apiURL := strings.TrimRight(cfg.APIURL, "/")
    if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
        return nil, fmt.Errorf("WAHA API URL inválida: deve começar com http:// ou https://")
    }
    
    // Validação 3: Timeout mínimo e default
    timeout := time.Duration(cfg.Timeout) * time.Second
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    if timeout < 5*time.Second {
        return nil, fmt.Errorf("timeout muito baixo: mínimo 5 segundos")
    }
    
    // Validação 4: Session default
    session := strings.TrimSpace(cfg.Session)
    if session == "" {
        session = "default"
    }
    
    // Cliente HTTP reutilizável (performance)
    client := &http.Client{
        Timeout: timeout,
        Transport: &http.Transport{
            MaxIdleConns:        10,
            IdleConnTimeout:     30 * time.Second,
            DisableCompression:  false,
            DisableKeepAlives:   false,
        },
    }
    
    return &wahaProvider{
        apiURL:  apiURL,
        session: session,
        apiKey:  cfg.APIKey,
        timeout: timeout,
        client:  client,
    }, nil
}
```

#### 1.1.3 Método Send (Núcleo do Provider)

**Estrutura de Requisição WAHA:**
- **Method:** `POST`
- **URL:** `<BASE_URL>/api/sendText`
- **Headers:**
  - `Content-Type: application/json`
  - `X-Api-Key: <API_KEY>` (se configurado)
- **Body JSON:**
  ```
  {
    "session": "default",
    "chatId": "5511999998888@c.us",
    "text": "Mensagem"
  }
  ```

**Implementação Completa:**

```
func (w *wahaProvider) Send(target, message string) error {
    // Validação 1: Target obrigatório
    if strings.TrimSpace(target) == "" {
        return fmt.Errorf("target vazio: forneça chatId no formato 5511999998888@c.us")
    }
    
    // Validação 2: Formato do chatId
    if err := w.validateChatID(target); err != nil {
        return err
    }
    
    // Validação 3: Mensagem obrigatória
    if strings.TrimSpace(message) == "" {
        return fmt.Errorf("mensagem vazia")
    }
    
    // Construir payload
    payload := map[string]interface{}{
        "session": w.session,
        "chatId":  target,
        "text":    message,
    }
    
    bodyBytes, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("erro ao serializar payload: %w", err)
    }
    
    // Construir request
    url := fmt.Sprintf("%s/api/sendText", w.apiURL)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return fmt.Errorf("erro ao criar request: %w", err)
    }
    
    // Headers
    req.Header.Set("Content-Type", "application/json")
    if w.apiKey != "" {
        req.Header.Set("X-Api-Key", w.apiKey)
    }
    
    // Executar request
    resp, err := w.client.Do(req)
    if err != nil {
        return fmt.Errorf("erro ao conectar com WAHA: %w. Verifique se está rodando em %s", err, w.apiURL)
    }
    defer resp.Body.Close()
    
    // Ler body da resposta
    respBody, _ := io.ReadAll(resp.Body)
    
    // Verificar status HTTP
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return w.handleErrorResponse(resp.StatusCode, respBody)
    }
    
    return nil
}

// validateChatID valida formato do Chat ID do WhatsApp
func (w *wahaProvider) validateChatID(chatID string) error {
    chatID = strings.TrimSpace(chatID)
    
    // Validação básica de formato
    if !strings.Contains(chatID, "@") {
        return fmt.Errorf(
            "chatId inválido: '%s'. Formato esperado: 5511999998888@c.us (contato) ou 120363XXX@g.us (grupo)",
            chatID,
        )
    }
    
    // Valida sufixo
    if !strings.HasSuffix(chatID, "@c.us") && !strings.HasSuffix(chatID, "@g.us") {
        return fmt.Errorf(
            "chatId inválido: deve terminar com @c.us (contato) ou @g.us (grupo). Recebido: %s",
            chatID,
        )
    }
    
    // Valida prefixo numérico (contatos)
    if strings.HasSuffix(chatID, "@c.us") {
        prefix := strings.Split(chatID, "@")
        if len(prefix) < 10 {
            return fmt.Errorf(
                "chatId muito curto: '%s'. Contatos devem ter código do país + DDD + número",
                chatID,
            )
        }
    }
    
    return nil
}

// handleErrorResponse processa erros HTTP do WAHA com mensagens amigáveis
func (w *wahaProvider) handleErrorResponse(statusCode int, body []byte) error {
    // Tentar parsear resposta JSON
    var errorResp map[string]interface{}
    if err := json.Unmarshal(body, &errorResp); err != nil {
        // Se não for JSON, retorna erro genérico
        return fmt.Errorf("WAHA retornou erro %d: %s", statusCode, string(body))
    }
    
    // Extrair mensagem de erro
    errorMsg := "erro desconhecido"
    if msg, ok := errorResp["error"].(string); ok {
        errorMsg = msg
    } else if msg, ok := errorResp["message"].(string); ok {
        errorMsg = msg
    }
    
    // Mensagens específicas por código
    switch statusCode {
    case 400:
        return fmt.Errorf("requisição inválida: %s", errorMsg)
    
    case 401:
        return fmt.Errorf("autenticação falhou: API Key incorreta ou ausente")
    
    case 404:
        if strings.Contains(errorMsg, "session") {
            return fmt.Errorf(
                "sessão '%s' não encontrada. Crie com: curl -X POST %s/api/sessions/start -d '{\"name\":\"%s\"}'",
                w.session,
                w.apiURL,
                w.session,
            )
        }
        return fmt.Errorf("endpoint não encontrado: verifique se WAHA está atualizado")
    
    case 500:
        if strings.Contains(strings.ToLower(errorMsg), "not connected") ||
           strings.Contains(strings.ToLower(errorMsg), "not authenticated") {
            return fmt.Errorf(
                "sessão '%s' não conectada. Escaneie o QR code em: %s",
                w.session,
                w.apiURL,
            )
        }
        return fmt.Errorf("erro interno do WAHA: %s", errorMsg)
    
    default:
        return fmt.Errorf("WAHA retornou erro %d: %s", statusCode, errorMsg)
    }
}

func (w *wahaProvider) Name() string {
    return "WAHA"
}
```

#### 1.1.4 Teste de Conectividade (Health Check)

**Endpoint de Status da Sessão:**
- **Method:** `GET`
- **URL:** `<BASE_URL>/api/sessions/<SESSION_NAME>`
- **Response:** JSON com status da sessão

```
// TestConnection verifica se WAHA está acessível e sessão está conectada
func (w *wahaProvider) TestConnection() error {
    url := fmt.Sprintf("%s/api/sessions/%s", w.apiURL, w.session)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return fmt.Errorf("erro ao criar request: %w", err)
    }
    
    if w.apiKey != "" {
        req.Header.Set("X-Api-Key", w.apiKey)
    }
    
    resp, err := w.client.Do(req)
    if err != nil {
        return fmt.Errorf("falha ao conectar: %w. Verifique se WAHA está rodando", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 404 {
        return fmt.Errorf(
            "sessão '%s' não existe. Crie com: curl -X POST %s/api/sessions/start -d '{\"name\":\"%s\"}'",
            w.session,
            w.apiURL,
            w.session,
        )
    }
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("WAHA retornou status %d", resp.StatusCode)
    }
    
    // Parse resposta
    var sessionInfo map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&sessionInfo); err != nil {
        return fmt.Errorf("erro ao parsear resposta: %w", err)
    }
    
    // Verificar status da sessão
    status, _ := sessionInfo["status"].(string)
    if status == "" {
        status = "UNKNOWN"
    }
    
    if status != "WORKING" {
        return fmt.Errorf(
            "sessão não está ativa (status: %s). Escaneie QR code em: %s",
            status,
            w.apiURL,
        )
    }
    
    return nil
}
```

---

## 2. WIZARD INTERATIVO (`cmd/cast/gateway.go`)

### 2.1 Função Principal do Wizard

O Wizard deve **educar o usuário** sobre a dependência externa e guiá-lo na configuração.

```
func runWAHAWizard(cfg *config.Config) error {
    // Cores para UX
    cyan := color.New(color.FgCyan, color.Bold)
    yellow := color.New(color.FgYellow)
    green := color.New(color.FgHiGreen, color.Bold)
    red := color.New(color.FgRed, color.Bold)
    
    // Banner educativo
    cyan.Println("╔════════════════════════════════════════════════════════════╗")
    cyan.Println("║   CONFIGURAÇÃO WAHA (WhatsApp HTTP API)                  ║")
    cyan.Println("╚════════════════════════════════════════════════════════════╝")
    fmt.Println()
    
    yellow.Println("⚠️  AVISOS IMPORTANTES:")
    yellow.Println("   -  WAHA deve estar RODANDO antes de configurar o CAST")
    yellow.Println("   -  Use Docker: docker run -d -p 3000:3000 devlikeapro/waha")
    yellow.Println("   -  WAHA NÃO é API oficial do WhatsApp (use por sua conta)")
    yellow.Println("   -  Ideal para: notificações pessoais e grupos pequenos")
    fmt.Println()
    
    // Perguntar se WAHA já está rodando
    var wahaRunning bool
    promptRunning := &survey.Confirm{
        Message: "WAHA já está rodando?",
        Default: false,
    }
    if err := survey.AskOne(promptRunning, &wahaRunning); err != nil {
        return err
    }
    
    if !wahaRunning {
        yellow.Println("\n📦 Para instalar WAHA, execute:")
        fmt.Println("   docker run -d --name waha -p 3000:3000 -v waha-data:/app/.sessions devlikeapro/waha")
        fmt.Println()
        yellow.Println("Após iniciar, acesse http://localhost:3000 e escaneie o QR code")
        fmt.Println()
        
        var continueAnyway bool
        promptContinue := &survey.Confirm{
            Message: "Continuar configuração mesmo assim?",
            Default: false,
        }
        if err := survey.AskOne(promptContinue, &continueAnyway); err != nil {
            return err
        }
        
        if !continueAnyway {
            cyan.Println("\n✋ Configuração cancelada. Instale WAHA e tente novamente.")
            return nil
        }
    }
    
    // Perguntas de configuração
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
                Message: "URL da API WAHA:",
                Default: "http://localhost:3000",
                Help:    "URL onde o WAHA está rodando (ex: http://localhost:3000 ou https://waha.exemplo.com)",
            },
            Validate: func(val interface{}) error {
                url, _ := val.(string)
                url = strings.TrimSpace(url)
                
                if url == "" {
                    return fmt.Errorf("URL obrigatória")
                }
                
                if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
                    return fmt.Errorf("URL deve começar com http:// ou https://")
                }
                
                // Teste básico de conectividade
                client := &http.Client{Timeout: 5 * time.Second}
                resp, err := client.Get(url + "/api/health")
                if err != nil {
                    yellow.Printf("\n⚠️  Não foi possível conectar em %s\n", url)
                    yellow.Println("   Verifique se WAHA está rodando")
                    
                    var continueAnyway bool
                    survey.AskOne(&survey.Confirm{
                        Message: "Continuar mesmo assim?",
                        Default: false,
                    }, &continueAnyway)
                    
                    if !continueAnyway {
                        return fmt.Errorf("conectividade falhou")
                    }
                }
                if resp != nil {
                    resp.Body.Close()
                }
                
                return nil
            },
        },
        {
            Name: "session",
            Prompt: &survey.Input{
                Message: "Nome da sessão WAHA:",
                Default: "default",
                Help:    "Nome da sessão criada no WAHA (geralmente 'default')",
            },
            Validate: func(val interface{}) error {
                session, _ := val.(string)
                session = strings.TrimSpace(session)
                
                if session == "" {
                    return fmt.Errorf("nome da sessão obrigatório")
                }
                
                // Validar caracteres permitidos
                if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(session) {
                    return fmt.Errorf("use apenas letras, números, hífen e underscore")
                }
                
                return nil
            },
        },
        {
            Name: "apikey",
            Prompt: &survey.Input{
                Message: "API Key (opcional - deixe vazio se não configurou):",
                Help:    "Se WAHA tiver autenticação habilitada (variável WHATSAPP_API_KEY)",
            },
        },
        {
            Name: "timeout",
            Prompt: &survey.Input{
                Message: "Timeout em segundos:",
                Default: "30",
                Help:    "Tempo máximo de espera por resposta (mínimo 5, recomendado 30)",
            },
            Validate: func(val interface{}) error {
                timeoutStr, _ := val.(string)
                timeout, err := strconv.Atoi(timeoutStr)
                if err != nil {
                    return fmt.Errorf("deve ser um número")
                }
                if timeout < 5 {
                    return fmt.Errorf("timeout mínimo: 5 segundos")
                }
                if timeout > 300 {
                    return fmt.Errorf("timeout máximo: 300 segundos (5 minutos)")
                }
                return nil
            },
        },
    }
    
    if err := survey.Ask(questions, &answers); err != nil {
        return err
    }
    
    // Processar respostas
    timeout, _ := strconv.Atoi(answers.Timeout)
    if timeout == 0 {
        timeout = 30
    }
    
    session := strings.TrimSpace(answers.Session)
    if session == "" {
        session = "default"
    }
    
    // Atualizar configuração
    cfg.WAHA.APIURL = strings.TrimRight(answers.APIURL, "/")
    cfg.WAHA.Session = session
    cfg.WAHA.APIKey = strings.TrimSpace(answers.APIKey)
    cfg.WAHA.Timeout = timeout
    
    // Resumo visual
    fmt.Println()
    cyan.Println("╔════════════════════════════════════════════════════════════╗")
    cyan.Println("║   RESUMO DA CONFIGURAÇÃO                                 ║")
    cyan.Println("╚════════════════════════════════════════════════════════════╝")
    fmt.Printf("  API URL:    %s\n", cfg.WAHA.APIURL)
    fmt.Printf("  Session:    %s\n", cfg.WAHA.Session)
    if cfg.WAHA.APIKey != "" {
        fmt.Printf("  API Key:    %s\n", maskToken(cfg.WAHA.APIKey))
    } else {
        fmt.Println("  API Key:    (não configurada)")
    }
    fmt.Printf("  Timeout:    %d segundos\n", cfg.WAHA.Timeout)
    fmt.Println()
    
    // Confirmação final
    var confirm bool
    promptConfirm := &survey.Confirm{
        Message: "Salvar esta configuração?",
        Default: true,
    }
    if err := survey.AskOne(promptConfirm, &confirm); err != nil {
        return err
    }
    
    if !confirm {
        yellow.Println("\n✋ Configuração cancelada")
        return nil
    }
    
    // Salvar
    if err := config.Save(cfg); err != nil {
        red.Printf("\n❌ Erro ao salvar: %v\n", err)
        return err
    }
    
    green.Println("\n✅ Configuração salva com sucesso!")
    fmt.Println()
    
    // Próximos passos
    cyan.Println("📋 PRÓXIMOS PASSOS:")
    fmt.Println("   1. Teste a conectividade:")
    fmt.Printf("      cast gateway test waha\n\n")
    fmt.Println("   2. Envie mensagem de teste:")
    fmt.Printf("      cast send waha SEUNUMERO@c.us \"Teste\"\n\n")
    fmt.Println("   3. Crie aliases para facilitar:")
    fmt.Printf("      cast alias add meu-zap waha SEUNUMERO@c.us\n\n")
    
    yellow.Println("💡 DICA: Para obter seu Chat ID, acesse:")
    yellow.Printf("   %s/api/%s/chats\n", cfg.WAHA.APIURL, cfg.WAHA.Session)
    
    return nil
}
```

### 2.2 Configuração via Flags

```
func addWAHAViaFlags(cmd *cobra.Command, cfg *config.Config) error {
    apiURL, _ := cmd.Flags().GetString("api-url")
    session, _ := cmd.Flags().GetString("session")
    apiKey, _ := cmd.Flags().GetString("api-key")
    timeout, _ := cmd.Flags().GetInt("timeout")
    
    // Validações
    if apiURL == "" {
        return fmt.Errorf("--api-url obrigatório. Exemplo: --api-url http://localhost:3000")
    }
    
    apiURL = strings.TrimRight(apiURL, "/")
    if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
        return fmt.Errorf("--api-url deve começar com http:// ou https://")
    }
    
    if session == "" {
        session = "default"
    }
    
    if timeout == 0 {
        timeout = 30
    }
    if timeout < 5 {
        return fmt.Errorf("--timeout mínimo: 5 segundos")
    }
    
    // Atualizar config
    cfg.WAHA.APIURL = apiURL
    cfg.WAHA.Session = session
    cfg.WAHA.APIKey = strings.TrimSpace(apiKey)
    cfg.WAHA.Timeout = timeout
    
    // Salvar
    if err := config.Save(cfg); err != nil {
        return fmt.Errorf("erro ao salvar: %w", err)
    }
    
    green := color.New(color.FgHiGreen, color.Bold)
    green.Println("✅ Configuração do WAHA salva com sucesso!")
    
    return nil
}
```

---

## 3. TESTE DE CONECTIVIDADE (`cmd/cast/gateway.go`)

### 3.1 Comando gateway test waha

```
func testWAHA(cfg config.WAHAConfig) error {
    cyan := color.New(color.FgCyan, color.Bold)
    green := color.New(color.FgHiGreen, color.Bold)
    yellow := color.New(color.FgYellow)
    red := color.New(color.FgRed, color.Bold)
    
    cyan.Println("╔════════════════════════════════════════════════════════════╗")
    cyan.Println("║   TESTE DE CONECTIVIDADE WAHA                            ║")
    cyan.Println("╚════════════════════════════════════════════════════════════╝")
    fmt.Println()
    
    // Teste 1: Health Check do WAHA
    fmt.Print("🔍 [1/3] Verificando se WAHA está respondendo... ")
    
    healthURL := fmt.Sprintf("%s/api/health", cfg.APIURL)
    client := &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Second}
    
    respHealth, err := client.Get(healthURL)
    if err != nil {
        red.Println("❌ FALHOU")
        red.Printf("\n   Erro: %v\n", err)
        red.Println("\n📋 DIAGNÓSTICO:")
        red.Println("   -  WAHA não está acessível")
        red.Println("   -  Verifique se o container está rodando: docker ps | grep waha")
        red.Println("   -  Verifique se a URL está correta")
        red.Printf("   -  URL configurada: %s\n", cfg.APIURL)
        return err
    }
    respHealth.Body.Close()
    
    if respHealth.StatusCode != 200 {
        red.Println("❌ FALHOU")
        red.Printf("\n   Status HTTP: %d\n", respHealth.StatusCode)
        return fmt.Errorf("health check retornou status %d", respHealth.StatusCode)
    }
    
    green.Println("✅ OK")
    
    // Teste 2: Verificar se sessão existe
    fmt.Print("🔍 [2/3] Verificando se sessão existe... ")
    
    sessionURL := fmt.Sprintf("%s/api/sessions/%s", cfg.APIURL, cfg.Session)
    req, err := http.NewRequest("GET", sessionURL, nil)
    if err != nil {
        red.Println("❌ FALHOU")
        return err
    }
    
    if cfg.APIKey != "" {
        req.Header.Set("X-Api-Key", cfg.APIKey)
    }
    
    respSession, err := client.Do(req)
    if err != nil {
        red.Println("❌ FALHOU")
        red.Printf("\n   Erro: %v\n", err)
        return err
    }
    defer respSession.Body.Close()
    
    if respSession.StatusCode == 401 {
        red.Println("❌ FALHOU")
        red.Println("\n   Erro: Autenticação falhou")
        red.Println("   -  API Key incorreta ou ausente")
        red.Println("   -  Verifique se WAHA foi iniciado com WHATSAPP_API_KEY")
        return fmt.Errorf("autenticação falhou")
    }
    
    if respSession.StatusCode == 404 {
        red.Println("❌ FALHOU")
        red.Println("\n   Erro: Sessão não encontrada")
        red.Println("\n📋 SOLUÇÃO:")
        red.Println("   Crie a sessão com:")
        red.Printf("   curl -X POST %s/api/sessions/start \\\n", cfg.APIURL)
        red.Printf("     -H 'Content-Type: application/json' \\\n")
        red.Printf("     -d '{\"name\": \"%s\"}'\n", cfg.Session)
        return fmt.Errorf("sessão '%s' não existe", cfg.Session)
    }
    
    if respSession.StatusCode != 200 {
        red.Println("❌ FALHOU")
        red.Printf("\n   Status HTTP: %d\n", respSession.StatusCode)
        return fmt.Errorf("status %d", respSession.StatusCode)
    }
    
    green.Println("✅ OK")
    
    // Parse info da sessão
    var sessionInfo map[string]interface{}
    if err := json.NewDecoder(respSession.Body).Decode(&sessionInfo); err != nil {
        yellow.Println("⚠️  Não foi possível parsear resposta")
        return nil
    }
    
    // Teste 3: Verificar status da sessão
    fmt.Print("🔍 [3/3] Verificando status da sessão... ")
    
    status, ok := sessionInfo["status"].(string)
    if !ok {
        yellow.Println("⚠️  Status desconhecido")
        status = "UNKNOWN"
    }
    
    switch status {
    case "WORKING":
        green.Println("✅ CONECTADA")
        
    case "SCAN_QR_CODE":
        yellow.Println("⚠️  AGUARDANDO QR CODE")
        fmt.Println()
        yellow.Println("📱 A sessão não está conectada:")
        yellow.Printf("   1. Acesse: %s\n", cfg.APIURL)
        yellow.Println("   2. Vá em 'Sessions' → clique na sessão")
        yellow.Println("   3. Escaneie o QR code com seu WhatsApp")
        
    case "FAILED", "STOPPED":
        red.Println("❌ INATIVA")
        red.Println("\n📋 SOLUÇÃO:")
        red.Println("   Reinicie a sessão:")
        red.Printf("   curl -X POST %s/api/sessions/%s/restart\n", cfg.APIURL, cfg.Session)
        return fmt.Errorf("sessão está inativa (status: %s)", status)
        
    default:
        yellow.Printf("⚠️  Status: %s\n", status)
    }
    
    // Resumo final
    fmt.Println()
    cyan.Println("╔════════════════════════════════════════════════════════════╗")
    cyan.Println("║   RESUMO DO TESTE                                        ║")
    cyan.Println("╚════════════════════════════════════════════════════════════╝")
    fmt.Printf("  URL:         %s\n", cfg.APIURL)
    fmt.Printf("  Session:     %s\n", cfg.Session)
    fmt.Printf("  Status:      %s\n", status)
    fmt.Printf("  Timeout:     %d segundos\n", cfg.Timeout)
    
    if cfg.APIKey != "" {
        fmt.Printf("  Auth:        Habilitada\n")
    } else {
        fmt.Printf("  Auth:        Desabilitada\n")
    }
    
    fmt.Println()
    
    if status == "WORKING" {
        green.Println("✅ TUDO OK! Pronto para enviar mensagens.")
        fmt.Println()
        cyan.Println("📋 TESTE DE ENVIO:")
        fmt.Println("   cast send waha SEUNUMERO@c.us \"Teste\"")
    } else {
        yellow.Println("⚠️  Configure a sessão antes de enviar mensagens")
    }
    
    return nil
}
```

---

## 4. INTEGRAÇÃO NA FACTORY (`internal/providers/factory.go`)

```
func GetProvider(name string, conf config.Config) (Provider, error) {
    // Resolve alias primeiro
    if alias := conf.GetAlias(name); alias != nil {
        name = alias.Provider
        // Se alias tiver target, substituir no contexto superior
    }
    
    // Normaliza nome
    normalized := normalizeProviderName(name)
    if normalized == "" {
        return nil, fmt.Errorf("provider desconhecido: %s", name)
    }
    
    // Factory
    switch normalized {
    case "tg":
        return NewTelegramProvider(conf.Telegram)
    
    case "mail":
        return NewEmailProvider(conf.Email)
    
    case "zap":
        return NewWhatsAppProvider(conf.WhatsApp)
    
    case "googlechat":
        return NewGoogleChatProvider(conf.GoogleChat)
    
    case "waha":  // NOVO
        return NewWAHAProvider(conf.WAHA)
    
    default:
        return nil, fmt.Errorf("provider não implementado: %s", normalized)
    }
}
```

---

## 5. ATUALIZAÇÃO DO CMD SEND (`cmd/cast/send.go`)

### 5.1 Help Atualizado

```
var sendCmd = &cobra.Command{
    Use:   "send provider target message",
    Short: "Envia mensagem via provider especificado",
    Long: `Envia mensagem através de um dos providers configurados.

PROVIDERS DISPONÍVEIS:
  -  tg, telegram      - Telegram Bot API
  -  mail, email       - SMTP Email
  -  zap, whatsapp     - WhatsApp Cloud API (Meta)
  -  googlechat        - Google Chat Webhook
  -  waha              - WAHA WhatsApp HTTP API (self-hosted)

FORMATOS DE TARGET:
  -  Telegram:   Chat ID numérico (ex: 123456789)
  -  Email:      Endereço de email (ex: user@exemplo.com)
  -  WhatsApp:   Número com código do país (ex: 5511999998888)
  -  GoogleChat: URL do webhook ou 'default'
  -  WAHA:       Chat ID WhatsApp (ex: 5511999998888@c.us)

ALIASES:
  Configure aliases para simplificar comandos:
    cast alias add meu-zap waha 5511999998888@c.us
    cast send waha meu-zap "Mensagem"
`,
    Example: `  # Telegram
  cast send tg 123456789 "Deploy finalizado"
  cast send telegram me "Lembrete pessoal"

  # Email
  cast send mail admin@empresa.com "Relatório anexo"
  cast send email team@empresa.com "Reunião às 15h"

  # WhatsApp Cloud API (Meta)
  cast send zap 5511999998888 "Alerta de sistema"

  # Google Chat
  cast send googlechat default "Build concluído"
  cast send googlechat https://chat.googleapis.com/v1/... "Erro crítico"

  # WAHA (WhatsApp self-hosted)
  cast send waha 5511999998888@c.us "Notificação via WAHA"
  cast send waha 120363XXXXX@g.us "Mensagem para grupo"
  cast send waha meu-zap "Usando alias"
`,
    Args: cobra.MinimumNArgs(3),
    RunE: runSend,
}
```

### 5.2 Função runSend (sem alterações)

O `runSend` já deve funcionar, pois usa a factory. Apenas garanta que:

```
func runSend(cmd *cobra.Command, args []string) error {
    provider := args
    target := args[1]
    message := strings.Join(args[2:], " ")
    
    // Carrega config
    cfg, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("erro ao carregar configuração: %w", err)
    }
    
    // Resolve alias se necessário
    if alias := cfg.GetAlias(target); alias != nil {
        provider = alias.Provider
        target = alias.Target
    }
    
    // Obtém provider via factory (WAHA incluído automaticamente)
    p, err := providers.GetProvider(provider, *cfg)
    if err != nil {
        return err
    }
    
    // Envia
    if err := p.Send(target, message); err != nil {
        color.Red("❌ Erro ao enviar: %v", err)
        return err
    }
    
    color.Green("✅ Mensagem enviada com sucesso via %s", p.Name())
    return nil
}
```

---

## 6. TESTES UNITÁRIOS (`internal/providers/waha_test.go`)

```
package providers

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/eduardoalcantara/cast/internal/config"
)

// TestWAHAProvider_NewProvider testa criação do provider com validações
func TestWAHAProvider_NewProvider(t *testing.T) {
    tests := []struct {
        name        string
        cfg         config.WAHAConfig
        expectError bool
        errorMsg    string
    }{
        {
            name: "configuração válida completa",
            cfg: config.WAHAConfig{
                APIURL:  "http://localhost:3000",
                Session: "test",
                APIKey:  "secret123",
                Timeout: 30,
            },
            expectError: false,
        },
        {
            name: "URL obrigatória",
            cfg: config.WAHAConfig{
                Session: "test",
            },
            expectError: true,
            errorMsg:    "não configurada",
        },
        {
            name: "URL inválida sem protocolo",
            cfg: config.WAHAConfig{
                APIURL: "localhost:3000",
            },
            expectError: true,
            errorMsg:    "http://",
        },
        {
            name: "session default aplicado",
            cfg: config.WAHAConfig{
                APIURL:  "http://localhost:3000",
                Session: "",
                Timeout: 30,
            },
            expectError: false,
        },
        {
            name: "timeout default aplicado",
            cfg: config.WAHAConfig{
                APIURL:  "http://localhost:3000",
                Timeout: 0,
            },
            expectError: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider, err := NewWAHAProvider(tt.cfg)
            
            if tt.expectError {
                if err == nil {
                    t.Errorf("Esperado erro, mas não ocorreu")
                } else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
                    t.Errorf("Erro não contém '%s': %v", tt.errorMsg, err)
                }
            } else {
                if err != nil {
                    t.Errorf("Erro inesperado: %v", err)
                }
                if provider == nil {
                    t.Error("Provider é nil")
                }
            }
        })
    }
}

// TestWAHAProvider_Send_Success testa envio bem-sucedido
func TestWAHAProvider_Send_Success(t *testing.T) {
    // Mock server que simula WAHA
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Valida endpoint
        if r.URL.Path != "/api/sendText" {
            t.Errorf("Endpoint incorreto: %s", r.URL.Path)
            w.WriteHeader(404)
            return
        }
        
        // Valida método
        if r.Method != "POST" {
            t.Errorf("Método incorreto: %s", r.Method)
            w.WriteHeader(405)
            return
        }
        
        // Valida Content-Type
        if r.Header.Get("Content-Type") != "application/json" {
            t.Error("Content-Type incorreto")
        }
        
        // Parse payload
        var payload map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            t.Errorf("Erro ao parsear payload: %v", err)
            w.WriteHeader(400)
            return
        }
        
        // Valida campos obrigatórios
        if payload["session"] != "test-session" {
            t.Errorf("Session incorreta: %v", payload["session"])
        }
        if payload["chatId"] != "5511999998888@c.us" {
            t.Errorf("ChatId incorreto: %v", payload["chatId"])
        }
        if payload["text"] != "Mensagem de teste" {
            t.Errorf("Text incorreto: %v", payload["text"])
        }
        
        // Resposta de sucesso
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(map[string]string{
            "id":      "msg-123",
            "status":  "sent",
        })
    }))
    defer server.Close()
    
    // Criar provider com mock URL
    cfg := config.WAHAConfig{
        APIURL:  server.URL,
        Session: "test-session",
        Timeout: 5,
    }
    
    provider, err := NewWAHAProvider(cfg)
    if err != nil {
        t.Fatalf("Erro ao criar provider: %v", err)
    }
    
    // Testar envio
    err = provider.Send("5511999998888@c.us", "Mensagem de teste")
    if err != nil {
        t.Errorf("Send falhou: %v", err)
    }
}

// TestWAHAProvider_Send_InvalidChatID testa validação de Chat ID
func TestWAHAProvider_Send_InvalidChatID(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        t.Error("Request não deveria ter sido enviada")
    }))
    defer server.Close()
    
    cfg := config.WAHAConfig{
        APIURL:  server.URL,
        Session: "test",
        Timeout: 5,
    }
    
    provider, _ := NewWAHAProvider(cfg)
    
    tests := []struct {
        name   string
        chatID string
    }{
        {"sem arroba", "5511999998888"},
        {"sufixo inválido", "5511999998888@invalid"},
        {"vazio", ""},
        {"só espaços", "   "},
        {"muito curto", "123@c.us"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := provider.Send(tt.chatID, "Teste")
            if err == nil {
                t.Errorf("Esperado erro para chatId '%s', mas não ocorreu", tt.chatID)
            }
        })
    }
}

// TestWAHAProvider_Send_SessionNotConnected testa erro de sessão
func TestWAHAProvider_Send_SessionNotConnected(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(500)
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
    
    if !strings.Contains(err.Error(), "não conectada") {
        t.Errorf("Mensagem de erro não é amigável: %v", err)
    }
}

// TestWAHAProvider_Send_SessionNotFound testa sessão inexistente
func TestWAHAProvider_Send_SessionNotFound(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(404)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Session not found",
        })
    }))
    defer server.Close()
    
    cfg := config.WAHAConfig{
        APIURL:  server.URL,
        Session: "inexistente",
        Timeout: 5,
    }
    
    provider, _ := NewWAHAProvider(cfg)
    err := provider.Send("5511999998888@c.us", "Teste")
    
    if err == nil {
        t.Error("Esperado erro, mas não ocorreu")
    }
    
    if !strings.Contains(err.Error(), "não encontrada") {
        t.Errorf("Mensagem não indica sessão inexistente: %v", err)
    }
}

// TestWAHAProvider_Send_WithAPIKey testa autenticação
func TestWAHAProvider_Send_WithAPIKey(t *testing.T) {
    expectedKey := "secret-api-key-123"
    
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Valida header de auth
        apiKey := r.Header.Get("X-Api-Key")
        if apiKey != expectedKey {
            t.Errorf("API Key incorreta. Esperado: %s, Recebido: %s", expectedKey, apiKey)
            w.WriteHeader(401)
            return
        }
        
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(map[string]string{"id": "msg-456"})
    }))
    defer server.Close()
    
    cfg := config.WAHAConfig{
        APIURL:  server.URL,
        Session: "test",
        APIKey:  expectedKey,
        Timeout: 5,
    }
    
    provider, _ := NewWAHAProvider(cfg)
    err := provider.Send("5511999998888@c.us", "Teste com auth")
    
    if err != nil {
        t.Errorf("Envio com API Key falhou: %v", err)
    }
}

// TestWAHAProvider_Name testa método Name
func TestWAHAProvider_Name(t *testing.T) {
    cfg := config.WAHAConfig{
        APIURL: "http://localhost:3000",
    }
    
    provider, _ := NewWAHAProvider(cfg)
    
    if provider.Name() != "WAHA" {
        t.Errorf("Nome incorreto: %s", provider.Name())
    }
}
```

---

## 7. CHECKLIST DE IMPLEMENTAÇÃO (DEFINITION OF DONE)

Use este checklist para validar que a implementação está completa:

### 7.1 Código Base
- [ ] `internal/config/config.go`: Struct `WAHAConfig` adicionada com tags `mapstructure`
- [ ] `internal/config/config.go`: Validação no método `Validate()` para WAHA
- [ ] `internal/providers/waha.go`: Provider completo implementando interface `Provider`
- [ ] `internal/providers/waha.go`: Validações robustas de Chat ID
- [ ] `internal/providers/waha.go`: Tratamento de erros com mensagens amigáveis
- [ ] `internal/providers/factory.go`: Case "waha" no switch de providers
- [ ] `internal/providers/factory.go`: Normalização "waha" adicionada

### 7.2 CLI Commands
- [ ] `cmd/cast/gateway.go`: Função `runWAHAWizard()` com UX educativa
- [ ] `cmd/cast/gateway.go`: Função `addWAHAViaFlags()` com validações
- [ ] `cmd/cast/gateway.go`: Função `testWAHA()` com diagnóstico completo
- [ ] `cmd/cast/gateway.go`: Flags `--api-url`, `--session`, `--api-key` adicionadas
- [ ] `cmd/cast/gateway.go`: Switch cases para WAHA em todos comandos (add/show/update/remove/test)
- [ ] `cmd/cast/alias.go`: Provider "waha" na função `normalizeProviderName`
- [ ] `cmd/cast/send.go`: Help atualizado com exemplos WAHA

### 7.3 Testes
- [ ] `internal/providers/waha_test.go`: Teste de criação com validações
- [ ] `internal/providers/waha_test.go`: Teste de envio bem-sucedido
- [ ] `internal/providers/waha_test.go`: Teste de Chat ID inválido
- [ ] `internal/providers/waha_test.go`: Teste de sessão não conectada
- [ ] `internal/providers/waha_test.go`: Teste de sessão não encontrada (404)
- [ ] `internal/providers/waha_test.go`: Teste com API Key
- [ ] `internal/providers/waha_test.go`: Cobertura mínima de 80%
- [ ] `go test ./...` passa 100% sem erros

### 7.4 Documentação
- [ ] `documents/05_TUTORIAL_WAHA.md`: Tutorial completo criado
- [ ] `documents/05_TUTORIAL_WAHA.md`: Instruções de instalação Docker
- [ ] `documents/05_TUTORIAL_WAHA.md`: Exemplos práticos de uso
- [ ] `documents/05_TUTORIAL_WAHA.md`: Seção de troubleshooting
- [ ] `documents/05_TUTORIAL_WAHA.md`: Avisos sobre riscos e API não-oficial
- [ ] `specifications/09_FASE_06_WAHA_IMPLEMENTATION_DEEP_SPECIFICATIONS.md`: Este arquivo
- [ ] `PROJECT_CONTEXT.md`: Atualizado com WAHA como 5º provider
- [ ] `PROJECT_CONTEXT.md`: Fase 06 marcada como concluída
- [ ] `README.md`: Atualizado com exemplo WAHA

### 7.5 Resultados e Evidências
- [ ] `results/06_RESULTS.md`: Relatório criado com métricas
- [ ] `results/06_RESULTS.md`: Log de testes incluído
- [ ] `results/06_RESULTS.md`: Exemplos de comandos funcionando
- [ ] `results/06_RESULTS.md`: Screenshot ou output de testes reais

### 7.6 Testes de Integração (Manual)
- [ ] `cast gateway add waha --interactive` executa wizard completo
- [ ] `cast gateway add waha --api-url http://localhost:3000` configura via flags
- [ ] `cast gateway show waha` exibe configuração corretamente
- [ ] `cast gateway test waha` valida conectividade e mostra status
- [ ] `cast gateway update waha --timeout 60` atualiza parcialmente
- [ ] `cast gateway remove waha` remove configuração
- [ ] `cast send waha 5511999998888@c.us "Teste"` envia mensagem real
- [ ] `cast alias add meu-zap waha 5511999998888@c.us` cria alias
- [ ] `cast send waha meu-zap "Via alias"` funciona com alias
- [ ] Arquivo `cast.yaml` persiste configuração corretamente
- [ ] Variáveis de ambiente `CAST_WAHA_*` funcionam

### 7.7 Qualidade e Padrões
- [ ] Código compila sem warnings: `go build ./...`
- [ ] Linter passa: `golangci-lint run` (se configurado)
- [ ] Formatação consistente: `gofmt -s -w .`
- [ ] Imports organizados: `goimports -w .`
- [ ] Sem hardcoded values (usar constantes)
- [ ] Comentários em funções públicas (godoc)
- [ ] Mensagens de erro user-friendly (português)
- [ ] Cores consistentes (cyan=info, green=success, red=error, yellow=warning)

---

## 8. DIFERENÇAS ARQUITETURAIS: WAHA vs WhatsApp Cloud API

Esta seção documenta as diferenças técnicas para justificar a existência de dois providers WhatsApp:

| Aspecto | WAHA (`waha`) | WhatsApp Cloud (`zap`) |
|---------|---------------|------------------------|
| **Tipo** | API HTTP sobre WhatsApp Web | API oficial Meta |
| **Dependência** | WAHA rodando externamente | Apenas credenciais Meta |
| **Autenticação** | QR Code (WhatsApp pessoal) | Business Account + Token |
| **Target Format** | `5511999998888@c.us` | `5511999998888` |
| **Grupos** | ✅ Suporta (`@g.us`) | ❌ Não suporta |
| **Limites** | Sem limite oficial (uso pessoal) | 250-1000-10000/dia (tiers) |
| **Sandbox** | Não precisa | Restrito a números verificados |
| **Status Oficial** | ⚠️ Não-oficial (risco de ban) | ✅ API oficial Meta |
| **Setup** | Docker + QR Code | Dashboard Meta + Aprovação |
| **Custo** | Gratuito (self-hosted) | Gratuito até 1000 conversas/mês |
| **Caso de Uso Ideal** | Notificações pessoais/pequenos grupos | Produção business crítica |

**Recomendação Arquitetural:**
- Use **WAHA** para: desenvolvimento, notificações pessoais, grupos pequenos, evitar burocracia Meta
- Use **WhatsApp Cloud** para: produção, alto volume, suporte oficial, compliance

---

## 9. MENSAGENS DE ERRO OBRIGATÓRIAS

Padronize as mensagens para facilitar troubleshooting:

```
// Erros de configuração
"WAHA API URL não configurada. Use: cast gateway add waha --interactive"
"WAHA API URL inválida: deve começar com http:// ou https://"
"timeout muito baixo: mínimo 5 segundos"

// Erros de target
"target vazio: forneça chatId no formato 5511999998888@c.us"
"chatId inválido: '%s'. Formato esperado: 5511999998888@c.us (contato) ou 120363XXX@g.us (grupo)"
"chatId deve terminar com @c.us (contato) ou @g.us (grupo)"
"chatId muito curto: contatos devem ter código do país + DDD + número"

// Erros de conectividade
"erro ao conectar com WAHA: %w. Verifique se está rodando em %s"
"WAHA retornou erro %d: %s"

// Erros de sessão
"sessão '%s' não encontrada. Crie com: curl -X POST %s/api/sessions/start -d '{\"name\":\"%s\"}'"
"sessão '%s' não conectada. Escaneie o QR code em: %s"
"sessão não está ativa (status: %s). Escaneie QR code em: %s"

// Erros de autenticação
"autenticação falhou: API Key incorreta ou ausente"
```

---

## 10. NOTAS FINAIS PARA O DESENVOLVEDOR

### ⚠️ ATENÇÃO CRÍTICA

1. **NEVER embutir browser/engine no CAST**: CAST é stateless, WAHA é stateful. São responsabilidades separadas.

2. **Validações são OBRIGATÓRIAS**: O WAHA falha silenciosamente se dados estiverem errados. O CAST deve validar antes de enviar.

3. **Mensagens de erro DEVEM ser educativas**: Não basta dizer "erro 500". Diga O QUE fazer para resolver.

4. **Wizard DEVE educar**: Usuário precisa entender que WAHA é dependência externa ANTES de configurar.

5. **Testes DEVEM usar httptest.NewServer**: Não dependa de WAHA rodando para testes unitários.

### ✅ PADRÕES DE QUALIDADE

- UX consistente com outros providers (cores, mensagens, flow)
- Código auto-explicativo com comentários onde necessário
- Erros com context chain (`fmt.Errorf("contexto: %w", err)`)
- Timeout generoso mas configurável (30s default)
- Validações client-side antes de request HTTP

### 📚 REFERÊNCIAS OBRIGATÓRIAS

- [WAHA Documentation](https://waha.devlike.pro/docs)
- [WAHA Send Messages](https://waha.devlike.pro/docs/how-to/send-messages/)
- [WAHA Sessions API](https://waha.devlike.pro/docs/how-to/sessions/)
- Fase 04 Implementation Specs (WhatsApp Cloud) - Usar como template
- Phase Implementation Protocol - Seguir rigorosamente

---

**Última atualização:** 2025-11-29  
**Versão do documento:** 1.0  
**Autor:** Arquiteto CAST + Eduardo (PO)  
**Status:** ✅ PRONTO PARA IMPLEMENTAÇÃO
```

***

## Resumo Executivo

Criei o documento **09_FASE_06_WAHA_IMPLEMENTATION_DEEP_SPECIFICATIONS.md** seguindo rigorosamente o padrão da Fase 04, incluindo:[2]

### ✅ Estrutura Completa

1. **Arquitetura do Driver** - Código Go completo com:
   - Validações robustas de Chat ID (`@c.us` vs `@g.us`)
   - Tratamento de erros HTTP com mensagens educativas
   - Cliente HTTP reutilizável para performance
   - Health check e teste de sessão

2. **Wizard Interativo** - UX educativa com:
   - Banner visual chamativo
   - Avisos sobre dependência externa ANTES de configurar
   - Validação de conectividade durante wizard
   - Próximos passos após configuração

3. **Teste de Conectividade** - Diagnóstico em 3 etapas:
   - Health check WAHA
   - Verificação de sessão
   - Status de conexão (QR code escaneado?)

4. **Checklist Definition of Done** - 50+ itens categorizados:
   - Código base (7 itens)
   - CLI commands (7 itens)
   - Testes unitários (8 itens)
   - Documentação (8 itens)
   - Testes de integração (11 itens)
   - Qualidade (9 itens)

5. **Testes Unitários Completos** - 8 cenários críticos:
   - Validações de criação
   - Envio bem-sucedido
   - Chat ID inválido
   - Sessão desconectada/inexistente
   - Autenticação com API Key

6. **Diferenças Arquiteturais** - Tabela comparativa WAHA vs WhatsApp Cloud

7. **Mensagens de Erro Padronizadas** - Copy-paste ready para o desenvolvedor

**Diferenciais vs Fase 04:**
- ⚠️ Ênfase em dependência externa (WAHA deve estar rodando)
- 🎓 Wizard mais educativo sobre riscos de API não-oficial
- 🔍 Validação de Chat ID específica do WhatsApp (`@c.us` vs `@g.us`)
- 🏥 Health check em 3 camadas (WAHA → Sessão → Status)

