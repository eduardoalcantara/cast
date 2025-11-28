# ESPECIFICAÇÃO: PROTOCOLO DE IMPLEMENTAÇÃO DE FASES

**Objetivo:** Estabelecer regras rígidas para a execução de cada fase de desenvolvimento do CAST, garantindo consistência, documentação atualizada e paridade de recursos entre os provedores.

## 1. PRINCÍPIO DE PARIDADE DE RECURSOS

**Regra de Ouro:** Não quebre a consistência.

Ao implementar uma nova funcionalidade (ex: "Suporte a Anexos" ou "Formatação Rich Text"), ela deve ser implementada para **TODOS** os providers suportados na versão atual, a menos que a API do provedor tecnicamente não suporte tal recurso.

* **Proibido:** Implementar uma feature apenas para Telegram e deixar Email quebrado ou desatualizado na mesma fase.

* **Obrigatório:** Se a Fase X toca na interface `Provider`, todos os arquivos em `internal/providers/*.go` devem ser atualizados e testados para conformidade.

## 2. CICLO DE VIDA DE UMA FASE

Para que uma fase seja considerada **CONCLUÍDA**, o programador deve seguir estritamente este checklist sequencial:

### 2.1 Implementação e Refatoração

* Escrever o código funcional.

* Garantir que não houve regressão em funcionalidades anteriores.

* Manter o padrão de código (Linter/Formatter).

### 2.2 Testes (Critério de Aceite)

* **Unitários:** Criar ou atualizar testes em `*_test.go`.

* **Integração:** Se possível, simular o fluxo completo.

* **A Regra:** A fase só termina quando `go test ./...` retornar `PASS` em todos os pacotes.

### 2.3 Atualização da CLI (Interface)

* **Help Geral:** O comando `cast --help` deve listar novos comandos se houver.

* **Help Específico:** O comando `cast <novo_comando> --help` deve conter:

  * `Short`: Descrição curta.

  * `Long`: Explicação detalhada do comportamento.

  * `Example`: **Pelo menos 3 exemplos práticos** e reais de uso.

## 3. ARTEFATOS DE DOCUMENTAÇÃO (OBRIGATÓRIOS)

Em cada fase, o programador **DEVE** gerar ou atualizar os seguintes arquivos:

### 3.1 Relatório de Resultados (`results/XX_RESULTS.md`)

Um novo arquivo markdown deve ser criado na pasta `results/` com o nome da fase (ex: `03_RESULTS.md`).
**Conteúdo obrigatório:**

* Resumo do que foi feito.

* Lista técnica de arquivos criados/alterados.

* Métricas (linhas de código, número de testes).

* **Log de Testes:** Evidência de que os testes passaram.

* Exemplos de comandos que agora funcionam.

### 3.2 Contexto do Projeto (`PROJECT_CONTEXT.md`)

O arquivo `PROJECT_CONTEXT.md` (anteriormente `PROJECT_STATUS.md`) é a fonte da verdade sobre o estado atual.
**Deve ser atualizado com:**

* Versão atual do software.

* Checkboxes marcados nas fases concluídas.

* Lista atualizada de funcionalidades ativas.

* Estrutura de diretórios atualizada (se houve mudança).

### 3.3 Tutoriais (`documents/`)

Se a fase introduziu um novo conceito (ex: "Como usar o Wizard"), um arquivo `XX_TUTORIAL_NOME.md` deve ser criado ou o README atualizado.

## 4. EXEMPLO DE FLUXO DE ENTREGA

Quando o programador entregar a tarefa, a resposta deve conter explicitamente:

1. **O Código Fonte** (atualizado).

2. **O Arquivo de Resultados** (`XX_RESULTS.md`).

3. **O Contexto Atualizado** (`PROJECT_CONTEXT.md`).

4. **Confirmação de Testes:** "Executei os testes e o resultado foi: \[Log\]".

## 5. SANÇÕES TÉCNICAS

O Pull Request (ou entrega de código) será **REJEITADO** pelo PO/Arquiteto se:

* O `--help` estiver genérico ou sem exemplos.

* Faltar o arquivo de resultados.

* O `PROJECT_CONTEXT.md` estiver desatualizado.

* Um provider estiver "mais esperto" que os outros sem justificativa técnica explicita no relatório.

**Última atualização:** 2025-01-XX
**Versão do documento:** 1.0
**Autor:** Arquiteto CAST
