# ROLE: Senior Go Engineer (Implementation Focus)
# PROJECT: CAST (Fase 03.5 - Execution)

# CONTEXTO ATUAL
Você realizou uma análise de gaps excelente no arquivo `07_FASE_03_5_STATUS.md`. O diagnóstico está correto: as funcionalidades de Export, Import, Update e Test estão pendentes.

# OBJECTIVE
**SAIR DO MODO DE ANÁLISE E ENTRAR NO MODO DE CODIFICAÇÃO.**
Sua tarefa única e exclusiva agora é implementar o código para transformar todos os "❌" do arquivo `07_FASE_03_5_STATUS.md` em "✅".

# ACTION PLAN (Checklist de Execução)

Execute esta lista sequencialmente:

## 1. Infraestrutura de Configuração (`internal/config/`)
- [ ] Implementar `MergeConfig(source, dest *Config)` em `manager.go` para suportar o import com merge.
- [ ] Implementar `BackupConfig()` em `manager.go` que copia `cast.yaml` para `cast.yaml.bak`.
- [ ] Implementar `Test()` (dummy ou real) nas structs de configuração se necessário, ou preparar o terreno para o teste de gateway.

## 2. Comandos de Configuração (`cmd/cast/config.go`)
- [ ] Implementar `cast config export`:
  - Usar `yaml.Marshal`.
  - Aplicar máscara se flag `--mask` estiver ativa.
  - Salvar em arquivo se `--output` for informado.
- [ ] Implementar `cast config import`:
  - Fazer backup primeiro.
  - Ler arquivo novo.
  - Se `--merge`, mesclar campos; senão, substituir.
  - Salvar resultado.
- [ ] Implementar `cast config reload`:
  - Apenas releia o arquivo do disco e imprima "Configuração recarregada: OK" ou erro.

## 3. Comandos de Gateway (`cmd/cast/gateway.go`)
- [ ] Implementar `cast gateway update`:
  - Carregar config.
  - Verificar se gateway existe (erro se não existir).
  - Atualizar *apenas* as flags passadas (não zerar as outras).
  - Salvar.
- [ ] Implementar `cast gateway test`:
  - **Telegram:** Fazer um `http.Get` em `api.telegram.org/bot<TOKEN>/getMe`.
  - **Email:** Fazer `smtp.Dial`, `Hello`, `Auth`, `Quit`.
  - Imprimir: "Conectividade OK (XXms)" ou "Erro: ...".

## 4. Comandos de Alias (`cmd/cast/alias.go`)
- [ ] Implementar `cast alias show`: Exibir formatado.
- [ ] Implementar `cast alias update`: Permitir mudar provider ou target sem deletar o alias.

## 5. Documentação Final
- [ ] **RENOMEAR** o arquivo `PROJECT_STATUS.md` para `PROJECT_CONTEXT.md` na raiz.
- [ ] Atualizar o `PROJECT_CONTEXT.md` marcando a Fase 03.5 como concluída.
- [ ] Criar o arquivo `results/03_5_RESULTS.md` listando as implementações.

# DELIVERABLE
**Código Fonte Go funcional.** Não quero mais análises ou planos. Quero o comando `cast config export` funcionando no meu terminal.
