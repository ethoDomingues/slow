package slow

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

	// Current App
	App *App

	Global map[string]any

	// Current Request
	Request *Request

	// Current Response
	Response *Response

	Session *Session

	// Contains information about the current request and the route
	MatchInfo *MatchInfo
}

func (ctx *Ctx) UrlFor(name string, external bool, args ...string) string {
	return ctx.App.UrlFor(name, external, args...)
}
