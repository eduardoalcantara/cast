package providers

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/fatih/color"

	"github.com/eduardoalcantara/cast/internal/config"
)

var (
	// ErrNoEmailResponse é retornado quando nenhuma resposta é encontrada dentro do prazo.
	ErrNoEmailResponse = errors.New("nenhuma resposta encontrada")
	// ErrIMAPConfigMissing é retornado quando a configuração IMAP não está completa.
	ErrIMAPConfigMissing = errors.New("configuração IMAP incompleta")
	// ErrIMAPAuth é retornado quando há falha na autenticação IMAP.
	ErrIMAPAuth = errors.New("falha na autenticação IMAP")
)

// WaitForEmailResponse aguarda por uma resposta de email via IMAP.
// Retorna nil se uma resposta for encontrada, ou um erro específico caso contrário.
func WaitForEmailResponse(
	cfg config.EmailConfig,
	messageID string,
	subject string,
	waitMinutes int,
	fullLayout bool,
	verbose bool,
) error {
	// Validação de configuração IMAP
	if cfg.IMAPHost == "" || cfg.IMAPPort == 0 || cfg.IMAPUsername == "" || cfg.IMAPPassword == "" {
		return fmt.Errorf("%w: para usar --wait-for-response é necessário configurar email.imap_* no cast.yaml", ErrIMAPConfigMissing)
	}

	// Validação de waitMinutes
	if waitMinutes <= 0 {
		return nil // Não aguarda se waitMinutes <= 0
	}

	if cfg.WaitForResponseMax > 0 && waitMinutes > cfg.WaitForResponseMax {
		return fmt.Errorf("waitMinutes (%d) excede o máximo configurado (%d minutos)", waitMinutes, cfg.WaitForResponseMax)
	}

	// Calcula deadline
	deadline := time.Now().Add(time.Duration(waitMinutes) * time.Minute)
	startTime := time.Now()

	// Intervalo de polling (configurável via imap_poll_interval_seconds, default 5s, max 60s)
	// REDUZIDO para 5s padrão para detecção mais rápida (importante para automação)
	pollInterval := 5 * time.Second
	if cfg.IMAPPollInterval > 0 {
		pollInterval = time.Duration(cfg.IMAPPollInterval) * time.Second
		// Limite máximo de 60 segundos para evitar polling muito lento
		if pollInterval > 60*time.Second {
			pollInterval = 60 * time.Second
		}
		// Limite mínimo de 3 segundos para evitar polling muito rápido (mas permite 3s para automação)
		if pollInterval < 3*time.Second {
			pollInterval = 3 * time.Second
		}
	}

	// Mensagem inicial
	yellow := color.New(color.FgYellow)
	yellow.Printf("⏳ Aguardando resposta por até %d minutos (IMAP: %s:%d, pasta %s)...\n",
		waitMinutes, cfg.IMAPHost, cfg.IMAPPort, cfg.IMAPFolder)

	if verbose {
		cyan := color.New(color.FgCyan)
		cyan.Printf("[DEBUG] Message-ID sendo buscado: %s\n", messageID)
		cyan.Printf("[DEBUG] Subject original: %s\n", subject)
		cyan.Printf("[DEBUG] Intervalo de polling: %v (entre cada ciclo de busca)\n", pollInterval)
	}

	cycle := 0
	for time.Now().Before(deadline) {
		cycle++
		if verbose {
			cyan := color.New(color.FgCyan)
			cyan.Printf("[DEBUG] Ciclo %d: verificando IMAP...\n", cycle)
		}

		// Conecta ao IMAP
		imapClient, err := connectIMAP(cfg, verbose)
		if err != nil {
			if verbose {
				red := color.New(color.FgRed)
				red.Printf("[DEBUG] Erro ao conectar IMAP: %v\n", err)
			}
			// Se for erro de autenticação, retorna erro específico
			if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "LOGIN") {
				return fmt.Errorf("%w: %v", ErrIMAPAuth, err)
			}
			// Outros erros de rede: continua tentando até o deadline
			time.Sleep(pollInterval)
			continue
		}

		// Busca resposta IMEDIATAMENTE (sem sleep antes do primeiro ciclo)
		// Sleep apenas ENTRE ciclos, não antes do primeiro
		// Fallback por Subject: após 1 ciclo (otimizado para automação rápida)
		// Com polling de 5s, 1 ciclo = ~5s (tempo mínimo para resposta chegar e ser indexada)
		// Se In-Reply-To/References não funcionarem, tenta Subject no próximo ciclo
		useSubjectFallback := cycle >= 1
		// Usa fullLayout da configuração se não foi especificado via flag
		fullLayoutToUse := fullLayout || cfg.WaitForResponseFullLayout
		found, response, err := searchEmailResponse(imapClient, cfg.IMAPFolder, messageID, subject, useSubjectFallback, fullLayoutToUse, verbose)
		if err != nil {
			imapClient.Logout()
			if verbose {
				red := color.New(color.FgRed)
				red.Printf("[DEBUG] Erro na busca: %v\n", err)
			}
			time.Sleep(pollInterval)
			continue
		}

		// Fecha conexão
		imapClient.Logout()

		if found {
			// Resposta encontrada! Retorna IMEDIATAMENTE (sem sleep)
			elapsed := time.Since(startTime)
			green := color.New(color.FgGreen, color.Bold)
			green.Printf("✓ Resposta recebida em %s\n", formatDuration(elapsed))

			// Exibe resposta
			printEmailResponse(response, cfg.WaitForResponseMaxLines, verbose)
			return nil
		}

		// Calcula tempo restante
		remaining := time.Until(deadline)
		if remaining > pollInterval {
			if verbose {
				cyan := color.New(color.FgCyan)
				cyan.Printf("[DEBUG] Ciclo %d: 0 respostas encontradas, aguardando %v antes da próxima verificação...\n", cycle, pollInterval)
			}
			time.Sleep(pollInterval)
		} else if remaining > 0 {
			// Última tentativa antes do deadline
			if verbose {
				cyan := color.New(color.FgCyan)
				cyan.Printf("[DEBUG] Última tentativa antes do deadline, aguardando %v...\n", remaining)
			}
			time.Sleep(remaining)
		}
	}

	// Timeout: nenhuma resposta encontrada
	yellow.Printf("\n⏰ Tempo de espera esgotado (%d minutos).\n", waitMinutes)
	red := color.New(color.FgRed, color.Bold)
	red.Printf("✗ O destinatário não respondeu à mensagem.\n")
	return ErrNoEmailResponse
}

