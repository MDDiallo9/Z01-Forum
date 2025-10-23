package services

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
)

func GenerateRandomString(length int) (string, error) {
	// Create a byte slice of the specified length
	// Note that a hex-encoded string is twice the lenght of the original byte
	numBytes := (length + 1) / 2
	actualLen := make([]byte, numBytes)

	// Read cryptographically secure random bytes into actualLen
	_, err := rand.Read(actualLen)
	if err != nil {
		return "", err
	}

	// Encode byte slice into a hexadecimal string and return
	return hex.EncodeToString(actualLen)[:length], nil
}

func ExtractIPFromRequest(r *http.Request) string {
	// Use r.Header.Get to request the headers
	// RemoteAddr allows HTTP servers to record the network address that sent the request
	IPAddress := r.Header.Get("X-Forwarded-For")
	if IPAddress != "" {
		return strings.Split(IPAddress, ",")[0]
	}

	IPAddress = r.Header.Get("X-Real-Ip")
	if IPAddress != "" {
		return IPAddress
	}
	// Fallback to get IP from RemoteAddr (host:port)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func ExtractUserAgent(r *http.Request) string {
	return r.UserAgent()
}
