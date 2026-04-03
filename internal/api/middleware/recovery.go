package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery middleware recovers from panics and returns 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic and stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())

				// Return 500 error to client
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				response := map[string]interface{}{
					"error":   "internal_server_error",
					"message": fmt.Sprintf("Internal server error: %v", err),
				}

				_ = json.NewEncoder(w).Encode(response) // nolint:errcheck // response already sent, can only log
			}
		}()

		next.ServeHTTP(w, r)
	})
}
