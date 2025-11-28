package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// getUpdatesResponse representa a resposta da API getUpdates do Telegram.
type getUpdatesResponse struct {
	OK     bool `json:"ok"`
	Result []struct {
		UpdateID int64 `json:"update_id"`
		Message  struct {
			MessageID int64 `json:"message_id"`
			From      struct {
				ID        int64  `json:"id"`
				IsBot     bool   `json:"is_bot"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID        int64  `json:"id"`
				Type      string `json:"type"`
				Title     string `json:"title,omitempty"`
				FirstName string `json:"first_name,omitempty"`
				Username  string `json:"username,omitempty"`
			} `json:"chat"`
			Date int64  `json:"date"`
			Text string `json:"text"`
		} `json:"message,omitempty"`
	} `json:"result"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run get-telegram-chat-id.go <BOT_TOKEN>")
		fmt.Println()
		fmt.Println("Este script obtém o Chat ID de usuários que iniciaram conversa com o bot.")
		fmt.Println("Envie uma mensagem para o bot antes de executar este script.")
		fmt.Println()
		fmt.Println("Exemplo:")
		fmt.Println("  go run scripts/get-telegram-chat-id.go AAGQz1StBQBlTc2b5...")
		os.Exit(1)
	}

	token := strings.TrimSpace(os.Args[1])
	if token == "" {
		fmt.Fprintf(os.Stderr, "Erro: Token não pode estar vazio\n")
		os.Exit(1)
	}

	// URL da API getUpdates
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", token)

	// Faz requisição
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Erro da API (status %d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	// Lê resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var updates getUpdatesResponse
	if err := json.Unmarshal(body, &updates); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao parsear JSON: %v\n", err)
		fmt.Fprintf(os.Stderr, "Resposta: %s\n", string(body))
		os.Exit(1)
	}

	if !updates.OK {
		fmt.Fprintf(os.Stderr, "API retornou ok=false\n")
		os.Exit(1)
	}

	if len(updates.Result) == 0 {
		fmt.Println("Nenhuma atualização encontrada.")
		fmt.Println()
		fmt.Println("Para obter seu Chat ID:")
		fmt.Println("  1. Abra o bot no Telegram")
		fmt.Println("  2. Clique em 'Iniciar' ou envie '/start'")
		fmt.Println("  3. Envie qualquer mensagem para o bot")
		fmt.Println("  4. Execute este script novamente")
		os.Exit(0)
	}

	// Agrupa por chat_id
	chats := make(map[int64]struct {
		ID        int64
		Type      string
		FirstName string
		Username  string
		Title     string
		LastMsg   string
	})

	for _, update := range updates.Result {
		if update.Message.Chat.ID == 0 {
			continue
		}

		chatID := update.Message.Chat.ID
		chat := chats[chatID]
		chat.ID = chatID
		chat.Type = update.Message.Chat.Type
		chat.FirstName = update.Message.Chat.FirstName
		chat.Username = update.Message.Chat.Username
		chat.Title = update.Message.Chat.Title
		chat.LastMsg = update.Message.Text
		chats[chatID] = chat
	}

	// Exibe resultados
	fmt.Println("Chat IDs encontrados:")
	fmt.Println(strings.Repeat("=", 80))

	for chatID, chat := range chats {
		fmt.Printf("\nChat ID: %d\n", chatID)
		fmt.Printf("Tipo:    %s\n", chat.Type)

		if chat.Type == "private" {
			if chat.FirstName != "" {
				fmt.Printf("Nome:    %s\n", chat.FirstName)
			}
			if chat.Username != "" {
				fmt.Printf("Username: @%s\n", chat.Username)
			}
		} else if chat.Type == "group" || chat.Type == "supergroup" {
			if chat.Title != "" {
				fmt.Printf("Título:  %s\n", chat.Title)
			}
		}

		if chat.LastMsg != "" {
			fmt.Printf("Última mensagem: %s\n", chat.LastMsg)
		}

		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
	fmt.Println("Para usar no CAST:")
	fmt.Printf("  cast gateway update telegram --default-chat-id \"%d\"\n", chats[0].ID)
	fmt.Println()
	fmt.Println("Ou configure via variável de ambiente:")
	fmt.Printf("  set CAST_TELEGRAM_DEFAULT_CHAT_ID=%d\n", chats[0].ID)
}
