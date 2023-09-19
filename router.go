package slow

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

func NewRouter(name string) *Router {
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
	is_main bool

	Name,
	Prefix,
	Subdomain string

	Cors        *Cors
	Routes      []*Route
	Middlewares []Func

	routesByName   map[string]*Route
	subdomainRegex *regexp.Regexp
}

func (r *Router) parse(servername string) {
	if r.routesByName == nil {
		r.routesByName = map[string]*Route{}
	}

	if r.Name == "" && !r.is_main {
		panic(fmt.Errorf("the routers must be named"))
	}
	if r.Subdomain != "" {
		if servername == "" {
			panic("to use subdomains you need to first add a ServerName in the 'app'")
		}
		sub := r.Subdomain
		if re.digit.MatchString(sub) {
			panic(fmt.Sprintf("router subdomain dont accept 'int' varible. Router: '%s'", r.Name))
		}
		if re.filepath.MatchString(sub) {
			panic(fmt.Sprintf("router subdomain dont accept 'filepath' varible. Router: '%s'", r.Name))
		}
		if re.str.MatchString(sub) {
			sub = re.str.ReplaceAllString(sub, `(\w+)`)
		} else {
			sub = "(" + sub + ")"
		}
		sub = sub + `(.` + servername + `)`
		r.subdomainRegex = regexp.MustCompile("^" + sub + "$")
	} else if servername != "" {
		r.subdomainRegex = regexp.MustCompile("^(" + servername + ")$")
	}

	for _, route := range r.Routes {
		if !route.parsed {
			r.parseRoute(route)
		}
	}
}

func (r *Router) parseRoute(route *Route) {
	if route.Name == "" {
		if route.Func == nil {
			l.err.Panic("route need be named")
		}
		route.Name = getFunctionName(route.Func)
	}
	if r.Name != "" {
		route.fullName = r.Name + "." + route.Name
	} else {
		route.fullName = route.Name
	}
	if r.Prefix != "" && !strings.HasPrefix(r.Prefix, "/") {
		panic(fmt.Errorf("Router '%v' Prefix must start with slash or be a null string ", r.Name))
	} else if route.Url != "" && (!strings.HasPrefix(route.Url, "/") && !strings.HasSuffix(r.Prefix, "/")) {
		panic(fmt.Errorf("Route '%v' Prefix must start with slash or be a null String", r.Name))
	}
	route.fullUrl = filepath.Join(r.Prefix, route.Url)
	if _, ok := r.routesByName[route.fullUrl]; ok {
		panic(fmt.Errorf("Route with name '%s' already registered", r.Name))
	}
	re.slash2.ReplaceAllString(route.fullName, "/")

	route.parse()
	r.routesByName[route.fullName] = route
	route.router = r
	route.parsed = true
}

func (r *Router) match(ctx *Ctx) bool {
	rq := ctx.Request

	if r.subdomainRegex != nil {
		if !r.subdomainRegex.MatchString(rq.URL.Host) {
			return false
		}
	}
	for _, route := range r.Routes {
		if route.match(ctx) {
			ctx.MatchInfo.Router = r
			return true
		}
	}
	return false
}

func (r *Router) AddRoute(route *Route) {
	if r.Routes == nil {
		r.Routes = []*Route{}
		r.routesByName = map[string]*Route{}
	}
	r.parseRoute(route)
	r.Routes = append(r.Routes, route)
}

func (r *Router) AddAll(routes ...*Route) {
	for _, route := range routes {
		r.AddRoute(route)
	}
}

func (r *Router) Add(url, name string, f Func, meths []string) {
	r.AddRoute(
		&Route{
			Name:    name,
			Url:     url,
			Func:    f,
			Methods: meths,
		})
}

func (r *Router) Get(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"GET"},
	})
}

func (r *Router) Head(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"HEAD"},
	})
}

func (r *Router) Post(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"POST"},
	})
}

func (r *Router) Put(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PUT"},
	})
}

func (r *Router) Delete(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"DELETE"},
	})
}

func (r *Router) Connect(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"CONNECT"},
	})
}

func (r *Router) Options(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"OPTIONS"},
	})
}

func (r *Router) Trace(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"TRACE"},
	})
}

func (r *Router) Patch(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PATCH"},
	})
}
