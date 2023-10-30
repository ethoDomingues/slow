package slow

import (
	"fmt"
	"log"
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
	Name,
	Prefix,
	Subdomain string
	StrictSlash bool

	Cors        *Cors
	Routes      []*Route
	Middlewares []Func

	is_main        bool
	routesByName   map[string]*Route
	subdomainRegex *regexp.Regexp

	// httpHandler *http.Handler
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
			sub = re.digit.ReplaceAllString(sub, `(\d+)`)
		}
		if re.all.MatchString(sub) {
			log.Panicf("'{*}' is not alowwed in router subdomain: '%s'", r.Name)
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
			panic("the route needs to be named or have a 'Route.Func'")
		}
		route.Name = getFunctionName(route.Func)
	}
	route.simpleName = r.Name
	if r.Name != "" {
		route.Name = r.Name + "." + route.Name
	}
	if r.Prefix != "" && !strings.HasPrefix(r.Prefix, "/") {
		panic(fmt.Errorf("Router '%v' Prefix must start with slash or be a null string ", r.Name))
	} else if route.Url != "" && (!strings.HasPrefix(route.Url, "/") && !strings.HasSuffix(r.Prefix, "/")) {
		panic(fmt.Errorf("Route '%v' Prefix must start with slash or be a null String", r.Name))
	}
	if route.Url == "" {
		route.Url = "/"
	}
	route.simpleUrl = route.Url
	route.Url = filepath.Join(r.Prefix, route.Url)
	if _, ok := r.routesByName[route.Url]; ok {
		panic(fmt.Errorf("Route with name '%s' already registered", r.Name))
	}
	route.parse()
	r.routesByName[route.Name] = route
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

func (r *Router) AddRoute(routes ...*Route) {
	r.Routes = append(r.Routes, routes...)
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

func (r *Router) GET(url string, f Func) {
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

func (r *Router) POST(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"POST"},
	})
}

func (r *Router) PUT(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PUT"},
	})
}

func (r *Router) DELETE(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"DELETE"},
	})
}

func (r *Router) CONNECT(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"CONNECT"},
	})
}

func (r *Router) OPTIONS(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"OPTIONS"},
	})
}

func (r *Router) TRACE(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"TRACE"},
	})
}

func (r *Router) PATCH(url string, f Func) {
	r.AddRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PATCH"},
	})
}

func (r *Router) AllMethods(url string, f Func) {
	r.AddRoute(&Route{
		Url:  url,
		Func: f,
		Name: getFunctionName(f),
		Methods: []string{
			"GET", "POST", "PUT", "DELETE",
			"CONNECT", "OPTIONS", "TRACE", "PATCH"},
	})
}
