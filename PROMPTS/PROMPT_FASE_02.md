# ROLE: Senior Go Engineer (Network Specialist)
# PROJECT: CAST (Fase 02 - Core Drivers)

# OBJECTIVE
Implementar a lógica real de envio para os provedores **Telegram** e **Email** (SMTP), substituindo o esqueleto atual.

# CONTEXT
O projeto já possui a CLI (Cobra) e Config (Viper) configurados. A interface `Provider` já existe em `internal/providers/provider.go`.
Consulte `specifications/04_GATEWAY_CONFIG_SPEC.md` para as structs de configuração exatas.

# REQUIREMENTS

## 1. Provider Factory (`internal/providers/factory.go`)
Crie uma função `GetProvider(name string, conf *config.Config) (Provider, error)` que retorne a implementação correta baseada no argumento CLI:
- "tg", "telegram" -> Retorna `TelegramProvider`
- "mail", "email" -> Retorna `EmailProvider`
- "zap", "whatsapp" -> Retorna erro "not implemented yet" (Fase 03)
- **Logica de Aliases (CRÍTICO):** Antes de resolver o provider, verifique se o argumento `name` é um alias definido em `conf.Aliases`.
  - Se for alias, substitua `name` pelo `alias.Provider` e o `target` pelo `alias.Target`.
  - **ATENÇÃO:** O unmarshal de mapas (`map[string]AliasConfig`) via Viper em Go pode ser delicado. Implemente um teste unitário em `internal/config/config_test.go` garantindo que aliases definidos em YAML/JSON sejam carregados corretamente. Use `mapstructure` tags corretamente.

## 2. Driver Telegram (`internal/providers/telegram.go`)
- **Implementação:** Use a stdlib `net/http`.
- **Config:** Leia `TelegramConfig` da struct principal.
- **Lógica:**
  - Endpoint: `https://api.telegram.org/bot<TOKEN>/sendMessage`
  - Payload JSON: `{"chat_id": "<TARGET>", "text": "<MESSAGE>"}`
  - Se `target` for "me", use o ChatID Default da config. Se não houver default, retorne erro.
  - Valide se o status code é 200. Se não, retorne o erro com o corpo da resposta para debug.

## 3. Driver Email (`internal/providers/email.go`)
- **Implementação:** Use a stdlib `net/smtp`.
- **Config:** Leia `EmailConfig` da struct principal.
- **Lógica:**
  - Monte a mensagem seguindo o padrão MIME básico ("Subject: ... \r\n\r\n Body").
  - O argumento `message` da CLI será o corpo. O Assunto (Subject) pode ser fixo "Notificação CAST" por enquanto (ou extraído de uma flag futura).
  - Use `smtp.SendMail` com autenticação `PlainAuth`.
  - **TLS:** Se `UseTLS` for true, certifique-se de que a conexão use `StartTLS` (padrão do `SendMail` na porta 587) ou wrapper `tls.Client` se for porta 465 (SMTPS).

## 4. Integração (`cmd/cast/send.go`)
- Atualize o `RunE` do comando `send`.
- Instancie o provider usando a Factory.
- Execute `provider.Send()`.
- Exiba mensagem de sucesso (Verde) ou Erro (Vermelho).

# DELIVERABLE
Código compilável.
- `cast send tg me "Ola Mundo"` deve funcionar (se o token estiver no `.env` ou `cast.properties`).
- `cast send mail user@exemplo.com "Teste"` deve funcionar (se SMTP estiver configurado).
- Teste unitário comprovando que `aliases` são carregados do config.
