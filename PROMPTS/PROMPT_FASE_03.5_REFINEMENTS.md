# ROLE: Senior Go Engineer (CLI Specialist)
# PROJECT: CAST (Fase 03.5 - Refinement & Gaps)

# OBJECTIVE
Implementar as funcionalidades pendentes de gerenciamento de configuração e refinar os comandos existentes conforme as decisões do arquiteto. O foco é fechar as lacunas deixadas na Fase 03.

# INPUTS
- `specifications/07_ARCHITECT_DECISIONS.md`: Regras de negócio e decisões técnicas (LEITURA OBRIGATÓRIA).
- `specifications/06_PHASE_IMPLEMENTATION_PROTOCOL.md`: Protocolo de entrega.
- Código atual em `cmd/cast/` e `internal/config/`.

# REQUIREMENTS

## 1. Comandos de Configuração (Export/Import)
Implemente em `cmd/cast/config.go` e `internal/config/manager.go`:

* **`cast config export`**:
    * **Padrão:** Imprime YAML no `stdout`.
    * **Flag `--output`:** Salva em arquivo. Deve falhar se o arquivo já existir, a menos que `--force` seja usado.
    * **Flag `--mask`:** (default true) Mascara tokens sensíveis antes de exportar.
    * **Validação:** Valide a configuração (`cfg.Validate()`) antes de exportar.

* **`cast config import`**:
    * **Flag `--merge`:** (default false).
        * Se `false`: Substituição total (carrega arquivo e salva).
        * Se `true`: Merge profundo (atualiza campos existentes, mantém outros).
    * **Backup:** OBRIGATÓRIO. Antes de sobrescrever `cast.yaml`, crie uma cópia `cast.yaml.bak`.
    * **Auto-detecção:** Tente detectar formato pela extensão. Se falhar, assuma YAML.

## 2. Comandos de Gateway (Update/Test)
Implemente em `cmd/cast/gateway.go`:

* **`cast gateway update`**:
    * **Diferença:** `add` falha se existe; `update` falha se NÃO existe.
    * **Patch:** Atualize apenas os campos fornecidos nas flags. Mantenha os outros intactos.
    * **Validação:** Valide o objeto resultante antes de salvar.

* **`cast gateway test`**:
    * **Telegram:** Chame `getMe` na API.
    * **Email:** Conecte ao SMTP, faça Auth e QUIT. Não envie email a menos que `--target` seja fornecido explicitamente.
    * **Feedback:** Imprima latência e status (Verde/Vermelho).

## 3. Comandos de Alias (Refinamento)
Refine `cmd/cast/alias.go`:

* **`cast alias show`**: Implemente formato "Ficha Técnica" (Key-Value vertical).
* **`cast alias update`**: Permita atualização parcial (ex: mudar só o target sem mudar o provider).

## 4. Protocolo e Documentação
* **Renomeie** `PROJECT_STATUS.md` para `PROJECT_CONTEXT.md` e atualize-o com o status atual.
* Atualize o help (`--help`) de todos os comandos novos com exemplos práticos.
* Crie `results/03_5_RESULTS.md` com o log do que foi feito.

# DELIVERABLE
* Código compilável.
* Funcionalidades de Export/Import e Test funcionando.
* Comandos de Update implementados.
* Testes unitários para a lógica de Merge e Backup.
