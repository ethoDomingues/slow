package slow

import (
	"log"
	"net/http"

	"github.com/ethodomingues/slow/routing"
)

func NewApp() *App {
	return &App{
		Router:  routing.Router{},
		routers: []*routing.Router{},
	}
}

type App struct {
	routing.Router
	routers []*routing.Router
}

func (app *App) parse() {
	app.Parse()
	for _, router := range app.routers {
		router.Parse()
	}
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
	log.Println(req.Method, req.URL.Path)
}

func (app *App) Listen() {
	app.parse()
	log.Println("Server is linsten in 0.0.0.0:5000")
	http.ListenAndServe("0.0.0.0:5000", app)
}
