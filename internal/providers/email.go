package providers

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"net/smtp"
	"path/filepath"
	"strings"

	"github.com/eduardoalcantara/cast/internal/config"
)

// EmailProviderExtended define interface estendida para email com assunto e anexos.
type EmailProviderExtended interface {
	Provider
	SendEmail(target string, message string, subject string, attachments []string) error
}

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

// NewEmailProviderExtended cria uma nova instância do EmailProvider como EmailProviderExtended.
func NewEmailProviderExtended(cfg *config.EmailConfig) EmailProviderExtended {
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
	return p.SendEmail(target, message, "", nil)
}

// SendEmail envia uma mensagem via Email (SMTP) com assunto e anexos opcionais.
func (p *emailProvider) SendEmail(target string, message string, subject string, attachments []string) error {
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

	// Define assunto (usa padrão se não fornecido)
	if subject == "" {
		subject = "Notificação CAST"
	}

	// Monta o corpo do email (com ou sem anexos)
	var emailBody []byte
	var err error

	if len(attachments) > 0 {
		// Email com anexos (multipart/mixed)
		emailBody, err = p.buildMultipartMessage(fromName, fromEmail, targets, subject, message, attachments)
		if err != nil {
			return fmt.Errorf("erro ao montar mensagem com anexos: %w", err)
		}
	} else {
		// Email simples (text/plain)
		headers := []string{
			fmt.Sprintf("From: %s <%s>", fromName, fromEmail),
			fmt.Sprintf("To: %s", strings.Join(targets, ", ")),
			fmt.Sprintf("Subject: %s", subject),
			"MIME-Version: 1.0",
			"Content-Type: text/plain; charset=UTF-8",
			"",
			message,
		}
		emailBody = []byte(strings.Join(headers, "\r\n"))
	}

	// Autenticação (apenas se username e password estiverem configurados)
	var auth smtp.Auth
	if p.config.Username != "" && p.config.Password != "" {
		auth = smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.SMTPHost)
	}

	// Envia email
	if p.config.UseSSL {
		// SSL (porta 465) - requer conexão TLS direta
		err = p.sendWithSSL(addr, auth, fromEmail, targets, emailBody)
	} else if p.config.UseTLS {
		// TLS (porta 587) - StartTLS
		err = p.sendWithTLS(addr, auth, fromEmail, targets, emailBody)
	} else {
		// Sem TLS/SSL (não recomendado, mas suportado) - usado para MailHog
		err = p.sendWithoutAuth(addr, auth, fromEmail, targets, emailBody)
	}

	if err != nil {
		return fmt.Errorf("erro ao enviar email: %w", err)
	}

	return nil
}

// buildMultipartMessage monta uma mensagem MIME multipart com anexos.
func (p *emailProvider) buildMultipartMessage(fromName, fromEmail string, targets []string, subject, message string, attachments []string) ([]byte, error) {
	boundary := "----=_Part_" + fmt.Sprintf("%d", len(attachments))

	var parts []string

	// Headers principais
	parts = append(parts, fmt.Sprintf("From: %s <%s>", fromName, fromEmail))
	parts = append(parts, fmt.Sprintf("To: %s", strings.Join(targets, ", ")))
	parts = append(parts, fmt.Sprintf("Subject: %s", subject))
	parts = append(parts, "MIME-Version: 1.0")
	parts = append(parts, fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"", boundary))
	parts = append(parts, "")

	// Corpo da mensagem
	parts = append(parts, fmt.Sprintf("--%s", boundary))
	parts = append(parts, "Content-Type: text/plain; charset=UTF-8")
	parts = append(parts, "Content-Transfer-Encoding: 8bit")
	parts = append(parts, "")
	parts = append(parts, message)
	parts = append(parts, "")

	// Anexos
	for _, attachmentPath := range attachments {
		// Lê o arquivo
		fileData, err := ioutil.ReadFile(attachmentPath)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler arquivo %s: %w", attachmentPath, err)
		}

		// Obtém o nome do arquivo
		fileName := filepath.Base(attachmentPath)

		// Detecta o tipo MIME
		mimeType := mime.TypeByExtension(filepath.Ext(attachmentPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		// Codifica em base64
		encodedData := base64.StdEncoding.EncodeToString(fileData)

		// Adiciona o anexo
		parts = append(parts, fmt.Sprintf("--%s", boundary))
		parts = append(parts, fmt.Sprintf("Content-Type: %s; name=\"%s\"", mimeType, fileName))
		parts = append(parts, "Content-Transfer-Encoding: base64")
		parts = append(parts, fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"", fileName))
		parts = append(parts, "")

		// Divide o conteúdo base64 em linhas de 76 caracteres (padrão MIME)
		for i := 0; i < len(encodedData); i += 76 {
			end := i + 76
			if end > len(encodedData) {
				end = len(encodedData)
			}
			parts = append(parts, encodedData[i:end])
		}
		parts = append(parts, "")
	}

	// Fecha o multipart
	parts = append(parts, fmt.Sprintf("--%s--", boundary))

	return []byte(strings.Join(parts, "\r\n")), nil
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