// connectIMAP conecta ao servidor IMAP e autentica.
func connectIMAP(cfg config.EmailConfig, verbose bool) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.IMAPHost, cfg.IMAPPort)

	if verbose {
		cyan := color.New(color.FgCyan)
		encryption := "sem criptografia"
		if cfg.IMAPUseSSL {
			encryption = "SSL"
		} else if cfg.IMAPUseTLS {
			encryption = "TLS"
		}
		cyan.Printf("[DEBUG] Conectando ao IMAP %s (%s)\n", addr, encryption)
	}

	var c *client.Client
	var err error

	if cfg.IMAPUseSSL {
		// SSL (porta 993)
		c, err = client.DialTLS(addr, nil)
	} else {
		// Conexão sem SSL inicial
		c, err = client.Dial(addr)
		if err == nil && cfg.IMAPUseTLS {
			// StartTLS
			err = c.StartTLS(nil)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao conectar IMAP: %w", err)
	}

	// Timeout
	if cfg.IMAPTimeout > 0 {
		c.Timeout = time.Duration(cfg.IMAPTimeout) * time.Second
	}

	// Autenticação
	if err := c.Login(cfg.IMAPUsername, cfg.IMAPPassword); err != nil {
		c.Logout()
		return nil, fmt.Errorf("falha na autenticação IMAP: %w", err)
	}

	if verbose {
		cyan := color.New(color.FgCyan)
		cyan.Printf("[DEBUG] Autenticado com sucesso\n")
	}

	return c, nil
}

