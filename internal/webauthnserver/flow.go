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
	listenAddr string
	landingURL string
	mobileURL  string
	rpID       string
	rpOrigin   string
	cert       *tls.Certificate
}

type runner struct {
	cfg        *config.Config
	mode       Mode
	landingURL string
	mobileURL  string
	rpID       string
	wa         *webauthnlib.WebAuthn
	errCh      chan error
	doneCh     chan struct{}
	sessionMu  sync.Mutex
	session    *webauthnlib.SessionData
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

	// Keep config aligned with active runtime host so future commands stay consistent.
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
		rpID:       plan.rpID,
		wa:         wa,
		errCh:      make(chan error, 1),
		doneCh:     make(chan struct{}, 1),
	}

	mux := http.NewServeMux()
	r.registerRoutes(mux)

	srv := &http.Server{
		Addr:              plan.listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if plan.cert != nil {
		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{*plan.cert},
			MinVersion:   tls.VersionTLS12,
		}
		go func() {
			if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
				r.errCh <- err
			}
		}()
		fmt.Printf("Open %s to continue %s flow\n", plan.landingURL, mode)
		fmt.Printf("Mobile QR target: %s\n", plan.mobileURL)
		fmt.Println("Note: this LAN TLS mode is experimental and may fail if certificate is not trusted.")
	} else {
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				r.errCh <- err
			}
		}()
		fmt.Printf("Open %s to continue %s flow (localhost mode)\n", plan.landingURL, mode)
		fmt.Println("Tip: use browser passkey option 'use a phone or tablet' for cross-device enrollment.")
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
	if os.Getenv("PASSKEY_SUDO_ENABLE_LAN_TLS") != "1" {
		lanIP := detectLANIP()
		fallbackOrigin := "http://localhost:" + defaultPort
		mobileURL := ""
		listenAddr := "localhost:" + defaultPort
		if lanIP != nil {
			ipStr := lanIP.String()
			// Expose the same flow page on LAN for QR scanning from phone.
			mobileURL = "http://" + ipStr + ":" + defaultPort + "/"
			listenAddr = "0.0.0.0:" + defaultPort
		}
		return servePlan{
			listenAddr: listenAddr,
			landingURL: fallbackOrigin + "/",
			mobileURL:  mobileURL,
			rpID:       "localhost",
			rpOrigin:   fallbackOrigin,
			cert:       nil,
		}
	}

	lanIP := detectLANIP()
	if lanIP != nil {
		ipStr := lanIP.String()
		cert, err := generateSelfSignedCert(
			[]net.IP{lanIP, net.ParseIP("127.0.0.1"), net.IPv6loopback},
			[]string{"localhost"},
		)
		if err == nil {
			rpOrigin := "https://" + ipStr + ":" + defaultPort
			return servePlan{
				listenAddr: "0.0.0.0:" + defaultPort,
				landingURL: rpOrigin + "/",
				mobileURL:  rpOrigin + "/",
				rpID:       ipStr,
				rpOrigin:   rpOrigin,
				cert:       &cert,
			}
		}
	}

	fallbackOrigin := "http://localhost:" + defaultPort
	return servePlan{
		listenAddr: "localhost:" + defaultPort,
		landingURL: fallbackOrigin + "/",
		mobileURL:  "",
		rpID:       "localhost",
		rpOrigin:   fallbackOrigin,
		cert:       nil,
	}
}

func (r *runner) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", r.handleIndex)
	mux.HandleFunc("/api/meta", r.handleMeta)
	mux.HandleFunc("/qr.png", r.handleQR)
	mux.HandleFunc("/api/begin", r.handleBegin)
	mux.HandleFunc("/api/finish", r.handleFinish)
}

func (r *runner) handleIndex(w http.ResponseWriter, req *http.Request) {
	if r.mobileURL != "" && shouldRedirectToRPHost(req.Host, r.rpID) {
		http.Redirect(w, req, r.mobileURL, http.StatusTemporaryRedirect)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(embeddedFlowPage)
}

func shouldRedirectToRPHost(hostPort string, rpID string) bool {
	if rpID == "localhost" {
		return false
	}
	host := hostPort
	if strings.Contains(hostPort, ":") {
		if h, _, err := net.SplitHostPort(hostPort); err == nil {
			host = h
		}
	}
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true
	}
	return host != rpID
}

func (r *runner) handleMeta(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"landing_url":  r.landingURL,
		"mobile_url":   r.mobileURL,
		"mobile_ready": r.mobileURL != "",
	})
}

func (r *runner) handleQR(w http.ResponseWriter, _ *http.Request) {
	if r.mobileURL == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "QR not available in localhost mode"})
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
