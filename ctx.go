package slow

import (
	"context"
	"time"
)

func NewCtx(app *App, ctx context.Context) *Ctx {
	c := &Ctx{
		id:      time.Now().String(),
		App:     app,
		Global:  map[string]any{},
		Session: NewSession(),
		Context: ctx,
	}
	c.Session.jwt.Secret = app.SecretKey
	return c
}

type Ctx struct {
	id       string
	App      *App
	Request  *Request
	Response *Response
	Session  *Session
	Global   map[string]any
	context.Context
}
