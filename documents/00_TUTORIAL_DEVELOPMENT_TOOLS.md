# TUTORIAL: Ferramentas de Desenvolvimento - CAST

Este tutorial guia você passo a passo para configurar o ambiente de desenvolvimento completo do CAST, incluindo todas as ferramentas necessárias para desenvolvimento e build.

## 📋 PRÉ-REQUISITOS

- Acesso à internet para download das ferramentas
- Permissões de administrador (para instalação)
- ~500 MB de espaço em disco

## 🎯 FERRAMENTAS NECESSÁRIAS

### Essenciais
- **Go** (versão 1.22 ou superior) - Linguagem de programação
- **Git** - Controle de versão
- **Editor de Código** - VS Code (recomendado) ou outro

### Recomendadas
- **goimports** - Formatação automática de imports
- **golangci-lint** - Linter para Go
- **gopls** - Language Server Protocol para Go

---

## 🪟 WINDOWS

### 1. Instalar Go

#### 1.1 Download
1. Acesse: https://go.dev/dl/
2. Baixe o instalador para Windows (ex: `go1.22.x.windows-amd64.msi`)
3. Execute o arquivo `.msi` baixado

#### 1.2 Instalação
1. Siga o assistente de instalação
2. A instalação padrão é em `C:\Program Files\Go`
3. O instalador configura automaticamente as variáveis de ambiente

#### 1.3 Verificar Instalação
Abra um novo **Prompt de Comando** (cmd) e execute:

```cmd
go version
```

Você deve ver algo como:
```
go version go1.22.x windows/amd64
```

#### 1.4 Configurar Variáveis de Ambiente (se necessário)
Se o comando `go` não funcionar:

1. Abra **Configurações do Sistema** → **Variáveis de Ambiente**
2. Verifique se `C:\Program Files\Go\bin` está em `PATH`
3. Se não estiver, adicione manualmente
4. Reinicie o terminal

#### 1.5 Configurar GOPATH e GOROOT
Normalmente não é necessário, mas se precisar:

```cmd
setx GOPATH "%USERPROFILE%\go"
setx GOROOT "C:\Program Files\Go"
```

Reinicie o terminal após configurar.

---

### 2. Instalar Git

#### 2.1 Download
1. Acesse: https://git-scm.com/download/win
2. Baixe o instalador (ex: `Git-2.x.x-64-bit.exe`)
3. Execute o instalador

#### 2.2 Instalação
1. Siga o assistente de instalação
2. Recomendações de configuração:
   - **Editor**: Escolha seu editor preferido (VS Code, Notepad++, etc.)
   - **Line ending**: "Checkout Windows-style, commit Unix-style line endings"
   - **Terminal**: "Use Windows' default console window"
   - **Git Credential Manager**: Deixe marcado

#### 2.3 Verificar Instalação
Abra um novo terminal e execute:

```cmd
git --version
```

Você deve ver algo como:
```
git version 2.x.x.windows.1
```

#### 2.4 Configurar Git (primeira vez)
```cmd
git config --global user.name "Seu Nome"
git config --global user.email "seu.email@example.com"
```

---

### 3. Instalar VS Code

#### 3.1 Download
1. Acesse: https://code.visualstudio.com/
2. Baixe o instalador para Windows
3. Execute o instalador

#### 3.2 Instalação
1. Siga o assistente de instalação
2. Recomendações:
   - Marque "Adicionar ao PATH"
   - Marque "Criar associação de arquivo .code"
   - Marque "Adicionar ação 'Abrir com Code' ao menu de contexto do Windows Explorer"

#### 3.3 Instalar Extensão Go
1. Abra o VS Code
2. Pressione `Ctrl+Shift+X` para abrir a aba de Extensões
3. Busque por "Go" (publicado por Go Team at Google)
4. Clique em **Instalar**
5. Aguarde a instalação

#### 3.4 Configuração Automática
A extensão Go instalará automaticamente:
- `gopls` (Language Server)
- `goimports` (formatação)
- Outras ferramentas necessárias

Aguarde a conclusão (barra de progresso no canto inferior direito).

---

