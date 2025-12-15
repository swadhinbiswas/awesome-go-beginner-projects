package functionality

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "crypto/subtle"
    "encoding/base64"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    "time"
)

var sessionKey []byte

func init() {
    // Prefer a base64-encoded environment variable SESSION_KEY.
    // If not present, generate a random key (note: sessions will not survive restarts).
    if v := os.Getenv("SESSION_KEY_BASE64"); v != "" {
        k, err := base64.RawURLEncoding.DecodeString(v)
        if err == nil && len(k) >= 16 {
            sessionKey = k
            return
        }
        log.Println("WARNING: SESSION_KEY_BASE64 provided but invalid; falling back to random key")
    }

    // generate random 32 bytes key
    k := make([]byte, 32)
    if _, err := rand.Read(k); err != nil {
        log.Fatalf("failed to generate session key: %v", err)
    }
    sessionKey = k
    log.Println("WARNING: using ephemeral session key; set SESSION_KEY_BASE64 in production to persist sessions")
}

// signValue creates token = base64(value) + "." + base64(hmac)
func signValue(secret []byte, value string) string {
    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(value))
    sig := mac.Sum(nil)
    return base64.RawURLEncoding.EncodeToString([]byte(value)) + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// verifyValue validates token and returns raw value if valid.
func verifyValue(secret []byte, token string) (string, bool) {
    parts := strings.Split(token, ".")
    if len(parts) != 2 {
        return "", false
    }
    rawVal, err := base64.RawURLEncoding.DecodeString(parts[0])
    if err != nil {
        return "", false
    }
    sig, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return "", false
    }
    mac := hmac.New(sha256.New, secret)
    mac.Write(rawVal)
    expected := mac.Sum(nil)
    if subtle.ConstantTimeCompare(expected, sig) != 1 {
        return "", false
    }
    return string(rawVal), true
}

// CreateSessionToken makes a signed token that encodes userID and expiry.
func CreateSessionToken(userID string, ttl time.Duration) (string, error) {
    exp := time.Now().Add(ttl).Unix()
    value := fmt.Sprintf("%s:%d", userID, exp)
    token := signValue(sessionKey, value)
    return token, nil
}

// ParseSessionToken verifies token, checks expiry, and returns userID if valid.
func ParseSessionToken(token string) (string, bool) {
    value, ok := verifyValue(sessionKey, token)
    if !ok {
        return "", false
    }
    parts := strings.SplitN(value, ":", 2)
    if len(parts) != 2 {
        return "", false
    }
    userID := parts[0]
    expUnix, err := strconv.ParseInt(parts[1], 10, 64)
    if err != nil {
        return "", false
    }
    if time.Now().Unix() > expUnix {
        return "", false
    }
    return userID, true
}
