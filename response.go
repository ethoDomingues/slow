package slow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
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
	h := r.Header
	if routeCors != nil {
		routeCors.parse(r.Header)
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

	if body == "" {
		body = fmt.Sprintf("%d %s", code, http.StatusText(code))
	}

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

func (r *Response) Ok(body ...any) { r.TEXT(fmt.Sprint(body...), 200) }

func (r *Response) Created(body ...any) { r.TEXT(fmt.Sprint(body...), 201) }

func (r *Response) NoContent() { r.TEXT("", 204) }

func (r *Response) BadRequest(body ...any) { r.TEXT(fmt.Sprint(body...), 400) }

func (r *Response) Unauthorized(body ...any) { r.TEXT(fmt.Sprint(body...), 401) }

func (r *Response) Forbidden(body ...any) { r.TEXT(fmt.Sprint(body...), 403) }

func (r *Response) NotFound(body ...any) { r.TEXT(fmt.Sprint(body...), 404) }

func (r *Response) MethodNotAllowed(body ...any) { r.TEXT(fmt.Sprint(body...), 405) }

func (r *Response) ImATaerpot(body ...any) { r.TEXT(fmt.Sprint(body...), 418) }

func (r *Response) InternalServerError(body ...any) { r.TEXT(fmt.Sprint(body...), 500) }

func (r *Response) internal_ServerError(err ...any) {
	if r.ctx.App.Env == "" || r.ctx.App.Env == "development" {
		r.InternalServerError(err...)
	} else {
		r.InternalServerError()
	}
}

func (r *Response) RenderTemplate(data any, pathToFile ...string) {

	template_Folder := r.ctx.App.TemplateFolder
	fullPath := path.Join(GetFullPath(), template_Folder)

	paths := []string{}
	for _, p := range pathToFile {
		paths = append(paths, path.Join(fullPath, p))
	}

	t, err := template.ParseFiles(paths...)

	if err != nil {
		r.internal_ServerError(err)
	}

	err = t.Execute(r.Body, data)
	if err != nil {
		if r.ctx.App.Env == "" || r.ctx.App.Env == "development" {
			r.InternalServerError(err)
		} else {
			r.InternalServerError()
		}
	}

	r.Close()
}

// Abort the current request. Server does not respond to client
func (r *Request) Cancel() { r.Context().Done() }

// Break execution, cleans up the response body, and writes the StatusCode to the response
func Abort(code int) { panic("abort:" + fmt.Sprint(code)) }
