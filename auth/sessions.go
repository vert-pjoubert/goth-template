package auth

import (
	"crypto/sha256"
	"encoding/pem"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/sessions"
)

type ISessionManager interface {
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(r *http.Request, w http.ResponseWriter, session *sessions.Session) error
}

type CookieSessionManager struct {
	store *sessions.CookieStore
}

// NewCookieSessionManager initializes a new CookieSessionManager with key pairs.
func NewCookieSessionManager(keyPairs ...[]byte) *CookieSessionManager {
	if len(keyPairs) == 1 {
		// Assume it's a simple string key
		return &CookieSessionManager{
			store: sessions.NewCookieStore(keyPairs...),
		}
	}

	// Assume keyPairs are X509 certificates
	hashedKeys := make([][]byte, len(keyPairs))
	for i, key := range keyPairs {
		block, _ := pem.Decode(key)
		if block == nil || (block.Type != "CERTIFICATE" && block.Type != "RSA PRIVATE KEY") {
			continue
		}
		hash := sha256.Sum256(block.Bytes)
		hashedKeys[i] = hash[:]
	}

	if len(hashedKeys) == 0 {
		// Fallback to WEAK_SESSION_STRING if no valid certificates were loaded
		weakSessionString := os.Getenv("WEAK_SESSION_STRING")
		if weakSessionString == "" {
			panic("WEAK_SESSION_STRING must be set in the environment")
		}
		hashedKeys = [][]byte{[]byte(weakSessionString)}
	}

	return &CookieSessionManager{
		store: sessions.NewCookieStore(hashedKeys...),
	}
}

func (c *CookieSessionManager) GetSession(r *http.Request) (*sessions.Session, error) {
	return c.store.Get(r, "auth-session")
}

func (c *CookieSessionManager) SaveSession(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return c.store.Save(r, w, session)
}

// LoadKeyPairsFromDir loads key pairs from PEM files in the specified directory.
func LoadKeyPairsFromDir(dir string) ([][]byte, error) {
	var keyPairs [][]byte

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		keyData, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		keyPairs = append(keyPairs, keyData)
	}

	return keyPairs, nil
}
