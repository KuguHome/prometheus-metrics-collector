package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Konfiguration ist die Wurzel der Steuerzentralen-Konfiguration
type Konfiguration struct {
	Steuerzentralen map[string]Steuerzentrale `toml:"steuerzentralen"`
}

// Steuerzentrale beschreibt eine einzelne Steuerzentrale
type Steuerzentrale struct {
	Bezeichnung       string        `toml:"bezeichnung"`
	SSHVerbindung     SSHVerbindung `toml:"ssh"`
	OpenHABVerbindung WebVerbindung `toml:"openhab"`
}

// SSHVerbindung ist eine SSH-Verbindung zu einer Steuerzentrale
type SSHVerbindung struct {
	Port     int    `toml:"port"`
	Benutzer string `toml:"benutzer"`
}

// WebVerbindung ist eine HTTP/HTTPS-Verbindung zu einer Steuerzentrale, aktuell die OpenHAB-Weboberfläche
type WebVerbindung struct {
	Port            int    `toml:"port"`
	Benutzer        string `toml:"benutzer"`
	Passwort        string `toml:"passwort"`
	Passworthinweis string `toml:"passwort_hinweis"`
}

func konfigurationEinlesen(pfad string) (konfig *Konfiguration, err error) {
	// Einlesen
	meta, err := toml.DecodeFile(pfad, &konfig)
	if err != nil {
		return nil, err
	}
	// Kontrollen
	if len(meta.Undecoded()) > 0 {
		return nil, fmt.Errorf("unbekannte Einträge in Konfiguration: %v", meta.Undecoded())
	}
	for szName := range konfig.Steuerzentralen {
		if !meta.IsDefined("steuerzentralen", szName, "bezeichnung") {
			return nil, fmt.Errorf("Steuerzentrale %s hat keine Bezeichnung", szName)
		}
		if !meta.IsDefined("steuerzentralen", szName, "ssh") {
			return nil, fmt.Errorf("Steuerzentrale %s hat keine SSH-Konfiguration", szName)
		}
		if !meta.IsDefined("steuerzentralen", szName, "ssh", "port") {
			return nil, fmt.Errorf("Steuerzentrale %s hat keinen SSH-Port", szName)
		}
		if !meta.IsDefined("steuerzentralen", szName, "ssh", "benutzer") {
			return nil, fmt.Errorf("Steuerzentrale %s hat keinen SSH-Benutzer", szName)
		}
		if !meta.IsDefined("steuerzentralen", szName, "openhab") {
			return nil, fmt.Errorf("Steuerzentrale %s hat keine OpenHAB-Konfiguration", szName)
		}
		if !meta.IsDefined("steuerzentralen", szName, "openhab", "port") {
			return nil, fmt.Errorf("Steuerzentrale %s hat keinen OpenHAB-Port", szName)
		}
		if !meta.IsDefined("steuerzentralen", szName, "openhab", "benutzer") {
			return nil, fmt.Errorf("Steuerzentrale %s hat keinen OpenHAB-Benutzer", szName)
		}
		if !meta.IsDefined("steuerzentralen", szName, "openhab", "passwort") {
			return nil, fmt.Errorf("Steuerzentrale %s hat kein OpenHAB-Passwort", szName)
		}
		if !meta.IsDefined("steuerzentralen", szName, "openhab", "passwort_hinweis") {
			return nil, fmt.Errorf("Steuerzentrale %s hat keinen OpenHAB-Passworthinweis", szName)
		}
	}
	// Ergebnis zurückliefern
	return konfig, nil
}
