// Package schema holds HTTP request/response JSON shapes — the only place
// that translates between domain entities and JSON.
package schema

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	return firstError(
		minLength("username", r.Username, 1),
		minLength("password", r.Password, 1),
	)
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func NewTokenResponse(accessToken string) TokenResponse {
	return TokenResponse{AccessToken: accessToken, TokenType: "bearer"}
}
