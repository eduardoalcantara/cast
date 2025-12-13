package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MCPRequest representa uma requisição MCP
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse representa uma resposta MCP
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError representa um erro MCP
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolCallParams parâmetros para chamada de ferramenta
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// CastMCPServer servidor MCP para cast.exe
type CastMCPServer struct {
	castPath      string
	defaultAlias  string
}

// NewCastMCPServer cria novo servidor MCP
func NewCastMCPServer(castPath string, defaultAlias string) *CastMCPServer {
	if defaultAlias == "" {
		defaultAlias = "me" // Fallback padrão
	}
	return &CastMCPServer{
		castPath:     castPath,
		defaultAlias: defaultAlias,
	}
}

// HandleRequest processa requisição MCP
func (s *CastMCPServer) HandleRequest(req MCPRequest) MCPResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req.ID)
	case "tools/list":
		return s.handleToolsList(req.ID)
	case "tools/call":
		return s.handleToolCall(req.ID, req.Params)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

// handleInitialize inicializa o servidor MCP
func (s *CastMCPServer) handleInitialize(id interface{}) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "cast-mcp-server",
				"version": "1.0.0",
			},
		},
	}
}

// handleToolsList lista ferramentas disponíveis
func (s *CastMCPServer) handleToolsList(id interface{}) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"tools": []map[string]interface{}{
				// Tool: cast_send
				{
					"name":        "cast_send",
					"description": "Envia uma mensagem através do provider especificado (telegram, whatsapp, email, etc). Suporta múltiplos destinatários, anexos, assunto customizado e aguardar resposta por email (IMAP).",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"alias": map[string]interface{}{
								"type":        "string",
								"description": "Nome do alias configurado (ex: me, team, alerts). Se fornecido, ignora provider e target.",
							},
							"provider": map[string]interface{}{
								"type":        "string",
								"description": "Nome do provider: tg (telegram), mail (email), zap (whatsapp), google_chat, waha",
								"enum":        []string{"tg", "mail", "zap", "google_chat", "waha"},
							},
							"target": map[string]interface{}{
								"type":        "string",
								"description": "Destinatário (chat_id, email, número, webhook_url) ou 'me' para padrão. Para múltiplos, separar por vírgula ou ponto-e-vírgula.",
							},
							"message": map[string]interface{}{
								"type":        "string",
								"description": "Mensagem a ser enviada",
							},
							"subject": map[string]interface{}{
								"type":        "string",
								"description": "Assunto (apenas para email)",
							},
							"attachments": map[string]interface{}{
								"type":        "array",
								"description": "Lista de caminhos de arquivos para anexar (apenas para email)",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"wait_for_response": map[string]interface{}{
								"type":        "number",
								"description": "Aguarda por N minutos por uma resposta via IMAP (apenas para email). Retorna a resposta recebida no campo 'response'.",
							},
							"wfr": map[string]interface{}{
								"type":        "number",
								"description": "Alias para wait_for_response",
							},
							"full_layout": map[string]interface{}{
								"type":        "boolean",
								"description": "Inclui HTML no corpo da resposta (padrão: false, apenas texto). Apenas válido se wait_for_response estiver ativo.",
							},
							"full": map[string]interface{}{
								"type":        "boolean",
								"description": "Alias para full_layout",
							},
						},
						"required": []string{"message"},
					},
				},
				// Tool: cast_alias_add
				{
					"name":        "cast_alias_add",
					"description": "Adiciona um novo alias (atalho para provider + target)",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Nome do alias (ex: me, team, alerts)",
							},
							"provider": map[string]interface{}{
								"type":        "string",
								"description": "Provider: tg, mail, zap, google_chat, waha",
								"enum":        []string{"tg", "mail", "zap", "google_chat", "waha"},
							},
							"target": map[string]interface{}{
								"type":        "string",
								"description": "Destinatário (chat_id, email, número, webhook_url)",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "Descrição opcional do alias",
							},
						},
						"required": []string{"name", "provider", "target"},
					},
				},
				// Tool: cast_alias_list
				{
					"name":        "cast_alias_list",
					"description": "Lista todos os aliases configurados",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				// Tool: cast_alias_show
				{
					"name":        "cast_alias_show",
					"description": "Mostra detalhes de um alias específico",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Nome do alias",
							},
						},
						"required": []string{"name"},
					},
				},
				// Tool: cast_alias_remove
				{
					"name":        "cast_alias_remove",
					"description": "Remove um alias",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Nome do alias a remover",
							},
						},
						"required": []string{"name"},
					},
				},
				// Tool: cast_gateway_test
				{
					"name":        "cast_gateway_test",
					"description": "Testa conectividade de um gateway",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"gateway": map[string]interface{}{
								"type":        "string",
								"description": "Nome do gateway (telegram, email, whatsapp, etc)",
							},
						},
						"required": []string{"gateway"},
					},
				},
				// Tool: cast_config_show
				{
					"name":        "cast_config_show",
					"description": "Mostra a configuração atual do CAST",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				// Tool: cast_config_validate
				{
					"name":        "cast_config_validate",
					"description": "Valida a configuração atual do CAST",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		},
	}
}

