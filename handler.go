package apcore

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gorilla/mux"
)

type handler struct {
	router *mux.Router
}

func newHandler(c *config, a Application, debug bool) (h *handler, err error) {
	r := mux.NewRouter()
	r.NotFoundHandler = a.NotFoundHandler()
	r.MethodNotAllowedHandler = a.MethodNotAllowedHandler()

	// Static assets
	if sd := c.ServerConfig.StaticRootDirectory; len(sd) == 0 {
		err = fmt.Errorf("static_root_directory is empty")
		return
	} else {
		InfoLogger.Infof("Serving static directory: %s", sd)
		fs := http.FileServer(http.Dir(sd))
		r.PathPrefix("/").Handler(fs)
	}

	// TODO: Webfinger
	// TODO: Node-info
	// TODO: Host-meta
	// TODO: Application-specific routes

	if debug {
		InfoLogger.Info("Adding request logging middleware for debugging")
		r.Use(requestLogger)
		InfoLogger.Info("Adding request timing middleware for debugging")
		r.Use(timingLogger)
	}

	h = &handler{
		router: r,
	}
	return
}

func (h handler) Handler() http.Handler {
	return h.router
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprintf("requestLogger debugging middleware failure: %s", err), http.StatusInternalServerError)
			return
		}
		InfoLogger.Infof("%s", dump)
		next.ServeHTTP(w, r)
	})
}

func timingLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()
		InfoLogger.Info("%s took %s", r.URL, end.Sub(start))
	})
}
