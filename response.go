package slow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func NewResponse(wr http.ResponseWriter, ctx *Ctx) *Response {
	return &Response{
		raw:        wr,
		ctx:        ctx,
		Body:       bytes.NewBufferString(""),
		Header:     Header{},
		StatusCode: 200,
	}
}

type Response struct {
	StatusCode int

	Body   *bytes.Buffer
	Header Header

	ctx *Ctx
	raw http.ResponseWriter
}

func (r *Response) parseHeaders() {
	ctx := r.ctx
	method := ctx.Request.Method
	routerCors := ctx.MatchInfo.Router.Cors
	if routerCors != nil {
		routerCors.parse(r.Header)
	}
	routeCors := ctx.MatchInfo.Route.Cors
	if routeCors != nil {
		routeCors.parse(r.Header)
	}
	if method == "OPTIONS" {
		h := r.Header
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
	} else if ctx.Request.Method == "HEAD" {
		r.Body.Reset()
	}
}

// Set a cookie in the Headers of Response
func (r *Response) SetCookie(cookie *http.Cookie) { r.Header.SetCookie(cookie) }

// Redirect to Following URL
func (r *Response) Redirect(url string) {
	r.Body.Reset()
	r.Header.Set("Location", url)
	r.StatusCode = 302

	r.Header.Set("Content-Type", "text/html; charset=utf-8")
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
	r.Header.Set("Content-Type", "application/json")
	r.Body.Write(j)
	panic(ErrHttpAbort)
}

func (r *Response) TEXT(body string, code int) {
	r.Body.Reset()

	r.StatusCode = code
	r.Header.Set("Content-Type", "text/plain")
	r.Body.WriteString(body)
	panic(ErrHttpAbort)
}

func (r *Response) HTML(body string, code int) {
	r.Body.Reset()

	r.StatusCode = code
	r.Header.Set("Content-Type", "text/html")
	r.Body.WriteString(body)
	panic(ErrHttpAbort)
}

// Halts execution and closes the "response".
// This does not clear the response body
func (r *Response) Close() { panic(ErrHttpAbort) }

// Send a StatusOk ( Status and Text )
func (r *Response) Ok(body ...any) { r.TEXT(fmt.Sprint(body...), 200) }

// Send a BadRequest ( Status and Text )
func (r *Response) BadRequest(body ...any) { r.TEXT(fmt.Sprint(body...), 400) }

// Send a Unauthorized ( Status and Text )
func (r *Response) Unauthorized(body ...any) { r.TEXT(fmt.Sprint(body...), 401) }

// Send a StatusForbidden ( Status and Text )
func (r *Response) Forbidden(body ...any) { r.TEXT(fmt.Sprint(body...), 403) }

// Send a StatusNotFound ( Status and Text )
func (r *Response) NotFound(body ...any) { r.TEXT(fmt.Sprint(body...), 404) }

// Send a StatusMethodNotAllowed ( Status and Text )
func (r *Response) MethodNotAllowed(body ...any) { r.TEXT(fmt.Sprint(body...), 405) }

// Send a StatusImATaerpot ( Status and Text )
func (r *Response) ImATaerpot(body ...any) { r.TEXT(fmt.Sprint(body...), 418) }

// Send a StatusInternalServerError ( Status and Text )
func (r *Response) InternalServerError(body ...any) { r.TEXT(fmt.Sprint(body...), 500) }

// Parse Html file and send to client
func (r *Response) RenderTemplate(pathToFile string, data ...any) {
	ctx := r.ctx
	fullPath := filepath.Join(GetFullPath(), ctx.App.TemplateFolder, pathToFile)

	if f, err := os.Open(fullPath); err == nil {
		defer f.Close()
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, f)
		t, err := template.New("").Parse(buf.String())
		if err != nil {
			l.err.Panic(err)
		}
		t.Execute(r.Body, data)
		r.Close()
	} else {
		l.err.Println(filepath.Join(ctx.App.TemplateFolder, pathToFile), " does not exist")
		r.NotFound("Template Not Found")
	}
}

// Abort the current request. Server does not respond to client
func (r *Request) Cancel() {
	r.Context().Done()
}

// Stops execution, cleans up the response body, and writes the StatusCode to the response
func Abort(code int) { panic("abort:" + fmt.Sprint(code)) }
