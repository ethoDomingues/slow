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
	str := c.Value
	if jwt, ok := ValidJWT(str, secret); ok {
		s.jwt = jwt
		if _, ok := s.jwt.Payload["_permanent"]; ok {
			s.Permanent = true
		}
	} else {
		if secret != "" {
			s.jwt = newJWT("")
		} else {
			s.jwt = NewJWT(secret)
		}
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
	exp := s.expires
	if len(s.jwt.Payload) == 0 {
		return &http.Cookie{
			Name:     "session",
			Value:    "",
			HttpOnly: true,
			Expires:  exp,
			MaxAge:   -1,
		}
	}
	if s.Permanent {
		if s.expiresPermanent.IsZero() {
			exp = time.Now().Add(time.Hour * 24 * 31)
		} else {
			exp = s.expiresPermanent
		}
		s.jwt.Payload["_permanent"] = "1"
	} else {
		if s.expires.IsZero() {
			exp = time.Now().Add(time.Hour)
		} else {
			exp = s.expires
		}
	}
	if len(s.jwt.Payload) == 0 {
		return &http.Cookie{
			Name:     "session",
			Value:    "",
			MaxAge:   -0,
			HttpOnly: true,
		}
	}
	s.jwt.Payload["iat"] = fmt.Sprint(exp.Unix())

	return &http.Cookie{
		Name:     "session",
		Value:    s.jwt.Sign(),
		HttpOnly: true,
		Expires:  exp,
	}
}

// Returns a JWT Token from session data
func (s *Session) GetSign() string {
	return s.save().Value
}

func (s *Session) CheckUpdate(ctx *Ctx) {
	update := false
	if s.changed {
		update = true
	}
	if !s.expires.IsZero() {
		if s.expires.After(time.Now()) {
			update = true
		}
	} else if !s.expiresPermanent.IsZero() {
		if s.expires.Before(time.Now()) {
			update = true
		}
	}
	if update {

	}
}
