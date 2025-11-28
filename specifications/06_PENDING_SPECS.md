# ESPECIFICAÇÕES PENDENTES PARA IMPLEMENTAÇÃO

**Objetivo:** Documentar as especificações necessárias do arquiteto de sistemas para implementar as funcionalidades pendentes da Fase 03.

**Status:** 🟡 Aguardando especificações do arquiteto

---

## 1. `cast config export/import`

### 1.1 Export

**O que já está especificado:**
- ✅ Flags: `--format`, `--output`, `--mask`
- ✅ Exemplos básicos

**O que precisa ser especificado:**

1. **Comportamento de `--output`:**
   - Se não especificado, usa stdout?
   - Se arquivo já existe, sobrescreve ou pergunta?
   - Deve criar backup do arquivo existente?

2. **Formato de saída:**
   - Quando usar `--format properties`, como converter estruturas aninhadas?
   - Deve incluir comentários no YAML exportado?
   - Ordem dos campos na saída (alfabética ou mantém ordem original)?

3. **Mascaramento:**
   - Quais campos devem ser mascarados? (tokens, senhas, webhook URLs?)
   - Padrão de mascaramento: `*****` ou `*****[últimos 4 chars]`?

4. **Validação:**
   - Deve validar configuração antes de exportar?
   - O que fazer se configuração estiver inválida?

### 1.2 Import

**O que já está especificado:**
- ✅ Flags: `--merge`, `--format`
- ✅ Exemplos básicos

**O que precisa ser especificado:**

1. **Comportamento de `--merge`:**
   - Quando `--merge=false` (padrão): sobrescreve completamente ou apenas campos fornecidos?
   - Quando `--merge=true`: como mesclar?
     - Gateways: substitui gateway inteiro ou mescla campos?
     - Aliases: adiciona novos ou substitui existentes?
   - O que acontece com campos não especificados no arquivo importado?

2. **Validação:**
   - Deve validar arquivo antes de importar?
   - O que fazer se arquivo estiver inválido?
   - Deve criar backup da configuração atual antes de importar?

3. **Auto-detecção de formato:**
   - Como detectar formato automaticamente? (extensão do arquivo?)
   - O que fazer se formato não puder ser detectado?

4. **Tratamento de erros:**
   - Se importação falhar parcialmente, deve reverter tudo ou manter o que foi importado?
   - Mensagens de erro específicas para cada tipo de problema?

---

## 2. `cast config reload`

**O que já está especificado:**
- ✅ Comando básico: `cast config reload`
- ✅ Descrição: "Útil após editar arquivo manualmente"

**O que precisa ser especificado:**

1. **Comportamento:**
   - O que exatamente "reload" faz?
     - Recarrega do arquivo de configuração?
     - Limpa cache do Viper?
     - Recarrega variáveis de ambiente?
   - Deve validar após recarregar?

2. **Feedback:**
   - Deve mostrar o que foi recarregado?
   - Deve mostrar diferenças entre configuração antiga e nova?
   - Mensagem de sucesso/erro?

3. **Uso:**
   - É apenas informativo ou tem efeito prático?
   - A configuração recarregada é usada imediatamente ou apenas na próxima execução?

---

## 3. `cast gateway update`

**O que já está especificado:**
- ✅ Sintaxe: `cast gateway update <provider> [flags]`
- ✅ Flags: Mesmas do comando `add`
- ✅ Exemplos básicos

**O que precisa ser especificado:**

1. **Diferença entre `add` e `update`:**
   - `add` cria nova configuração ou atualiza se já existe?
   - `update` apenas atualiza campos fornecidos ou substitui tudo?
   - O que acontece se tentar `update` em gateway não configurado?

2. **Mesclagem de configurações:**
   - Se apenas `--timeout` for fornecido, mantém outros campos (token, etc.)?
   - Como mesclar campos opcionais vs obrigatórios?
   - Deve validar configuração completa após update?

3. **Comportamento:**
   - Deve mostrar configuração atual antes de atualizar?
   - Deve pedir confirmação para atualizações?
   - Deve criar backup antes de atualizar?

4. **Validação:**
   - Deve validar apenas campos fornecidos ou configuração completa?
   - O que fazer se update deixar configuração inválida?

---

## 4. `cast gateway test`

**O que já está especificado:**
- ✅ Sintaxe: `cast gateway test <provider> [flags]`
- ✅ Flag: `--target <target>` (opcional)
- ✅ Exemplos básicos

**O que precisa ser especificado:**

1. **O que testa:**
   - Conectividade com o servidor/API?
   - Autenticação (token válido, credenciais corretas)?
   - Envio de mensagem de teste real?
   - Apenas validação de configuração?

2. **Comportamento por provider:**
   - **Telegram:** Envia mensagem de teste? Para qual chat_id?
   - **Email:** Testa conexão SMTP? Autenticação? Envia email de teste?
   - **WhatsApp:** Testa API? Envia mensagem de teste?
   - **Google Chat:** Testa webhook? Envia mensagem de teste?

3. **Flag `--target`:**
   - Quando usar? (para providers que precisam de target)
   - O que fazer se `--target` não for fornecido? (usa default_chat_id, etc.)

4. **Saída:**
   - Formato de saída (sucesso/erro detalhado)?
   - Deve mostrar tempo de resposta?
   - Deve mostrar detalhes da conexão?

5. **Mensagens de teste:**
   - Qual mensagem enviar? (fixa ou configurável?)
   - Deve deletar mensagem de teste após enviar?

---

## 5. `cast alias show`