// searchEmailResponse busca uma resposta de email usando Message-ID ou Subject como fallback.
// messageID é o Message-ID da mensagem enviada que estamos aguardando resposta.
// useSubjectFallback: se true, tenta fallback por Subject (só após alguns ciclos)
// fullLayout: se true, inclui HTML; se false, apenas texto
func searchEmailResponse(
	c *client.Client,
	folder string,
	messageID string,
	subject string,
	useSubjectFallback bool,
	fullLayout bool,
	verbose bool,
) (bool, *EmailResponse, error) {
	// Seleciona pasta
	mbox, err := c.Select(folder, false)
	if err != nil {
		return false, nil, fmt.Errorf("erro ao selecionar pasta %s: %w", folder, err)
	}

	if verbose {
		cyan := color.New(color.FgCyan)
		cyan.Printf("[DEBUG] Pasta selecionada: %s (%d mensagens)\n", folder, mbox.Messages)
	}

	// OTIMIZAÇÃO: Busca apenas mensagens RECENTES (últimas 2 minutos) para acelerar
	// Isso reduz drasticamente o tempo de busca quando há muitas mensagens na pasta
	// 2 minutos é suficiente pois estamos buscando resposta logo após envio
	recentTime := time.Now().Add(-2 * time.Minute)
	criteria := imap.NewSearchCriteria()
	criteria.Since = recentTime

	// Busca primária: In-Reply-To (apenas em mensagens recentes)
	messageIDClean := strings.Trim(messageID, "<>")
	criteria.Header.Add("In-Reply-To", messageID)
	if messageIDClean != messageID {
		criteria.Header.Add("In-Reply-To", messageIDClean)
	}
	uids, err := c.Search(criteria)
	if err != nil {
		if verbose {
			cyan := color.New(color.FgCyan)
			cyan.Printf("[DEBUG] Erro na busca In-Reply-To: %v\n", err)
		}
		} else if len(uids) > 0 {
			if verbose {
				cyan := color.New(color.FgCyan)
				cyan.Printf("[DEBUG] SEARCH HEADER In-Reply-To encontrou %d mensagem(ns): %v\n", len(uids), uids)
			}
			return fetchLatestMessage(c, uids, fullLayout, verbose)
		} else if verbose {
			cyan := color.New(color.FgCyan)
			cyan.Printf("[DEBUG] SEARCH HEADER In-Reply-To não encontrou mensagens (Message-ID: %s)\n", messageID)
		}

	if verbose {
		cyan := color.New(color.FgCyan)
		cyan.Printf("[DEBUG] Nenhuma mensagem correspondente encontrada, tentando References...\n")
	}

	// Busca secundária: References (apenas em mensagens recentes)
	criteria = imap.NewSearchCriteria()
	criteria.Since = recentTime
	criteria.Header.Add("References", messageID)
	if messageIDClean != messageID {
		criteria.Header.Add("References", messageIDClean)
	}
	uids, err = c.Search(criteria)
	if err != nil {
		if verbose {
			cyan := color.New(color.FgCyan)
			cyan.Printf("[DEBUG] Erro na busca References: %v\n", err)
		}
		} else if len(uids) > 0 {
			if verbose {
				cyan := color.New(color.FgCyan)
				cyan.Printf("[DEBUG] SEARCH HEADER References encontrou %d mensagem(ns): %v\n", len(uids), uids)
			}
			return fetchLatestMessage(c, uids, fullLayout, verbose)
		} else if verbose {
			cyan := color.New(color.FgCyan)
			cyan.Printf("[DEBUG] SEARCH HEADER References não encontrou mensagens\n")
		}

	// Fallback: Subject (Re: <subject>)
	// IMPORTANTE: Só usa fallback após alguns ciclos (dá tempo para o destinatário responder)
	// E quando usar, valida que a mensagem encontrada realmente responde ao Message-ID correto
	if useSubjectFallback && subject != "" {
		if verbose {
			cyan := color.New(color.FgCyan)
			cyan.Printf("[DEBUG] Nenhuma mensagem correspondente por Message-ID, tentando fallback por Subject...\n")
		}
		criteria = imap.NewSearchCriteria()
		// OTIMIZAÇÃO: Busca apenas mensagens RECENTES (últimas 2 minutos) também no fallback
		criteria.Since = recentTime
		// Tenta variações de "Re: <subject>"
		reVariations := []string{
			fmt.Sprintf("Re: %s", subject),
			fmt.Sprintf("RE: %s", subject),
			fmt.Sprintf("re: %s", subject),
			subject, // Também tenta o subject original
		}
		for _, reSubject := range reVariations {
			criteria.Header.Add("Subject", reSubject)
		}
		uids, err = c.Search(criteria)
		if err == nil && len(uids) > 0 {
			if verbose {
				cyan := color.New(color.FgCyan)
				cyan.Printf("[DEBUG] SEARCH HEADER Subject encontrou %d mensagem(ns), validando InReplyTo...\n", len(uids))
			}
			// Valida que a mensagem encontrada realmente responde ao Message-ID correto
			return fetchAndValidateMessage(c, uids, messageID, messageIDClean, fullLayout, verbose)
		}
	} else if !useSubjectFallback && verbose {
		cyan := color.New(color.FgCyan)
		cyan.Printf("[DEBUG] Fallback por Subject desabilitado (aguardando mais ciclos para dar tempo de resposta)...\n")
	}

	return false, nil, nil
}

