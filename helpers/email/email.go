package email

import (
	"errors"
	"log"
	"net/smtp"
	"strconv"
)

type plainAuth struct {
	identity, username, password string
	host                         string
}

func PlainAuth(identity, username, password, host string) smtp.Auth {
	return &plainAuth{identity, username, password, host}
}

func (a *plainAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *plainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// We've already sent everything.
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}

func Send(tos []string, subject string, body string, html bool) {
	fullserver := EmailServer + ":" + strconv.Itoa(EmailPort)
	mimetype := "text/plain"
	if html {
		mimetype = "text/html"
	}
	mime := "MIME-version: 1.0;\nContent-Type: " + mimetype + "; charset=\"UTF-8\";\n\n"
	subject = "Subject: " + subject + "\n"
	msg := []byte(subject + mime + body)

	// Set up authentication information.
	auth := PlainAuth(
		"",
		EmailUsername,
		EmailPassword,
		EmailServer,
	)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		fullserver,
		auth,
		EmailAddress,
		tos,
		msg,
	)
	if err != nil {
		log.Println(err)
	}
}
