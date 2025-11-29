# TUTORIAL: Configurando WAHA no CAST

Este tutorial explica como configurar o **WAHA** (WhatsApp HTTP API) como provider do CAST para enviar notificações via WhatsApp pessoal/grupos [web:19][web:23].

## ⚠️ AVISOS IMPORTANTES

- **WAHA não é oficial**: Use por sua conta e risco [web:23]
- **Risco de ban**: Automação não-oficial pode bloquear sua conta WhatsApp
- **Casos de uso ideais**: Notificações para você mesmo ou grupos pequenos/controlados
- **Dependência externa**: WAHA deve rodar separadamente do CAST (Docker recomendado)

## PRÉ-REQUISITOS

- CAST instalado
- Docker instalado (ou Node.js para rodar WAHA nativo)
- WhatsApp pessoal para escanear QR code
- Porta 3000 disponível (ou outra à sua escolha)

---

## PASSO 1: Instalar e Rodar WAHA

### 1.1 Via Docker (Recomendado)

**Opção A - Engine WEBJS (mais estável):**

```
docker run -d \
  --name waha \
  -p 3000:3000 \
  -v waha-data:/app/.sessions \
  devlikeapro/waha
```

**Opção B - Engine NOWEB (mais leve):**

```
docker run -d \
  --name waha \
  -p 3000:3000 \
  -e WHATSAPP_DEFAULT_ENGINE=NOWEB \
  -v waha-data:/app/.sessions \
  devlikeapro/waha
```

**Verificar se está rodando:**

```
curl http://localhost:3000/api/health
# Deve retornar: {"status":"ok"}
```

### 1.2 Via Docker Compose (Produção)

Crie `docker-compose.yml`:

```
version: '3.8'
services:
  waha:
    image: devlikeapro/waha:latest
    container_name: waha
    ports:
      - "3000:3000"
    environment:
      - WHATSAPP_DEFAULT_ENGINE=WEBJS
      - WAHA_LOG_LEVEL=info
      # Opcional: Adicionar autenticação
      # - WHATSAPP_API_KEY=seu-api-key-secreto
    volumes:
      - waha-sessions:/app/.sessions
    restart: unless-stopped

volumes:
  waha-sessions:
```

Iniciar:

```
docker-compose up -d
docker-compose logs -f waha  # Ver logs
```

---

## PASSO 2: Conectar WhatsApp (Escanear QR Code)

### 2.1 Criar Sessão

```
curl -X POST http://localhost:3000/api/sessions/start \
  -H "Content-Type: application/json" \
  -d '{
    "name": "default"
  }'
```

### 2.2 Obter QR Code

**Opção A - Via API:**

```
curl http://localhost:3000/api/default/auth/qr
# Retorna JSON com base64 da imagem
```

**Opção B - Painel Web (mais fácil):**

1. Acesse: `http://localhost:3000`
2. Navegue até "Sessions" → "default"
3. Clique em "Show QR Code"
4. Escaneie com seu WhatsApp:
   - Android: WhatsApp → ⋮ → Dispositivos conectados → Conectar dispositivo
   - iOS: WhatsApp → Configurações → Dispositivos conectados → Conectar dispositivo

### 2.3 Verificar Conexão

```
curl http://localhost:3000/api/sessions/default
# Status deve ser "WORKING"
```

---

## PASSO 3: Configurar WAHA no CAST

### 3.1 Modo Wizard (Interativo)

```
cast gateway add waha --interactive
```

**O wizard perguntará:**
1. URL da API WAHA: `http://localhost:3000`
2. Nome da sessão: `default`
3. API Key (opcional): deixe vazio se não configurou
4. Timeout: `30` segundos

**Saída esperada:**
```
✅ Configuração a ser salva:
  API URL:  http://localhost:3000
  Session:  default
  Timeout:  30 segundos

Confirmar e salvar? (Y/n): Y

✅ Configuração do WAHA salva com sucesso!
⚠️  Lembre-se: WAHA deve estar rodando e com sessão conectada
```

