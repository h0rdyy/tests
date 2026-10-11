package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

type Config struct {
	Host, User, Password, Receiver string
	Port                           string
}

type Mailer struct {
	config Config
	log    *log.Logger
}

func New(config Config, logger *log.Logger) *Mailer {
	return &Mailer{config: config, log: logger}
}

func (m *Mailer) Send(phone, source, when string) {
	if m.config.Host == "" || m.config.User == "" || m.config.Password == "" || m.config.Receiver == "" {
		m.log.Printf("SMTP is not configured; notification skipped")
		return
	}

	address := fmt.Sprintf("%s:%s", m.config.Host, m.config.Port)
	tlsConfig := &tls.Config{ServerName: m.config.Host, MinVersion: tls.VersionTLS12}
	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		m.log.Printf("SMTP TLS connection failed: %v", err)
		return
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.config.Host)
	if err != nil {
		m.log.Printf("SMTP client creation failed: %v", err)
		return
	}
	defer client.Close()

	auth := smtp.PlainAuth("", m.config.User, m.config.Password, m.config.Host)
	if err = client.Auth(auth); err != nil {
		m.log.Printf("SMTP authentication failed: %v", err)
		return
	}
	if err = client.Mail(m.config.User); err != nil {
		m.log.Printf("SMTP sender rejected: %v", err)
		return
	}
	if err = client.Rcpt(m.config.Receiver); err != nil {
		m.log.Printf("SMTP recipient rejected: %v", err)
		return
	}

	writer, err := client.Data()
	if err != nil {
		m.log.Printf("SMTP data command failed: %v", err)
		return
	}
	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: Новая заявка с сайта!\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nУ вас новая заявка!\nТелефон: %s\nФорма: %s\nВремя: %s\n", m.config.Receiver, m.config.User, phone, source, when)
	if _, err = writer.Write([]byte(message)); err != nil {
		m.log.Printf("SMTP message write failed: %v", err)
		return
	}
	if err = writer.Close(); err != nil {
		m.log.Printf("SMTP message delivery failed: %v", err)
		return
	}
	if err = client.Quit(); err != nil {
		m.log.Printf("SMTP quit failed: %v", err)
	}
}
