package utils

import "net"

// FreePort asks the OS for an ephemeral TCP port, then releases it
// immediately — there's a small window where another process could grab it
// first, but that's an accepted tradeoff for a test, not a correctness
// requirement.
func FreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
