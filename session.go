package slow

import (
	"fmt"
	"net/http"
	"time"
)

func NewSession() *Session {
	return &Session{
		jwt: NewJWT(),
		del: []string{},
	}
}

type Session struct {
	jwt              *JWT
	Permanent        bool
	expires          time.Time
	expiresPermanent time.Time
	del              []string
}

func (s *Session) validate(c *http.Cookie, secret string) {
	str := c.Value
	if jwt, ok := ValidJWT(str, secret); ok {
		s.jwt = jwt
		if _, ok := s.jwt.Payload["_permanent"]; ok {
			s.Permanent = true
		}
	} else {
		s.jwt = NewJWT()
	}
}

func (s *Session) Set(key, value string) {
	s.jwt.Payload[key] = value
}

func (s *Session) Get(key string) (string, bool) {
	v, ok := s.jwt.Payload[key]
	return v, ok
}

func (s *Session) Del(key string) {
	s.del = append(s.del, key)
	delete(s.jwt.Payload, key)
}

func (s *Session) Save() *http.Cookie {
	if s.jwt.Secret == "" {
		l.warn.Println("Para usar as session, vc precisda add uma secretKey no app")
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
	s.jwt.Payload["exp"] = fmt.Sprint(exp.Unix())

	return &http.Cookie{
		Name:     "session",
		Value:    s.jwt.Sign(),
		HttpOnly: true,
		Expires:  exp,
	}
}

func (s *Session) GetSign() string {
	return s.Save().Value
}