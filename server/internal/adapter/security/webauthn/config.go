// Package webauthn wraps github.com/go-webauthn/webauthn into the
// Begin/Finish ceremony this app needs, translating between our stored
// domain.WebAuthnCredential rows and the library's own types. It never
// creates a user — see internal/service/user for that.
package webauthn

import (
	"strings"

	"jww/utils/configuration"
)

// Config is the WebAuthn Relying Party configuration. RPID and RPOrigins
// must exactly match the domain/origin the PWA is served from — there is no
// sane default, so both are required; get them wrong and every ceremony
// fails closed (not silently insecure).
type Config struct {
	// RPID is the effective domain of the Relying Party, e.g.
	// "example.com" (no scheme, no port).
	RPID string `env:"WEBAUTHN_RP_ID" bind:"required"`
	// RPDisplayName is shown to the user by the platform's biometric UI.
	RPDisplayName string `env:"WEBAUTHN_RP_DISPLAY_NAME" bind:"optional" default:"Joint Wallet Watch"`
	// RPOrigins is the comma-separated list of fully-qualified origins
	// (scheme+host[:port]) a ceremony may be bound to, e.g.
	// "https://example.com". Split via RPOriginsList.
	RPOrigins string `env:"WEBAUTHN_RP_ORIGINS" bind:"required"`
}

func LoadConfig() (Config, error) {
	return configuration.Load[Config]()
}

// RPOriginsList splits the comma-separated RPOrigins, mirroring
// http_server.Config.CorsOriginsList.
func (c Config) RPOriginsList() []string {
	var origins []string
	for _, origin := range strings.Split(c.RPOrigins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
