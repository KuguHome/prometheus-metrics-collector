package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// MailMessage enthält die Eigenschaften einer Nachricht / eines E-Briefes (e-mail)
type MailMessage struct {
	Sender  string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Body    string
}

// SMTPServer ist ein SMTP-Server, an den versandt werden kann
type SMTPServer struct {
	Host      string
	Port      string
	TLSConfig *tls.Config
}

// Address liefert eine serialisierte Form der Server-Adresse
func (s *SMTPServer) Address() string {
	return s.Host + ":" + s.Port
}

// Serialize erstellt eine versandfertige SMTP-Nachricht
func (mail *MailMessage) Serialize() string {
	header := ""
	header += fmt.Sprintf("From: %s\r\n", mail.Sender)
	if len(mail.To) > 0 {
		header += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	}
	if len(mail.Cc) > 0 {
		header += fmt.Sprintf("Cc: %s\r\n", strings.Join(mail.Cc, ";"))
	}

	header += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	header += "\r\n" + mail.Body

	return header
}

//TODO aktuell nicht verwendet, erlaubt ev. mehr Kontrolle
func sendMailManual() error {
	// Nachricht erstellen
	mail := MailMessage{}
	mail.Sender = "abc@gmail.com"
	mail.To = []string{"def@yahoo.com", "xyz@outlook.com"}
	mail.Cc = []string{"mnp@gmail.com"}
	mail.Bcc = []string{"a69@outlook.com"}
	mail.Subject = "I am Harry Potter!!"
	mail.Body = "Harry Potter and threat to Israel\n\nGood editing!!"

	// Nachricht versandbereit machen
	messageBody := mail.Serialize()

	// SMTP-Server anlegen
	smtpServer := SMTPServer{Host: "smtp.gmail.com", Port: "465"}
	smtpServer.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer.Host,
	}

	// Authentifierungsmethode vorbereiten (Benutzername und Passwort)
	auth := smtp.PlainAuth("", mail.Sender, "password", smtpServer.Host)

	// Verbindung aufbauen über TLS
	conn, err := tls.Dial("tcp", smtpServer.Address(), smtpServer.TLSConfig)
	if err != nil {
		panic(err)
	}

	// SMTP-Verbindung aufbauen
	client, err := smtp.NewClient(conn, smtpServer.Host)
	if err != nil {
		panic(err)
	}

	// Anmelden / authentifizieren
	if err = client.Auth(auth); err != nil {
		panic(err)
	}

	// Absenderadresse angeben
	if err = client.Mail(mail.Sender); err != nil {
		panic(err)
	}
	// alle Empfänger angeben
	receivers := append(mail.To, mail.Cc...)
	receivers = append(receivers, mail.Bcc...)
	for _, k := range receivers {
		fmt.Println("sending to: ", k)
		if err = client.Rcpt(k); err != nil {
			panic(err)
		}
	}

	// in Datenmodus wechseln, verpackte Nachricht senden, Datenmodus schließen
	w, err := client.Data()
	if err != nil {
		panic(err)
	}
	_, err = w.Write([]byte(messageBody))
	if err != nil {
		panic(err)
	}
	err = w.Close()
	if err != nil {
		panic(err)
	}

	// SMTP- (und TLS-)Verbindung schließen
	client.Quit()

	//TODO
	return nil
}
