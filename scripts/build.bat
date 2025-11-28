@echo off
REM Script de build do CAST
REM Compila o projeto e copia o executável para ./run/

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_ROOT=%SCRIPT_DIR%..
set BUILD_DIR=%PROJECT_ROOT%\run
set LOG_DIR=%PROJECT_ROOT%\logs
REM Gera nome do log com timestamp simples
set timestamp=%date:~-4,4%%date:~-7,2%%date:~-10,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set timestamp=%timestamp: =0%
set LOG_FILE=%LOG_DIR%\build_%timestamp%.log
set EXE_NAME=cast.exe
set EXE_PATH=%BUILD_DIR%\%EXE_NAME%

REM Remove espaços do nome do log
set LOG_FILE=%LOG_FILE: =%

echo ======================================== > "%LOG_FILE%"
echo CAST - Build Script >> "%LOG_FILE%"
echo Data/Hora: %date% %time% >> "%LOG_FILE%"
echo ======================================== >> "%LOG_FILE%"
echo. >> "%LOG_FILE%"

echo [INFO] Iniciando build do CAST...
echo [INFO] Iniciando build do CAST... >> "%LOG_FILE%"

REM Verifica se Go está instalado
echo [INFO] Verificando instalacao do Go... >> "%LOG_FILE%"
go version >> "%LOG_FILE%" 2>&1
if errorlevel 1 (
    echo [ERRO] Go nao encontrado! Instale o Go 1.22+ primeiro.
    echo [ERRO] Go nao encontrado! Instale o Go 1.22+ primeiro. >> "%LOG_FILE%"
    exit /b 1
)

REM Cria diretórios se não existirem
if not exist "%BUILD_DIR%" (
    echo [INFO] Criando diretorio: %BUILD_DIR% >> "%LOG_FILE%"
    mkdir "%BUILD_DIR%"
)

if not exist "%LOG_DIR%" (
    echo [INFO] Criando diretorio: %LOG_DIR% >> "%LOG_FILE%"
    mkdir "%LOG_DIR%"
)

REM Navega para o diretório do projeto
cd /d "%PROJECT_ROOT%"
echo [INFO] Diretorio de trabalho: %CD% >> "%LOG_FILE%"

REM Limpa build anterior
if exist "%EXE_PATH%" (
    echo [INFO] Removendo executavel anterior... >> "%LOG_FILE%"
    del /f "%EXE_PATH%" >> "%LOG_FILE%" 2>&1
)

REM Executa go mod tidy
echo [INFO] Executando go mod tidy... >> "%LOG_FILE%"
go mod tidy >> "%LOG_FILE%" 2>&1
if errorlevel 1 (
    echo [ERRO] Falha ao executar go mod tidy
    echo [ERRO] Falha ao executar go mod tidy >> "%LOG_FILE%"
    exit /b 1
)

REM Compila o projeto
echo [INFO] Compilando projeto... >> "%LOG_FILE%"
go build -v -o "%EXE_PATH%" ./cmd/cast >> "%LOG_FILE%" 2>&1
if errorlevel 1 (
    echo [ERRO] Falha na compilacao!
    echo [ERRO] Falha na compilacao! >> "%LOG_FILE%"
    echo. >> "%LOG_FILE%"
    echo Detalhes do erro: >> "%LOG_FILE%"
    type "%LOG_FILE%" | findstr /i "error" >> "%LOG_FILE%"
    exit /b 1
)

REM Verifica se o executável foi criado
if not exist "%EXE_PATH%" (
    echo [ERRO] Executavel nao foi criado!
    echo [ERRO] Executavel nao foi criado! >> "%LOG_FILE%"
    exit /b 1
)

REM Obtém informações do executável
echo [INFO] Executavel criado com sucesso! >> "%LOG_FILE%"
echo [INFO] Caminho: %EXE_PATH% >> "%LOG_FILE%"
for %%A in ("%EXE_PATH%") do (
    echo [INFO] Tamanho: %%~zA bytes >> "%LOG_FILE%"
    echo [INFO] Data: %%~tA >> "%LOG_FILE%"
)

REM Testa o executável
echo [INFO] Testando executavel... >> "%LOG_FILE%"
"%EXE_PATH%" --help >> "%LOG_FILE%" 2>&1
if errorlevel 1 (
    echo [AVISO] Executavel pode ter problemas (exit code: %errorlevel%) >> "%LOG_FILE%"
) else (
    echo [INFO] Executavel testado com sucesso! >> "%LOG_FILE%"
)

echo. >> "%LOG_FILE%"
echo ======================================== >> "%LOG_FILE%"
echo Build concluido com sucesso! >> "%LOG_FILE%"
echo Executavel: %EXE_PATH% >> "%LOG_FILE%"
echo Log: %LOG_FILE% >> "%LOG_FILE%"
echo ======================================== >> "%LOG_FILE%"

echo.
echo [SUCESSO] Build concluido!
echo [INFO] Executavel: %EXE_PATH%
echo [INFO] Log: %LOG_FILE%
echo.

endlocal
