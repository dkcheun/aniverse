package util

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"math/rand"
	"time"
)

// GenerateCodeVerifier creates a random string for PKCE (min length: 43, max length: 128)
func GenerateCodeVerifier() string {
	rand.Seed(time.Now().UnixNano())
	verifier := make([]byte, 43)
	for i := range verifier {
		verifier[i] = byte(rand.Intn(26) + 97) // Generate random lowercase letters (a-z)
	}
	return string(verifier)
}

// GenerateCodeChallenge creates the code_challenge using SHA256 hash (for enhanced security)
// Though MAL supports plain, using a challenge with SHA256 is more secure
func GenerateCodeChallenge(verifier string) (string, error) {
	if len(verifier) < 43 || len(verifier) > 128 {
		return "", errors.New("code_verifier must be between 43 and 128 characters")
	}
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return challenge, nil
}

// In-memory storage for code verifiers (could replace with a session store or database)
var codeVerifierStore = make(map[string]string)

// StoreCodeVerifier stores the code verifier in memory (associated with the given state)
func StoreCodeVerifier(state, verifier string) {
	codeVerifierStore[state] = verifier
}

// GetCodeVerifier retrieves the code verifier for the given state (returns a boolean indicating existence)
func GetCodeVerifier(state string) (string, bool) {
	verifier, exists := codeVerifierStore[state]
	return verifier, exists
}
