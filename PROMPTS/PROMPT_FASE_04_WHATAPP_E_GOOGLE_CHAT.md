# ROLE: Senior Go Engineer (API Integration Specialist)
# PROJECT: CAST (Fase 04 - Advanced Drivers)

# CONTEXTO
Os drivers básicos (Telegram/Email) e o sistema de configuração (Fase 03) estão prontos. Agora precisamos implementar os "pesos pesados": WhatsApp e Google Chat.
Você hesitou anteriormente sugerindo configuração manual. **Isso não é aceitável.** O CAST deve oferecer Wizards e Flags completos para todos os providers.

# INPUTS
- `specifications/08_FASE_04_ZAP_GOOGLE_IMPLEMENTATION_DEEP_SPECIFICATIONS.md`: Guia técnico detalhado (LEITURA OBRIGATÓRIA).
- `specifications/04_GATEWAY_CONFIG_SPEC.md`: Estruturas de dados.

# REQUIREMENTS

## 1. Driver WhatsApp (`internal/providers/whatsapp.go`)
- Implemente a estrutura `whatsappProvider` usando `net/http`.
- **Send:** Use a API `messages` com payload JSON para texto (`type: text`).
- **Tratamento de Erro:** Se a API retornar erro (ex: token inválido, janela de 24h fechada), parseie o JSON de erro do Facebook e retorne uma mensagem amigável ao usuário.

## 2. Driver Google Chat (`internal/providers/googlechat.go`)
- Implemente `googleChatProvider`.
- **Lógica de Target:**
  - O `target` pode ser a URL completa do Webhook (para envio ad-hoc).
  - Ou pode ser uma string vazia/alias, caso em que deve usar a URL configurada no `cast.yaml`.

## 3. Integração na Factory (`internal/providers/factory.go`)
- Atualize o `switch` para instanciar `NewWhatsAppProvider` e `NewGoogleChatProvider`.
- Valide se a configuração necessária existe antes de instanciar.

## 4. Expansão da CLI (`cmd/cast/gateway.go`)
- Implemente `runWhatsAppWizard`: Pergunte `PhoneID` e `Token`.
- Implemente `runGoogleChatWizard`: Pergunte `Webhook URL`.
- Atualize `gatewayAddCmd` para suportar flags `--phone-id`, `--access-token`, `--webhook-url`.

## 5. Testes Unitários
- Crie `whatsapp_test.go` e `googlechat_test.go`.
- Use `httptest.NewServer` para simular as respostas da Meta e do Google.
- Teste o cenário de sucesso (200 OK) e de erro (400/401).

# DELIVERABLE
Código compilável onde:
1. `cast gateway add whatsapp --interactive` funciona.
2. `cast send zap 551199999 "Teste"` tenta enviar via API da Meta.
3. `cast send google_chat <URL> "Teste"` envia para o chat.
