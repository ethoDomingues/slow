package slow

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/yaml.v2"
)

func NewFile(p *multipart.Part) *File {
	byts, err := io.ReadAll(p)
	if err != nil {
		panic(nil)
	}
	buf := bytes.NewBuffer(byts)

	return &File{
		Filename:     p.FileName(),
		ContentType:  p.Header.Get("Content-Type"),
		ContentLeght: buf.Len(),
		Stream:       buf,
	}
}

type File struct {
	Filename     string
	ContentType  string
	ContentLeght int
	Stream       *bytes.Buffer
}

func NewRequest(req *http.Request, ctx *Ctx) *Request {
	req.URL.Host = req.Host
	rq := &Request{
		ctx:        ctx,
		raw:        req,
		URL:        req.URL,
		Method:     req.Method,
		RemoteAddr: req.RemoteAddr,
		RequestURI: req.RequestURI,

		ContentLength: int(req.ContentLength),

		Args:    map[string]string{},
		Mime:    map[string]string{},
		Form:    map[string]any{},
		Files:   map[string][]*File{},
		Cookies: map[string]*http.Cookie{},

		Header: Header(req.Header),
	}
	return rq
}

type Request struct {
	raw *http.Request

	ctx    *Ctx
	Header Header

	Body,
	Method,
	RemoteAddr,
	RequestURI,
	ContentType string

	ContentLength int

	URL     *url.URL
	Form    map[string]any
	Args    map[string]string
	Mime    map[string]string
	Query   map[string][]string
	Files   map[string][]*File
	Cookies map[string]*http.Cookie

	TransferEncoding []string

	Proto      string // "HTTP/1.0"
	ProtoMajor int    // 1
	ProtoMinor int    // 0
}

func (r *Request) parseHeaders() {
	ctx := r.ctx
	ct := r.raw.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		if err.Error() != "mime: no media type" {
			ctx.Response.BadRequest()
		}
	}
	r.ContentType = mediaType
	r.Mime = params
}

func (r *Request) parseCookies() {
	cs := r.raw.Cookies()
	for _, c := range cs {
		r.Cookies[c.Name] = c
	}
}

func (r *Request) parseBody() {
	ctx := r.ctx
	body := bytes.NewBuffer(nil)
	body.Grow(int(r.ContentLength))
	io.CopyBuffer(body, r.raw.Body, nil)
	switch {
	case r.ContentType == "", strings.HasPrefix(r.ContentType, "application/json"):
		json.Unmarshal(body.Bytes(), &r.Form)
	case strings.HasPrefix(r.ContentType, "application/xml"):
		xml.Unmarshal(body.Bytes(), &r.Form)
	case strings.HasPrefix(r.ContentType, "application/yaml"):
		yaml.Unmarshal(body.Bytes(), &r.Form)
	case strings.HasPrefix(r.ContentType, "multipart/"):
		mp := multipart.NewReader(body, r.Mime["boundary"])
		for {
			p, err := mp.NextPart()
			if err == io.EOF {
				return
			}
			if err != nil {
				ctx.Response.BadRequest()
			}
			if p.FileName() != "" {
				file := NewFile(p)
				r.Files[p.FormName()] = append(r.Files[p.FormName()], file)
			} else {
				val, err := io.ReadAll(p)
				if err != nil {
					continue
				}
				r.Form[p.FormName()] = string(val)
			}
		}
	}
	r.Body = body.String()
}

func (r *Request) parseRequest() {
	r.parseHeaders()
	r.parseCookies()
	r.parseBody()
}

// Returns the current url
func (r *Request) RequestURL() string {
	route := r.ctx.MatchInfo.Route
	args := []string{}
	for k, v := range r.Args {
		args = append(args, k, v)
	}
	return r.ctx.App.UrlFor(route.fullName, true, args...)
}

/*
URL Builder

	app.GET("/", index)
	app.GET("/login", login)

	app.UrlFor("login", false, "next", "currentUrl"})
	// results: /login?next=currentUrl

	app.UrlFor("login", true, "token", "foobar"})
	// results: http://yourAddress/login?token=foobar

	// example
	func index(ctx *slow.Ctx) {
		req := ctx.Request
		rsp := ctx.Response
		userID, ok := ctx.Global["user"]
		if !ok {
			next := r.RequestUrl()
			rsp.Redirect(req.UrlFor("login", true, "next", next))
			//  redirect to: http://youraddress/login?next=http://yourhost:port/
		}
		... you code here
	}
*/
func (r *Request) UrlFor(name string, external bool, args ...string) string {
	return r.ctx.App.UrlFor(name, external, args...)
}

/* Returns a '*slow.Ctx' of the current request */
func (r *Request) Ctx() *Ctx { return r.ctx }

// Returns a 'context.Context' of the current request
func (r *Request) Context() context.Context { return r.raw.Context() }

func (r *Request) WithContext(ctx context.Context) *Request {
	return NewRequest(r.raw.WithContext(ctx), r.ctx)
}

func (r *Request) Clone(ctx context.Context) *Request {
	return NewRequest(r.raw.Clone(ctx), r.ctx)
}

func (r *Request) ProtoAtLeast(major, minor int) bool {
	return r.ProtoMajor > major ||
		r.ProtoMajor == major && r.ProtoMinor >= minor
}

func (r *Request) UserAgent() string {
	return r.Header.Get("User-Agent")
}

func (r *Request) BasicAuth() (username, password string, ok bool) {
	return r.raw.BasicAuth()
}

func (r *Request) RawRequest() *http.Request {
	return r.raw
}
func (r *Request) Referer() string {
	return r.raw.Referer()
}
