package slow

import (
	"fmt"
	"net"
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
	Middlewares Middlewares

	routesByName   map[string]*Route
	subdomainRegex *regexp.Regexp
}

func (r *Router) parse() {
	if r.Name == "" && !r.is_main {
		panic(fmt.Errorf("the routers must be named"))
	}

	// DYNAMIC SUBDOMAINS ARE NOT A GOOD IDEA
	if r.Subdomain != "" {
		if servername == "" {
			l.err.Fatal("to use subdomains you need to first add a ServerName in the 'app'")
		}
		sub := r.Subdomain
		if re.digit.MatchString(sub) {
			l.err.Fatalf("router subdomain dont accept 'int' varible. Router: '%s'", r.Name)
		}
		if re.filepath.MatchString(sub) {
			l.err.Fatalf("router subdomain dont accept 'filepath' varible. Router: '%s'", r.Name)
		}
		r.subdomainRegex = regexp.MustCompile(sub)
	}

	for _, route := range r.Routes {
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
		route.fullUrl = r.Prefix + route.Url
		re.slash2.ReplaceAllString(route.fullName, "/")

		route.parse()
		r.routesByName[route.fullName] = route
	}
}

func (r *Router) Match(ctx *Ctx) bool {
	rq := ctx.Request
	rqUrl := rq.Raw.Host
	if servername != "" {
		if !strings.Contains(rqUrl, servername) {
			return false
		}
	}
	if r.subdomainRegex != nil {
		if net.ParseIP(rqUrl) != nil {
			return false
		}
		// só para garantir q a proxima etapa não quebre
		if !strings.Contains(rqUrl, ".") {
			return false
		}
		u := strings.Split(rqUrl, ".")[0]
		// se o subdominio não der match
		if !r.subdomainRegex.MatchString(u) {
			return false
		}
	} else {
		if servername != "" && r.Subdomain == "" {
			if rqUrl != servername {
				return false
			}
		}
	}

	for _, route := range r.Routes {
		if route.Match(ctx) {
			ctx.MatchInfo.Router = r
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
		l.err.Panic(route.Name + " already registered!")
	}
	r.Routes = append(r.Routes, route)
	if r.Name != "" {
		r.routesByName[r.Name+"."+route.Name] = route
	} else {
		r.routesByName[route.Name] = route
	}
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

func (r *Router) ALL(url string, f Func) {
	r.addRoute(&Route{
		Url:  url,
		Func: f,
		Name: getFunctionName(f),
		Methods: []string{
			"GET", "HEAD", "POST",
			"PUT", "DELETE", "CONNECT",
			"OPTIONS", "TRACE", "PATCH",
		},
	})
}

func (r *Router) GET(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"GET"},
	})
}

func (r *Router) HEAD(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"HEAD"},
	})
}

func (r *Router) POST(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"POST"},
	})
}

func (r *Router) PUT(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PUT"},
	})
}

func (r *Router) DELETE(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"DELETE"},
	})
}

func (r *Router) CONNECT(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"CONNECT"},
	})
}

func (r *Router) OPTIONS(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"OPTIONS"},
	})
}

func (r *Router) TRACE(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"TRACE"},
	})
}

func (r *Router) PATCH(url string, f Func) {
	r.addRoute(&Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PATCH"},
	})
}
