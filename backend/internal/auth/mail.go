package auth

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// Mailer sends the OTP email.
type Mailer interface {
	SendOTP(ctx context.Context, to, code string) error
}

// LogMailer logs OTP emails instead of sending them (development / no SMTP).
type LogMailer struct{}

// SendOTP logs the code.
func (LogMailer) SendOTP(_ context.Context, to, code string) error {
	slog.Info("OTP email (not sent, logged)", "to", to, "code", code)
	return nil
}

// DevMailer logs the code and also tries SMTP, so email setup can be tested in development
// without ever blocking sign-in when the mail server is misconfigured.
type DevMailer struct{ SMTP *SMTPMailer }

// SendOTP logs the code, then attempts real delivery; delivery errors are only logged.
func (m DevMailer) SendOTP(ctx context.Context, to, code string) error {
	_ = LogMailer{}.SendOTP(ctx, to, code)
	if err := m.SMTP.SendOTP(ctx, to, code); err != nil {
		slog.Warn("SMTP delivery failed (development, ignored)", "to", to, "err", err)
	}
	return nil
}

// SMTPMailer sends mail over SMTP with STARTTLS (587) or implicit TLS (465).
type SMTPMailer struct {
	Host, Username, Password, From string
	Port                           int
}

func otpBody(code string) (subject, body string) {
	subject = "Your Sawyun sign-in code: " + code
	body = fmt.Sprintf(`Your Sawyun Ecommerce sign-in code is:

    %s

It expires in 5 minutes. If you did not request this code, you can ignore this email.

---

Sawyun Ecommerce သို့ ဝင်ရောက်ရန် ကုဒ်နံပါတ်မှာ -

    %s

ဤကုဒ်သည် ၅ မိနစ်အတွင်း သက်တမ်းကုန်ဆုံးပါမည်။ သင်မတောင်းဆိုခဲ့ပါက ဤအီးမေးလ်ကို လျစ်လျူရှုနိုင်ပါသည်။
`, code, code)
	return subject, body
}

// SendOTP sends the bilingual OTP email.
func (m *SMTPMailer) SendOTP(ctx context.Context, to, code string) error {
	subject, body := otpBody(code)
	return m.Send(ctx, to, subject, body)
}

// Send delivers a plain-text UTF-8 email.
func (m *SMTPMailer) Send(ctx context.Context, to, subject, body string) error {
	from, err := mail.ParseAddress(m.From)
	if err != nil {
		return fmt.Errorf("invalid SMTP_FROM: %w", err)
	}
	addr := net.JoinHostPort(m.Host, strconv.Itoa(m.Port))
	deadline := time.Now().Add(20 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	dialer := &net.Dialer{Deadline: deadline}
	tlsCfg := &tls.Config{ServerName: m.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	if m.Port == 465 {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(deadline)
	c, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()
	if m.Port != 465 {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsCfg); err != nil {
				return err
			}
		} else if m.Host != "localhost" && m.Host != "127.0.0.1" {
			return fmt.Errorf("smtp server %s does not support STARTTLS", m.Host)
		}
	}
	if m.Username != "" {
		if err := c.Auth(smtp.PlainAuth("", m.Username, m.Password, m.Host)); err != nil {
			return err
		}
	}
	if err := c.Mail(from.Address); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write(buildMessage(from, to, subject, body)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func buildMessage(from *mail.Address, to, subject, body string) []byte {
	var b strings.Builder
	h := func(k, v string) { b.WriteString(k + ": " + v + "\r\n") }
	h("From", from.String())
	h("To", to)
	h("Subject", mime.BEncoding.Encode("UTF-8", subject))
	h("Date", time.Now().Format(time.RFC1123Z))
	domain := "sawyun.local"
	if i := strings.LastIndexByte(from.Address, '@'); i >= 0 {
		domain = from.Address[i+1:]
	}
	h("Message-ID", fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), strconv.Itoa(len(to)), domain))
	h("MIME-Version", "1.0")
	h("Content-Type", "text/plain; charset=UTF-8")
	h("Content-Transfer-Encoding", "base64")
	b.WriteString("\r\n")
	enc := base64.StdEncoding.EncodeToString([]byte(body))
	for len(enc) > 76 {
		b.WriteString(enc[:76] + "\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc + "\r\n")
	return []byte(b.String())
}
