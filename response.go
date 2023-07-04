package slow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	r.WriteString("<a href=\"" + HtmlEscape(url) + "\"> Manual Redirect </a>.\n")
	panic(ErrHttpAbort)
}

func (r *Response) JSON(body any, code int) {
	r.Reset()

	j, err := json.Marshal(body)
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
	} else {
		body = http.StatusText(code)
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

func (r *Response) Ok(body ...any) { r.textCode(fmt.Sprint(body...), 200) }

func (r *Response) Created(body ...any) { r.textCode(fmt.Sprint(body...), 201) }

func (r *Response) NoContent() { r.textCode("", 204) }

func (r *Response) BadRequest(body ...any) { r.textCode(fmt.Sprint(body...), 400) }

func (r *Response) Unauthorized(body ...any) { r.textCode(fmt.Sprint(body...), 401) }

func (r *Response) Forbidden(body ...any) { r.textCode(fmt.Sprint(body...), 403) }

func (r *Response) NotFound(body ...any) { r.textCode(fmt.Sprint(body...), 404) }

func (r *Response) MethodNotAllowed(body ...any) { r.textCode(fmt.Sprint(body...), 405) }

func (r *Response) ImATaerpot(body ...any) { r.textCode(fmt.Sprint(body...), 418) }

func (r *Response) InternalServerError(body ...any) { r.textCode(fmt.Sprint(body...), 500) }

func (r *Response) checkErrByEnv(err ...any) {
	if r.ctx.App.Env == "" || r.ctx.App.Env == "development" {
		r.InternalServerError(err...)
	} else {
		r.InternalServerError()
	}
}

// Abort the current request. Server does not respond to client
func (r *Request) Cancel() { r.Context().Done() }

// Break execution, cleans up the response body, and writes the StatusCode to the response
func Abort(code int) { panic("abort:" + fmt.Sprint(code)) }
