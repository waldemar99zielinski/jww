package signup

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GeneratedPasswordLength is the length of passwords returned by
// GeneratePassword. It is long enough for a real credential while still being
// easy to copy from the log of a one-off command.
const GeneratedPasswordLength = 24

// The character groups used by GeneratePassword. Look-alike characters (l, I,
// O, 0, 1) are omitted so a generated password is safe to retype by hand.
const (
	passwordLower   = "abcdefghijkmnopqrstuvwxyz"
	passwordUpper   = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	passwordDigits  = "23456789"
	passwordSymbols = "!@#$%^&*()-_=+"

	passwordAlphabet = passwordLower + passwordUpper + passwordDigits + passwordSymbols
)

// GeneratePassword returns a cryptographically random password containing at
// least one lower-case letter, upper-case letter, digit, and symbol. It reads
// from crypto/rand, so the result is suitable for a real credential.
func GeneratePassword() (string, error) {
	chars := make([]byte, 0, GeneratedPasswordLength)
	for _, group := range []string{passwordLower, passwordUpper, passwordDigits, passwordSymbols} {
		c, err := randomChar(group)
		if err != nil {
			return "", err
		}
		chars = append(chars, c)
	}
	for len(chars) < GeneratedPasswordLength {
		c, err := randomChar(passwordAlphabet)
		if err != nil {
			return "", err
		}
		chars = append(chars, c)
	}
	if err := shuffle(chars); err != nil {
		return "", err
	}
	return string(chars), nil
}

// randomChar picks one byte from alphabet uniformly at random.
func randomChar(alphabet string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
	if err != nil {
		return 0, fmt.Errorf("generating password: %w", err)
	}
	return alphabet[n.Int64()], nil
}

// shuffle reorders chars in place with a Fisher–Yates shuffle, so the
// guaranteed one-of-each group from GeneratePassword does not land at fixed
// positions.
func shuffle(chars []byte) error {
	for i := len(chars) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return fmt.Errorf("generating password: %w", err)
		}
		chars[i], chars[j.Int64()] = chars[j.Int64()], chars[i]
	}
	return nil
}
