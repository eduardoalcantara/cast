# 🚀 CAST MCP Server

Servidor MCP (Model Context Protocol) para integração do `cast.exe` com Cursor IDE e outros clientes MCP.

## 📋 Funcionalidades Expostas

O servidor MCP expõe as seguintes ferramentas do `cast.exe`:

### 1. **cast_send** - Enviar Mensagens
Envia mensagens através de múltiplos providers (Telegram, WhatsApp, Email, etc).

**Parâmetros**:
- `message` (obrigatório): Mensagem a ser enviada
- `alias` (opcional): Nome do alias configurado (ex: "me", "team")
- `provider` (opcional): Provider (tg, mail, zap, google_chat, waha)
- `target` (opcional): Destinatário ou "me" para padrão
- `subject` (opcional): Assunto (apenas para email)
- `attachments` (opcional): Lista de arquivos para anexar (apenas para email)
- `wait_for_response` ou `wfr` (opcional, bool): Ativa espera por resposta via IMAP (usa tempo do config ou 30min, apenas para email). Retorna a resposta no campo `response`.
- `wfr_minutes` (opcional, int): Especifica tempo de espera em minutos (sobrescreve config, apenas para email). Se usado sozinho, ativa automaticamente a espera.
- `full_layout` ou `full` (opcional): Inclui HTML no corpo da resposta (padrão: false, apenas texto). Apenas válido se `wait_for_response` estiver ativo.

**Exemplos de uso no Cursor**:
- "Notifique-me quando a tarefa terminar"
- "Envie email para admin@empresa.com com o relatório"
- "Envie mensagem para o time no Telegram"

### 2. **cast_alias_add** - Adicionar Alias
Adiciona um novo alias (atalho para provider + target).

**Parâmetros**:
- `name` (obrigatório): Nome do alias
- `provider` (obrigatório): Provider (tg, mail, zap, etc)
- `target` (obrigatório): Destinatário
- `description` (opcional): Descrição do alias

### 3. **cast_alias_list** - Listar Aliases
Lista todos os aliases configurados.

### 4. **cast_alias_show** - Mostrar Alias
Mostra detalhes de um alias específico.

**Parâmetros**:
- `name` (obrigatório): Nome do alias

### 5. **cast_alias_remove** - Remover Alias
Remove um alias.

**Parâmetros**:
- `name` (obrigatório): Nome do alias a remover

### 6. **cast_gateway_test** - Testar Gateway
Testa conectividade de um gateway.

**Parâmetros**:
- `gateway` (obrigatório): Nome do gateway (telegram, email, etc)

### 7. **cast_config_show** - Mostrar Configuração
Mostra a configuração atual do CAST.

### 8. **cast_config_validate** - Validar Configuração
Valida a configuração atual do CAST.

---

## 🔧 Configuração no Cursor IDE

### Método 1: Via UI

1. Abra Cursor IDE
2. Vá em **Settings** → **Agents** → **Tools & MCP**
3. Clique em **Add Custom MCP**
4. Configure:
   - **Name**: `cast`
   - **Command**: `go`
   - **Args**: `["run", "D:\\proj\\ia\\tools\\mcp\\cast-mcp-server\\main.go"]`
   - **Env**: Adicione `CAST_PATH` = `D:\proj\cast\run\cast.exe`

### Método 2: Via Arquivo de Configuração

Crie/edite o arquivo `.cursor/mcp.json` (na raiz do projeto ou em `%APPDATA%\Cursor\User\globalStorage\`):

```json
{
  "mcpServers": {
    "cast": {
      "command": "go",
      "args": [
        "run",
        "D:\\proj\\ia\\tools\\mcp\\cast-mcp-server\\main.go"
      ],
      "env": {
        "CAST_PATH": "D:\\proj\\cast\\run\\cast.exe"
      }
    }
  }
}
```

**Ou usando o executável compilado**:

```json
{
  "mcpServers": {
    "cast": {
      "command": "D:\\proj\\ia\\tools\\mcp\\cast-mcp-server\\cast-mcp-server.exe",
      "env": {
        "CAST_PATH": "D:\\proj\\cast\\run\\cast.exe"
      }
    }
  }
}
```

---

## 🧪 Teste Manual

Para testar o servidor MCP manualmente:

```bash
# Teste 1: Initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | go run tools/mcp/cast-mcp-server/main.go

# Teste 2: List Tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | go run tools/mcp/cast-mcp-server/main.go

# Teste 3: Call Tool
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"cast_send","arguments":{"alias":"me","message":"Teste MCP"}}}' | go run tools/mcp/cast-mcp-server/main.go
```

---

## 📝 Exemplos de Uso no Cursor

### Exemplo 1: Notificação Simples
```
Usuário: "Notifique-me quando terminar a refatoração"
Cursor: Chama cast_send({"alias": "me", "message": "✅ Refatoração concluída"})
```

### Exemplo 2: Notificação com Detalhes
```
Usuário: "Envie email para admin@empresa.com informando que o deploy foi concluído"
Cursor: Chama cast_send({
  "provider": "mail",
  "target": "admin@empresa.com",
  "message": "Deploy concluído com sucesso",
  "subject": "Deploy - Sistema Principal"
})
```

### Exemplo 3: Múltiplos Destinatários
```
Usuário: "Notifique o time no Telegram sobre a conclusão"
Cursor: Chama cast_send({
  "provider": "tg",
  "target": "123456789,987654321",
  "message": "✅ Tarefa concluída: Implementação do módulo X"
})
```

### Exemplo 4: Email com Anexo
```
Usuário: "Envie o relatório por email para admin@empresa.com"
Cursor: Chama cast_send({
  "provider": "mail",
  "target": "admin@empresa.com",
  "message": "Segue o relatório em anexo",
  "subject": "Relatório Diário",
  "attachments": ["relatorio.pdf"]
})
```

### Exemplo 5: Email Aguardando Resposta
```
Usuário: "Envie email para admin@empresa.com perguntando se posso prosseguir e aguarde a resposta"
Cursor: Chama cast_send({
  "provider": "mail",
  "target": "admin@empresa.com",
  "message": "Posso prosseguir com a refatoração?",
  "subject": "Confirmação: Refatoração",
  "wfr": true,
  "wfr_minutes": 30
})
// Retorna: { "content": [...], "response": "Sim, pode prosseguir", "has_response": true }
```

### Exemplo 6: Email Aguardando Resposta com HTML
```
Usuário: "Envie email e aguarde resposta incluindo HTML"
Cursor: Chama cast_send({
  "provider": "mail",
  "target": "admin@empresa.com",
  "message": "Pergunta importante",
  "subject": "Confirmação",
  "wfr": true,
  "wfr_minutes": 15,
  "full": true
})
```

---

## 🔍 Troubleshooting

### Problema: "cast.exe not found"
**Solução**: Configure a variável de ambiente `CAST_PATH` ou coloque `cast.exe` no PATH.

### Problema: "Method not found"
**Solução**: Verifique se está usando o nome correto da tool (ex: `cast_send`, não `cast-send`).

### Problema: "Invalid params"
**Solução**: Verifique se todos os parâmetros obrigatórios estão presentes e no formato correto.

### Problema: Cursor não detecta o servidor MCP
**Solução**:
1. Verifique se o arquivo de configuração está no local correto
2. Reinicie o Cursor IDE
3. Verifique os logs do Cursor para erros

---

## 📚 Referências

- **MCP Specification**: https://modelcontextprotocol.io
- **Cursor IDE MCP**: Documentação oficial do Cursor
- **cast.exe**: Documentação do projeto cast

---

**Versão**: 1.0.0
**Data**: 2025-12-11
