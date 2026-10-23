package web

import (
	"net/http"
	"net/url"
	"strings"
)

func securityHeaders(cfg Config) func(http.Handler) http.Handler {
	csp := contentSecurityPolicy(cfg.APIBaseURL)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// CSP is the primary XSS boundary. The separately configured API origin
			// is allowed for data and previews, but never as a script source.
			w.Header().Set("Content-Security-Policy", csp)

			// Isolate the application from cross-origin windows and prevent other
			// origins from loading its static responses as subresources.
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

			// These device APIs are not used. File inputs, including capture-enabled
			// inputs on mobile, do not require getUserMedia permissions.
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			// Avoid leaking transaction routes or identifiers in outbound requests.
			w.Header().Set("Referrer-Policy", "no-referrer")

			// Browsers honor HSTS only over HTTPS. includeSubDomains is intentionally
			// omitted because this service cannot make guarantees about sibling hosts.
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")

			// Do not let browsers reinterpret HTML or JSON as executable content.
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// CSP frame-ancestors is authoritative; X-Frame-Options protects older
			// clients and makes the clickjacking policy explicit to intermediaries.
			w.Header().Set("X-Frame-Options", "DENY")
			next.ServeHTTP(w, r)
		})
	}
}

func contentSecurityPolicy(apiBaseURL string) string {
	apiOrigin := ""
	if parsed, err := url.Parse(apiBaseURL); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		apiOrigin = " " + parsed.Scheme + "://" + parsed.Host
	}

	return strings.Join([]string{
		"default-src 'self'",
		"base-uri 'self'",
		"object-src 'none'",
		"frame-ancestors 'none'",
		"form-action 'self'",
		"script-src 'self'",
		// React uses style attributes for a small number of dynamic values. Keep
		// the inline exception scoped to styles; scripts remain strictly external.
		"style-src 'self' 'unsafe-inline'",
		"font-src 'self'",
		"manifest-src 'self'",
		"worker-src 'self'",
		"connect-src 'self'" + apiOrigin,
		// Attachment previews are created as blob URLs; design-mode placeholders
		// use data URLs. The API origin supports separately hosted attachments.
		"img-src 'self' data: blob:" + apiOrigin,
		"media-src 'self' blob:" + apiOrigin,
		"frame-src 'self' blob:" + apiOrigin,
	}, "; ")
}
