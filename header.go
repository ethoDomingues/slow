package slow

import (
	"net/http"
	"net/textproto"
)

type Header http.Header

// Set a Header
func (h *Header) Set(key string, value string) {
	textproto.MIMEHeader(*(h)).Set(key, value)
}

// Add value in a in Header Key. If the key does not exist, it is created
func (h *Header) Add(key string, value string) {
	textproto.MIMEHeader(*(h)).Add(key, value)
}

// Return a value of Header Key. If the key does not exist, return a empty string
func (h *Header) Get(key string) string {
	return textproto.MIMEHeader(*(h)).Get(key)
}

func (h *Header) Del(key string) {
	textproto.MIMEHeader(*(h)).Del(key)
}

// Set a Cookie. Has the same effect as 'Response.SetCookie'
func (h *Header) SetCookie(cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		h.Add("Set-Cookie", v)
	}
}

// Write the headers in the response
func (h *Header) Save(w http.ResponseWriter) {
	for key := range *(h) {
		w.Header().Set(key, h.Get(key))
	}
}

// If present on route or router, allows resource sharing between origins
type Cors struct {
	MaxAge           string // Access-Control-Max-Age
	AllowOrigin      string // Access-Control-Allow-Origin
	AllowMethods     string // Access-Control-Allow-Methods
	AllowHeaders     string // Access-Control-Allow-Headers
	ExposeHeaders    string // Access-Control-Expose-Headers
	RequestMethod    string // Access-Control-Request-Method
	AllowCredentials bool   // Access-Control-Allow-Credentials
}

func (c *Cors) parse(h Header) {
	if c.MaxAge != "" {
		h.Set("Access-Control-Max-Age", c.MaxAge)
	}
	if c.AllowOrigin != "" {
		h.Set("Access-Control-Allow-Origin", c.AllowOrigin)
	}
	if len(c.AllowHeaders) > 0 {
		h.Set("Access-Control-Allow-Headers", c.AllowHeaders)
	}
	if len(c.ExposeHeaders) > 0 {
		h.Set("Access-Control-Expose-Headers", c.ExposeHeaders)
	}
	if c.RequestMethod != "" {
		h.Set("Access-Control-Request-Method", c.RequestMethod)
	}
	if c.AllowCredentials {
		h.Set("Access-Control-Allow-Credentials", "true")
	}
}
