# RELATÓRIO TÉCNICO: Bug na Flag --wfr Sem Valor

**Data:** 2025-12-13
**Severidade:** Alta
**Status:** Não Resolvido
**Responsável pela Análise:** Cursor AI Agent

---

## PROBLEMA REPORTADO

Quando o comando `cast send mail ... --wfr` é executado **sem valor numérico**, o CAST:
1. Envia a mensagem com sucesso
2. **NÃO aguarda resposta via IMAP** (deveria aguardar usando config ou padrão de 30 minutos)
3. **Sai silenciosamente** após o envio

**Comportamento Esperado:**
- Se `--wfr` for usado sem valor, deve usar `wait_for_response_default_minutes` do `cast.yaml`
- Se não houver valor no config, deve usar padrão de 30 minutos
- Deve aguardar resposta via IMAP conforme configurado

---

## ANÁLISE DO CÓDIGO ATUAL

### Localização do Problema
**Arquivo:** `cmd/cast/send.go`
**Linhas:** 216-313 (lógica de detecção de flags) e 376-395 (chamada a WaitForEmailResponse)

### Lógica Implementada (Atual)

```go
// Linha 216-254: Detecção de flags
waitForResponse := -1
flagWasSet := false
flagUsedWithoutValue := false

if cmd.Flags().Changed("wait-for-response") {
    flagWasSet = true
    waitForResponse, _ = cmd.Flags().GetInt("wait-for-response")
    if waitForResponse == -1 {
        flagUsedWithoutValue = true
    }
}
// ... mesma lógica para --wfr

// Linha 265-289: Cálculo de waitMinutes
waitMinutes := waitForResponse
if waitMinutes < 0 {
    if flagUsedWithoutValue {
        if cfg.Email.WaitForResponseDefault > 0 {
            waitMinutes = cfg.Email.WaitForResponseDefault
        } else {
            waitMinutes = 30 // Padrão
        }
    } else {
        waitMinutes = 0 // Desabilitado
    }
}

// Linha 376: Verificação antes de aguardar
if waitMinutes > 0 && (actualProviderName == "email" || actualProviderName == "mail") {
    err = providers.WaitForEmailResponse(...)
}
```

### Problema Identificado

**HIPÓTESE PRINCIPAL:** O método `cmd.Flags().Changed("wfr")` do Cobra **pode não estar retornando `true`** quando a flag é usada sem valor, fazendo com que:
1. `flagWasSet` permaneça `false`
2. `flagUsedWithoutValue` permaneça `false`
3. `waitMinutes` seja calculado como `0` (desabilitado)
4. O bloco `if waitMinutes > 0` nunca seja executado

**EVIDÊNCIA:**
- O comportamento observado (sai silenciosamente) indica que `waitMinutes` está sendo calculado como `0`
- Isso só acontece quando `flagUsedWithoutValue == false` E `cfg.Email.WaitForResponseDefault == 0`

---

## TENTATIVAS DE CORREÇÃO REALIZADAS

### Tentativa 1: Verificação de Valor
- **O que foi feito:** Verificar se `waitForResponse == -1` para detectar flag sem valor
- **Resultado:** Não funcionou
- **Motivo provável:** `Changed()` pode não estar retornando `true` quando flag é usada sem valor

### Tentativa 2: Verificação em Argumentos da Linha de Comando
- **O que foi feito:** Adicionar verificação em `os.Args` para detectar `--wfr` mesmo se `Changed()` falhar
- **Resultado:** Não funcionou (código adicionado mas não testado adequadamente)
- **Problema:** A verificação pode estar incorreta ou o problema está em outro lugar

### Tentativa 3: Logs de Debug
- **O que foi feito:** Adicionar logs verbosos para diagnosticar
- **Resultado:** Não testado (usuário não executou com `--verbose`)
- **Valor:** Poderia ajudar a identificar o problema real

---

## POSSÍVEIS CAUSAS RAÍZ

