package slow

import (
	"net/http"

	"github.com/ethodomingues/slow/routing"
)

type App struct {
	routing.Router
	routers []routing.Router
}

func (app *App) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	mi := &routing.MatchInfo{}
	for _, router := range app.routers {
		if router.Match(req, mi) {
			break
		}
	}
	if mi.Route != nil {
		mi.Route.Func(req, wr)
	} else if mi.MethodNotAllowed != nil {
		wr.WriteHeader(405)
		wr.Write([]byte("405 " + http.StatusText(405)))
	} else {
		wr.WriteHeader(404)
		wr.Write([]byte("404 " + http.StatusText(404)))
	}
}

func (app *App) Listen() {
	http.ListenAndServe("0.0.0.0:5000", app)
}
