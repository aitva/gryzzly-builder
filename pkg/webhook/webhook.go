package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
)

// Webhook represents a webhook HTTP handler.
type Webhook struct {
	l      *log.Logger
	secret string
}

type WebhookConfig struct {
	L *log.Logger
}

// New instanciates a new Webhook with the given secret.
func New(secret string, config WebhookConfig) *Webhook {
	if config.L == nil {
		config.L = log.New(ioutil.Discard, "", 0)
	}
	return &Webhook{
		l:      config.L,
		secret: secret,
	}
}

// ServeHTTP handles the HTTP call to the hook.
func (wh *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	signature := r.Header.Get("X-Hub-Signature")
	if signature == "" {
		wh.l.Println("signature is missing")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		wh.l.Println("fail to read body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if !wh.isValidSignature(signature, body) {
		wh.l.Println("invalid signature:", signature)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	event := r.Header.Get("X-Github-Event")
	switch event {
	case "ping":
		w.Write([]byte("pong"))
	case "push":
		// TODO: handle push
	case "release":
		// TODO: handle release
	default:
		// return 400 if we do not handle the event type.
		// This is to visually show the user a configuration error in the GH ui.
		wh.l.Println("unexpected event:", event)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
}

func (wh *Webhook) isValidSignature(signature string, body []byte) bool {
	mac := hmac.New(sha1.New, []byte(wh.secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return signature[5:] == expected
}
