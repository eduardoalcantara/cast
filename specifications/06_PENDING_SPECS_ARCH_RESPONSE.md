# DECISÕES DE ARQUITETURA E ESPECIFICAÇÕES COMPLEMENTARES

**Referência:** `06_PENDING_SPECS.md`
**Autor:** Arquiteto CAST
**Data:** 2025-11-28

Este documento define as regras de negócio e comportamentos técnicos para as funcionalidades pendentes da Fase 03 e futuras.

## 1. COMANDOS DE CONFIGURAÇÃO (EXPORT/IMPORT)

### 1.1 `cast config export`

* **Output:**

  * Padrão: `stdout` (imprime no terminal).

  * Com flag `--output <file>`: Salva em arquivo.

  * **Sobrescrita:** Se o arquivo existir e não houver flag `--force`, deve perguntar ao usuário ou falhar.

* **Formato:**

  * Padrão: YAML.

  * Comentários: Não são suportados na exportação (limitação técnica de parsers padrão).

* **Mascaramento:**

  * Padrão (`--mask=true`): Substituir valores sensíveis por `*****`.

  * Campos a mascarar: `token`, `password`, `access_token`, `webhook_url` (parcialmente).

* **Validação:** Sempre validar (`cfg.Validate()`) antes de exportar. Se inválido, alertar mas permitir exportar (pode ser útil para debug).

### 1.2 `cast config import`

* **Merge (`--merge`):**

  * `false` (Padrão): **Substituição Total**. A configuração em memória é limpa e carregada do arquivo.

  * `true` (Merge Profundo):

    * **Gateways:** Campos presentes no arquivo importado sobrescrevem os atuais. Campos ausentes são mantidos.

    * **Aliases:** Novos aliases são adicionados. Aliases com mesmo nome são atualizados.

* **Backup:** OBRIGATÓRIO. Antes de qualquer importação que altere o arquivo local, criar `cast.yaml.bak`.

* **Auto-detecção:** Baseada na extensão (`.json`, `.yaml`, `.yml`). Se falhar, tentar YAML.

* **Validação:** Validar a struct resultante **antes** de salvar no disco. Se inválido, abortar operação e não salvar.

### 1.3 `cast config reload`

**Decisão:** Como o CAST é uma CLI (stateless) que carrega a config a cada execução, um comando `reload` só faz sentido se o usuário estiver em uma sessão interativa (REPL) ou para testar se o arquivo é legível.

* **Comportamento:** Força uma releitura do arquivo do disco, valida e imprime "Configuração recarregada e válida" ou o erro encontrado. Útil para verificar sintaxe após edição manual.

## 2. COMANDOS DE GATEWAY (UPDATE/TEST)

### 2.1 `cast gateway update`

* **Diferença Add/Update:**

  * `add`: Falha se o gateway já tiver configuração (token/host preenchido).

  * `update`: Falha se o gateway NÃO tiver configuração prévia.

* **Comportamento:** Atualização parcial (Patch).

  * Ex: `cast gateway telegram update --timeout 60` -> Mantém token, atualiza apenas timeout.

* **Validação:** Valida o objeto completo resultante. Não permite salvar estado inconsistente.

### 2.2 `cast gateway test`

**Objetivo:** Verificar conectividade e autenticação.

* **Telegram:** Chamar endpoint `getMe`.

* **Email:** Fazer a conexão SMTP, `EHLO`, `StartTLS` (se aplicável), Autenticação e `QUIT`. **Não enviar email** a menos que `--target` seja fornecido.

* **WhatsApp:** Chamar endpoint de metadados da conta (ex: `GET /<phone_id>`).

* **Google Chat:** Não é possível testar sem enviar mensagem.

  * Sem `--target`: Validar apenas formato da URL.

  * Com `--target`: Enviar mensagem "CAST Test Message".

## 3. COMANDOS DE ALIAS

### 3.1 `cast alias show`

* **Formato:** Exibir estilo "Ficha":

  ```
  Alias:      me
  Provider:   tg (Telegram)
  Target:     123456789
  Descrição:  Meu Telegram Pessoal

  ```

* **Erro:** Se não existir, retornar erro não-zero (exit code 1).

### 3.2 `cast alias update`

* **Mesclagem:** Aceitar atualização parcial.

  * `cast alias update me --target 999` -> Mantém provider e nome.

* **Validação:** Verificar se o provider referenciado existe/é válido.

## 4. WIZARDS PENDENTES

### 4.1 WhatsApp Wizard

**Ordem de Perguntas:**

1. Phone Number ID (Obrigatório).

2. Access Token (Obrigatório, input tipo password).

3. Business Account ID (Opcional).

4. API Version (Default: v18.0).

### 4.2 Google Chat Wizard

**Ordem de Perguntas:**

1. Webhook URL (Obrigatório).

   * *Validação:* Deve começar com `https://chat.googleapis.com`.

## 5. TESTES E QUALIDADE

* **Cobertura:** Focar testes na lógica de **Config Manager** (Save/Load/Merge) e **Factory**.

* **Mocking:**

  * Mockar o sistema de arquivos não é estritamente necessário se usar arquivos temporários (`t.TempDir()`) nos testes, o que é mais confiável.

  * Mockar APIs externas (HTTP/SMTP) é obrigatório para testes de unidade dos drivers.

## 6. TABELAS

* **Lib:** Continuar tentando usar `tablewriter` ou `text/tabwriter` nativo.

* **Estilo:** Bordas simples ou sem bordas (apenas espaçamento), priorizando legibilidade.

**Instrução ao Programador:** Implemente as funcionalidades pendentes seguindo estas diretrizes. Se houver conflito com specs anteriores, este documento prevalece.
