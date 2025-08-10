package config

import (
	"os"
)

// Config 結構定義了應用程式的所有組態選項。
type Config struct {
	DBPath     string
	ServerPort string
	CertFile   string
	KeyFile    string
}

// Load 函式從環境變數讀取組態，並在未設定時提供預設值。
func Load() *Config {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "voip.db"
	}

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	certFile := os.Getenv("CERT_FILE")
	if certFile == "" {
		certFile = "scripts/cert.pem"
	}

	keyFile := os.Getenv("KEY_FILE")
	if keyFile == "" {
		keyFile = "scripts/key.pem"
	}

	return &Config{
		DBPath:     dbPath,
		ServerPort: ":" + serverPort,
		CertFile:   certFile,
		KeyFile:    keyFile,
	}
}
