/**
 * Ebase frame for daemon program
 * Author Jonsen Yang
 * Date 2013-07-05
 * Copyright (c) 2013 ForEase Times Technology Co., Ltd. All rights reserved.
 */
package ebase

import (
	"bytes"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// SMTP setup
type Smtp struct {
	SmtpUserName                     string
	SmtpHost, SmtpUser, SmtpPassword string
	SmtpPort                         int
	SmtpAuth, SmtpTLS, SmtpDaemon    bool
	mailChan                         chan *Mailer
}

type Mailer struct {
	Subject     string
	Content     string
	To, Cc, Bcc string
}

func NewSmtp() *Smtp {
	// SMTP
	s := new(Smtp)
	s.SmtpHost, _ = Config.String("smtp.host", "")
	s.SmtpUser, _ = Config.String("smtp.user", "")
	s.SmtpPassword, _ = Config.String("smtp.password", "")
	s.SmtpPort, _ = Config.Int("smtp.port", 25)
	s.SmtpAuth, _ = Config.Bool("smtp.auth", false)
	s.SmtpTLS, _ = Config.Bool("smtp.tls", false)
	s.SmtpDaemon, _ = Config.Bool("smtp.daemon", false)

	s.mailChan = make(chan *Mailer)

	return s
}

// 运行一个goroutine 监听发送邮件任务
func (s *Smtp) MailSendServer() {
	//    mailChan = make(chan *Mailer)
	Log.Info("Running Mail Send Server...")

	for {

		mailer := <-s.mailChan

		if mailer == nil {
			continue
		}

		m := s.NewMailMessage(mailer)
		if err := m.Send(); err != nil {
			Log.Errorf("send mail to "+mailer.To+" error %s", err)
		}
	}

}

func (s *Smtp) MailSender(subject, content, to, cc, bcc string) (err error) {

	if subject != "" && content != "" && to != "" {
		m := &Mailer{Subject: subject, Content: content, To: to, Cc: cc, Bcc: bcc}
		if s.SmtpDaemon {
			s.mailChan <- m
		} else {
			send := s.NewMailMessage(m)
			if err = send.Send(); err != nil {
				Log.Errorf("send mail to "+to+" error %s", err)
				return err
			}
		}

		return nil
	}

	return errors.New("input is null")
}

func (s *Smtp) NewMailMessage(m *Mailer) *MailMessage {
	tos := strings.Split(m.To, ",")
	ccs := strings.Split(m.Cc, ",")
	bccs := strings.Split(m.Bcc, ",")
	message := &MailMessage{Subject: m.Subject, Content: m.Content,
		To:  make([]mail.Address, len(tos)),
		Cc:  make([]mail.Address, len(ccs)),
		Bcc: make([]mail.Address, len(bccs)),
		S:   s,
	}

	for k, v := range tos {
		message.To[k].Address = v
	}
	for k, v := range ccs {
		message.Cc[k].Address = v
	}
	for k, v := range bccs {
		message.Bcc[k].Address = v
	}

	//fmt.Println( message.To )
	return message
}

/*
func NewMailMessageFrom(subject, content, from, to string) *MailMessage {
	message := NewMailMessage(subject, content, to)
	message.From.Address = from
	return message
}
*/

const crlf = "\r\n"

type MailMessage struct {
	From    mail.Address // if From.Address is empty, Config.DefaultFrom will be used
	To      []mail.Address
	Cc      []mail.Address
	Bcc     []mail.Address
	Subject string
	Content string
	S       *Smtp
}

// http://tools.ietf.org/html/rfc822
// http://tools.ietf.org/html/rfc2821
func (self *MailMessage) String() string {
	var buf bytes.Buffer

	write := func(what string, recipients []mail.Address) {
		if len(recipients) == 0 {
			return
		}
		for i := range recipients {
			if i == 0 {
				buf.WriteString(what)
			} else {
				buf.WriteString(", ")
			}
			buf.WriteString(recipients[i].String())
		}
		buf.WriteString(crlf)
	}

	from := &self.From
	if from.Address == "" {
		from = &mail.Address{self.S.SmtpUserName, self.S.SmtpUser} //&Config.From
	}

	//if cfg.adminMail != "" {
	//    self.Bcc = make([]mail.Address, 1)
	//    self.Bcc[0] = mail.Address{ "adminer", cfg.adminMail }
	//}

	fmt.Fprintf(&buf, "From: %s%s", from.String(), crlf)
	write("To: ", self.To)
	write("Cc: ", self.Cc)
	write("Bcc: ", self.Bcc)
	fmt.Fprintf(&buf, "Date: %s%s", time.Now().UTC().Format(time.RFC822), crlf)
	fmt.Fprintf(&buf, "Subject: %s%s%s", self.Subject, crlf, self.Content)
	return buf.String()
}

// Returns the first error
func (self *MailMessage) Validate() error {
	if len(self.To) == 0 {
		return errors.New("Missing email recipient (email.Message.To)")
	}
	return nil
}

type fakeAuth struct {
	smtp.Auth
}

func (a fakeAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	server.TLS = true
	return a.Auth.Start(server)
}

func (self *MailMessage) Send() error {
	var auth smtp.Auth

	if err := self.Validate(); err != nil {
		return err
	}

	to := make([]string, len(self.To))
	for i := range self.To {
		to[i] = self.To[i].Address
	}

	from := self.From.Address
	if from == "" {
		from = self.S.SmtpUser // Config.From.Address
	}

	addr := fmt.Sprintf("%s:%d", self.S.SmtpHost, self.S.SmtpPort)

	if self.S.SmtpTLS {
		auth = fakeAuth{smtp.PlainAuth("", self.S.SmtpUser,
			self.S.SmtpPassword, self.S.SmtpHost)}
	} else {
		auth = smtp.PlainAuth("", self.S.SmtpUser,
			self.S.SmtpPassword, self.S.SmtpHost)
	}

	return smtp.SendMail(addr, auth, from, to, []byte(self.String()))
}
