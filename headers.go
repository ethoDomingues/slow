package slow

import (
	"net/http"
	"net/textproto"
	"strings"
)

type Headers map[string][]string

// Set a Header
func (h *Headers) Set(key string, value string) {
	textproto.MIMEHeader(*(h)).Set(key, value)
}

// Add value in a in Header Key. If the key does not exist, it is created
func (h *Headers) Add(key string, value string) {
	textproto.MIMEHeader(*(h)).Add(key, value)
}

// Return a value of Header Key. If the key does not exist, return a empty string
func (h *Headers) Get(key string) string {
	return textproto.MIMEHeader(*(h)).Get(key)
}

func (h *Headers) Del(key string) {
	textproto.MIMEHeader(*(h)).Del(key)
}

// Set a Cookie. Has the same effect as 'Response.SetCookie'
func (h *Headers) SetCookie(cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		h.Add("Set-Cookie", v)
	}
}

// Write the headers in the response
func (h *Headers) Save(w http.ResponseWriter) {
	for key := range *(h) {
		w.Header().Set(key, h.Get(key))
	}
}

// If present on route or router, allows resource sharing between origins
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
