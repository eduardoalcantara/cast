@echo off
REM Script para obter Chat ID do Telegram via API
REM Uso: get-telegram-chat-id.bat <BOT_TOKEN>

if "%~1"=="" (
    echo Uso: get-telegram-chat-id.bat ^<BOT_TOKEN^>
    echo.
    echo Este script obtem o Chat ID de usuarios que iniciaram conversa com o bot.
    echo Envie uma mensagem para o bot antes de executar este script.
    echo.
    echo Exemplo:
    echo   get-telegram-chat-id.bat AAGQz1StBQBlTc2b5...
    exit /b 1
)

set TOKEN=%~1
set URL=https://api.telegram.org/bot%TOKEN%/getUpdates

echo Obtendo atualizacoes do bot...
echo.

REM Usa curl se disponivel, caso contrario usa PowerShell
where curl >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    curl -s "%URL%" | go run scripts\format-chat-id.go
) else (
    powershell -Command "Invoke-RestMethod -Uri '%URL%' | ConvertTo-Json -Depth 10"
)

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Erro ao obter Chat ID.
    echo.
    echo Certifique-se de que:
    echo   1. O token do bot esta correto
    echo   2. Voce enviou uma mensagem para o bot
    echo   3. curl ou PowerShell esta disponivel
    exit /b 1
)
