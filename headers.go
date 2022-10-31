package slow

import (
	"net/http"
	"net/textproto"
	"strings"
)

type Headers map[string][]string

func (h *Headers) Set(key string, value string) {
	_h := *h
	textproto.MIMEHeader(_h).Set(key, value)
}

func (h *Headers) Add(key string, value string) {
	_h := *h
	textproto.MIMEHeader(_h).Add(key, value)
}

func (h *Headers) Get(key string) string {
	_h := *h
	return textproto.MIMEHeader(_h).Get(key)
}

func (h *Headers) Del(key string) {
	_h := *h
	textproto.MIMEHeader(_h).Del(key)
}

func (h *Headers) SetCookie(cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		h.Add("Set-Cookie", v)
	}
}

func (h *Headers) Save(w http.ResponseWriter) {
	for key := range *(h) {
		w.Header().Set(key, h.Get(key))
	}
}

type Cors struct {
	MaxAge           string   // Access-Control-Max-Age
	AllowOrigin      string   // Access-Control-Allow-Origin
	AllowMethods     []string // Access-Control-Allow-Methods
	AllowHeaders     []string // Access-Control-Allow-Headers
	ExposeHeaders    []string // Access-Control-Expose-Headers
	RequestMethod    string   // Access-Control-Request-Method
	AllowCredentials bool     // Access-Control-Allow-Credentials
}

func (c *Cors) parse(h *Headers) {
	if c.MaxAge != "" {
		h.Set("Access-Control-Max-Age", c.MaxAge)
	}
	if c.AllowOrigin != "" {
		h.Set("Access-Control-Allow-Origin", c.AllowOrigin)
	}
	if len(c.AllowHeaders) > 0 {
		h.Set("Access-Control-Allow-Headers", strings.Join(c.AllowHeaders, ", "))
	}
	if len(c.ExposeHeaders) > 0 {
		h.Set("Access-Control-Expose-Headers", strings.Join(c.ExposeHeaders, ", "))
	}
	if c.RequestMethod != "" {
		h.Set("Access-Control-Request-Method", c.RequestMethod)
	}
	if c.AllowCredentials {
		h.Set("Access-Control-Allow-Credentials", "true")
	}
}
