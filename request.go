package slow

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/yaml.v2"
	"nhooyr.io/websocket"
)

func NewFile(p *multipart.Part) *File {
	b, err := io.ReadAll(p)
	if err != nil {
		panic(nil)
	}
	buf := bytes.NewBuffer(b)

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

		Body:    bytes.NewBuffer(nil),
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

	Body *bytes.Buffer
	Method,
	RemoteAddr,
	RequestURI,
	ContentType string

	isWebsocket bool

	ContentLength int

	URL     *url.URL
	Form    map[string]any
	Args    map[string]string
	Mime    map[string]string
	Query   url.Values
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
	mi := r.ctx.MatchInfo
	r.Query = r.URL.Query()
	r.Args = re.getUrlValues(mi.Route.Url, r.URL.Path)

	head := r.Header
	if head.Get("Connection") == "Upgrade" && head.Get("Upgrade") == "websocket" {
		r.isWebsocket = true
	}
}

func (r *Request) parseCookies() {
	cs := r.raw.Cookies()
	for _, c := range cs {
		r.Cookies[c.Name] = c
	}
	if c, ok := r.Cookies["_session"]; ok {
		r.ctx.Session.validate(c, r.ctx.App.SecretKey)
	}
}

func (r *Request) parseBody() {
	ctx := r.ctx
	r.Body.Grow(r.ContentLength)
	io.Copy(r.Body, r.raw.Body)
	switch {
	case r.ContentType == "", strings.HasPrefix(r.ContentType, "application/json"):
		json.Unmarshal(r.Body.Bytes(), &r.Form)
	case strings.HasPrefix(r.ContentType, "application/xml"):
		xml.Unmarshal(r.Body.Bytes(), &r.Form)
	case strings.HasPrefix(r.ContentType, "application/yaml"):
		yaml.Unmarshal(r.Body.Bytes(), &r.Form)
	case strings.HasPrefix(r.ContentType, "application/x-www-form-urlencoded"):
		v, err := url.ParseQuery(r.Body.String())
		if err == nil {
			for k, _v := range v {
				r.Form[k] = _v[0]
			}
		}
	case strings.HasPrefix(r.ContentType, "multipart/"):
		mp := multipart.NewReader(r.Body, r.Mime["boundary"])
		b := bytes.NewBufferString("")
		for {
			p, err := mp.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				ctx.Response.BadRequest()
			}
			if p.FileName() != "" {
				file := NewFile(p)
				r.Files[p.FormName()] = append(r.Files[p.FormName()], file)
			} else if p.FormName() != "" {
				b.ReadFrom(p)
				r.Form[p.FormName()] = b.String()
				b.Reset()
			}
		}
	}
}

func (r *Request) parseSchema() {
	if r.ctx.SchemaFielder == nil {
		return
	}
	sch := r.ctx.SchemaFielder
	nSch, err := MountSchemaFromRequest(sch, r)
	if err != nil {
		r.ctx.Response.JSON(err, 400)
	}
	r.ctx.Schema = nSch
}

func (r *Request) parse() {
	r.parseHeaders()
	r.parseCookies()
	r.parseBody()
	r.parseSchema()
}

// Returns the current url
func (r *Request) RequestURL() string {
	route := r.ctx.MatchInfo.Route
	args := []string{}
	for k, v := range r.Args {
		args = append(args, k, v)
	}
	return r.ctx.App.UrlFor(route.Name, true, args...)
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

func (r *Request) Ctx() *Ctx                          { return r.ctx }
func (r *Request) Context() context.Context           { return r.raw.Context() }
func (r *Request) UserAgent() string                  { return r.Header.Get("User-Agent") }
func (r *Request) Referer() string                    { return r.Header.Get("Referer") }
func (r *Request) RawRequest() *http.Request          { return r.raw }
func (r *Request) Clone(ctx context.Context) *Request { return NewRequest(r.raw.Clone(ctx), r.ctx) }
func (r *Request) WithContext(ctx context.Context) *Request {
	return NewRequest(r.raw.WithContext(ctx), r.ctx)
}

func (r *Request) ProtoAtLeast(major, minor int) bool {
	return r.ProtoMajor > major ||
		r.ProtoMajor == major && r.ProtoMinor >= minor
}

func (r *Request) BasicAuth() (username, password string, ok bool) {
	if r.ctx.App.BasicAuth != nil {
		return r.ctx.App.BasicAuth(r.ctx)
	}
	return r.raw.BasicAuth()
}

func (r *Request) Websocket() (*websocket.Conn, error) {
	return websocket.Accept(r.ctx, r.raw, nil)
}
