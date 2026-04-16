package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ---- XOAUTH2 smtp.Auth implementation ----

type xoauth2Auth struct {
	email, tokenFile string
}

func XOAuth2Auth(email, tokenFile string) smtp.Auth {
	return &xoauth2Auth{email, tokenFile}
}

func (a *xoauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	token, err := loadOAuth2Token(a.tokenFile)
	if err != nil {
		return "", nil, fmt.Errorf("xoauth2: load token: %w", err)
	}

	if !token.Valid() {
		token, err = refreshOAuth2Token(token)
		if err != nil {
			return "", nil, fmt.Errorf("xoauth2: refresh token: %w", err)
		}
		if err := saveOAuth2Token(a.tokenFile, token); err != nil {
			return "", nil, fmt.Errorf("xoauth2: save token: %w", err)
		}
	}

	blob := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.email, token.AccessToken)
	return "XOAUTH2", []byte(blob), nil
}

func (a *xoauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	// Gmail sends a base64-encoded JSON error in the 334 challenge on failure.
	// Returning an empty response causes the server to send the final 535 error.
	if more {
		return []byte{}, nil
	}
	return nil, nil
}

// ---- Token file helpers ----

func loadOAuth2Token(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t oauth2.Token
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func saveOAuth2Token(path string, t *oauth2.Token) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func refreshOAuth2Token(t *oauth2.Token) (*oauth2.Token, error) {
	cfg := gmailOAuth2Config()
	ts := cfg.TokenSource(context.Background(), t)
	return ts.Token()
}

// ---- Gmail OAuth2 config ----

func gmailOAuth2Config() *oauth2.Config {
	redirectURL := *oauth2RedirectURL
	if redirectURL == "" {
		redirectURL = "http://" + *webListen + "/oauth2/callback"
	}
	return &oauth2.Config{
		ClientID:     *oauth2ClientID,
		ClientSecret: *oauth2ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://mail.google.com/"},
		RedirectURL:  redirectURL,
	}
}

// ---- OAuth2 state map (CSRF protection) ----

type oauth2FlowState struct {
	Email     string
	TokenFile string
	ExpiresAt time.Time
}

var oauth2States = struct {
	sync.Mutex
	m map[string]oauth2FlowState
}{m: make(map[string]oauth2FlowState)}

func newOAuth2State(email, tokenFile string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(b)

	oauth2States.Lock()
	defer oauth2States.Unlock()
	oauth2States.m[nonce] = oauth2FlowState{
		Email:     email,
		TokenFile: tokenFile,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	return nonce, nil
}

func consumeOAuth2State(nonce string) (oauth2FlowState, bool) {
	oauth2States.Lock()
	defer oauth2States.Unlock()
	s, ok := oauth2States.m[nonce]
	if !ok {
		return oauth2FlowState{}, false
	}
	delete(oauth2States.m, nonce)
	if time.Now().After(s.ExpiresAt) {
		return oauth2FlowState{}, false
	}
	return s, true
}
