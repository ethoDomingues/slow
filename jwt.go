package slow

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func computeMac(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}

func validMac(message, messageMac, secret []byte) bool {
	expectedMac := computeMac(message, secret)
	return hmac.Equal(messageMac, expectedMac)
}

func SignJWT(headers, payload map[string]string, secret string) string {
	if headers == nil && len(headers) == 0 {
		headers = map[string]string{
			"alg": "HS256",
			"typ": "JWT",
		}
	}
	bytsH, _ := json.Marshal(headers)

	b64H := base64.RawURLEncoding.EncodeToString(bytsH)
	bytsP, _ := json.Marshal(payload)
	b64P := base64.RawURLEncoding.EncodeToString(bytsP)

	b64HB := fmt.Sprintf("%s.%s", b64H, b64P)
	sig := computeMac([]byte(b64HB), []byte(secret))

	return fmt.Sprintf("%s.%s", b64HB, base64.RawURLEncoding.EncodeToString(sig))
}

func ValidJWT(jwt, secret string) (*JWT, bool) {
	if jwt != "" {
		jwt = strings.TrimPrefix(jwt, "Bearer ")
		hps := strings.Split(jwt, ".")
		if len(hps) == 3 {
			hb := fmt.Sprintf("%s.%s", hps[0], hps[1])
			sig, err := base64.RawURLEncoding.DecodeString(hps[2])
			if err == nil {
				if validMac([]byte(hb), sig, []byte(secret)) {
					h := map[string]string{}
					p := map[string]string{}

					hB64 := hps[0]
					pB64 := hps[1]
					if hB64 != "" {
						hJSON, err := base64.RawURLEncoding.DecodeString(hB64)
						if err == nil {
							json.Unmarshal(hJSON, &h)
						}
					}
					if pB64 != "" {
						pJSON, err := base64.RawURLEncoding.DecodeString(pB64)
						if err == nil {
							json.Unmarshal(pJSON, &p)
						}
					}
					if exp, ok := p["exp"]; ok {
						i, err := strconv.ParseInt(exp, 10, 64)
						if err != nil {
							return &JWT{h, p, secret}, false
						}
						t := time.Unix(i, 0)
						return &JWT{h, p, secret}, time.Now().UTC().Before(t)
					}
					return &JWT{h, p, secret}, true
				}
			}
		}
	}
	return nil, false
}

func NewJWT(secret string) *JWT {
	if secret == "" {
		panic("for use JWT, a 'secret' is required")
	}
	return newJWT(secret)
}

func newJWT(secret string) *JWT {
	return &JWT{
		Payload: map[string]string{},
		Headers: map[string]string{
			"alg": "HS256",
			"typ": "JWT",
		},
		Secret: secret,
	}
}

type JWT struct {
	Headers, Payload map[string]string
	Secret           string
}

func (j *JWT) Sign() string {
	return SignJWT(j.Headers, j.Payload, j.Secret)
}