### 4. Instalar Ferramentas Go Adicionais

#### 4.1 goimports
```cmd
go install golang.org/x/tools/cmd/goimports@latest
```

#### 4.2 golangci-lint
**Opção 1: Via Chocolatey (recomendado)**
```cmd
choco install golangci-lint
```

**Opção 2: Download Manual**
1. Acesse: https://golangci-lint.run/usage/install/#windows
2. Baixe o binário para Windows
3. Extraia e adicione ao PATH

**Opção 3: Via Go Install**
```cmd
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

#### 4.3 Verificar Instalação
```cmd
goimports --version
golangci-lint --version
```

---

### 5. Configurar VS Code para o Projeto CAST

#### 5.1 Abrir o Projeto
1. Abra o VS Code
2. File → Open Folder
3. Selecione a pasta do projeto CAST

#### 5.2 Verificar Configurações
O projeto já possui `.vscode/settings.json` configurado. As configurações incluem:

- **Go Language Server**: Habilitado
- **Formatação**: `goimports` no save
- **Linting**: `golangci-lint` no save
- **Terminal padrão**: Command Prompt
- **Build on Save**: Habilitado

#### 5.3 Verificar se Está Funcionando
1. Abra qualquer arquivo `.go` do projeto
2. Faça uma pequena alteração
3. Salve o arquivo (`Ctrl+S`)
4. O VS Code deve formatar automaticamente

Se houver erros, verifique:
- Extensão Go instalada
- `gopls` instalado (verifique na aba "Output" → "gopls")

---

### 6. Testar o Ambiente

#### 6.1 Clonar/Baixar o Projeto
Se ainda não tiver o projeto:

```cmd
git clone <url-do-repositorio>
cd cast
```

#### 6.2 Baixar Dependências
```cmd
go mod download
go mod tidy
```

#### 6.3 Compilar o Projeto
```cmd
go build -o run\cast.exe ./cmd/cast
```

Ou use o script de build:

```cmd
scripts\build.bat
```

#### 6.4 Executar Testes
```cmd
go test ./...
```

#### 6.5 Executar o CAST
```cmd
run\cast.exe --help
```

Se tudo funcionar, você verá o banner e o help do CAST.

---

### 7. Estrutura de Diretórios

Após o primeiro build, você terá:

```
cast/
├── cmd/cast/          # Código fonte principal
├── internal/          # Código interno
├── run/               # Executável compilado (cast.exe)
├── logs/              # Logs de build
├── scripts/           # Scripts de build
└── .vscode/           # Configurações do VS Code
```

---

## 🐧 LINUX

### 1. Instalar Go

#### 1.1 Via Gerenciador de Pacotes (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install golang-go
```

**Nota:** A versão do repositório pode ser antiga. Para Go 1.22+, use a instalação manual.

#### 1.2 Instalação Manual (Recomendado)

**Download:**
```bash
cd /tmp
wget https://go.dev/dl/go1.22.x.linux-amd64.tar.gz
```

**Instalar:**
```bash
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.x.linux-amd64.tar.gz
```

**Configurar PATH:**
Adicione ao `~/.bashrc` ou `~/.zshrc`:

