package slow

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
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
		ctx:       ctxID,
		Raw:       req,
		Method:    req.Method,
		MatchInfo: &MatchInfo{},

		Args:    map[string]string{},
		Mime:    map[string]string{},
		Form:    map[string]any{},
		Files:   map[string][]*File{},
		Cookies: map[string]*http.Cookie{},
	}
}

type Request struct {
	*MatchInfo
	Raw         *http.Request
	ctx         string
	Body        string
	Method      string
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
	case r.ContentType == "", r.ContentType == "application/json":
		r.Body = body.String()
		json.Unmarshal(body.Bytes(), &r.Form)
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
}

func (r *Request) RequestURL() string {
	return r.UrlFor(r.MatchInfo.Route.fullName, true, r.Args)
}

func (r *Request) Parse() {
	r.parseHeaders()
	r.parseCookies()
	r.parseBody()
}

func (r *Request) Ctx() *Ctx { return contextsNamed[r.ctx] }

func (r *Request) UrlFor(name string, external bool, params map[string]string) string {
	return r.Ctx().App.UrlFor(name, external, params)
}
