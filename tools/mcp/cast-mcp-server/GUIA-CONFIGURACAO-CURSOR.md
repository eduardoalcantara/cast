# 🚀 Guia de Configuração: cast.exe no Cursor IDE

## ✅ Configuração Concluída!

O arquivo `C:\Users\Eduardo\.cursor\mcp.json` foi configurado com o servidor MCP do cast.exe.

---

## 📋 O que foi configurado

```json
{
  "mcpServers": {
    "cast": {
      "command": "D:\\proj\\ia\\tools\\mcp\\cast-mcp-server\\cast-mcp-server.exe",
      "env": {
        "CAST_PATH": "D:\\proj\\cast\\run\\cast.exe",
        "CAST_DEFAULT_ALIAS": "me"
      }
    }
  }
}
```

**Variáveis de Ambiente:**
- `CAST_PATH`: Caminho para o executável `cast.exe` (obrigatório)
- `CAST_DEFAULT_ALIAS`: Alias padrão a ser usado quando nenhum alias/provider/target for especificado (padrão: "me")

---

## 🔄 Próximos Passos

### 1. Reiniciar o Cursor IDE

**IMPORTANTE**: O Cursor precisa ser reiniciado para carregar a nova configuração MCP.

1. Feche completamente o Cursor IDE
2. Abra novamente
3. A configuração será carregada automaticamente

### 2. Verificar se está funcionando

Após reiniciar, você pode verificar se o servidor MCP está ativo:

1. Abra o Cursor IDE
2. Vá em **Settings** → **Agents** → **Tools & MCP**
3. Você deve ver o servidor `cast` listado
4. As ferramentas `cast_send`, `cast_alias_*`, etc. devem estar disponíveis

### 3. Testar Notificação

Peça ao Cursor para testar:

```
"Envie uma notificação de teste para mim via Telegram"
```

O Cursor deve:
1. Detectar que precisa usar `cast_send`
2. Chamar a ferramenta automaticamente
3. Você receberá a notificação no Telegram

---

## 🎯 Como Funciona

### Fluxo Automático

```
1. Você completa uma tarefa no Cursor
   ↓
2. Cursor detecta: "Tarefa concluída"
   ↓
3. Cursor chama automaticamente: cast_send({
     "alias": "me",
     "message": "✅ Tarefa concluída: [descrição]"
   })
   ↓
4. cast.exe envia notificação via Telegram
   ↓
5. Você recebe no celular! 📱
```

### Exemplos de Uso

**Exemplo 1: Notificação Automática**
```
Você: "Quando terminar a refatoração, me notifique"
Cursor: [completa tarefa] → [chama cast_send automaticamente]
Você: Recebe notificação no Telegram
```

**Exemplo 2: Notificação com Detalhes**
```
Você: "Notifique-me quando o build terminar com sucesso"
Cursor: [build completa] → [chama cast_send com detalhes]
Você: Recebe: "✅ Build concluído com sucesso"
```

**Exemplo 3: Notificação para Múltiplos**
```
Você: "Notifique o time quando o deploy terminar"
Cursor: [deploy completa] → [chama cast_send para múltiplos]
Time: Todos recebem notificação
```

---

## 🔧 Troubleshooting

### Problema: Cursor não detecta o servidor MCP

**Solução 1**: Reinicie o Cursor IDE completamente

**Solução 2**: Verifique se o arquivo está no local correto:
- Windows: `C:\Users\Eduardo\.cursor\mcp.json`
- Ou: `%APPDATA%\Cursor\User\globalStorage\`

**Solução 3**: Verifique os logs do Cursor:
- Abra Developer Tools (Ctrl+Shift+I)
- Veja se há erros relacionados a MCP

### Problema: "cast.exe not found"

**Solução**: Verifique se o caminho está correto:
- `CAST_PATH`: `D:\proj\cast\run\cast.exe`
- O executável deve existir nesse caminho

### Problema: Notificações não são enviadas

**Solução 1**: Teste o cast.exe manualmente:
```bash
cast.exe send me "Teste"
```

**Solução 2**: Verifique se o alias "me" está configurado:
```bash
cast.exe alias show me
```

**Solução 3**: Verifique a configuração do gateway:
```bash
cast.exe gateway test telegram
```

---

## 📝 Notas Importantes

1. **Reinício Obrigatório**: O Cursor precisa ser reiniciado para carregar a configuração MCP
2. **Alias "me"**: Certifique-se de que o alias "me" está configurado no cast.exe
3. **Gateway Configurado**: O gateway do Telegram deve estar configurado e funcionando
4. **Notificações Automáticas**: O Cursor decidirá automaticamente quando usar o cast_send

---

## ✅ Status

- ✅ Servidor MCP criado
- ✅ Executável compilado
- ✅ Arquivo de configuração atualizado
- ⏳ **Próximo passo**: Reiniciar o Cursor IDE

---

**Data**: 2025-12-11
**Versão**: 1.0.0
