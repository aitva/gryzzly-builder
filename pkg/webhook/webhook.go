package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// emptyCallback is the default callback.
var emptyCallback Callback = func(interface{}) error {
	return nil
}

// Push represents a push event from Github.
type Push struct {
	Ref string `json:"ref"`
}

// Release represents a release event from Github.
type Release struct {
	Action  string `json:"action"`
	Release struct {
		TagName string      `json:"tag_name"`
		Name    interface{} `json:"name"`
	} `json:"release"`
}

// Callback represents a callback function for Github event.
type Callback func(interface{}) error

// Webhook represents a webhook HTTP handler.
type Webhook struct {
	l               *log.Logger
	secret          string
	pushCallback    Callback
	releaseCallback Callback
}

type WebhookConfig struct {
	L               *log.Logger
	PushCallback    Callback
	ReleaseCallback Callback
}

// New instanciates a new Webhook with the given secret.
func New(secret string, config WebhookConfig) *Webhook {
	if config.L == nil {
		config.L = log.New(ioutil.Discard, "", 0)
	}
	if config.PushCallback == nil {
		config.PushCallback = emptyCallback
	}
	if config.ReleaseCallback == nil {
		config.ReleaseCallback = emptyCallback
	}
	return &Webhook{
		l:               config.L,
		secret:          secret,
		pushCallback:    config.PushCallback,
		releaseCallback: config.ReleaseCallback,
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
		// return with code 200
		return
	case "push":
		push := &Push{}
		err = json.Unmarshal(body, push)
		if err != nil {
			wh.l.Println("fail to unmarshal json:", err)
			http.Error(w, "bad request", http.StatusBadRequest)
		}
		err = wh.pushCallback(push)
		if err != nil {
			wh.l.Println("fail to execute push callback:", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	case "release":
		release := &Release{}
		err = json.Unmarshal(body, release)
		if err != nil {
			wh.l.Println("fail to unmarshal json:", err)
			http.Error(w, "bad request", http.StatusBadRequest)
		}
		err := wh.releaseCallback(release)
		if err != nil {
			wh.l.Println("fail to execute push callback:", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
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
