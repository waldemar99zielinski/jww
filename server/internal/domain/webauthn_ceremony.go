package domain

// WebAuthnCeremonyOptions is the publicKey options a Begin call produces —
// forwarded verbatim to the browser's navigator.credentials.create()/get().
// Nothing in this app ever reads a field out of it.
type WebAuthnCeremonyOptions []byte

// WebAuthnCeremonySession is the opaque ceremony state a Begin call hands
// back, round-tripped to the matching Finish call via a signed challenge
// token (this app has no session store — see
// service/webauthn_credential.ChallengeTokens).
type WebAuthnCeremonySession []byte

// WebAuthnCeremonyResponse is the browser's signed
// navigator.credentials.create()/get() response. Opaque to everything
// except the Ceremony adapter, which parses it to verify the ceremony.
type WebAuthnCeremonyResponse []byte
