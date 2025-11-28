# TUTORIAIS DE CONFIGURAÇÃO - CAST

Este diretório contém tutoriais completos para configurar cada gateway suportado pelo CAST e o ambiente de desenvolvimento.

## 📚 ÍNDICE DE TUTORIAIS

### 0. [Ferramentas de Desenvolvimento](00_TUTORIAL_DEVELOPMENT_TOOLS.md)
- Instalação do Go
- Configuração do Git
- Setup do VS Code
- Ferramentas Go (goimports, golangci-lint)
- Scripts de build
- Windows e Linux

### 1. [Telegram](01_TUTORIAL_TELEGRAM.md)
- Criar bot via @BotFather
- Obter token e Chat ID
- Configurar no CAST
- Configurar aliases

### 2. [WhatsApp (Meta Cloud API)](02_TUTORIAL_WHATSAPP.md)
- Criar app no Meta for Developers
- Configurar WhatsApp Business API
- Sandbox vs Produção
- Obter Phone Number ID e Access Token

### 3. [Email (SMTP)](03_TUTORIAL_EMAIL.md)
- Gmail (App Password)
- Outlook/Hotmail
- SendGrid
- Resend
- Outros provedores SMTP

### 4. [Google Chat](04_TUTORIAL_GOOGLE_CHAT.md)
- Criar Incoming Webhook
- Obter URL do webhook
- Configurar no CAST
- Múltiplos espaços

## 🚀 INÍCIO RÁPIDO

1. Escolha o gateway que deseja configurar
2. Siga o tutorial correspondente
3. Teste a configuração com `cast send`
4. Configure aliases para facilitar o uso

## 📖 ESPECIFICAÇÃO TÉCNICA

Para detalhes técnicos completos, consulte:
- [Especificação de Configuração de Gateways](../specifications/04_GATEWAY_CONFIG_SPEC.md)

## 🔗 LINKS ÚTEIS

- [Documentação Oficial do CAST](../README.md)
- [Especificações Técnicas](../specifications/)
- [Master Plan](../specifications/00_MASTER_PLAN.md)