### 3.2 Modo Flags (Direto)

```
cast gateway add waha \
  --api-url http://localhost:3000 \
  --session default \
  --timeout 30
```

### 3.3 Via Arquivo (cast.yaml)

```
waha:
  apiurl: http://localhost:3000
  session: default
  # apikey: seu-key-opcional
  timeout: 30
```

### 3.4 Via Variáveis de Ambiente

```
# Linux/Mac
export CAST_WAHA_APIURL=http://localhost:3000
export CAST_WAHA_SESSION=default
export CAST_WAHA_TIMEOUT=30

# Windows CMD
set CAST_WAHA_APIURL=http://localhost:3000
set CAST_WAHA_SESSION=default

# Windows PowerShell
$env:CAST_WAHA_APIURL="http://localhost:3000"
$env:CAST_WAHA_SESSION="default"
```

---

## PASSO 4: Testar Conectividade

```
cast gateway test waha
```

**Saída esperada (sessão conectada):**
```
🔍 Testando conectividade com WAHA...
✅ Conectividade OK!
   URL:     http://localhost:3000
   Session: default
   Status:  WORKING
```

**Saída esperada (sessão desconectada):**
```
✅ Conectividade OK!
   URL:     http://localhost:3000
   Session: default
   Status:  SCAN_QR_CODE

⚠️  Sessão não está ativa!
   Escaneie o QR code no painel WAHA
```

---

## PASSO 5: Enviar Mensagem de Teste

### 5.1 Descobrir Seu Chat ID

**Para você mesmo:**
1. Envie mensagem para **você mesmo** no WhatsApp Web
2. Veja o ID no painel WAHA ou via API:

```
curl http://localhost:3000/api/default/chats
# Busque seu número no formato: 5511999998888@c.us
```

**Para grupos:**
1. Abra o grupo no WhatsApp Web
2. URL terá formato: `chat/120363XXXXX@g.us`
3. Copie o ID: `120363XXXXX@g.us`

### 5.2 Enviar Mensagem

**Para contato individual:**

```
cast send waha 5511999998888@c.us "🎉 Teste via WAHA funcionou!"
```

**Para grupo:**

```
cast send waha 120363XXXXX@g.us "🤖 Notificação do CAST via WAHA"
```

**Usando alias (mais prático):**

```
# Criar alias
cast alias add meu-zap waha 5511999998888@c.us --name "Meu WhatsApp"
cast alias add team waha 120363XXXXX@g.us --name "Grupo da Equipe"

# Usar alias
cast send waha meu-zap "Mensagem pessoal"
cast send waha team "🚨 Deploy concluído com sucesso!"
```

---

## CASOS DE USO RECOMENDADOS

### 1. Notificações do Cursor/IA

```
# Em scripts de automação
cast send waha meu-zap "✅ Cursor finalizou refatoração"
cast send waha team "📊 Relatório semanal disponível"
```

### 2. Monitoramento de Servidor

```
#!/bin/bash
# health-check.sh
if ! curl -f http://meuapp.com/health; then
  cast send waha meu-zap "🚨 ALERTA: App fora do ar!"
fi
```

### 3. CI/CD

```
# .github/workflows/deploy.yml
- name: Notificar Deploy
  run: |
    cast send waha ${{ secrets.TEAM_CHAT_ID }} \
      "🚀 Deploy v${{ github.ref }} concluído"
```

---

## CONFIGURAÇÕES AVANÇADAS

### Timeout Customizado

```
waha:
  apiurl: http://localhost:3000
  session: default
  timeout: 60  # 60 segundos para conexões lentas
```

### API Key (Segurança)

No Docker:

```
docker run -d \
  -e WHATSAPP_API_KEY=meu-secret-key \
  -p 3000:3000 \
  devlikeapro/waha
```

No CAST:

```
cast gateway add waha \
  --api-url http://localhost:3000 \
  --api-key meu-secret-key
```

