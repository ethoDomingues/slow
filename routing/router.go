package routing

import "net/http"

type Router struct {
	routes       []Route
	routesByName map[string]*Route
}

func (r *Router) Parse() {
	for _, route := range r.routes {
		route.parse()
	}
}

func (r *Router) Match(req *http.Request, mi *MatchInfo) bool {
	for _, route := range r.routes {
		if route.match(req, mi) {
			mi.Router = r
			return true
		}
	}
	return false
}

func (r *Router) Add(route *Route) {
	if r.routes == nil {
		r.routes = []Route{}
		r.routesByName = map[string]*Route{}
	}
	if _, ok := r.routesByName[route.Name]; ok {
		panic(route.Name + " already registered!")
	}

	r.routes = append(r.routes, *route)
	r.routesByName[route.Name] = route
}

func (r *Router) Get(url string, f Func) {
	r.Add(&Route{
		Url:     url,
		Func:    f,
		Name:    GetFunctionName(f),
		Methods: []string{"GET"},
	})
}
