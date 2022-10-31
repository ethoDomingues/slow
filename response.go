package slow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (r *Response) Ctx() *Ctx { return contextsNamed[r.ctx] }

func (r *Response) _afterRequest() {
	ctx := r.Ctx()
	method := ctx.Request.Method
	if ctx.Request.MatchInfo.Router.Cors != nil {
		ctx.Request.MatchInfo.Router.Cors.parse(r.Headers)
	}
	if ctx.Request.MatchInfo.Route.Cors != nil {
		ctx.Request.MatchInfo.Route.Cors.parse(r.Headers)
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

func (r *Response) SetCookie(cookie *http.Cookie) { r.Headers.SetCookie(cookie) }

func (r *Response) Abort(code int) {
	r.Body.Reset()
	r.Body.WriteString(fmt.Sprint(code, " ", http.StatusText(code)))
	r.StatusCode = code
	panic(HttpAbort)
}
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
	panic(HttpAbort)
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
	panic(HttpAbort)
}

func (r *Response) HTML(body string, code int) {
	r.Body.Reset()

	r.StatusCode = code
	r.Headers.Set("Content-Type", "text/html")
	r.Body.WriteString(body)
	panic(HttpAbort)
}

func (r *Response) BadRequest()          { r.Abort(400) }
func (r *Response) Unauthorized()        { r.Abort(401) }
func (r *Response) Forbidden()           { r.Abort(403) }
func (r *Response) NotFound()            { r.Abort(404) }
func (r *Response) MethodNotAllowed()    { r.Abort(405) }
func (r *Response) ImATaerpot()          { r.Abort(418) }
func (r *Response) InternalServerError() { r.Abort(500) }

func Abort(code int) { panic(fmt.Errorf("abort:%d", code)) }
