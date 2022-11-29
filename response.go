package slow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

func Abort(code int) { panic(fmt.Errorf("abort:%d", code)) }

func NewResponse(wr http.ResponseWriter, ctxID string) *Response {
	return &Response{
		raw:        wr,
		ctx:        ctxID,
		Body:       bytes.NewBufferString(""),
		Headers:    &Headers{},
		StatusCode: 200,
	}
}

type Response struct {
	ctx        string
	StatusCode int

	raw     http.ResponseWriter
	Body    *bytes.Buffer
	Headers *Headers
}

func (r *Response) parseHeaders() {
	ctx := r.Ctx()
	method := ctx.Request.Method
	routerCors := ctx.MatchInfo.Router().Cors
	if routerCors != nil {
		routerCors.parse(r.Headers)
	}
	routeCors := ctx.MatchInfo.Route().Cors
	if routeCors != nil {
		routeCors.parse(r.Headers)
	}
	if method == "OPTIONS" {
		h := *(r.Headers)
		if _, ok := h["Access-Control-Allow-Origin"]; !ok {
			h.Set("Access-Control-Allow-Origin", ctx.App.Servername)
		}
		if _, ok := h["Access-Control-Allow-Headers"]; !ok {
			h.Set("Access-Control-Allow-Headers", "")
		}
		if _, ok := h["Access-Control-Expose-Headers"]; !ok {
			h.Set("Access-Control-Expose-Headers", "")
		}
		if _, ok := h["Access-Control-Request-Method"]; !ok {
			h.Set("Access-Control-Request-Method", method)
		}
		if _, ok := h["Access-Control-Allow-Credentials"]; !ok {
			h.Set("Access-Control-Allow-Credentials", "false")
		}
	} else if ctx.Request.Raw.Method == "HEAD" {
		r.Body.Reset()
	}
}

// Halts execution and closes the "response".
// This does not clear the response body
func (r *Response) Close() {
	panic(ErrHttpAbort)
}

// Returns a current "*slow.Ctx"
func (r *Response) Ctx() *Ctx { return contextsNamed[r.ctx] }

// Set a cookie in the Headers of Response
func (r *Response) SetCookie(cookie *http.Cookie) { r.Headers.SetCookie(cookie) }

// Stops execution, cleans up the response body, and writes the StatusCode to the response
func (r *Response) Abort(code int) {
	r.Body.Reset()
	r.Body.WriteString(fmt.Sprint(code, " ", http.StatusText(code)))
	r.StatusCode = code
	panic(ErrHttpAbort)
}

// Redirect to Following URL
func (r *Response) Redirect(url string) {
	ctx := r.Ctx()
	rq := ctx.Request

	r.Body.Reset()
	r.Headers.Set("Location", url)
	if rq.Raw.Method == "GET" || rq.Raw.Method == "HEAD" {
		r.Headers.Set("Content-Type", "text/html; charset=utf-8")
	}
	r.StatusCode = 302
	r.Body.WriteString("<a href=\"" + HtmlEscape(url) + "\"> Manual Redirect </a>.\n")
	panic(ErrHttpAbort)
}

func (r *Response) JSON(body any, code int) {
	r.Body.Reset()

	j, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	r.StatusCode = code
	r.Headers.Set("Content-Type", "application/json")
	r.Body.Write(j)
	panic(ErrHttpAbort)
}

func (r *Response) HTML(body string, code int) {
	r.Body.Reset()

	r.StatusCode = code
	r.Headers.Set("Content-Type", "text/html")
	r.Body.WriteString(body)
	panic(ErrHttpAbort)
}

// Send a StatusOk, but without the body
func (r *Response) Ok() { r.Abort(200) }

// Send a BadRequest
func (r *Response) BadRequest() { r.Abort(400) }

// Send a Unauthorized
func (r *Response) Unauthorized() { r.Abort(401) }

// Send a StatusForbidden
func (r *Response) Forbidden() { r.Abort(403) }

// Send a StatusNotFound
func (r *Response) NotFound() { r.Abort(404) }

// Send a StatusMethodNotAllowed
func (r *Response) MethodNotAllowed() { r.Abort(405) }

// Send a StatusImATaerpot
func (r *Response) ImATaerpot() { r.Abort(418) }

// Send a StatusInternalServerError
func (r *Response) InternalServerError() { r.Abort(500) }

// Parse Html file and send to client
func (r *Response) RenderTemplate(pathToFile string, data ...any) {
	ctx := r.Ctx()
	dir, file := filepath.Split(ctx.App.TemplateFolder + pathToFile)
	d := http.Dir(dir)
	if f, err := d.Open(file); err == nil {
		defer f.Close()
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, f)
		t, err := template.New(file).Parse(buf.String())
		if err != nil {
			l.err.Panic(err)
		}
		t.Execute(r.Body, data)
		r.Close()
	} else {
		l.err.Panicln(err)
	}
}
