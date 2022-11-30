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

func NewRequest(req *http.Request, ctxID string) *Request {
	return &Request{
		ctx:        ctxID,
		Raw:        req,
		Method:     req.Method,
		RemoteAddr: req.RemoteAddr,

		Args:    map[string]string{},
		Mime:    map[string]string{},
		Form:    map[string]any{},
		Files:   map[string][]*File{},
		Cookies: map[string]*http.Cookie{},
	}
}

type Request struct {
	Raw *http.Request

	ctx,
	Body,
	Method,
	RemoteAddr,
	ContentType string

	Form    map[string]any
	Args    map[string]string
	Mime    map[string]string
	Query   map[string][]string
	Files   map[string][]*File
	Cookies map[string]*http.Cookie
}

func (r *Request) parseHeaders() {
	ctx := r.Ctx()
	ct := r.Raw.Header.Get("Content-Type")
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
	cs := r.Raw.Cookies()
	for _, c := range cs {
		r.Cookies[c.Name] = c
	}
}

func (r *Request) parseBody() {
	ctx := r.Ctx()
	body := bytes.NewBuffer(nil)

	body.Grow(int(r.Raw.ContentLength))
	io.CopyBuffer(body, r.Raw.Body, nil)

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
	route := r.Ctx().Route()
	args := []string{}
	for k, v := range r.Args {
		args = append(args, k, v)
	}
	return UrlFor(route.fullName, true, args...)
}

// Returns a '*Slow.Ctx' of the current request
func (r *Request) Ctx() *Ctx { return contextsNamed[r.ctx] }

// Returns a 'context.Context' of the current request
func (r *Request) Context() context.Context { return r.Raw.Context() }

// Abort the current request. Server does not respond to client
func (r *Request) Cancel() {
	r.Raw.Context().Done()
}
