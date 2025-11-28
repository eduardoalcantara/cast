package providers

// Provider define o contrato para provedores de envio de mensagens.
type Provider interface {
	// Name retorna o nome do provider (ex: "telegram", "email").
	Name() string

	// Send envia a mensagem para o target especificado.
	// Retorna erro se a operação falhar.
	Send(target string, message string) error
}
