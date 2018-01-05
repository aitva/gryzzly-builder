package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aitva/gryzzly-builder/pkg/webhook"
	"github.com/gorilla/mux"
	flag "github.com/spf13/pflag"
)

type configuration struct {
	addr          string
	webhookSecret string
}

func loadConfiguration() *configuration {
	c := &configuration{}

	flag.StringVar(&c.addr, "addr", ":8080", "server address. Env: BUILDER_ADDR")
	flag.StringVar(&c.webhookSecret, "webhook", "", "mandatory webhook secret. Env: BUILDER_WEBHOOK")
	flag.Parse()

	if env := os.Getenv("BUILDER_ADDR"); env != "" {
		c.addr = env
	}
	if env := os.Getenv("BUILDER_WEBHOOK"); env != "" {
		c.webhookSecret = env
	}

	if c.webhookSecret == "" {
		flag.Usage()
		return nil
	}
	return c
}

func main() {
	config := loadConfiguration()
	if config == nil {
		os.Exit(1)
	}

	l := log.New(os.Stdout, "MAIN  ", log.LstdFlags|log.Lshortfile)

	l.Println("init components")
	hook := webhook.New(config.webhookSecret, webhook.WebhookConfig{
		L: log.New(os.Stdout, "HOOK  ", log.LstdFlags|log.Lshortfile),
	})

	mux := mux.NewRouter()
	mux.Handle("/", hook)

	server := http.Server{
		Addr:     config.addr,
		Handler:  mux,
		ErrorLog: log.New(os.Stdout, "SRV   ", log.LstdFlags|log.Lshortfile),
	}

	l.Printf("listening on %v...", config.addr)
	server.ListenAndServe()
}
