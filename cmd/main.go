package main

import (
	s "dev-support-schedule/pkg/mode/chatbot"
	c "dev-support-schedule/pkg/mode/cli"
	"flag"
	"fmt"
	"os"

	"dev-support-schedule/pkg/storage"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	googleClientSecretPath = "data/google_client_secret.json"
	googleTokenPath        = "data/google_token.json"
	sqlitePath             = "data/db.sqlite"
	configPath             = "data/config.yaml"
)

const (
	modeHTTPS = "https"
	modeHTTP  = "http"
	modeCLI   = "cli"
)

var mode = flag.String("mode", modeHTTP, "Режим запуска: https, http или cli")

type Config struct {
	Domain        string `yaml:"domain"`
	PachkaWebhook string `yaml:"pachka_webhook"`
}

func main() {
	var cfg Config
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// db InitDB
	db := storage.InitDB(sqlitePath)

	flag.Parse()

	switch *mode {
	case modeHTTPS:
		server := s.NewServer(db, cfg.PachkaWebhook, cfg.Domain, googleClientSecretPath, googleTokenPath)
		server.Start(true)
	case modeHTTP:
		server := s.NewServer(db, cfg.PachkaWebhook, cfg.Domain, googleClientSecretPath, googleTokenPath)
		server.Start(false)
	case modeCLI:
		cli := c.NewCLI(db, googleClientSecretPath, googleTokenPath)
		cli.Start()
	default:
		fmt.Printf("Неизвестный режим: %s. Допустимые режимы: %s, %s, %s\n", *mode, modeHTTPS, modeHTTP, modeCLI)
	}
}
