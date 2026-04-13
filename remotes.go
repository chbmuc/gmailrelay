package main

import (
	"fmt"
	"net/smtp"
	"net/url"
)

type Remote struct {
	SkipVerify      bool
	Auth            smtp.Auth
	Scheme          string
	Hostname        string
	Port            string
	Addr            string
	Sender          string
	ClientCertPath  string
	ClientKeyPath   string
	OAuth2Email     string
	OAuth2TokenFile string
}

// ParseRemote creates a remote from a given url in the following format:
//
// smtp://[user[:password]@][netloc][:port][/remote_sender][?param1=value1&...]
// smtps://[user[:password]@][netloc][:port][/remote_sender][?param1=value1&...]
// starttls://[user[:password]@][netloc][:port][/remote_sender][?param1=value1&...]
//
// Supported Params:
// - skipVerify: can be "true" or empty to prevent ssl verification of remote server's certificate.
// - auth: can be "login" to trigger "LOGIN" auth instead of "PLAIN" auth
func ParseRemote(remoteURL string) (*Remote, error) {
	u, err := url.Parse(remoteURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "smtp" && u.Scheme != "smtps" && u.Scheme != "starttls" {
		return nil, fmt.Errorf("'%s' is not a supported relay scheme", u.Scheme)
	}

	hostname, port := u.Hostname(), u.Port()

	if port == "" {
		switch u.Scheme {
		case "smtp":
			port = "25"
		case "smtps":
			port = "465"
		case "starttls":
			port = "587"
		}
	}

	q := u.Query()
	r := &Remote{
		Scheme:   u.Scheme,
		Hostname: hostname,
		Port:     port,
		Addr:     fmt.Sprintf("%s:%s", hostname, port),
	}

	if hasAuth, authVal := q.Has("auth"), q.Get("auth"); hasAuth && authVal == "xoauth2" {
		if *oauth2ClientID == "" || *oauth2ClientSecret == "" {
			return nil, fmt.Errorf("auth=xoauth2 requires oauth2_client_id and oauth2_client_secret to be set")
		}
		email := q.Get("email")
		tokenFile := q.Get("token_file")
		if email == "" || tokenFile == "" {
			return nil, fmt.Errorf("auth=xoauth2 requires email and token_file query parameters")
		}
		r.OAuth2Email = email
		r.OAuth2TokenFile = tokenFile
		r.Auth = XOAuth2Auth(email, tokenFile)
	} else if u.User != nil {
		pass, _ := u.User.Password()
		user := u.User.Username()

		if hasAuth && authVal == "login" {
			r.Auth = LoginAuth(user, pass)
		} else if hasAuth {
			return nil, fmt.Errorf("Auth must be login, xoauth2, or not present, received '%s'", authVal)
		} else {
			r.Auth = smtp.PlainAuth("", user, pass, u.Hostname())
		}
	}

	if hasVal, skipVerify := q.Has("skipVerify"), q.Get("skipVerify"); hasVal && skipVerify != "false" {
		r.SkipVerify = true
	}

	if u.Path != "" {
		r.Sender = u.Path[1:]
	}

	return r, nil
}
