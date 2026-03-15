package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/netip"
	neturl "net/url"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

const (
	consoleCSRFCookie = "homelabwatch_console_csrf"
	consoleCSRFHeader = "X-Homelabwatch-CSRF"
)

func parseTrustedNetworks(values []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

func (r *Router) withTrustedConsole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := r.validateTrustedConsole(req); err != nil {
			status := http.StatusForbidden
			if strings.Contains(err.Error(), "csrf") {
				status = http.StatusUnauthorized
			}
			writeError(w, status, err)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (r *Router) withExternalToken(requiredScope domain.TokenScope, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ok, err := r.app.ValidateAPIToken(req.Context(), apiToken(req), requiredScope)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, errors.New("missing or invalid api token"))
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (r *Router) validateTrustedConsole(req *http.Request) error {
	if !r.isTrustedNetwork(req) {
		return errors.New("ui write access is limited to trusted networks")
	}
	if !sameOrigin(req) {
		return errors.New("browser origin is not allowed")
	}
	if !validCSRF(req) {
		return errors.New("missing or invalid console csrf token")
	}
	return nil
}

func (r *Router) isTrustedNetwork(req *http.Request) bool {
	if len(r.trustedNetworks) == 0 {
		return true
	}
	clientAddr, err := clientIP(req)
	if err != nil {
		return false
	}
	for _, prefix := range r.trustedNetworks {
		if prefix.Contains(clientAddr) {
			return true
		}
	}
	return false
}

func clientIP(req *http.Request) (netip.Addr, error) {
	host := strings.TrimSpace(req.RemoteAddr)
	if host == "" {
		return netip.Addr{}, errors.New("missing remote address")
	}
	if addrPort, err := netip.ParseAddrPort(host); err == nil {
		return addrPort.Addr().Unmap(), nil
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		return addr.Unmap(), nil
	}
	if host, _, err := net.SplitHostPort(host); err == nil {
		if addr, err := netip.ParseAddr(host); err == nil {
			return addr.Unmap(), nil
		}
	}
	return netip.Addr{}, errors.New("invalid remote address")
}

func sameOrigin(req *http.Request) bool {
	for _, raw := range []string{req.Header.Get("Origin"), req.Header.Get("Referer")} {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		parsed, err := neturl.Parse(raw)
		if err != nil {
			return false
		}
		return parsed.Host == req.Host
	}
	return false
}

func validCSRF(req *http.Request) bool {
	headerToken := strings.TrimSpace(req.Header.Get(consoleCSRFHeader))
	if headerToken == "" {
		return false
	}
	cookie, err := req.Cookie(consoleCSRFCookie)
	if err != nil {
		return false
	}
	return cookie.Value == headerToken
}

func issueConsoleCSRF(w http.ResponseWriter, req *http.Request) (string, error) {
	if cookie, err := req.Cookie(consoleCSRFCookie); err == nil && strings.TrimSpace(cookie.Value) != "" {
		return cookie.Value, nil
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     consoleCSRFCookie,
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   req.TLS != nil,
		Expires:  time.Now().Add(12 * time.Hour),
	})
	return token, nil
}

func apiToken(req *http.Request) string {
	if token := strings.TrimSpace(req.Header.Get("X-Admin-Token")); token != "" {
		return token
	}
	if auth := strings.TrimSpace(req.Header.Get("Authorization")); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return ""
}

func randomToken() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
