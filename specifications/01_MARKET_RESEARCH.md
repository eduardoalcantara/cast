# PESQUISA DE MERCADO: GATEWAYS DE MENSAGERIA

## 1. TELEGRAM (ESCOLHIDO: BOT API)
- **Custo:** Grátis.
- **Método:** HTTP POST simples.
- **Auth:** Bot Token no Header/URL.
- **Restrição:** Usuário deve iniciar conversa (`/start`) com o bot antes.

## 2. WHATSAPP (ESCOLHIDO: META CLOUD API SANDBOX)
- **Custo:** Grátis (Sandbox) / Pago (Produção).
- **Método:** HTTP REST oficial.
- **Por que não web-scraping?** Instabilidade e risco de banimento em ferramentas CLI.
- **Uso:** Ideal para alertas críticos pessoais.

## 3. EMAIL (ESCOLHIDO: SMTP STANDARD)
- **Compatibilidade:** Gmail (App Password), Resend, SendGrid.
- **Método:** Lib `net/smtp` (Go) ou `gomail`.
- **Foco:** Notificações ricas (HTML/Anexos).

## 4. GOOGLE CHAT (GOOGLE WORKSPACE)
- **Método:** Incoming Webhook.
- **Custo:** Grátis (incluso no G-Suite/Workspace).
- **Uso:** Notificações institucionais.