**O que já está especificado:**
- ✅ Sintaxe: `cast alias show <nome>`
- ✅ Saída esperada (formato básico)

**O que precisa ser especificado:**

1. **Formato de saída:**
   - Apenas texto simples ou formatado?
   - Deve incluir informações adicionais? (data de criação, última modificação?)
   - Suporte a `--format json/yaml`?

2. **Comportamento:**
   - O que fazer se alias não existir? (erro ou mensagem amigável?)
   - Deve validar se alias ainda está válido? (provider existe, target válido?)

---

## 6. `cast alias update`

**O que já está especificado:**
- ✅ Sintaxe: `cast alias update <nome> [flags]`
- ✅ Flags: `--provider`, `--target`, `--name`
- ✅ Exemplos básicos

**O que precisa ser especificado:**

1. **Mesclagem:**
   - Se apenas `--target` for fornecido, mantém provider e name?
   - Se apenas `--name` for fornecido, mantém provider e target?
   - Como mesclar campos parciais?

2. **Validação:**
   - Deve validar provider antes de atualizar?
   - Deve validar target antes de atualizar?
   - O que fazer se validação falhar?

3. **Comportamento:**
   - Deve mostrar alias atual antes de atualizar?
   - Deve pedir confirmação?
   - Mensagem de sucesso deve mostrar o que mudou?

---

## 7. Flag `--source` no `cast config show`

**O que já está especificado:**
- ✅ Flag: `--source` - "Mostra origem (ENV ou File)"
- ✅ Exemplo básico

**O que precisa ser especificado:**

1. **Formato de saída:**
   - Como mostrar origem? (prefixo em cada campo? Tabela separada?)
   - Exemplo de saída esperada:
     ```yaml
     telegram:
       token: "*****"  # source: ENV
       default_chat_id: "123456789"  # source: cast.yaml
     ```

2. **Granularidade:**
   - Mostra origem por campo ou por seção (gateway)?
   - O que fazer se campo vier de múltiplas fontes? (ENV tem precedência)

3. **Comportamento:**
   - Deve mostrar apenas campos configurados ou todos?
   - Como mostrar campos com valores padrão?

---

## 8. Wizard para WhatsApp e Google Chat

**O que já está especificado:**
- ✅ Estrutura de configuração (04_GATEWAY_CONFIG_SPEC.md)
- ✅ Campos obrigatórios e opcionais

**O que precisa ser especificado:**

1. **WhatsApp - Campos do Wizard:**
   - Ordem das perguntas?
   - Validações específicas:
     - Phone Number ID: formato esperado? (numérico, tamanho?)
     - Access Token: formato esperado? (prefixo EAA?)
     - Business Account ID: obrigatório ou opcional? (depende de Sandbox vs Produção?)

2. **Google Chat - Campos do Wizard:**
   - Ordem das perguntas?
   - Validações específicas:
     - Webhook URL: formato esperado? (deve começar com https://chat.googleapis.com?)
     - Como validar se URL é válida antes de salvar?

3. **Fluxo do Wizard:**
   - Deve perguntar sobre Sandbox vs Produção para WhatsApp?
   - Deve testar conexão após configurar? (opcional?)

4. **Mensagens de ajuda:**
   - Onde obter Phone Number ID?
   - Onde obter Access Token?
   - Como criar webhook do Google Chat?

---

## 9. Testes Unitários Completos

**O que já está especificado:**
- ✅ Estrutura básica de testes (manager_test.go existe)

**O que precisa ser especificado:**

1. **Cobertura esperada:**
   - Quais funções devem ter testes?
   - Qual nível de cobertura é esperado? (80%, 90%, 100%?)

2. **Casos de teste específicos:**
   - Testes de edge cases (arquivo corrompido, permissões, etc.)?
   - Testes de validação de inputs do wizard?
   - Testes de merge em import?

3. **Mocks:**
   - Deve mockar operações de arquivo?
   - Deve mockar Viper?

---

## 10. Melhorias na Formatação de Tabelas

**O que já está especificado:**
- ✅ Formato esperado (tabelas ASCII)

**O que precisa ser especificado:**

1. **Biblioteca:**
   - Continuar sem tablewriter ou implementar solução própria?
   - Se usar tablewriter, qual versão/API usar?

2. **Formato:**
   - Bordas (sim/não)?
   - Alinhamento de colunas?
   - Cores nas tabelas?

3. **Prioridade:**
   - É crítico ou pode ser melhorado depois?

---

## RESUMO DAS ESPECIFICAÇÕES NECESSÁRIAS

### Alta Prioridade (bloqueiam implementação):
1. ✅ Comportamento de `--merge` no import
2. ✅ Diferença entre `add` e `update` no gateway
3. ✅ O que `cast gateway test` deve testar exatamente
4. ✅ Comportamento de `cast config reload`

### Média Prioridade (melhoram UX):
5. ✅ Formato de saída do `--source` no config show
6. ✅ Validações específicas do wizard WhatsApp/Google Chat
7. ✅ Comportamento de mesclagem no alias update

### Baixa Prioridade (pode ser melhorado depois):
8. ⚠️ Formatação de tabelas (funciona sem, mas pode melhorar)
9. ⚠️ Testes unitários completos (já tem básicos)

---

**Próximos Passos:**
1. Arquiteto revisa este documento
2. Arquiteto preenche as especificações faltantes
3. Implementação das funcionalidades conforme especificações

---

**Última atualização:** 2025-01-XX
**Versão:** 1.0
**Status:** 🟡 Aguardando especificações
