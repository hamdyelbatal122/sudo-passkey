package webauthnserver

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	webauthnlib "github.com/go-webauthn/webauthn/webauthn"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/hamdyelbatal122/sudo-passkey/internal/config"
	"github.com/hamdyelbatal122/sudo-passkey/internal/execx"
)

//go:embed flow.html
var embeddedFlowPage []byte

type Mode string

const (
	ModeRegister Mode = "register"
	ModeLogin    Mode = "login"
)

const defaultPort = "14141"

type user struct {
	cfg *config.Config
}

func (u user) WebAuthnID() []byte {
	decoded, err := base64.RawURLEncoding.DecodeString(u.cfg.UserID)
	if err == nil && len(decoded) > 0 {
		return decoded
	}
	return []byte(u.cfg.UserID)
}

func (u user) WebAuthnName() string {
	return u.cfg.Username
}

func (u user) WebAuthnDisplayName() string {
	return u.cfg.Username
}

func (u user) WebAuthnCredentials() []webauthnlib.Credential {
	return u.cfg.Credentials
}

type servePlan struct {
	listenAddr  string
	landingURL  string
	mobileURL   string
	mobileReady bool
	mobileHint  string
	rpID        string
	rpOrigin    string
}

const sessionTTL = 3 * time.Minute

type sessionEntry struct {
	data      *webauthnlib.SessionData
	expiresAt time.Time
}

type runner struct {
	cfg        *config.Config
	mode       Mode
	landingURL string
	mobileURL  string
	mobileHint string
	wa         *webauthnlib.WebAuthn
	errCh      chan error
	doneCh     chan struct{}
	sessionMu  sync.Mutex
	sessions   map[string]*sessionEntry
}

func Run(ctx context.Context, cfg *config.Config, mode Mode) error {
	if mode == ModeLogin && len(cfg.Credentials) == 0 {
		return errors.New("no passkey enrolled yet. run `passkey-sudo enroll` first")
	}

	plan := buildServePlan()

	wa, err := webauthnlib.New(&webauthnlib.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          plan.rpID,
		RPOrigins:     []string{plan.rpOrigin},
	})
	if err != nil {
		return fmt.Errorf("init webauthn: %w", err)
	}

	if cfg.RPID != plan.rpID || cfg.RPOrigin != plan.rpOrigin {
		cfg.RPID = plan.rpID
		cfg.RPOrigin = plan.rpOrigin
		_ = config.SaveDefault(cfg)
	}

	r := &runner{
		cfg:        cfg,
		mode:       mode,
		landingURL: plan.landingURL,
		mobileURL:  plan.mobileURL,
		mobileHint: plan.mobileHint,
		wa:         wa,
		errCh:      make(chan error, 1),
		doneCh:     make(chan struct{}, 1),
		sessions:   make(map[string]*sessionEntry),
	}

	mux := http.NewServeMux()
	r.registerRoutes(mux)

	srv := &http.Server{
		Addr:              plan.listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.errCh <- err
		}
	}()

	fmt.Printf("Open %s to continue %s flow\n", plan.landingURL, mode)
	if plan.mobileReady {
		fmt.Printf("Mobile QR target: %s\n", plan.mobileURL)
	} else {
		fmt.Printf("Mobile mode: %s\n", plan.mobileHint)
	}

	if cfg.OpenBrowserOnPrompt {
		_ = execx.OpenBrowser(plan.landingURL)
	}

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-r.doneCh:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return nil
	case err := <-r.errCh:
		return err
	case <-sigCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return errors.New("interrupted")
	}
}

func buildServePlan() servePlan {
	localhostOrigin := "http://localhost:" + defaultPort
	plan := servePlan{
		listenAddr:  "0.0.0.0:" + defaultPort,
		landingURL:  localhostOrigin + "/",
		mobileURL:   "",
		mobileReady: false,
		mobileHint:  "To use phone passkey, start a trusted HTTPS tunnel (for example ngrok), then refresh this page.",
		rpID:        "localhost",
		rpOrigin:    localhostOrigin,
	}

	origin, hint := detectTrustedPublicOrigin()
	if origin == "" {
		if hint != "" {
			plan.mobileHint = hint
		}
		return plan
	}

	u, err := url.Parse(origin)
	if err != nil {
		return plan
	}
	host := u.Hostname()
	if host == "" {
		return plan
	}
	plan.landingURL = strings.TrimRight(origin, "/") + "/"
	plan.mobileURL = plan.landingURL
	plan.mobileReady = true
	plan.mobileHint = "Phone passkey is ready. Scan QR and continue on your phone."
	plan.rpID = host
	plan.rpOrigin = strings.TrimRight(origin, "/")
	return plan
}

