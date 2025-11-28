# UX SPEC: CAST INTERFACE
**Foco:** Developer Experience, Feedback Visual e Help Rico.

## 1. IDENTIDADE VISUAL (BANNER)
Sempre que o comando for invocado sem argumentos ou com `--help`, exibir:
(Cor sugerida: Verde Neon)

```text
┏┓┏┓┏┓┏┳┓
┃ ┣┫┗┓ ┃
┗┛┛┗┗┛ ┻
CAST Automates Sending Tasks
2025 Ⓒ Eduardo Alcântara
```

## 2. COMANDO `send`
**Sintaxe:** `cast send [provider] [target] [message] [subject] [files]`

**Help Contextual (Examples):**
O help deve conter exemplos práticos copiáveis:
```text
Examples:
  # Telegram (usando alias 'me' definido no config)
  cast send tg me "Deploy finalizado com sucesso"

  # WhatsApp (formato internacional)
  cast send zap 5511999998888 "Alerta: Disco cheio"

  # Email
  cast send mail admin@empresa.com "Bom dia!" "Relatório Diário" c:\rel.txt
```

## 3. FEEDBACK DE ERRO
- Erros de configuração ou envio devem ser impressos em **Vermelho**.
- Sucesso deve ser silencioso (exit code 0) ou mensagem mínima em **Verde** (configurável via flag `--verbose`).