```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

**Aplicar:**
```bash
source ~/.bashrc
# ou
source ~/.zshrc
```

#### 1.3 Verificar Instalação
```bash
go version
```

Você deve ver:
```
go version go1.22.x linux/amd64
```

---

### 2. Instalar Git

#### 2.1 Ubuntu/Debian
```bash
sudo apt update
sudo apt install git
```

#### 2.2 Fedora/RHEL
```bash
sudo dnf install git
```

#### 2.3 Arch Linux
```bash
sudo pacman -S git
```

#### 2.4 Verificar Instalação
```bash
git --version
```

#### 2.5 Configurar Git (primeira vez)
```bash
git config --global user.name "Seu Nome"
git config --global user.email "seu.email@example.com"
```

---

### 3. Instalar VS Code

#### 3.1 Ubuntu/Debian (via Snap)
```bash
sudo snap install --classic code
```

#### 3.2 Ubuntu/Debian (via .deb)
1. Acesse: https://code.visualstudio.com/
2. Baixe o `.deb` para Linux
3. Instale:
```bash
sudo dpkg -i code_*.deb
sudo apt-get install -f  # Instala dependências faltantes
```

#### 3.3 Fedora/RHEL
```bash
sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc
sudo sh -c 'echo -e "[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc" > /etc/yum.repos.d/vscode.repo'
sudo dnf install code
```

#### 3.4 Arch Linux
```bash
sudo pacman -S code
```

#### 3.5 Instalar Extensão Go
1. Abra o VS Code
2. Pressione `Ctrl+Shift+X`
3. Busque por "Go" (publicado por Go Team at Google)
4. Clique em **Instalar**

---

### 4. Instalar Ferramentas Go Adicionais

#### 4.1 goimports
```bash
go install golang.org/x/tools/cmd/goimports@latest
```

#### 4.2 golangci-lint

**Opção 1: Via Script (Recomendado)**
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

**Opção 2: Via Go Install**
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Opção 3: Via Gerenciador de Pacotes**
```bash
# Ubuntu/Debian
sudo apt install golangci-lint

# Fedora
sudo dnf install golangci-lint

# Arch (AUR)
yay -S golangci-lint-bin
```

#### 4.3 Verificar Instalação
```bash
goimports --version
golangci-lint --version
```

---

### 5. Configurar VS Code para o Projeto CAST

#### 5.1 Abrir o Projeto
```bash
code /caminho/para/cast
```

Ou:
1. Abra o VS Code
2. File → Open Folder
3. Selecione a pasta do projeto CAST

#### 5.2 Verificar Configurações
O projeto já possui `.vscode/settings.json` configurado.

#### 5.3 Verificar se Está Funcionando
1. Abra qualquer arquivo `.go`
2. Faça uma alteração
3. Salve (`Ctrl+S`)
4. Deve formatar automaticamente

---

### 6. Criar Script de Build (Opcional)

Crie `scripts/build.sh`:

```bash
#!/bin/bash
# Script de build do CAST para Linux

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/run"
LOG_DIR="$PROJECT_ROOT/logs"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="$LOG_DIR/build_$TIMESTAMP.log"
EXE_NAME="cast"
EXE_PATH="$BUILD_DIR/$EXE_NAME"

mkdir -p "$BUILD_DIR"
mkdir -p "$LOG_DIR"

echo "========================================" | tee -a "$LOG_FILE"
echo "CAST - Build Script" | tee -a "$LOG_FILE"
echo "Data/Hora: $(date)" | tee -a "$LOG_FILE"
echo "========================================" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

echo "[INFO] Iniciando build do CAST..."
echo "[INFO] Verificando instalação do Go..." | tee -a "$LOG_FILE"
go version | tee -a "$LOG_FILE"

cd "$PROJECT_ROOT"

echo "[INFO] Executando go mod tidy..." | tee -a "$LOG_FILE"
go mod tidy | tee -a "$LOG_FILE"

echo "[INFO] Compilando projeto..." | tee -a "$LOG_FILE"
go build -v -o "$EXE_PATH" ./cmd/cast | tee -a "$LOG_FILE"

if [ ! -f "$EXE_PATH" ]; then
    echo "[ERRO] Executável não foi criado!" | tee -a "$LOG_FILE"
    exit 1
fi

chmod +x "$EXE_PATH"

echo "[INFO] Executável criado com sucesso!" | tee -a "$LOG_FILE"
echo "[INFO] Caminho: $EXE_PATH" | tee -a "$LOG_FILE"
echo "[INFO] Tamanho: $(stat -f%z "$EXE_PATH" 2>/dev/null || stat -c%s "$EXE_PATH") bytes" | tee -a "$LOG_FILE"

echo "[INFO] Testando executável..." | tee -a "$LOG_FILE"
"$EXE_PATH" --help | tee -a "$LOG_FILE"

echo "" | tee -a "$LOG_FILE"
echo "========================================" | tee -a "$LOG_FILE"
echo "Build concluído com sucesso!" | tee -a "$LOG_FILE"
echo "Executável: $EXE_PATH" | tee -a "$LOG_FILE"
echo "Log: $LOG_FILE" | tee -a "$LOG_FILE"
echo "========================================" | tee -a "$LOG_FILE"

