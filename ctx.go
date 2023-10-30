package slow

import (
	"net/http"

	"github.com/ethoDomingues/c3po"
)

// Returns a new *Slow.Ctx
func newCtx(app *App) *Ctx {
	return &Ctx{
		App:       app,
		Global:    map[string]any{},
		MatchInfo: &MatchInfo{},
		Session:   newSession(app.SecretKey),
	}
}

type Ctx struct {

	// Clone Current App
	App *App

	Global map[string]any

	*Response          // Current Response
	Request   *Request // Current Request

	// Current Cookie Session
	Session *Session

	// New Schema valid from route schema
	Schema        any
	SchemaFielder *c3po.Fielder

	// Contains information about the current request, route, etc...
	MatchInfo *MatchInfo
	mids      Middlewares
	c_mid     int
}

// executes the next middleware or main function of the request
func (ctx *Ctx) Next() {
	if ctx.c_mid < len(ctx.mids) {
		n := ctx.mids[ctx.c_mid]
		ctx.c_mid += 1
		n(ctx)
	} else {
		panic(ErrHttpAbort)
	}
}

func (ctx *Ctx) parseMids() {
	ctx.mids = append(ctx.mids, ctx.MatchInfo.Router.Middlewares...)
	ctx.mids = append(ctx.mids, ctx.MatchInfo.Route.Middlewares...)
	ctx.mids = append(ctx.mids, ctx.MatchInfo.Func)
}

func (ctx *Ctx) UrlFor(name string, external bool, args ...string) string {
	return ctx.App.UrlFor(name, external, args...)
}

func (app *App) TestCtx(req ...*http.Request) *Ctx {
	if !app.built {
		l.err.Panic("Vc ta tentando usar o Ctx fora do contexto. Vc precisa Buildar o app primeiro")
	}
	var rq *http.Request
	if len(req) > 0 {
		rq = req[0]
	}
	ctx := newCtx(app)
	ctx.Request = NewRequest(rq, ctx)
	ctx.Response = NewResponse(Response{}, ctx)
	return ctx
}
