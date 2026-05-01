package webauthnserver

import (
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
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

// Mode controls whether this is a registration or authentication flow.
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

func (u user) WebAuthnName() string        { return u.cfg.Username }
func (u user) WebAuthnDisplayName() string { return u.cfg.Username }
func (u user) WebAuthnCredentials() []webauthnlib.Credential {
	return u.cfg.Credentials
}

type runner struct {
	cfg       *config.Config
	mode      Mode
	mobileURL string
	wa        *webauthnlib.WebAuthn
	errCh     chan error
	doneCh    chan struct{}
	sessionMu sync.Mutex
	session   *webauthnlib.SessionData
}

// Run starts the local WebAuthn HTTP(S) server. When a LAN IP is detected it
// automatically generates a self-signed TLS certificate and serves HTTPS so
// mobile devices can complete the WebAuthn ceremony (HTTP on non-localhost is
// not a secure context and browsers will refuse WebAuthn).
func Run(ctx context.Context, cfg *config.Config, mode Mode) error {
	if mode == ModeLogin && len(cfg.Credentials) == 0 {
		return errors.New("no passkey enrolled yet — run `passkey-sudo enroll` first")
	}

	var (
		listenAddr string
		targetURL  string
		mobileURL  string
		tlsCert    *tls.Certificate
		rpID       string
		rpOrigin   string
	)

	lanIP := detectLANIP()

	if lanIP != nil {
		ipStr := lanIP.String()
		cert, err := generateSelfSignedCert(
			[]net.IP{lanIP, net.ParseIP("127.0.0.1"), net.IPv6loopback},
			[]string{"localhost"},
		)
		if err == nil {
			tlsCert = &cert
			rpID = ipStr
			rpOrigin = fmt.Sprintf("https://%s:%s", ipStr, defaultPort)
			mobileURL = rpOrigin + "/"
			listenAddr = "0.0.0.0:" + defaultPort
			// Laptop opens via localhost so cert is trusted for "localhost" SAN.
			targetURL = fmt.Sprintf("https://localhost:%s/", defaultPort)
			// Persist so `passkey-sudo run` / `check` stay in sync.
			if cfg.RPID != rpID || cfg.RPOrigin != rpOrigin {
				cfg.RPID = rpID
				cfg.RPOrigin = rpOrigin
				_ = config.SaveDefault(cfg)
			}
		}
	}

	// Fallback: plain HTTP on localhost when no LAN or cert error.
	if tlsCert == nil {
		rpID = "localhost"
		rpOrigin = fmt.Sprintf("http://localhost:%s", defaultPort)
		listenAddr = "localhost:" + defaultPort
		targetURL = rpOrigin + "/"
		mobileURL = ""
	}

	wa, err := webauthnlib.New(&webauthnlib.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          rpID,
		RPOrigins:     []string{rpOrigin},
	})
	if err != nil {
		return fmt.Errorf("init webauthn: %w", err)
	}

	r := &runner{
		cfg:       cfg,
		mode:      mode,
		mobileURL: mobileURL,
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

	if tlsCert != nil {
		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{*tlsCert},
			MinVersion:   tls.VersionTLS12,
		}
		go func() {
			if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
				r.errCh <- err
			}
		}()
		printMobileMode(targetURL, mobileURL, mode)
	} else {
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				r.errCh <- err
			}
		}()
		fmt.Printf("Open %s to continue %s flow (laptop only)\n", targetURL, mode)
	}

	if cfg.OpenBrowserOnPrompt {
		_ = execx.OpenBrowser(targetURL)
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

func printMobileMode(laptopURL, mobileURL string, mode Mode) {
	fmt.Println()
	fmt.Println("  ┌─ Passkey-Sudo ───────────────────────────────────────────┐")
	fmt.Printf("  │  Mode    : %s\n", mode)
	fmt.Printf("  │  Laptop  : %s\n", laptopURL)
	if mobileURL != "" {
		fmt.Printf("  │  Mobile  : %s\n", mobileURL)
	}
	fmt.Println("  │  ⚠  Self-signed cert — browser will show a security warning.")
	fmt.Println("  │     Click  Advanced → Proceed  to continue (only needed once).")
	fmt.Println("  └──────────────────────────────────────────────────────────┘")
	fmt.Println()
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
		"mobile_url":   r.mobileURL,
		"mobile_ready": r.mobileURL != "",
	})
}

func (r *runner) handleQR(w http.ResponseWriter, _ *http.Request) {
	if r.mobileURL == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "QR not available in localhost mode"})
		return
	}
	png, err := qrcode.Encode(r.mobileURL, qrcode.Medium, 280)
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

	r.sessionMu.Lock()
	defer r.sessionMu.Unlock()

	var (
		opts interface{}
		err  error
	)
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session missing — refresh the page"})
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
		if _, err := r.wa.FinishLogin(u, *session, req); err != nil {
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