echo ""
echo "[SUCESSO] Build concluído!"
echo "[INFO] Executável: $EXE_PATH"
echo "[INFO] Log: $LOG_FILE"
```

Torne executável:
```bash
chmod +x scripts/build.sh
```

---

### 7. Testar o Ambiente

#### 7.1 Clonar/Baixar o Projeto
```bash
git clone <url-do-repositorio>
cd cast
```

#### 7.2 Baixar Dependências
```bash
go mod download
go mod tidy
```

#### 7.3 Compilar o Projeto
```bash
go build -o run/cast ./cmd/cast
```

Ou use o script de build:

```bash
./scripts/build.sh
```

#### 7.4 Executar Testes
```bash
go test ./...
```

#### 7.5 Executar o CAST
```bash
./run/cast --help
```

---

## 🔧 VERIFICAÇÃO FINAL DO AMBIENTE

Execute estes comandos para verificar se tudo está configurado:

### Windows
```cmd
go version
git --version
code --version
goimports --version
golangci-lint --version
```

### Linux
```bash
go version
git --version
code --version
goimports --version
golangci-lint --version
```

Todos os comandos devem retornar versões sem erros.

---

## 🐛 SOLUÇÃO DE PROBLEMAS

### Go não encontrado
- **Windows**: Verifique se `C:\Program Files\Go\bin` está no PATH
- **Linux**: Verifique se `/usr/local/go/bin` está no PATH
- Reinicie o terminal após alterar PATH

### gopls não funciona no VS Code
1. Abra Command Palette (`Ctrl+Shift+P`)
2. Digite "Go: Install/Update Tools"
3. Selecione todas as ferramentas
4. Aguarde a instalação

### golangci-lint não encontrado
- Verifique se `$GOPATH/bin` está no PATH
- No Linux, pode ser necessário `~/.local/bin` ou `/usr/local/bin`

### Erros de compilação
1. Execute `go mod tidy`
2. Execute `go mod download`
3. Verifique se Go 1.22+ está instalado

### VS Code não formata automaticamente
1. Verifique se a extensão Go está instalada
2. Verifique se `goimports` está instalado
3. Verifique as configurações em `.vscode/settings.json`

---

## 📚 RECURSOS ADICIONAIS

### Documentação Oficial
- **Go**: https://go.dev/doc/
- **Git**: https://git-scm.com/doc
- **VS Code**: https://code.visualstudio.com/docs
- **golangci-lint**: https://golangci-lint.run/

### Tutoriais do CAST
- [Tutorial Telegram](01_TUTORIAL_TELEGRAM.md)
- [Tutorial WhatsApp](02_TUTORIAL_WHATSAPP.md)
- [Tutorial Email](03_TUTORIAL_EMAIL.md)
- [Tutorial Google Chat](04_TUTORIAL_GOOGLE_CHAT.md)

### Especificações Técnicas
- [Master Plan](../specifications/00_MASTER_PLAN.md)
- [Tech Spec](../specifications/02_TECH_SPEC.md)
- [CLI UX](../specifications/03_CLI_UX.md)

---

## ✅ CHECKLIST DE INSTALAÇÃO

### Windows
- [ ] Go 1.22+ instalado e funcionando
- [ ] Git instalado e configurado
- [ ] VS Code instalado
- [ ] Extensão Go instalada no VS Code
- [ ] goimports instalado
- [ ] golangci-lint instalado
- [ ] Projeto compila sem erros
- [ ] Testes executam com sucesso

### Linux
- [ ] Go 1.22+ instalado e funcionando
- [ ] Git instalado e configurado
- [ ] VS Code instalado
- [ ] Extensão Go instalada no VS Code
- [ ] goimports instalado
- [ ] golangci-lint instalado
- [ ] Script de build criado (opcional)
- [ ] Projeto compila sem erros
- [ ] Testes executam com sucesso

---

**Última atualização:** 2025-01-XX
**Versão:** 1.0
**Autor:** CAST Development Team
