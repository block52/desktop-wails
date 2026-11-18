package main

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// appDataDir is where the wallet vault lives: ~/Library/Application Support/Block52
// (or the OS user-config dir equivalent).
func appDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	return filepath.Join(dir, "Block52")
}

// noopKeychain is a placeholder OsKeyStore for the spike: it does not use a real
// OS keychain, so the vault is protected by the user's password only. A real
// build wires the platform keychain here.
type noopKeychain struct{}

func (noopKeychain) IsSecure() bool                      { return false }
func (noopKeychain) Encrypt(plain string) ([]byte, error) { return []byte(plain), nil }
func (noopKeychain) Decrypt(c []byte) (string, error)     { return string(c), nil }

func hexEncode(b []byte) string { return hex.EncodeToString(b) }

// parseGRPC turns grpc://host:port / grpcs://host:port (or a bare host:port)
// into a dial target and a TLS flag.
func parseGRPC(endpoint string) (target string, useTLS bool) {
	switch {
	case strings.HasPrefix(endpoint, "grpcs://"):
		return strings.TrimPrefix(endpoint, "grpcs://"), true
	case strings.HasPrefix(endpoint, "grpc://"):
		return strings.TrimPrefix(endpoint, "grpc://"), false
	case strings.HasPrefix(endpoint, "https://"):
		return strings.TrimPrefix(endpoint, "https://"), true
	case strings.HasPrefix(endpoint, "http://"):
		return strings.TrimPrefix(endpoint, "http://"), false
	default:
		return endpoint, false
	}
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 20*time.Second)
}
