package gateway

import (
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Allow CORS from any origin. TODO: Fix me to restrict CORS.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				handler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// Handler that adds necessary headers. CORS from any origin using the methods "GET", "POST" (Not allowed, "HEAD, "PUT", "DELETE")
func handler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "POST"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
}

// Gateway health check. Returns ok if the grpc server connection is ready.
func healthServer(conn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if s := conn.GetState(); s != connectivity.Ready {
			http.Error(w, fmt.Sprintf("grpc server is %s", s), http.StatusBadGateway)
			return
		}
		fmt.Fprintln(w, "ok")
	}
}
