# smtprelay

[![Go Report Card](https://goreportcard.com/badge/github.com/decke/smtprelay)](https://goreportcard.com/report/github.com/decke/smtprelay)
[![OpenSSF Scorecard](https://img.shields.io/ossf-scorecard/github.com/decke/smtprelay?label=openssf%20scorecard&style=flat)](https://scorecard.dev/viewer/?uri=github.com/decke/smtprelay)

Simple Golang based SMTP relay/proxy server that accepts mail via SMTP
and forwards it directly to another SMTP server.


## Why another SMTP server?

Outgoing mails are usually send via SMTP to an MTA (Mail Transfer Agent)
which is one of Postfix, Exim, Sendmail or OpenSMTPD on UNIX/Linux in most
cases. You really don't want to setup and maintain any of those full blown
kitchensinks yourself because they are complex, fragile and hard to
configure.

My use case is simple. I need to send automatically generated mails from
cron via msmtp/sSMTP/dma, mails from various services and network printers
via a remote SMTP server without giving away my mail credentials to each
device which produces mail.


## Main features

* Simple configuration with ini file .env file or environment variables
* Supports SMTPS/TLS (465), STARTTLS (587) and unencrypted SMTP (25)
* Checks for sender, receiver, client IP
* Authentication support with file (LOGIN, PLAIN)
* Enforce encryption for authentication
* Forwards all mail to a smarthost (any SMTP server)
* Small codebase
* IPv6 support
* Aliases support (dynamic reload when alias file changes)
* Web UI for live configuration editing
* Gmail OAuth2 / XOAUTH2 support with browser-based authorization and automatic token refresh


## Web UI

smtprelay includes an optional embedded web server for viewing and editing configuration through a browser. Enable it with three config options:

```ini
web_listen   = 127.0.0.1:8080
web_username = admin
web_password = secret
```

Or via flags: `--web_listen 127.0.0.1:8080 --web_username admin --web_password secret`

Opening `http://127.0.0.1:8080` in a browser prompts for the credentials and then shows a form with all relay settings. Saving the form writes a new config file and restarts the process automatically.

**Note:** The `--config` flag must point to a writable INI file for saving to work. If smtprelay was started without `--config`, the web UI is read-only.


## Gmail OAuth2

Google requires OAuth2 (XOAUTH2) for SMTP access instead of plain passwords. smtprelay supports this natively.

### 1. Create Google OAuth2 credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com/) and create a project.
2. Enable the **Gmail API**.
3. Under **APIs & Services → Credentials**, create an **OAuth 2.0 Client ID** of type *Desktop app*.
4. Note the **Client ID** and **Client Secret**.
5. Add `http://127.0.0.1:8080/oauth2/callback` (replace port with your `web_listen`) to the list of **Authorized redirect URIs**.

### 2. Configure smtprelay

Add the credentials to your config file:

```ini
oauth2_client_id     = 123456789-abc.apps.googleusercontent.com
oauth2_client_secret = GOCSPX-...
oauth2_redirect_url  = http://myhost:8080/oauth2/callback
```

The web UI must also be enabled (see above) because the authorization flow runs through it.

### 3. Authorize a Gmail account

1. Open `http://<web_listen>/oauth2` in your browser.
2. Enter the Gmail address and a path where the token file should be saved (e.g. `/etc/smtprelay/gmail.json`).
3. Click **Authorize with Google** and complete the Google consent screen.
4. smtprelay writes the token file (access token + refresh token) and updates the config file with `oauth2_email` and `oauth2_token_file` automatically.
5. The relay restarts and a Gmail remote (`smtps://smtp.gmail.com:465`) is configured for you — no manual remote URL needed.

### Token refresh

smtprelay checks the token expiry before every outgoing connection. If the access token has expired it uses the stored refresh token to obtain a new one and updates the token file on disk — no manual intervention required.