# gmailrelay

A small SMTP relay that accepts local mail and forwards it to Gmail over
XOAUTH2. Fork of [smtprelay](https://github.com/decke/smtprelay), trimmed
and reshaped around a single outbound path (`smtp.gmail.com:465`).

Authorization is handled by a built-in web UI: paste your Google OAuth2
client credentials, click through the consent screen, and the resulting
refresh token is written to disk. The relay takes care of token refresh
automatically on every outgoing connection.


## Use case

Daemons, cron jobs, routers, NAS boxes, and assorted devices that want to
send mail via `sendmail`/`msmtp`/`sSMTP`/`dma` to a local SMTP endpoint,
without each of them holding Gmail credentials or implementing OAuth2.
Point them at gmailrelay on localhost; gmailrelay handles Gmail.


## Features

* Listens for SMTP, STARTTLS, and SMTPS from local clients
* Outbound to Gmail via SMTPS/XOAUTH2, with automatic token refresh
* Browser-based OAuth2 authorization flow
* Web UI for live configuration editing (writes config + restarts)
* Allow/deny by client network, sender regex, recipient regex
* Optional local SMTP AUTH (LOGIN/PLAIN) backed by a bcrypt password file
* Aliases file with live reload
* Optional pipe-to-command delivery in addition to (or instead of) Gmail
* Configuration via ini file, `.env`, or `GMAILRELAY_*` environment variables


## Install and run

```sh
go install github.com/chbmuc/gmailrelay@latest
gmailrelay --config /etc/gmailrelay/gmailrelay.ini
```

A sample `gmailrelay.ini` is shipped with each release. Without `--config`
the process still runs but the web UI cannot persist changes.


## Web UI

Enable the web UI with three config options:

```ini
web_listen   = 127.0.0.1:8080
web_username = admin
web_password = secret
```

Or via flags: `--web_listen 127.0.0.1:8080 --web_username admin --web_password secret`

Open `http://127.0.0.1:8080/` and authenticate. The form exposes every
config key; saving writes a fresh ini file and re-execs the process so
changes take effect immediately.


## Gmail OAuth2 setup

### 1. Create Google OAuth2 credentials

1. In the [Google Cloud Console](https://console.cloud.google.com/), create (or pick) a project.
2. Enable the **Gmail API** under **APIs & Services → Library**.
3. Under **APIs & Services → Credentials**, create an **OAuth 2.0 Client ID** of type *Web application*.
4. Add your redirect URL (e.g. `https://myhost:8080/oauth2/callback`) to **Authorized redirect URIs**. This value must match `oauth2_redirect_url` exactly (Google accepts http-only reditect URLs only for 127.0.0.1).
5. Note the generated **Client ID** and **Client Secret**.

### 2. Configure credentials

Either edit the config file directly:

```ini
oauth2_client_id     = 123456789-abc.apps.googleusercontent.com
oauth2_client_secret = GOCSPX-...
oauth2_redirect_url  = http://myhost:8080/oauth2/callback
```

…or fill the same fields in the web UI and click **Save & Restart**.

### 3. Authorize a Gmail account

1. On the config page, fill in **oauth2_email** (the Gmail address to send as) and **oauth2_token_file** (where to persist the token, e.g. `/etc/gmailrelay/gmail.json`).
2. Click **Authorize with Google →** and complete the consent screen.
3. gmailrelay writes the token file, updates the config with `oauth2_email` / `oauth2_token_file`, and restarts.
4. A Gmail remote at `smtps://smtp.gmail.com:465` is configured automatically

### Token refresh

Before every outgoing connection gmailrelay checks the token expiry. If the
access token has expired it uses the stored refresh token to obtain a new
one and rewrites the token file on disk. No manual intervention required.


## Pipe command

If `command` is set, every accepted message is also piped to that program
on stdin. The following environment variables are exported for the child:

| Variable            | Contents                    |
|---------------------|-----------------------------|
| `GMAILRELAY_FROM`   | Envelope sender             |
| `GMAILRELAY_TO`     | Envelope recipients         |
| `GMAILRELAY_PEER`   | Client IP                   |


## License

See [LICENSE](LICENSE).
