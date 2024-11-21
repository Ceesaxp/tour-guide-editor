package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

func SessionID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check for existing session ID in cookie
        cookie, err := r.Cookie("session_id")
        var sessionID string

        if err == http.ErrNoCookie {
            // Generate new session ID
            sessionID = generateSessionID()

            // Set cookie
            http.SetCookie(w, &http.Cookie{
                Name:     "session_id",
                Value:    sessionID,
                Path:     "/",
                HttpOnly: true,
                Secure:   r.TLS != nil,
                SameSite: http.SameSiteStrictMode,
            })
        } else {
            sessionID = cookie.Value
        }

        // Add session ID to context
        ctx := context.WithValue(r.Context(), "sessionID", sessionID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func generateSessionID() string {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    return base64.URLEncoding.EncodeToString(b)
}

// Add middleware chain helper
func Chain(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middleware) - 1; i >= 0; i-- {
        handler = middleware[i](handler)
    }
    return handler
}