### Múltiplas Sessões

```
waha:
  apiurl: http://localhost:3000
  session: pessoal  # Use outra sessão criada no WAHA
  timeout: 30
```

Criar nova sessão:

```
curl -X POST http://localhost:3000/api/sessions/start \
  -H "Content-Type: application/json" \
  -d '{"name": "pessoal"}'
```

---

## SOLUÇÃO DE PROBLEMAS

### Erro: "Session is not connected"

**Causa:** QR code não foi escaneado ou expirou.

**Solução:**
1. Acesse `http://localhost:3000`
2. Vá em Sessions → default
3. Clique em "Logout" e depois "Restart"
4. Escaneie novo QR code

### Erro: "Connection refused"

**Causa:** WAHA não está rodando.

**Solução:**
```
docker ps | grep waha  # Verificar se está rodando
docker start waha      # Iniciar se parou
docker logs waha       # Ver logs de erro
```

### Erro: "Invalid chatId"

**Causa:** Formato do Chat ID incorreto.

**Solução:**
- Contatos devem terminar em `@c.us`
- Grupos devem terminar em `@g.us`
- Exemplo correto: `5511999998888@c.us`

### Timeout ao Enviar

**Causa:** WAHA está processando ou rede lenta.

**Solução:**
```
cast gateway update waha --timeout 60
```

### Mensagem não Chega

**Causa:** Número bloqueou você ou não usa WhatsApp.

**Solução:**
1. Verifique se consegue enviar manualmente no WhatsApp
2. Teste enviando para você mesmo primeiro
3. Veja logs do WAHA: `docker logs waha`

---

## SEGURANÇA

### ⚠️ Nunca Compartilhe

- Token/API Key do WAHA
- URL do WAHA se exposta publicamente
- Não commite credenciais no Git

### ✅ Boas Práticas

```
# Use variáveis de ambiente
export CAST_WAHA_APIURL=$WAHA_URL_SECRET
export CAST_WAHA_APIKEY=$WAHA_KEY_SECRET

# No .gitignore
echo "cast.yaml" >> .gitignore
echo ".env" >> .gitignore
```

### 🔒 Produção

Se WAHA for exposto na internet:

1. **Use HTTPS** (nginx com Let's Encrypt)
2. **Configure API Key** obrigatório
3. **Firewall** limitando IPs
4. **Monitoramento** de acessos

---

## DIFERENÇAS: WAHA vs WhatsApp Cloud API

| Aspecto | WAHA | WhatsApp Cloud (Meta) |
|---------|------|---------------------|
| **Autenticação** | QR Code (pessoal) | Business Account |
| **Limites** | Sem limite oficial | 250-1000/dia (tier) |
| **Custo** | Gratuito | Gratuito até 1000 conversas/mês |
| **Sandbox** | Não precisa | Números pré-verificados |
| **Aprovação** | Não precisa | Requer revisão Meta |
| **Risco** | Possível ban | API oficial |
| **Grupos** | ✅ Suporta | ❌ Não suporta |
| **Status** | Não oficial | Oficial |

**Use WAHA quando:**
- Notificações pessoais ou pequenos grupos
- Desenvolvimento/testes sem burocracia
- Não quer depender de aprovações Meta
- Precisa enviar para grupos

**Use WhatsApp Cloud quando:**
- Produção com alto volume
- Necessita suporte oficial
- Aplicação business crítica

---

## REFERÊNCIAS

- [Documentação Oficial WAHA](https://waha.devlike.pro)
- [GitHub WAHA](https://github.com/devlikeapro/waha)
- [Tutorial WAHA Send Messages](https://waha.devlike.pro/docs/how-to/send-messages/)
- [Especificação CAST - WAHA](../specifications/09_FASE_06_WAHA_IMPLEMENTATION_DEEP_SPECIFICATIONS.md)

---

**Última atualização:** 2025-11-29  
**Versão:** 1.0  
**Autor:** Equipe CAST
