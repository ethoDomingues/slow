package slow

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ethoDomingues/c3po"
)

func NewResponse(wr http.ResponseWriter, ctx *Ctx) *Response {
	return &Response{
		Buffer:     bytes.NewBufferString(""),
		raw:        wr,
		ctx:        ctx,
		Headers:    Header{},
		StatusCode: 200,
	}
}

type Response struct {
	*bytes.Buffer
	StatusCode int
	Headers    Header
	ctx        *Ctx
	raw        http.ResponseWriter
}

func (r *Response) _write(v any) {
	if reader, ok := v.(io.Reader); ok {
		io.Copy(r, reader)
	} else {
		r.WriteString(fmt.Sprint(v))
	}
}

func (r *Response) parseHeaders() {
	ctx := r.ctx
	method := ctx.Request.Method
	routerCors := ctx.MatchInfo.Router.Cors
	if routerCors != nil {
		routerCors.parse(r.Headers, ctx.Request)
	}
	routeCors := ctx.MatchInfo.Route.Cors
	h := r.Headers
	if routeCors != nil {
		routeCors.parse(r.Headers, ctx.Request)
	}
	if routerCors != nil || routeCors != nil {
		if _, ok := h["Access-Control-Request-Method"]; !ok {
			h.Set("Access-Control-Request-Method", method)
		}
	}
	if method == "OPTIONS" {
		if _, ok := h["Access-Control-Allow-Origin"]; !ok {
			h.Set("Access-Control-Allow-Origin", ctx.App.Servername)
		}
		if _, ok := h["Access-Control-Allow-Headers"]; !ok {
			h.Set("Access-Control-Allow-Headers", "")
		}
		if _, ok := h["Access-Control-Expose-Headers"]; !ok {
			h.Set("Access-Control-Expose-Headers", "")
		}
		if _, ok := h["Access-Control-Allow-Credentials"]; !ok {
			h.Set("Access-Control-Allow-Credentials", "false")
		}
	} else if ctx.Request.Method == "HEAD" {
		r.Reset()
	}
}

func (r *Response) textCode(body any, code int) {
	r.StatusCode = code
	if body == nil {
		body = fmt.Sprintf("%d %s", code, http.StatusText(code))
	}
	r.Headers.Set("Content-Type", "text/html")
	fmt.Fprint(r, body)
	panic(ErrHttpAbort)
}

func (r Response) Header() http.Header            { return http.Header(r.Headers) }
func (r Response) Write(b []byte) (int, error)    { return r.Buffer.Write(b) }
func (r Response) WriteHeader(statusCode int)     { (&r).StatusCode = statusCode }
func (r *Response) SetCookie(cookie *http.Cookie) { r.Headers.SetCookie(cookie) }

// Redirect to Following URL
func (r *Response) Redirect(url string) {
	r.Reset()
	r.Headers.Set("Location", url)
	r.StatusCode = 302

	r.Headers.Set("Content-Type", "text/html; charset=utf-8")
	r.WriteString("<a href=\"" + c3po.HtmlEscape(url) + "\"> Manual Redirect </a>.\n")
	panic(ErrHttpAbort)
}

func (r *Response) JSON(body any, code int) {
	r.Reset()
	r.StatusCode = code
	r.Headers.Set("Content-Type", "application/json")
	if b, ok := body.(string); ok {
		r.WriteString(b)
		panic(ErrHttpAbort)
	} else if b, ok := body.(error); ok {
		r.WriteString(b.Error())
		panic(ErrHttpAbort)
	}
	if j, err := json.Marshal(body); err != nil {
		panic(err)
	} else {
		r.Write(j)
		panic(ErrHttpAbort)
	}
}

func (r *Response) TEXT(body any, code int) {
	r.Reset()
	r.StatusCode = code
	r.Headers.Set("Content-Type", "text/plain")
	r._write(body)
	panic(ErrHttpAbort)
}

func (r *Response) HTML(body any, code int) {
	r.Reset()
	r.StatusCode = code
	r.Headers.Set("Content-Type", "text/html")
	r._write(body)
	panic(ErrHttpAbort)
}

func Abort(code int)                     { panic("abort:" + fmt.Sprint(code)) }
func (r *Response) Close()               { panic(ErrHttpAbort) }
func (r *Response) Ok()                  { r.textCode(nil, 200) }
func (r *Response) Created()             { r.textCode(nil, 201) }
func (r *Response) NoContent()           { r.textCode(nil, 204) }
func (r *Response) BadRequest()          { r.textCode(nil, 400) }
func (r *Response) Unauthorized()        { r.textCode(nil, 401) }
func (r *Response) Forbidden()           { r.textCode(nil, 403) }
func (r *Response) NotFound()            { r.textCode(nil, 404) }
func (r *Response) MethodNotAllowed()    { r.textCode(nil, 405) }
func (r *Response) ImATaerpot()          { r.textCode(nil, 418) }
func (r *Response) InternalServerError() { r.textCode(nil, 500) }

func (r *Response) RenderTemplate(tmpl string, data ...any) {
	var (
		t     *template.Template
		ok    bool
		value any
	)

	if t, ok = htmlTemplates[tmpl]; !ok || r.ctx.App.Env == "development" {
		f, err := os.Open(filepath.Join(r.ctx.App.TemplateFolder, tmpl))
		r.CheckErr(err)
		buf := bytes.NewBufferString("")
		_, err = io.Copy(buf, f)
		r.CheckErr(err)
		t, err = template.New(tmpl).
			Funcs(r.ctx.App.TemplateFuncs).
			Parse(buf.String())
		r.CheckErr(err)
		if htmlTemplates == nil {
			htmlTemplates = make(map[string]*template.Template)
		}
		htmlTemplates[tmpl] = t
	}
	if len(data) == 1 {
		value = data[0]
	}
	t.Execute(r, value)
	r.Close()
}

func (r *Response) CheckErr(err error) {
	if err != nil {
		l.err.Println(err)
		if r.ctx.App.Env == "developement" {
			r.HTML(err, 500)
		}
		r.InternalServerError()
	}
}

func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.raw.(http.Hijacker).Hijack()
}
