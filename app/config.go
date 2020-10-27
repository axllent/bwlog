package app

import "time"

// Config struct
var Config struct {
	Interfaces   []string
	DatabaseDir  string
	SaveInterval time.Duration
	Listen       string
	AuthUser     string
	AuthPass     string
	SSLCert      string
	SSLKey       string
}
