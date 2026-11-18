package speedtest

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func SignTicket(secret, nodeID string, ttl time.Duration) string {
	if secret == "" {
		return ""
	}
	var nonceBytes [8]byte
	_, _ = rand.Read(nonceBytes[:])
	nonce := hex.EncodeToString(nonceBytes[:])
	exp := time.Now().Add(ttl).Unix()

	payload := fmt.Sprintf("%s|%d|%s", nodeID, exp, nonce)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	raw := payload + "|" + sig
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func VerifyTicket(secret, expectedNodeID, ticket string, now time.Time) bool {
	if secret == "" || ticket == "" {
		return false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(ticket)
	if err != nil {
		return false
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 4 {
		return false
	}
	nodeID, expStr, nonce, gotSigHex := parts[0], parts[1], parts[2], parts[3]
	if expectedNodeID != "" && nodeID != expectedNodeID {
		return false
	}
	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || now.Unix() > expUnix {
		return false
	}

	payload := fmt.Sprintf("%s|%s|%s", nodeID, expStr, nonce)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	wantSig := mac.Sum(nil)

	gotSig, err := hex.DecodeString(gotSigHex)
	if err != nil {
		return false
	}
	return hmac.Equal(gotSig, wantSig)
}
