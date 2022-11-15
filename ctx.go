package slow

import (
	"context"
	"time"
)

func NewCtx(app *App, ctx context.Context) *Ctx {
	c := &Ctx{
		id:        time.Now().String(),
		App:       app,
		Global:    map[string]any{},
		Session:   NewSession(),
		Context:   ctx,
		MatchInfo: &MatchInfo{},
	}
	c.MatchInfo.ctx = c.id

	c.Session.jwt.Secret = app.SecretKey
	return c
}

type Ctx struct {
	id string
	context.Context

	App       *App
	Global    map[string]any
	Request   *Request
	Response  *Response
	Session   *Session
	MatchInfo *MatchInfo
}

func (c *Ctx) Router() *Router { return c.MatchInfo.Router() }
func (c *Ctx) Route() *Route   { return c.MatchInfo.Route() }
