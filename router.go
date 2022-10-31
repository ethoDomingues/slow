package slow

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func NewRouter(name string, cfg map[string]string) *Router {
	if name == "" {
		panic(fmt.Errorf("the routers must be named"))
	}
	return &Router{
		Name:         name,
		Routes:       []*Route{},
		routesByName: map[string]*Route{},
	}
}

type Router struct {
	_main          bool
	Name           string
	Prefix         string
	Subdomain      string
	Routes         []*Route
	subdomainRegex *regexp.Regexp
	routesByName   map[string]*Route
	Middlewares
	*Cors
}

func (r *Router) parse() {
	if r.Name == "" && !r._main {
		panic(fmt.Errorf("the routers must be named"))
	}
	for _, route := range r.Routes {
		route.fullName = r.Name + "." + route.Name
		if r.Prefix != "" && !strings.HasPrefix(r.Prefix, "/") {
			panic(fmt.Errorf("Router '%v' Prefix must start with slash or be a null string ", r.Name))
		} else if route.Url != "" && (!strings.HasPrefix(route.Url, "/") && !strings.HasSuffix(r.Prefix, "/")) {
			panic(fmt.Errorf("Route '%v' Prefix must start with slash or be a null String", r.Name))
		}
		route.fullUrl = r.Prefix + route.Url
		re.slash2.ReplaceAllString(route.fullName, "/")

		route.parse()
		r.routesByName[route.fullName] = route

	}

	if r.Subdomain != "" {
		sub := r.Subdomain
		if re.isVar.MatchString(r.Subdomain) {
			sub = re.str.ReplaceAllString(r.Subdomain, `(\w+)`)
			sub = re.digit.ReplaceAllString(r.Subdomain, `(\d+)`)
		}
		r.subdomainRegex = regexp.MustCompile(`^(` + sub + "[.]" + servername + `)$`)
	} else {
		r.subdomainRegex = regexp.MustCompile(`^(` + servername + `)$`)
	}
}

func (r *Router) Match(req *http.Request, mi *MatchInfo) bool {
	if !r.subdomainRegex.MatchString(req.Host) {
		return false
	}
	for _, route := range r.Routes {
		if route.Match(req, mi) {
			mi.Router = r
			return true
		}
	}
	return false
}

func (r *Router) addRoute(route *Route) {
	if r.Routes == nil {
		r.Routes = []*Route{}
		r.routesByName = map[string]*Route{}
	}
	if _, ok := r.routesByName[route.Name]; ok {
		panic(route.Name + " already registered!")
	}
	r.Routes = append(r.Routes, route)
	r.routesByName[r.Name+"."+route.Name] = route
}

func (r *Router) AddAll(routes ...*Route) {
	for _, route := range routes {
		r.addRoute(route)
	}
}

func (r *Router) Add(url, name string, f Func, meths []string) {
	r.addRoute(
		&Route{
			Name:    name,
			Url:     url,
			Func:    f,
			Methods: meths,
		})
}

func (r *Router) Get(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    GetFunctionName(f),
		Methods: []string{"GET"},
	})
}

func (r *Router) Post(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    GetFunctionName(f),
		Methods: []string{"POST"},
	})
}
