package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// leer el header de auth
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.logger.Debugw("Basic auth: missing authorization header", "path", r.URL.Path)
				app.unauthorizedBasicError(w, r, fmt.Errorf("Missing authorization header"))
				return
			}

			// parseaarlo => obtener el base64
			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Basic" {
				app.logger.Debugw("Basic auth: invalid authorization header format", "path", r.URL.Path, "header", authHeader)
				app.unauthorizedBasicError(w, r, fmt.Errorf("Invalid authorization header"))
				return
			}

			// decodificar el base64
			decoded, err := base64.StdEncoding.DecodeString(headerParts[1])
			if err != nil {
				app.logger.Debugw("Basic auth: failed to decode base64", "path", r.URL.Path, "error", err)
				app.unauthorizedBasicError(w, r, fmt.Errorf("Invalid authorization header"))
				return
			}

			// checkear credenciales
			username := app.config.auth.basic.username
			password := app.config.auth.basic.password

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 {
				app.logger.Debugw("Basic auth: invalid credentials format", "path", r.URL.Path, "decoded", string(decoded))
				app.unauthorizedBasicError(w, r, fmt.Errorf("Invalid credentials"))
				return
			}

			if creds[0] != username || creds[1] != password {
				app.logger.Debugw("Basic auth: invalid credentials",
					"path", r.URL.Path,
					"provided_username", creds[0],
					"expected_username", username)
				app.unauthorizedBasicError(w, r, fmt.Errorf("Invalid credentials"))
				return
			}

			app.logger.Debugw("Basic auth: authentication successful", "path", r.URL.Path, "username", username)
			next.ServeHTTP(w, r)
		})
	}
}
