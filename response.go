package slow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ethoDomingues/c3po"
)

func NewResponse(wr http.ResponseWriter, ctx *Ctx) *Response {
	return &Response{
		Buffer:     bytes.NewBufferString(""),
		raw:        wr,
		ctx:        ctx,
		Header:     Header{},
		StatusCode: 200,
	}
}

type Response struct {
	*bytes.Buffer

	StatusCode int

	Header Header

	ctx *Ctx
	raw http.ResponseWriter
}

func (r *Response) RawResponse() http.ResponseWriter {
	return r.raw
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
		r.Reset()
	}
}

// Set a cookie in the Headers of Response
func (r *Response) SetCookie(cookie *http.Cookie) { r.Header.SetCookie(cookie) }

// Redirect to Following URL
func (r *Response) Redirect(url string) {
	r.Reset()
	r.Header.Set("Location", url)
	r.StatusCode = 302

	r.Header.Set("Content-Type", "text/html; charset=utf-8")
	r.WriteString("<a href=\"" + c3po.HtmlEscape(url) + "\"> Manual Redirect </a>.\n")
	panic(ErrHttpAbort)
}

func (r *Response) JSON(body any, code int) {
	r.Reset()

	j, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		panic(err)
	}

	r.StatusCode = code
	r.Header.Set("Content-Type", "application/json")
	r.Write(j)
	panic(ErrHttpAbort)
}

func (r *Response) TEXT(body any, code int) {
	r.Reset()
	r.StatusCode = code
	r.Header.Set("Content-Type", "text/plain")
	r._write(body)
	panic(ErrHttpAbort)
}

func (r *Response) textCode(body any, code int) {
	r.StatusCode = code
	if body == nil {
		body = fmt.Sprintf("%d %s", code, http.StatusText(code))
	}
	fmt.Fprint(r, body)
	panic(ErrHttpAbort)
}

func (r *Response) HTML(body any, code int) {
	r.Reset()
	r.StatusCode = code
	r.Header.Set("Content-Type", "text/html")
	r._write(body)
	panic(ErrHttpAbort)
}

// Halts execution and closes the "response".
// This does not clear the response body
func (r *Response) Close() { panic(ErrHttpAbort) }

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

func (r *Response) checkErrByEnv(err ...any) {
	r.InternalServerError()
}

// Abort the current request. Server does not respond to client
func (r *Request) Cancel() { r.Context().Done() }

// Break execution, cleans up the response body, and writes the StatusCode to the response
func Abort(code int) { panic("abort:" + fmt.Sprint(code)) }
