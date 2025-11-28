# ESPECIFICAÇÃO TÉCNICA PROFUNDA: IMPLEMENTAÇÃO WHATSAPP & GOOGLE CHAT

**Objetivo:** Fornecer o caminho das pedras técnico para implementação dos drivers complexos da Fase 04, garantindo paridade de recursos (Wizard, Test, Send) com os drivers simples.

## 1. WHATSAPP (META CLOUD API)

### 1.1 Arquitetura do Driver (`internal/providers/whatsapp.go`)
Diferente do Telegram, a API do WhatsApp é mais estrita. O driver deve suportar dois modos implícitos: **Sandbox** e **Produção**.

**Payload de Envio (Template Message):**
Para garantir compatibilidade (especialmente no Sandbox, que só aceita templates), o driver deve padronizar o envio usando o template `hello_world` (padrão) ou um template genérico de notificação se configurado.
*Por enquanto, vamos focar no envio de TEXTO LIVRE*, mas cientes de que em produção isso exige que a janela de conversa esteja aberta (24h). Se falhar, o erro deve ser claro.

**Estrutura de Requisição:**
- **Method:** `POST`
- **URL:** `https://graph.facebook.com/<API_VERSION>/<PHONE_NUMBER_ID>/messages`
- **Headers:**
  - `Authorization: Bearer <ACCESS_TOKEN>`
  - `Content-Type: application/json`
- **Body JSON (Texto Livre):**
  ```json
  {
    "messaging_product": "whatsapp",
    "to": "<TARGET_PHONE>",
    "type": "text",
    "text": { "body": "<MESSAGE>" }
  }
  ```

### 1.2 Wizard Interativo (`cmd/cast/gateway.go`)
O Wizard deve educar o usuário sobre onde pegar os dados.
1. **Phone Number ID:** "ID do número (não o número em si). Ex: 1059..."
2. **Access Token:** "Token (Começa com EAA...). Se for teste, lembre que expira em 24h."
3. **Business Account ID:** (Opcional, mas bom ter).

### 1.3 Teste de Conectividade (`cast gateway test whatsapp`)
Não envie mensagem. Chame o endpoint de debug ou verificação:
- **GET** `https://graph.facebook.com/<API_VERSION>/<PHONE_NUMBER_ID>`
- **Header:** `Authorization: Bearer <ACCESS_TOKEN>`
- **Sucesso:** Status 200 e JSON com dados do número.

---

## 2. GOOGLE CHAT (INCOMING WEBHOOKS)

### 2.1 Arquitetura do Driver (`internal/providers/googlechat.go`)
É o mais simples de todos, mas precisa de validação de URL robusta.

**Estrutura de Requisição:**
- **Method:** `POST`
- **URL:** `<WEBHOOK_URL>` (O target ou o configurado no gateway)
- **Body JSON:**
  ```json
  { "text": "<MESSAGE>" }
  ```

**Lógica de Target:**
- Se o comando for `cast send google_chat <URL_DO_WEBHOOK> "msg"`, use a URL.
- Se o comando for `cast send google_chat default "msg"`, use a URL configurada no `cast.yaml`.

### 2.2 Wizard Interativo
1. **Webhook URL:** Validar se começa com `https://chat.googleapis.com/`.

### 2.3 Teste de Conectividade (`cast gateway test google_chat`)
- Como webhooks são "write-only", o teste ideal é enviar uma mensagem "dry-run" ou validar o formato da URL.
- **Decisão:** O comando `test` sem target deve apenas validar a sintaxe da URL. Com target, envia "CAST Connectivity Test".

---

## 3. CHECKLIST DE IMPLEMENTAÇÃO (DEFINITION OF DONE)

Para cada um dos dois novos providers:
1. [ ] Criar arquivo `internal/providers/<nome>.go` implementando a interface `Provider`.
2. [ ] Adicionar ao `internal/providers/factory.go` (switch case).
3. [ ] Criar função `run<Nome>Wizard` em `cmd/cast/gateway.go`.
4. [ ] Adicionar flags correspondentes em `add<Nome>ViaFlags`.
5. [ ] Implementar Teste Unitário com Mock HTTP (`<nome>_test.go`).
6. [ ] Atualizar `cast gateway test` para suportar o novo provider.

**NÃO ACEITAMOS "HARDCODED" ou "DUMMY".** O código deve fazer requisições HTTP reais.