// EmailResponse representa uma resposta de email encontrada.
type EmailResponse struct {
	From    string
	Date    time.Time
	Subject string
	Body    string
}

// fetchAndValidateMessage busca a mensagem mais recente e valida que ela responde ao Message-ID correto.
// Usado no fallback por Subject para evitar pegar mensagens antigas.
func fetchAndValidateMessage(c *client.Client, uids []uint32, messageID string, messageIDClean string, fullLayout bool, verbose bool) (bool, *EmailResponse, error) {
	if len(uids) == 0 {
		return false, nil, nil
	}

	// Ordena UIDs (do mais antigo para o mais recente)
	// Pega a mensagem mais recente (última da lista)
	latestUID := uids[len(uids)-1]

	if verbose {
		cyan := color.New(color.FgCyan)
		cyan.Printf("[DEBUG] Validando mensagem UID=%d contra Message-ID: %s\n", latestUID, messageID)
	}

	// Fetch apenas do Envelope primeiro para validar InReplyTo
	seqset := new(imap.SeqSet)
	seqset.AddNum(latestUID)
	items := []imap.FetchItem{imap.FetchEnvelope}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	msg := <-messages
	if msg == nil {
		<-done
		return false, nil, fmt.Errorf("mensagem não encontrada para validação")
	}

	<-done

	// Valida InReplyTo
	if msg.Envelope != nil && msg.Envelope.InReplyTo != "" {
		inReplyTo := msg.Envelope.InReplyTo
		// Remove < > para comparação
		inReplyToClean := strings.Trim(inReplyTo, "<>")
		messageIDCleanForCompare := strings.Trim(messageID, "<>")

		if verbose {
			cyan := color.New(color.FgCyan)
			cyan.Printf("[DEBUG] InReplyTo da mensagem: %s (limpo: %s)\n", inReplyTo, inReplyToClean)
			cyan.Printf("[DEBUG] Message-ID buscado: %s (limpo: %s)\n", messageID, messageIDCleanForCompare)
		}

		// Verifica se corresponde
		if inReplyTo == messageID || inReplyTo == messageIDClean ||
		   inReplyToClean == messageIDCleanForCompare || strings.Contains(inReplyTo, messageIDCleanForCompare) {
			if verbose {
				cyan := color.New(color.FgCyan)
				cyan.Printf("[DEBUG] ✓ Mensagem validada! InReplyTo corresponde ao Message-ID buscado.\n")
			}
			// Agora busca a mensagem completa
			return fetchLatestMessage(c, []uint32{latestUID}, fullLayout, verbose)
		} else {
			if verbose {
				yellow := color.New(color.FgYellow)
				yellow.Printf("[DEBUG] ✗ Mensagem rejeitada: InReplyTo (%s) não corresponde ao Message-ID buscado (%s)\n", inReplyTo, messageID)
				yellow.Printf("[DEBUG] Tentando mensagens mais antigas na lista...\n")
			}
			// Tenta mensagens anteriores na lista (mais antigas)
			for i := len(uids) - 2; i >= 0; i-- {
				uid := uids[i]
				if verbose {
					cyan := color.New(color.FgCyan)
					cyan.Printf("[DEBUG] Validando mensagem UID=%d...\n", uid)
				}
				seqset2 := new(imap.SeqSet)
				seqset2.AddNum(uid)
				messages2 := make(chan *imap.Message, 1)
				done2 := make(chan error, 1)
				go func() {
					done2 <- c.Fetch(seqset2, items, messages2)
				}()
				msg2 := <-messages2
				<-done2
				if msg2 != nil && msg2.Envelope != nil && msg2.Envelope.InReplyTo != "" {
					inReplyTo2 := strings.Trim(msg2.Envelope.InReplyTo, "<>")
					if inReplyTo2 == messageIDCleanForCompare || strings.Contains(msg2.Envelope.InReplyTo, messageIDCleanForCompare) {
						if verbose {
							cyan := color.New(color.FgCyan)
							cyan.Printf("[DEBUG] ✓ Mensagem UID=%d validada!\n", uid)
						}
						return fetchLatestMessage(c, []uint32{uid}, fullLayout, verbose)
					}
				}
			}
			// Nenhuma mensagem correspondeu
			if verbose {
				yellow := color.New(color.FgYellow)
				yellow.Printf("[DEBUG] Nenhuma mensagem encontrada por Subject corresponde ao Message-ID buscado.\n")
			}
			return false, nil, nil
		}
	}

	// Se não tem InReplyTo, não podemos validar - retorna false para continuar buscando
	if verbose {
		yellow := color.New(color.FgYellow)
		yellow.Printf("[DEBUG] Mensagem não tem InReplyTo, não é possível validar. Ignorando.\n")
	}
	return false, nil, nil
}

