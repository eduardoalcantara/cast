# CAST - MASTER PLAN
**Project Name:** CAST (CAST Automates Sending Tasks)
**Description:** Ferramenta CLI standalone para envio agnóstico de mensagens (Fire & Forget).
**Stack:** Go (Golang) 1.22+
**Architect:** Gemini 3 (AI)
**Product Owner:** User

## VISÃO GERAL
O `cast.exe` é um utilitário de linha de comando projetado para ser invocado em scripts, pipelines CI/CD ou manualmente, com o objetivo de enviar notificações rápidas via múltiplos canais (Telegram, WhatsApp, Email) sem travar o processo principal.

## FASES DO PROJETO
- [x] **Fase 00: Pesquisa & Naming**
  - Nome definido: CAST.
  - Stack definida: Go + Cobra + Viper.
  - Análise de Gateways (Telegram Free, WhatsApp Sandbox).
- [ ] **Fase 01: Bootstrap & CLI Skeleton (PRÓXIMA)**
  - Criação da estrutura de pastas.
  - Configuração do Viper (ENV > File).
  - UX Básica (Banner + Help).
- [ ] **Fase 02: Implementação de Drivers (Core)**
  - Driver: Telegram (HTTP Nativo).
  - Driver: Email (SMTP com suporte a TLS).
- [ ] **Fase 03: Integração Avançada**
  - Driver: WhatsApp (Meta Cloud API).
  - Webhooks (Google Chat).
- [ ] **Fase 04: Build & Release**
  - Cross-compilation (Windows/Linux).
