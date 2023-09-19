package slow

import (
	"github.com/ethoDomingues/c3po"
)

// Returns a new *Slow.Ctx
func newCtx(app *App) *Ctx {
	c := &Ctx{
		App:       app,
		Global:    map[string]any{},
		MatchInfo: &MatchInfo{},
	}
	c.MatchInfo.ctx = c
	c.Session = newSession(app.SecretKey)

	return c
}

type Ctx struct {

	// Clone Current App
	App *App

	Global map[string]any

	Request  *Request  // Current Request
	Response *Response // Current Response

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
	if ctx.c_mid >= len(ctx.mids) {
		panic(ErrHttpAbort)
	}
	n := ctx.mids[ctx.c_mid]
	ctx.c_mid += 1
	n(ctx)
}

func (ctx *Ctx) UrlFor(name string, external bool, args ...string) string {
	return ctx.App.UrlFor(name, external, args...)
}