// fetchLatestMessage busca a mensagem mais recente da lista de UIDs e retorna seus dados.
func fetchLatestMessage(c *client.Client, uids []uint32, fullLayout bool, verbose bool) (bool, *EmailResponse, error) {
	if len(uids) == 0 {
		return false, nil, nil
	}

	// Pega o UID mais recente (último da lista, assumindo que estão ordenados)
	latestUID := uids[len(uids)-1]

	if verbose {
		cyan := color.New(color.FgCyan)
		cyan.Printf("[DEBUG] Mensagem encontrada, UID=%d\n", latestUID)
	}

	// Fetch da mensagem completa (headers + corpo)
	seqset := new(imap.SeqSet)
	seqset.AddNum(latestUID)

	// Busca mensagem completa - BODY[] retorna toda a mensagem (headers + corpo)
	// Usa BodySectionName vazio para BODY[]
	bodySectionFull := &imap.BodySectionName{}
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchBodyStructure,
		bodySectionFull.FetchItem(), // BODY[] - mensagem completa (mais confiável)
		section.FetchItem(),          // BODY[section] - seção específica (fallback)
	}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	msg := <-messages
	if msg == nil {
		<-done
		return false, nil, fmt.Errorf("mensagem não encontrada")
	}

	if verbose {
		cyan := color.New(color.FgCyan)
		cyan.Printf("[DEBUG] Mensagem recebida, Envelope: %+v\n", msg.Envelope)
		if msg.Body != nil {
			cyan.Printf("[DEBUG] Body sections disponíveis: %d\n", len(msg.Body))
			// Lista todas as seções disponíveis para debug
			for sec := range msg.Body {
				cyan.Printf("[DEBUG] Seção disponível no map: %+v\n", sec)
			}
		} else {
			cyan.Printf("[DEBUG] Body é nil\n")
		}
		if msg.BodyStructure != nil {
			cyan.Printf("[DEBUG] BodyStructure tipo: %s\n", msg.BodyStructure.MIMEType)
		}
	}

	// Parse da mensagem (passa bodySectionFull para tentar BODY[] primeiro)
	response, err := parseEmailMessage(msg, bodySectionFull, section, fullLayout, verbose)
	if err != nil {
		<-done
		return false, nil, fmt.Errorf("erro ao parsear mensagem: %w", err)
	}

	<-done

	return true, response, nil
}

