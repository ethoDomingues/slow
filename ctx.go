package slow

import (
	"time"
)

// Returns a new *Slow.Ctx
func newCtx(app *App) *Ctx {
	c := &Ctx{
		id:        time.Now().String(),
		App:       app,
		Global:    map[string]any{},
		MatchInfo: &MatchInfo{},
	}
	if app.SecretKey != "" {
		c.Session = newSession(app.SecretKey)
	} else {
		c.Session = &Session{}
	}

	c.MatchInfo.ctx = c.id

	return c
}

type Ctx struct {

	// Ctx ID
	id string

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

// Get Current Route
func (c *Ctx) Router() *Router { return c.MatchInfo.Router() }

// Get Current Router
func (c *Ctx) Route() *Route { return c.MatchInfo.Route() }