func detectTrustedPublicOrigin() (string, string) {
	if env := strings.TrimSpace(os.Getenv("PASSKEY_SUDO_PUBLIC_HTTPS_ORIGIN")); env != "" {
		u, err := url.Parse(env)
		if err != nil {
			return "", "Public HTTPS origin value is invalid."
		}
		if u.Scheme != "https" || !isPublicHost(u.Hostname()) {
			return "", "Public HTTPS origin must use a trusted https host."
		}
		return strings.TrimRight(env, "/"), ""
	}

	origin, err := detectNgrokHTTPSOrigin()
	if err == nil && origin != "" {
		return origin, ""
	}
	return "", "Phone passkey needs trusted HTTPS tunnel. Start ngrok and refresh."
}

func detectNgrokHTTPSOrigin() (string, error) {
	client := &http.Client{Timeout: 900 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:4040/api/tunnels")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ngrok api status %d", resp.StatusCode)
	}

	var payload struct {
		Tunnels []struct {
			PublicURL string `json:"public_url"`
			Proto     string `json:"proto"`
			Config    struct {
				Addr string `json:"addr"`
			} `json:"config"`
		} `json:"tunnels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	for _, t := range payload.Tunnels {
		if !strings.HasPrefix(t.PublicURL, "https://") {
			continue
		}
		addr := strings.ToLower(t.Config.Addr)
		if strings.Contains(addr, ":14141") || strings.Contains(addr, "localhost:14141") || strings.Contains(addr, "127.0.0.1:14141") {
			return strings.TrimRight(t.PublicURL, "/"), nil
		}
	}
	return "", errors.New("no https tunnel for :14141")
}

func isPublicHost(host string) bool {
	if host == "" || host == "localhost" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return !ip.IsLoopback() && !ip.IsPrivate()
	}
	return true
}

func (r *runner) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", r.handleIndex)
	mux.HandleFunc("/api/meta", r.handleMeta)
	mux.HandleFunc("/qr.png", r.handleQR)
	mux.HandleFunc("/api/begin", r.handleBegin)
	mux.HandleFunc("/api/finish", r.handleFinish)
}

func (r *runner) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(embeddedFlowPage)
}

func (r *runner) handleMeta(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"landing_url":  r.landingURL,
		"mobile_url":   r.mobileURL,
		"mobile_ready": r.mobileURL != "",
		"mobile_hint":  r.mobileHint,
		"mode":         string(r.mode),
	})
}

func (r *runner) handleQR(w http.ResponseWriter, _ *http.Request) {
	if r.mobileURL == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "trusted mobile URL not available"})
		return
	}
	png, err := qrcode.Encode(r.mobileURL, qrcode.Medium, 256)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(png)
}

func (r *runner) handleBegin(w http.ResponseWriter, _ *http.Request) {
	u := user{cfg: r.cfg}
	var (
		opts    interface{}
		session *webauthnlib.SessionData
		err     error
	)

	switch r.mode {
	case ModeRegister:
		opts, session, err = r.wa.BeginRegistration(u)
	case ModeLogin:
		opts, session, err = r.wa.BeginLogin(u)
	default:
		err = fmt.Errorf("unsupported mode: %s", r.mode)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	token, err := randomSessionToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	r.sessionMu.Lock()
	r.sessions[token] = &sessionEntry{data: session, expiresAt: time.Now().Add(sessionTTL)}
	r.pruneExpiredSessions()
	r.sessionMu.Unlock()
	w.Header().Set("X-Session-Token", token)
	writeJSON(w, http.StatusOK, opts)
}

func (r *runner) handleFinish(w http.ResponseWriter, req *http.Request) {
	u := user{cfg: r.cfg}
	token := strings.TrimSpace(req.Header.Get("X-Session-Token"))
	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session token"})
		return
	}

	r.sessionMu.Lock()
	entry := r.sessions[token]
	delete(r.sessions, token)
	r.sessionMu.Unlock()
	if entry == nil || time.Now().After(entry.expiresAt) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session expired. refresh and retry"})
		return
	}
	session := entry.data

	switch r.mode {
	case ModeRegister:
		credential, err := r.wa.FinishRegistration(u, *session, req)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		r.cfg.Credentials = append(r.cfg.Credentials, *credential)
		if err := config.Save(config.DefaultPath(), r.cfg); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	case ModeLogin:
		_, err := r.wa.FinishLogin(u, *session, req)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported mode"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	select {
	case r.doneCh <- struct{}{}:
	default:
	}
}

// pruneExpiredSessions must be called with sessionMu held.
func (r *runner) pruneExpiredSessions() {
	now := time.Now()
	for k, e := range r.sessions {
		if now.After(e.expiresAt) {
			delete(r.sessions, k)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func randomSessionToken() (string, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
