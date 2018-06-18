package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
	"strings"
	"time"

	gomail "gopkg.in/gomail.v2"
)

var (
	log *syslog.Writer
)

func main() {
	// mit Systemprotokoll verbinden
	protokollVerbinden()
	log.Info("Sitzung beginnt")

	// Parameter einlesen
	var checkConfig bool
	flag.BoolVar(&checkConfig, "checkconfig", false, "only check configuration, then exit")
	flag.Parse()
	if checkConfig {
		log.Info("Aufruf mit -checkconfig")
	}
	if flag.NArg() == 0 && !checkConfig {
		// keine Empfänger angegeben
		fmt.Println("FEHLER: Keine Empfänger angegeben")
		log.Err("FEHLER: Aufruf ohne Empfängerliste")
		os.Exit(1)
	}
	// Empfänger-Adressen für die Berichte
	smtpReceivers := flag.Args()
	log.Debug(fmt.Sprintf("Liste der Empfänger: %v", smtpReceivers))

	// Liste der Steuerzentralen holen aus Konfigurationsdatei (TOML-Format)
	// -> SSH-Port, Name
	//TODO Wohlgeformtheits-Kontrollprogramm / -parameter (liest nur Datei ein -> ja/nein)
	// ist enthalten -> installieren TODO
	const konfigPfad = "/home/ubuntu/sz.conf"
	var konfig *Konfiguration
	if konfigErhalten, err := konfigurationEinlesen(konfigPfad); err != nil {
		log.Err(fmt.Sprintf("FEHLER beim Einlesen der SZ-Liste aus '%s': %s", konfigPfad, err))
		os.Exit(2)
	} else {
		konfig = konfigErhalten
	}
	log.Info("OK Habe Liste der Steuerzentralen:")
	for szName, sz := range konfig.Steuerzentralen {
		log.Debug(fmt.Sprintf("    SZ %s mit Benutzer %s auf Port %d\n", szName, sz.SSHVerbindung.Benutzer, sz.SSHVerbindung.Port))
	}
	log.Info("Ende der Liste.")
	if checkConfig {
		// Konfiguration geprüft, aussteigen
		return
	}

	// an Steuerzentralen anmelden und Tagesberichte holen
	const szSSHPass = "[FIXME]"
	const szSSHUser = "pi"
	szStatus := map[string]string{}
	for szName, sz := range konfig.Steuerzentralen {
		// Werte extrahieren
		//szBezeichnung := sz.Bezeichnung
		//benutzerName := sz.SSHVerbindung.Benutzer
		port := sz.SSHVerbindung.Port
		log.Info(fmt.Sprintf("[%s] Hole Berichte\n", szName))
		// rsync soll Abgleich machen
		// rsync holt auch Berichte nach, falls z.B. gestern keine ZK-Verbindung bestand
		// /usr/bin/rsync -avx --timeout=30 --rsh="/usr/bin/sshpass -p [FIXME] ssh -o StrictHostKeyChecking=no -l pi -p 2010" localhost:/home/pi/tagesbericht/*pdf ~/tagesbericht/kunde1/
		cmdName := "/usr/bin/rsync"
		cmdArgs := []string{"-avx",
			"--timeout=30",
			"--rsh",
			fmt.Sprintf(`/usr/bin/sshpass -p %s ssh -o StrictHostKeyChecking=no -l %s -p %d`, szSSHPass, szSSHUser, port),
			fmt.Sprintf("localhost:/home/%s/tagesbericht/*pdf", szSSHUser),
			fmt.Sprintf("/home/ubuntu/tagesbericht/%s/", szName),
		}
		cmdline := cmdName + "_" + strings.Join(cmdArgs, "_")
		cmdline = strings.Replace(cmdline, szSSHPass, "[Passwort]", 1)
		log.Debug(fmt.Sprintf("[%s] Befehlszeile = %s", szName, cmdline))
		// Befehl ausführen
		if cmdOutBytes, err := exec.Command(cmdName, cmdArgs...).CombinedOutput(); err != nil {
			// Fehler protokollieren
			status := fmt.Sprintf("[%s] FEHLER bein Holen der Berichte: %s - überspringe. Ausgabe:\n", szName, err)
			log.Err(status)
			logMultiIndent(string(cmdOutBytes))
			// Status merken
			szStatus[szName] = status + "\n" + string(cmdOutBytes)
			// nächste SZ bearbeiten
			continue
		} else {
			// Erfolg protokollieren
			status := fmt.Sprintf("[%s] OK Berichte geholt. Ausgabe:\n", szName)
			log.Info(status)
			logMultiIndent(string(cmdOutBytes))
			// Status merken
			szStatus[szName] = status + "\n" + string(cmdOutBytes)
		}
	}
	log.Info("Berichte geholt")

	// neueste Tagesberichte per E-Post über SMTP versenden
	// Bibliotheken: https://stackoverflow.com/questions/11075749/how-to-send-an-email-with-attachments-in-go
	// -> https://github.com/jordan-wright/email
	// -> https://github.com/go-gomail/gomail
	// Nur mit net/smtp: https://hackernoon.com/golang-sendmail-sending-mail-through-net-smtp-package-5cadbe2670e0
	// von Befehlszeile aus: http://backreference.org/2013/05/22/send-email-with-attachments-from-script-or-command-line/

	// Verbindungseinstellungen
	const smtpServer = "exchangeserver.kugu-home.com"
	const smtpPort = 587
	const smtpUser = "[FIXME]"
	const smtpPass = "[FIXME]"
	const smtpSender = "KUGU Home <[FIXME]>"

	// Vorbereiten, Verbindung herstellen
	//TODO TLSConfig genauer anschauen - Sicherheitsprüfung
	versuch := 1
erneutVerbinden:
	log.Info(fmt.Sprintf("Verbinde mit SMTP-Server (Versuch %d)...", versuch))
	mitteleuropa, _ := time.LoadLocation("Europe/Vienna")
	datumStr := time.Now().AddDate(0, 0, -1).In(mitteleuropa).Format("2006-01-02") // gestrigen Bericht versenden
	d := gomail.NewDialer(smtpServer, smtpPort, smtpUser, smtpPass)
	s, err := d.Dial()
	if err != nil {
		if versuch < 4 {
			log.Err(fmt.Sprintf("FEHLER beim Verbinden zum SMTP-Server: %s - Versuche erneut in 5 Minuten.\n", err))
			time.Sleep(5 * time.Minute)
			versuch++
			goto erneutVerbinden
		} else {
			log.Err(fmt.Sprintf("FEHLER beim Verbinden zum SMTP-Server: %s - Abbruch.\n", err))
			os.Exit(2)
		}
	}
	msg := gomail.NewMessage()
	// Bericht jeder SZ versenden
	log.Info("Versende Berichte...")
	for szName, sz := range konfig.Steuerzentralen {
		// relevante Werte extrahieren
		status := szStatus[szName]
		log.Info(fmt.Sprintf("[%s] Erstelle Nachricht...\n", szName))
		// Nachricht erstellen
		msg.SetHeader("From", smtpSender)
		msg.SetHeader("To", smtpReceivers...)
		//msg.SetAddressHeader("Cc", "dan@example.com", "Dan")
		//msg.SetAddressHeader("To", address, name)
		msg.SetHeader("Subject", fmt.Sprintf("SZ-Tagesbericht %s", szName))

		// gibt es einen Bericht für diese SZ?
		// Existiert eine Datei: https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
		dateiName := fmt.Sprintf("/home/ubuntu/tagesbericht/%s/tagesbericht_%s.pdf", szName, datumStr)
		if _, err := os.Stat(dateiName); os.IsNotExist(err) {
			// kein Bericht vorhanden
			log.Err(fmt.Sprintf("[%s] FEHLER Kein aktueller Tagesbericht für %s gefunden.\n", szName, datumStr))
			msg.SetBody("text/plain", fmt.Sprintf("Kein aktueller Tagesbericht für die Steuerzentrale %s (%s) vorhanden.\nStatus war:\n%s", szName, sz.Bezeichnung, status))
		} else {
			// aktueller Bericht vorhanden
			log.Info(fmt.Sprintf("[%s] OK Aktueller Tagesbericht gefunden.\n", szName))
			msg.Attach(dateiName)
			msg.SetBody("text/plain", fmt.Sprintf("Anbei finden Sie den aktuellen Tagesbericht der Steuerzentrale %s (%s).", szName, sz.Bezeichnung))
		}

		// Send the email to Bob, Cora and Dan.
		/*
			d := gomail.NewDialer(smtpServer, smtpPort, smtpUser, smtpPass)
			if err := d.DialAndSend(m); err != nil {
				panic(err)
			}
		*/
		if err := gomail.Send(s, msg); err != nil {
			log.Err(fmt.Sprintf("[%s] FEHLER beim Senden der Nachricht: %s\n", szName, err))
		} else {
			log.Info(fmt.Sprintf("[%s] OK Nachricht erfolgreich abgeschickt\n", szName))
		}

		// Nachricht leeren -> wiederverwendbar
		msg.Reset()
	}

	// Erfolg
	log.Info("OK Nachrichten erfolgreich gesendet")
	log.Info("Sitzung beendet.")
}

func protokollVerbinden() {
	var err error
	log, err = syslog.New(syslog.LOG_ERR|syslog.LOG_DAEMON, "tagesbericht-versand")
	if err != nil {
		fmt.Println("FEHLER: keine Verbindung zu Syslog möglich")
		os.Exit(1)
	}
}

func logMultiIndent(ausgabe string) {
	teile := strings.Split(ausgabe, "\n")
	for _, teil := range teile {
		log.Info("    " + teil)
	}
}
