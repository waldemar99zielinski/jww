package jwt

import jwtlib "github.com/golang-jwt/jwt/v5"

const tokenIssuer = "jww"

// codec is the single shared implementation of JWT signing and parsing. Token
// services provide their own claims, audience, issuer, and expiry rules.
type codec struct {
	secretKey     string
	algorithm     string
	signingMethod jwtlib.SigningMethod
}

func newCodec(cfg Config) codec {
	return codec{
		secretKey:     cfg.SecretKey,
		algorithm:     cfg.Algorithm,
		signingMethod: jwtlib.GetSigningMethod(cfg.Algorithm),
	}
}

func (c codec) sign(claims jwtlib.Claims) (string, error) {
	return jwtlib.NewWithClaims(c.signingMethod, claims).SignedString([]byte(c.secretKey))
}

func (c codec) parse(tokenString string, claims jwtlib.Claims, options ...jwtlib.ParserOption) (*jwtlib.Token, error) {
	options = append(options, jwtlib.WithValidMethods([]string{c.algorithm}))
	return jwtlib.ParseWithClaims(tokenString, claims, func(*jwtlib.Token) (any, error) {
		return []byte(c.secretKey), nil
	}, options...)
}