// parseEmailMessage parseia uma mensagem IMAP e extrai headers e corpo.
// fullLayout: se true, inclui HTML; se false, apenas texto (padrão)
func parseEmailMessage(msg *imap.Message, bodySectionFull *imap.BodySectionName, section *imap.BodySectionName, fullLayout bool, verbose bool) (*EmailResponse, error) {
	response := &EmailResponse{}

	// Usa Envelope se disponível (mais confiável)
	if msg.Envelope != nil {
		if len(msg.Envelope.From) > 0 {
			response.From = msg.Envelope.From[0].Address()
		}
		if !msg.Envelope.Date.IsZero() {
			response.Date = msg.Envelope.Date
		}
		if msg.Envelope.Subject != "" {
			response.Subject = msg.Envelope.Subject
		}
	}

	// Tenta ler o corpo
	if msg.Body == nil {
		if verbose {
			yellow := color.New(color.FgYellow)
			yellow.Printf("[DEBUG] Body é nil, tentando buscar novamente com FETCH completo...\n")
		}
		// Se não tem body, retorna pelo menos os headers do Envelope
		if response.From != "" {
			return response, nil
		}
		return nil, fmt.Errorf("corpo da mensagem não disponível e Envelope vazio")
	}

	// Tenta ler o corpo - tenta várias formas de acesso
	var body io.Reader

	// 1. Tenta BODY[] completo primeiro (mais confiável)
	body = msg.Body[bodySectionFull]

	// 2. Se não encontrou, tenta a seção específica
	if body == nil {
		body = msg.Body[section]
	}

	// 3. Se ainda não encontrou, tenta qualquer seção disponível
	if body == nil && msg.Body != nil {
		if verbose {
			yellow := color.New(color.FgYellow)
			yellow.Printf("[DEBUG] Seções específicas não encontradas, tentando todas as seções disponíveis...\n")
		}
		for sec, bodyReader := range msg.Body {
			if verbose {
				cyan := color.New(color.FgCyan)
				cyan.Printf("[DEBUG] Tentando seção: %+v (reader: %v)\n", sec, bodyReader != nil)
			}
			if bodyReader != nil {
				body = bodyReader
				if verbose {
					cyan := color.New(color.FgCyan)
					cyan.Printf("[DEBUG] Seção encontrada! Usando esta seção.\n")
				}
				break
			}
		}
	}

	if body == nil {
		if verbose {
			yellow := color.New(color.FgYellow)
			yellow.Printf("[DEBUG] Seção do corpo não disponível, mas temos Envelope. Retornando headers apenas.\n")
		}
		// Se não tem corpo mas tem Envelope, retorna pelo menos os headers
		if response.From != "" {
			response.Body = "[Corpo não disponível via IMAP, mas mensagem encontrada]"
			return response, nil
		}
		return nil, fmt.Errorf("seção do corpo não disponível")
	}

	// Parse usando go-message
	mr, err := mail.CreateReader(body)
	if err != nil {
		if verbose {
			yellow := color.New(color.FgYellow)
			yellow.Printf("[DEBUG] Erro ao criar mail reader: %v, tentando ler como texto simples...\n", err)
		}
		// Se falhar, tenta ler como texto simples
		b, readErr := io.ReadAll(body)
		if readErr == nil {
			response.Body = string(b)
			return response, nil
		}
		return nil, fmt.Errorf("erro ao criar reader: %w", err)
	}

	// Headers (só preenche se não foram preenchidos pelo Envelope)
	if response.From == "" {
		header := mr.Header
		if from, err := header.AddressList("From"); err == nil && len(from) > 0 {
			response.From = from[0].Address
		}
		if response.Date.IsZero() {
			if date, err := header.Date(); err == nil {
				response.Date = date
			}
		}
		if response.Subject == "" {
			if subject, err := header.Text("Subject"); err == nil {
				response.Subject = subject
			}
		}
	}

	// Corpo (text/plain)
	var bodyParts []string
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			contentType, _, _ := h.ContentType()

			// Se fullLayout=false, ignora HTML e pega apenas text/plain
			if !fullLayout && contentType == "text/html" {
				if verbose {
					cyan := color.New(color.FgCyan)
					cyan.Printf("[DEBUG] Ignorando parte HTML (fullLayout=false)\n")
				}
				continue
			}

			// Se fullLayout=true, inclui tudo (HTML e texto)
			b, err := io.ReadAll(p.Body)
			if err == nil {
				bodyParts = append(bodyParts, string(b))
			}
		case *mail.AttachmentHeader:
			// Ignora anexos por enquanto
			continue
		default:
			_ = h
		}
	}

	if len(bodyParts) > 0 {
		response.Body = strings.Join(bodyParts, "\n")
		// Limpa a resposta: remove a mensagem original do CAST que vem após "escreveu:"
		response.Body = cleanEmailResponse(response.Body)
	}

	return response, nil
}

