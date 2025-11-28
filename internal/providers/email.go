package providers

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/eduardoalcantara/cast/internal/config"
)

// emailProvider implementa o Provider para Email (SMTP).
type emailProvider struct {
	config *config.EmailConfig
}

// NewEmailProvider cria uma nova instância do EmailProvider.
func NewEmailProvider(cfg *config.EmailConfig) Provider {
	return &emailProvider{
		config: cfg,
	}
}

// Name retorna o nome do provider.
func (p *emailProvider) Name() string {
	return "email"
}

// Send envia uma mensagem via Email (SMTP).
func (p *emailProvider) Send(target string, message string) error {
	// Parseia múltiplos targets usando função do config
	targets := config.ParseTargets(target)

	if len(targets) == 0 {
		return fmt.Errorf("nenhum destinatário especificado")
	}

	// Monta o endereço do servidor SMTP
	addr := fmt.Sprintf("%s:%d", p.config.SMTPHost, p.config.SMTPPort)

	// Monta a mensagem MIME básica
	fromEmail := p.config.FromEmail
	if fromEmail == "" {
		if p.config.Username != "" {
			fromEmail = p.config.Username
		} else {
			// Fallback: usa o primeiro destinatário ou um email genérico
			if len(targets) > 0 {
				fromEmail = "noreply@cast.local"
			} else {
				fromEmail = "noreply@cast.local"
			}
		}
	}

	fromName := p.config.FromName
	if fromName == "" {
		fromName = "CAST Notifications"
	}

	// Monta headers
	headers := []string{
		fmt.Sprintf("From: %s <%s>", fromName, fromEmail),
		fmt.Sprintf("To: %s", strings.Join(targets, ", ")),
		"Subject: Notificação CAST",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		message,
	}

	emailBody := strings.Join(headers, "\r\n")

	// Autenticação (apenas se username e password estiverem configurados)
	var auth smtp.Auth
	if p.config.Username != "" && p.config.Password != "" {
		auth = smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.SMTPHost)
	}

	// Envia email
	var err error

	if p.config.UseSSL {
		// SSL (porta 465) - requer conexão TLS direta
		err = p.sendWithSSL(addr, auth, fromEmail, targets, []byte(emailBody))
	} else if p.config.UseTLS {
		// TLS (porta 587) - StartTLS
		err = p.sendWithTLS(addr, auth, fromEmail, targets, []byte(emailBody))
	} else {
		// Sem TLS/SSL (não recomendado, mas suportado) - usado para MailHog
		err = p.sendWithoutAuth(addr, auth, fromEmail, targets, []byte(emailBody))
	}

	if err != nil {
		return fmt.Errorf("erro ao enviar email: %w", err)
	}

	return nil
}

// sendWithSSL envia email usando SSL (porta 465).
func (p *emailProvider) sendWithSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Cria conexão TLS
	tlsConfig := &tls.Config{
		ServerName: p.config.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("erro ao conectar via TLS: %w", err)
	}
	defer conn.Close()

	// Cria cliente SMTP sobre a conexão TLS
	client, err := smtp.NewClient(conn, p.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("erro ao criar cliente SMTP: %w", err)
	}
	defer client.Close()

	// Autentica (apenas se auth não for nil)
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("erro na autenticação: %w", err)
		}
	}

	// Define remetente
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("erro ao definir remetente: %w", err)
	}

	// Define destinatários
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("erro ao definir destinatário %s: %w", recipient, err)
		}
	}

	// Envia dados
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("erro ao iniciar envio de dados: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		writer.Close()
		return fmt.Errorf("erro ao escrever mensagem: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("erro ao fechar writer: %w", err)
	}

	return nil
}

// sendWithTLS envia email usando TLS (porta 587) com StartTLS.
func (p *emailProvider) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Conecta ao servidor SMTP
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("erro ao conectar: %w", err)
	}
	defer client.Close()

	// EHLO
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("erro no EHLO: %w", err)
	}

	// StartTLS
	tlsConfig := &tls.Config{
		ServerName: p.config.SMTPHost,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("erro no StartTLS: %w", err)
	}

	// Autentica (apenas se auth não for nil)
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("erro na autenticação: %w", err)
		}
	}

	// Define remetente
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("erro ao definir remetente: %w", err)
	}

	// Define destinatários
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("erro ao definir destinatário %s: %w", recipient, err)
		}
	}

	// Envia dados
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("erro ao iniciar envio de dados: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		writer.Close()
		return fmt.Errorf("erro ao escrever mensagem: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("erro ao fechar writer: %w", err)
	}

	return nil
}

// sendWithoutAuth envia email sem autenticação (para servidores como MailHog).
func (p *emailProvider) sendWithoutAuth(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Se auth estiver configurado, usa smtp.SendMail (que pode usar auth)
	// Se não, cria conexão manual sem auth
	if auth != nil {
		return smtp.SendMail(addr, auth, from, to, msg)
	}

	// Conecta sem autenticação
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("erro ao conectar: %w", err)
	}
	defer client.Close()

	// EHLO
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("erro no EHLO: %w", err)
	}

	// Define remetente
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("erro ao definir remetente: %w", err)
	}

	// Define destinatários
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("erro ao definir destinatário %s: %w", recipient, err)
		}
	}

	// Envia dados
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("erro ao iniciar envio de dados: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		writer.Close()
		return fmt.Errorf("erro ao escrever mensagem: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("erro ao fechar writer: %w", err)
	}

	return nil
}
