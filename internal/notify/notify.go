package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config is loaded from environment.
// Email (SMTP):
//   NEXO_SMTP_HOST, NEXO_SMTP_PORT (default 587), NEXO_SMTP_USER, NEXO_SMTP_PASS
//   NEXO_NOTIFY_EMAIL (default lhmolam@gmail.com)
// SMS HTTP gateway (optional):
//   NEXO_SMS_URL  - endpoint that accepts JSON {"to","message"}
//   NEXO_SMS_TOKEN - optional Bearer token
//   NEXO_NOTIFY_PHONE (default +5350937508)
// Always writes a local log under data/notifications.log as audit trail.
type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	NotifyEmail  string
	SMSURL       string
	SMSToken     string
	NotifyPhone  string
	LogDir       string
}

func LoadConfig() Config {
	port := os.Getenv("NEXO_SMTP_PORT")
	if port == "" {
		port = "587"
	}
	email := os.Getenv("NEXO_NOTIFY_EMAIL")
	if email == "" {
		email = "lhmolam@gmail.com"
	}
	phone := os.Getenv("NEXO_NOTIFY_PHONE")
	if phone == "" {
		phone = "+5350937508"
	}
	dir := os.Getenv("NEXO_DATA_DIR")
	if dir == "" {
		dir = "data"
	}
	return Config{
		SMTPHost:    os.Getenv("NEXO_SMTP_HOST"),
		SMTPPort:    port,
		SMTPUser:    os.Getenv("NEXO_SMTP_USER"),
		SMTPPass:    os.Getenv("NEXO_SMTP_PASS"),
		NotifyEmail: email,
		SMSURL:      os.Getenv("NEXO_SMS_URL"),
		SMSToken:    os.Getenv("NEXO_SMS_TOKEN"),
		NotifyPhone: phone,
		LogDir:      dir,
	}
}

type Notifier struct {
	cfg Config
}

func New(cfg Config) *Notifier {
	return &Notifier{cfg: cfg}
}

// Event is a form submission notification.
type Event struct {
	Kind    string            // request | provider | abroad
	Subject string
	Body    string
	Fields  map[string]string
}

func (n *Notifier) Notify(ev Event) {
	// Always persist locally so nothing is lost.
	n.appendLog(ev)

	var errs []string
	if err := n.sendEmail(ev); err != nil {
		errs = append(errs, "email: "+err.Error())
	}
	if err := n.sendSMS(ev); err != nil {
		errs = append(errs, "sms: "+err.Error())
	}
	if len(errs) > 0 {
		log.Printf("[nexo-notify] partial failure: %s", strings.Join(errs, "; "))
	} else {
		log.Printf("[nexo-notify] delivered kind=%s subject=%q", ev.Kind, ev.Subject)
	}
}

func (n *Notifier) appendLog(ev Event) {
	_ = os.MkdirAll(n.cfg.LogDir, 0o755)
	path := filepath.Join(n.cfg.LogDir, "notifications.log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("[nexo-notify] log open: %v", err)
		return
	}
	defer f.Close()
	line := fmt.Sprintf("---- %s | %s | %s ----\n%s\n\n",
		time.Now().Format(time.RFC3339), ev.Kind, ev.Subject, ev.Body)
	if _, err := f.WriteString(line); err != nil {
		log.Printf("[nexo-notify] log write: %v", err)
	}
}

func (n *Notifier) sendEmail(ev Event) error {
	if n.cfg.SMTPHost == "" || n.cfg.SMTPUser == "" || n.cfg.SMTPPass == "" {
		// Not configured: skip silently (log already saved).
		return nil
	}
	from := n.cfg.SMTPUser
	to := []string{n.cfg.NotifyEmail}
	msg := bytes.Buffer{}
	msg.WriteString("From: NEXO <" + from + ">\r\n")
	msg.WriteString("To: " + n.cfg.NotifyEmail + "\r\n")
	msg.WriteString("Subject: " + sanitizeHeader(ev.Subject) + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(ev.Body)

	addr := n.cfg.SMTPHost + ":" + n.cfg.SMTPPort
	auth := smtp.PlainAuth("", n.cfg.SMTPUser, n.cfg.SMTPPass, n.cfg.SMTPHost)
	return smtp.SendMail(addr, auth, from, to, msg.Bytes())
}

func (n *Notifier) sendSMS(ev Event) error {
	if n.cfg.SMSURL == "" {
		return nil
	}
	// Keep SMS short.
	text := ev.Subject
	if len(ev.Body) > 0 {
		// Prefer phone + key lines
		lines := strings.Split(ev.Body, "\n")
		for _, ln := range lines {
			if strings.Contains(strings.ToLower(ln), "tel") || strings.Contains(strings.ToLower(ln), "nombre") {
				text += " | " + strings.TrimSpace(ln)
			}
		}
	}
	if len(text) > 300 {
		text = text[:300]
	}
	payload := map[string]string{
		"to":      n.cfg.NotifyPhone,
		"message": text,
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, n.cfg.SMSURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if n.cfg.SMSToken != "" {
		req.Header.Set("Authorization", "Bearer "+n.cfg.SMSToken)
	}
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sms gateway status %d", resp.StatusCode)
	}
	return nil
}

func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// Helpers to build bodies from domain data.

func RequestEvent(kind string, fields map[string]string) Event {
	subj := "NEXO · Nueva solicitud"
	if kind == "abroad" {
		subj = "NEXO · Solicitud desde el exterior"
	}
	var b strings.Builder
	b.WriteString("Nueva solicitud recibida en NEXO\n")
	b.WriteString("--------------------------------\n")
	order := []string{
		"id", "tipo", "servicio", "municipio", "fecha", "descripcion",
		"nombre", "telefono", "desde_exterior", "nombre_exterior", "contacto_exterior",
		"familiar", "tel_familiar",
	}
	for _, k := range order {
		if v, ok := fields[k]; ok && strings.TrimSpace(v) != "" {
			b.WriteString(k + ": " + v + "\n")
		}
	}
	// any remaining
	for k, v := range fields {
		found := false
		for _, o := range order {
			if o == k {
				found = true
				break
			}
		}
		if !found && strings.TrimSpace(v) != "" {
			b.WriteString(k + ": " + v + "\n")
		}
	}
	return Event{Kind: kind, Subject: subj, Body: b.String(), Fields: fields}
}

func ProviderEvent(fields map[string]string) Event {
	var b strings.Builder
	b.WriteString("Nuevo proveedor registrado en NEXO\n")
	b.WriteString("----------------------------------\n")
	for _, k := range []string{"id", "nombre", "telefono", "zona", "servicios", "experiencia"} {
		if v := fields[k]; strings.TrimSpace(v) != "" {
			b.WriteString(k + ": " + v + "\n")
		}
	}
	return Event{
		Kind:    "provider",
		Subject: "NEXO · Nuevo proveedor",
		Body:    b.String(),
		Fields:  fields,
	}
}
