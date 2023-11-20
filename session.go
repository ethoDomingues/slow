package slow

import (
	"fmt"
	"net/http"
	"time"
)

func newSession(secretKey string) *Session {
	return &Session{
		jwt: newJWT(secretKey),
		del: []string{},
	}
}

type Session struct {
	jwt              *JWT
	Permanent        bool
	expires          time.Time
	expiresPermanent time.Time
	del              []string
	changed          bool
}

// validate a cookie session
func (s *Session) validate(c *http.Cookie, secret string) {
	if secret == "" {
		return
	}
	str := c.Value
	if jwt, ok := ValidJWT(str, secret); ok {
		s.jwt = jwt
		if p, ok := jwt.Payload["_permanent"]; ok && p == "true" {
			s.Permanent = true
		}
	} else {
		s.jwt = NewJWT(secret)
	}
}

// This inserts a value into the session
func (s *Session) Set(key, value string) {
	s.jwt.Payload[key] = value
	s.changed = true
}

// Returns a session value based on the key. If key does not exist, returns an empty string
func (s *Session) Get(key string) (string, bool) {
	v, ok := s.jwt.Payload[key]
	return v, ok
}

// Delete a Value from Session
func (s *Session) Del(key string) {
	s.del = append(s.del, key)
	delete(s.jwt.Payload, key)
	s.changed = true
}

// Returns a cookie, with the value being a jwt
func (s *Session) save() *http.Cookie {
	if s.jwt == nil {
		l.warn.Println("to use the session you need to set a secretKey. rejecting session")
		return nil
	}
	if len(s.jwt.Payload) == 0 {
		return &http.Cookie{
			Name:     "_session",
			Value:    "",
			HttpOnly: true,
			MaxAge:   -1,
		}
	}
	var exp time.Time
	if s.Permanent {
		if s.expiresPermanent.IsZero() {
			exp = time.Now().Add(time.Hour * 24 * 31)
		} else {
			exp = s.expiresPermanent
		}
		s.jwt.Payload["_permanent"] = "true"
	} else {
		if s.expires.IsZero() {
			exp = time.Now().Add(time.Hour)
		} else {
			exp = s.expires
		}
	}
	s.jwt.Payload["exp"] = fmt.Sprint(exp.UTC().Unix())
	s.jwt.Payload["iat"] = fmt.Sprint(time.Now().UTC().Unix())
	return &http.Cookie{
		Name:     "_session",
		Value:    s.jwt.Sign(),
		HttpOnly: true,
		Expires:  exp,
	}
}

// Returns a JWT Token from session data
func (s *Session) GetSign() string {
	return s.save().Value
}