// handleToolCall executa uma ferramenta
func (s *CastMCPServer) handleToolCall(id interface{}, params json.RawMessage) MCPResponse {
	var callParams ToolCallParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	switch callParams.Name {
	case "cast_send":
		return s.executeCastSend(id, callParams.Arguments)
	case "cast_alias_add":
		return s.executeCastAliasAdd(id, callParams.Arguments)
	case "cast_alias_list":
		return s.executeCastAliasList(id)
	case "cast_alias_show":
		return s.executeCastAliasShow(id, callParams.Arguments)
	case "cast_alias_remove":
		return s.executeCastAliasRemove(id, callParams.Arguments)
	case "cast_gateway_test":
		return s.executeCastGatewayTest(id, callParams.Arguments)
	case "cast_config_show":
		return s.executeCastConfigShow(id)
	case "cast_config_validate":
		return s.executeCastConfigValidate(id)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Tool not found: %s", callParams.Name),
			},
		}
	}
}

// executeCastSend executa cast.exe send
func (s *CastMCPServer) executeCastSend(id interface{}, args map[string]interface{}) MCPResponse {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "Message parameter is required",
			},
		}
	}

	// Construir comando
	var cmdArgs []string
	cmdArgs = append(cmdArgs, "send")

	// Verificar se usa alias
	if alias, ok := args["alias"].(string); ok && alias != "" {
		// Formato: cast send [alias] [message]
		cmdArgs = append(cmdArgs, alias, message)
	} else {
		// Se não especificou alias, mas também não especificou provider/target,
		// usar o alias padrão configurado
		provider, hasProvider := args["provider"].(string)
		target, hasTarget := args["target"].(string)

		if !hasProvider && !hasTarget {
			// Nenhum parâmetro especificado: usar alias padrão
			cmdArgs = append(cmdArgs, s.defaultAlias, message)
		} else {
			// Formato: cast send [provider] [target] [message]
			if provider == "" {
				provider = "tg" // Default: telegram
			}
			if target == "" {
				target = "me" // Default: me
			}
			cmdArgs = append(cmdArgs, provider, target, message)
		}
	}

	// Adicionar flags opcionais (apenas para email)
	if subject, ok := args["subject"].(string); ok && subject != "" {
		cmdArgs = append(cmdArgs, "--subject", subject)
	}

	if attachments, ok := args["attachments"].([]interface{}); ok && len(attachments) > 0 {
		for _, att := range attachments {
			if attPath, ok := att.(string); ok && attPath != "" {
				cmdArgs = append(cmdArgs, "--attachment", attPath)
			}
		}
	}

	// Verificar se deve aguardar resposta (wait-for-response)
	var waitForResponse int = -1
	if wfr, ok := args["wait_for_response"].(float64); ok && wfr > 0 {
		waitForResponse = int(wfr)
	} else if wfr, ok := args["wfr"].(float64); ok && wfr > 0 {
		waitForResponse = int(wfr)
	}

	if waitForResponse > 0 {
		cmdArgs = append(cmdArgs, "--wfr", fmt.Sprintf("%d", waitForResponse))

		// Verificar se deve incluir HTML (full_layout)
		fullLayout := false
		if full, ok := args["full_layout"].(bool); ok && full {
			fullLayout = true
		} else if full, ok := args["full"].(bool); ok && full {
			fullLayout = true
		}

		if fullLayout {
			cmdArgs = append(cmdArgs, "--full")
		}
	}

	// Executar comando
	cmd := exec.Command(s.castPath, cmdArgs...)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	// Se wait_for_response foi usado, tentar extrair a resposta do email
	var emailResponse string
	if waitForResponse > 0 && err == nil {
		// Procura por "=== EMAIL RESPONSE ===" no output
		lines := strings.Split(outputStr, "\n")
		inResponse := false
		var responseLines []string

		for _, line := range lines {
			if strings.Contains(line, "=== EMAIL RESPONSE ===") {
				inResponse = true
				continue
			}
			if strings.Contains(line, "=== END EMAIL RESPONSE ===") {
				break
			}
			if inResponse {
				responseLines = append(responseLines, line)
			}
		}

		if len(responseLines) > 0 {
			emailResponse = strings.Join(responseLines, "\n")
		}
	}

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("Failed to execute cast send: %v", err),
				Data:    outputStr,
			},
		}
	}

	// Construir resposta
	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": outputStr,
			},
		},
	}

	// Se uma resposta de email foi recebida, adicionar ao resultado
	if emailResponse != "" {
		result["response"] = emailResponse
		result["has_response"] = true
	} else if waitForResponse > 0 {
		result["has_response"] = false
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// executeCastAliasAdd executa cast.exe alias add
func (s *CastMCPServer) executeCastAliasAdd(id interface{}, args map[string]interface{}) MCPResponse {
	name, _ := args["name"].(string)
	provider, _ := args["provider"].(string)
	target, _ := args["target"].(string)

	if name == "" || provider == "" || target == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "name, provider, and target are required",
			},
		}
	}

	cmdArgs := []string{"alias", "add", name, provider, target}

	if desc, ok := args["description"].(string); ok && desc != "" {
		cmdArgs = append(cmdArgs, "--name", desc)
	}

	cmd := exec.Command(s.castPath, cmdArgs...)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("Failed to add alias: %v", err),
				Data:    outputStr,
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": outputStr,
				},
			},
		},
	}
}

