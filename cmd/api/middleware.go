package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/marceterrone10/social/internal/store"
)

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.logger.Debugw("missing authorization header", "path", r.URL.Path)
			app.unauthorizedError(w, r, fmt.Errorf("Missing authorization header"))
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.logger.Debugw("invalid authorization header format", "path", r.URL.Path, "header", authHeader)
			app.unauthorizedError(w, r, fmt.Errorf("Invalid authorization header"))
			return
		}
		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("Token is invalid"))
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.getUserFromCache(ctx, userID)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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

func (app *application) CheckPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r.Context())
		post := getPostFromCtx(r.Context())
		// chequear si el usuario es el propietario del post
		if post.UserID != user.ID {
			next.ServeHTTP(w, r)
			return
		}

		// chequear si el usuario tiene el rol necesario
		allowedRoles, err := app.checkRole(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowedRoles {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRole(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}
	return user.Role.Level >= role.Level, nil
}

func (app *application) getUserFromCache(ctx context.Context, userID int64) (*store.User, error) {
	user, err := app.cacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err := app.store.Users.GetById(ctx, userID)
		if err != nil {
			return nil, err
		}

		if err := app.cacheStorage.Users.Set(ctx, user); err != nil {
			return nil, err
		}

	}

	return user, nil
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.tooManyRequestsError(w, r, retryAfter)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