### 1. Comportamento do Cobra com Flags Int Sem Valor
O Cobra pode ter comportamento inesperado quando uma flag `Int` é usada sem valor:
- `Changed()` pode retornar `false` se o parsing falhar
- `GetInt()` pode retornar valor padrão (`-1`) mas `Changed()` pode não detectar

### 2. Ordem de Processamento
A flag pode estar sendo processada **depois** do envio da mensagem, fazendo com que a lógica de wait não seja executada.

### 3. Configuração do Cast.yaml
Se `wait_for_response_default_minutes: 0` no config, e a flag não for detectada, `waitMinutes` fica `0`.

### 4. Erro Silencioso
Pode haver um erro sendo retornado mas não exibido, fazendo o programa sair antes de aguardar.

---

## DIAGNÓSTICO NECESSÁRIO

Para identificar a causa real, é necessário:

1. **Executar com `--verbose`** para ver os logs de debug:
   ```bash
   cast send mail ... --wfr --verbose
   ```

2. **Verificar o valor de `waitMinutes` calculado:**
   - Adicionar `fmt.Printf("waitMinutes=%d\n", waitMinutes)` antes da linha 376

3. **Verificar se `Changed()` está retornando `true`:**
   - Adicionar log: `fmt.Printf("Changed('wfr')=%v\n", cmd.Flags().Changed("wfr"))`

4. **Verificar o valor retornado por `GetInt()`:**
   - Já existe log em modo verbose, mas precisa ser testado

---

## SOLUÇÕES PROPOSTAS

### Solução 1: Usar Flag Bool + Flag Int Opcional (RECOMENDADO)
Mudar a abordagem:
- `--wfr` vira flag bool (indica que deve aguardar)
- `--wfr-minutes N` vira flag int opcional (tempo específico)
- Se `--wfr` for usado sem `--wfr-minutes`, usa config/padrão

**Vantagem:** Comportamento mais previsível, flags bool sempre funcionam

### Solução 2: Verificar Argumentos Antes do Parsing do Cobra
Antes de processar flags, verificar manualmente em `os.Args`:
```go
wfrFound := false
for i, arg := range os.Args {
    if arg == "--wfr" || arg == "--wait-for-response" {
        // Verifica se próximo argumento é número
        if i+1 >= len(os.Args) || !isNumeric(os.Args[i+1]) {
            wfrFound = true
            flagUsedWithoutValue = true
        }
        break
    }
}
```

### Solução 3: Usar Flag String e Fazer Parsing Manual
Mudar flag para `String` e fazer parsing manual:
```go
sendCmd.Flags().String("wfr", "", "Aguarda resposta (valor em minutos ou vazio para usar config)")
```

---

## RECOMENDAÇÃO TÉCNICA

**Para o Supervisor:**

1. **Investigar comportamento do Cobra:** Verificar documentação/testes do `spf13/cobra` sobre flags `Int` usadas sem valor
2. **Implementar Solução 1:** Mudar para flag bool + flag int opcional (mais robusto)
3. **Adicionar testes unitários:** Testar especificamente o caso de flag sem valor
4. **Melhorar tratamento de erros:** Garantir que erros sejam sempre exibidos, nunca silenciosos

---

## ARQUIVOS MODIFICADOS (Tentativas)

- `cmd/cast/send.go` (linhas 216-395)
  - Múltiplas tentativas de correção
  - Código pode estar com lógica redundante ou incorreta
  - **Recomendação:** Revisar e simplificar

---

## CONCLUSÃO

O problema persiste após múltiplas tentativas. A causa raiz provável é o comportamento do Cobra com flags `Int` sem valor, mas isso não foi confirmado por falta de diagnóstico adequado (logs verbose não foram testados).

**Próximos Passos Recomendados:**
1. Executar diagnóstico com `--verbose`
2. Considerar mudança de arquitetura (Solução 1)
3. Adicionar testes unitários para este caso específico
4. Revisar e simplificar o código atual (pode ter ficado confuso após múltiplas tentativas)

---

**Assinado:** Cursor AI Agent
**Data:** 2025-12-13
