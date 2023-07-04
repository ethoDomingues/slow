package slow

import "github.com/ethoDomingues/c3po"

// Returns a new *Slow.Ctx
func newCtx(app *App) *Ctx {
	c := &Ctx{
		App:       app.clone(),
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

	// Current Request
	Request *Request

	// Current Response
	Response *Response

	// Current Cookie Session
	Session *Session

	// New Schema valid from route schema
	Schema any

	SchemaFielder *c3po.Fielder

	// Contains information about the current request, route, etc...
	MatchInfo *MatchInfo
}

func (ctx *Ctx) UrlFor(name string, external bool, args ...string) string {
	return ctx.App.UrlFor(name, external, args...)
}