// executeCastAliasList executa cast.exe alias list
func (s *CastMCPServer) executeCastAliasList(id interface{}) MCPResponse {
	cmd := exec.Command(s.castPath, "alias", "list")
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("Failed to list aliases: %v", err),
				Data:    outputStr,
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": outputStr,
				},
			},
		},
	}
}

// executeCastAliasShow executa cast.exe alias show
func (s *CastMCPServer) executeCastAliasShow(id interface{}, args map[string]interface{}) MCPResponse {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "name parameter is required",
			},
		}
	}

	cmd := exec.Command(s.castPath, "alias", "show", name)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("Failed to show alias: %v", err),
				Data:    outputStr,
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": outputStr,
				},
			},
		},
	}
}

// executeCastAliasRemove executa cast.exe alias remove
func (s *CastMCPServer) executeCastAliasRemove(id interface{}, args map[string]interface{}) MCPResponse {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "name parameter is required",
			},
		}
	}

	cmd := exec.Command(s.castPath, "alias", "remove", name)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("Failed to remove alias: %v", err),
				Data:    outputStr,
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": outputStr,
				},
			},
		},
	}
}

// executeCastGatewayTest executa cast.exe gateway test
func (s *CastMCPServer) executeCastGatewayTest(id interface{}, args map[string]interface{}) MCPResponse {
	gateway, ok := args["gateway"].(string)
	if !ok || gateway == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32602,
				Message: "gateway parameter is required",
			},
		}
	}

	cmd := exec.Command(s.castPath, "gateway", "test", gateway)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("Failed to test gateway: %v", err),
				Data:    outputStr,
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": outputStr,
				},
			},
		},
	}
}

// executeCastConfigShow executa cast.exe config show
func (s *CastMCPServer) executeCastConfigShow(id interface{}) MCPResponse {
	cmd := exec.Command(s.castPath, "config", "show")
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("Failed to show config: %v", err),
				Data:    outputStr,
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": outputStr,
				},
			},
		},
	}
}

// executeCastConfigValidate executa cast.exe config validate
func (s *CastMCPServer) executeCastConfigValidate(id interface{}) MCPResponse {
	cmd := exec.Command(s.castPath, "config", "validate")
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("Config validation failed: %v", err),
				Data:    outputStr,
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": outputStr,
				},
			},
		},
	}
}

func main() {
	// Detectar caminho do cast.exe
	castPath := os.Getenv("CAST_PATH")
	if castPath == "" {
		// Tentar caminhos padrão
		possiblePaths := []string{
			"D:\\proj\\cast\\run\\cast.exe",
			".\\cast.exe",
			"cast.exe",
		}

		for _, path := range possiblePaths {
			if absPath, err := filepath.Abs(path); err == nil {
				if _, err := os.Stat(absPath); err == nil {
					castPath = absPath
					break
				}
			}
		}

		if castPath == "" {
			fmt.Fprintf(os.Stderr, "Error: cast.exe not found. Set CAST_PATH environment variable or place cast.exe in PATH.\n")
			os.Exit(1)
		}
	}

	// Obter alias padrão da variável de ambiente
	defaultAlias := os.Getenv("CAST_DEFAULT_ALIAS")

	server := NewCastMCPServer(castPath, defaultAlias)

	// Ler requisições do stdin (MCP usa stdio)
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req MCPRequest
		if err := decoder.Decode(&req); err != nil {
			// EOF significa que o cliente fechou a conexão
			break
		}

		resp := server.HandleRequest(req)
		if err := encoder.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
			break
		}
	}
}