// cleanEmailResponse remove a mensagem original do CAST que vem na resposta.
// Procura por padrões como "escreveu:" ou "wrote:" e remove tudo a partir daí.
// Também remove linhas de cabeçalho de citação que precedem "escreveu:" (ex: "Em sáb., 13 de dez. de 2025 15:44, NOME <email>")
func cleanEmailResponse(body string) string {
	lines := strings.Split(body, "\n")
	var cleanLines []string
	foundMarker := false

	for _, line := range lines {
		lowerLine := strings.ToLower(strings.TrimSpace(line))

		// Detecta linha de cabeçalho de citação (ex: "Em sáb., 13 de dez. de 2025 15:44, NOME <email>")
		// Padrões: "em [dia], [data], [nome] <email>" ou "on [day], [date], [name] <email>"
		// Detecta linhas que contêm data + email (mesmo sem "escreveu:" na mesma linha)
		hasDatePattern := strings.Contains(lowerLine, "em ") || strings.Contains(lowerLine, "on ")
		hasEmailPattern := strings.Contains(lowerLine, "@") || (strings.Contains(lowerLine, "<") && strings.Contains(lowerLine, ">"))
		hasDateInfo := strings.Contains(lowerLine, "de ") || (strings.Count(lowerLine, ",") >= 2)

		isCitationHeader := hasDatePattern && hasEmailPattern && hasDateInfo

		// Detecta marcadores de citação (português e inglês)
		isCitationMarker := strings.Contains(lowerLine, "escreveu:") ||
			strings.Contains(lowerLine, "wrote:") ||
			strings.Contains(lowerLine, "escreveu")

		if isCitationHeader || isCitationMarker {
			foundMarker = true
			// Remove a linha imediatamente anterior se for vazia ou tiver padrão relacionado
			if len(cleanLines) > 0 {
				prevLine := strings.TrimSpace(cleanLines[len(cleanLines)-1])
				if prevLine == "" || strings.ToLower(prevLine) == "---" || strings.ToLower(prevLine) == "---original message---" {
					cleanLines = cleanLines[:len(cleanLines)-1]
				}
			}
			// Para na linha do marcador (não inclui ela nem as seguintes)
			break
		}
		cleanLines = append(cleanLines, line)
	}

	// Se não encontrou marcador, retorna tudo (pode ser resposta sem citação)
	if !foundMarker {
		return body
	}

	// Remove linhas vazias no final
	for len(cleanLines) > 0 && strings.TrimSpace(cleanLines[len(cleanLines)-1]) == "" {
		cleanLines = cleanLines[:len(cleanLines)-1]
	}

	return strings.Join(cleanLines, "\n")
}

// printEmailResponse exibe a resposta de email no formato especificado.
func printEmailResponse(response *EmailResponse, maxLines int, verbose bool) {
	// Delimitadores apenas em modo verbose
	if verbose {
		fmt.Println("=== EMAIL RESPONSE ===")
		fmt.Printf("From: %s\n", response.From)
		fmt.Printf("Date: %s\n", response.Date.Format("2006-01-02 15:04:05"))
		fmt.Printf("Subject: %s\n\n", response.Subject)
	}

	// Corpo (já limpo, sem mensagem original)
	bodyLines := strings.Split(response.Body, "\n")
	if maxLines > 0 && len(bodyLines) > maxLines {
		// Trunca
		for i := 0; i < maxLines; i++ {
			fmt.Println(bodyLines[i])
		}
		yellow := color.New(color.FgYellow)
		yellow.Printf("\n[... corpo truncado em %d linhas (ajuste email.wait_for_response_max_lines para 0 se quiser mostrar tudo) ...]\n", maxLines)
	} else {
		// Mostra completo
		fmt.Println(response.Body)
	}

	// Delimitadores apenas em modo verbose
	if verbose {
		fmt.Println("=== END EMAIL RESPONSE ===")
	}
}

// formatDuration formata uma duração de forma legível.
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
