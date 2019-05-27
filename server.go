package apcore

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	_ "github.com/gorilla/schema"
)

type server struct {
	a           Application
	handler     *handler
	db          *database
	sessions    *sessions
	config      *config
	httpServer  *http.Server
	httpsServer *http.Server
	debug       bool
}

func newServer(configFileName string, a Application, debug bool) (s *server, err error) {
	// Load the configuration
	var c *config
	c, err = loadConfigFile(configFileName, a)
	if err != nil {
		return
	}

	// Connect to database
	var db *database
	db, err = newDatabase(c, a)
	if err != nil {
		return
	}

	// Build application routes
	var h *handler
	h, err = newHandler(c, a, debug)
	if err != nil {
		return
	}

	// Prepare sessions
	var ses *sessions
	ses, err = newSessions(c)
	if err != nil {
		return
	}

	// Prepare HTTPS server. No option to run the server as HTTP, because
	// we're living in the future.
	httpsServer := &http.Server{
		Addr:         ":https",
		Handler:      h.Handler(),
		ReadTimeout:  time.Duration(c.ServerConfig.HttpsReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(c.ServerConfig.HttpsWriteTimeoutSeconds) * time.Second,
		TLSConfig:    createTlsConfig(),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	// Prepare redirection server. HTTP is only allowed to redirect.
	httpServer := createRedirectServer(c)

	// Create the apcore server
	s = &server{
		a:           a,
		handler:     h,
		db:          db,
		sessions:    ses,
		config:      c,
		httpServer:  httpServer,
		httpsServer: httpsServer,
		debug:       debug,
	}

	// Post-creation hooks
	httpsServer.RegisterOnShutdown(s.stop)
	return
}

// Do not let clients downgrade connections to use insecure, older
// cryptographic functions or curves.
func createTlsConfig() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.X25519},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}

func createRedirectServer(c *config) *http.Server {
	return &http.Server{
		Addr:         ":http",
		ReadTimeout:  time.Duration(c.ServerConfig.RedirectReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(c.ServerConfig.RedirectWriteTimeoutSeconds) * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			http.Redirect(w, req, fmt.Sprintf("https://%s%s", req.Host, req.URL), http.StatusMovedPermanently)
		}),
	}
}

func (s *server) start() {
	// TODO
}

func (s *server) stop() {
	// TODO
}
