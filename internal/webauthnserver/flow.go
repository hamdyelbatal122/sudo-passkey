package webauthnserver

import (
	"context"
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

type runner struct {
	cfg       *config.Config
	mode      Mode
	targetURL string
	wa        *webauthnlib.WebAuthn
	errCh     chan error
	doneCh    chan struct{}
	sessionMu sync.Mutex
	session   *webauthnlib.SessionData
}

func Run(ctx context.Context, cfg *config.Config, mode Mode) error {
	if mode == ModeLogin && len(cfg.Credentials) == 0 {
		return errors.New("no passkey enrolled yet. run `passkey-sudo enroll` first")
	}

	wa, err := webauthnlib.New(&webauthnlib.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          cfg.RPID,
		RPOrigins:     []string{cfg.RPOrigin},
	})
	if err != nil {
		return fmt.Errorf("init webauthn: %w", err)
	}

	listenAddr, targetURL, err := resolveAddress(cfg.RPOrigin)
	if err != nil {
		return err
	}

	r := &runner{
		cfg:       cfg,
		mode:      mode,
		targetURL: targetURL,
		wa:        wa,
		errCh:     make(chan error, 1),
		doneCh:    make(chan struct{}, 1),
	}

	mux := http.NewServeMux()
	r.registerRoutes(mux)

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.errCh <- err
		}
	}()

	if cfg.OpenBrowserOnPrompt {
		_ = execx.OpenBrowser(targetURL)
	}

	fmt.Printf("Open %s to continue %s flow\n", targetURL, mode)
	if !isMobileReady(targetURL) {
		fmt.Println("Tip: For mobile passkeys, use your laptop LAN IP/hostname as both rp-id and rp-origin host.")
		fmt.Println("Example: passkey-sudo settings set rp-id 192.168.1.10")
		fmt.Println("Example: passkey-sudo settings set rp-origin http://192.168.1.10:14141")
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

func resolveAddress(origin string) (listenAddr string, targetURL string, err error) {
	u, err := url.Parse(origin)
	if err != nil {
		return "", "", fmt.Errorf("invalid rp-origin %q: %w", origin, err)
	}
	if u.Host == "" {
		return "", "", fmt.Errorf("invalid rp-origin %q: missing host", origin)
	}
	if !strings.Contains(u.Host, ":") {
		if u.Scheme == "https" {
			u.Host += ":443"
		} else {
			u.Host += ":80"
		}
	}
	return u.Host, strings.TrimRight(origin, "/") + "/", nil
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
		"target_url":   r.targetURL,
		"mobile_ready": isMobileReady(r.targetURL),
	})
}

func (r *runner) handleQR(w http.ResponseWriter, _ *http.Request) {
	png, err := qrcode.Encode(r.targetURL, qrcode.Medium, 256)
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
		opts interface{}
		err  error
	)

	r.sessionMu.Lock()
	defer r.sessionMu.Unlock()

	switch r.mode {
	case ModeRegister:
		opts, r.session, err = r.wa.BeginRegistration(u)
	case ModeLogin:
		opts, r.session, err = r.wa.BeginLogin(u)
	default:
		err = fmt.Errorf("unsupported mode: %s", r.mode)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, opts)
}

func (r *runner) handleFinish(w http.ResponseWriter, req *http.Request) {
	u := user{cfg: r.cfg}
	r.sessionMu.Lock()
	session := r.session
	r.sessionMu.Unlock()
	if session == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session"})
		return
	}

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

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func isMobileReady(target string) bool {
	u, err := url.Parse(target)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "" {
		return false
	}
	if host == "localhost" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return false
	}
	return true
}
