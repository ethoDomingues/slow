package slow

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/ethoDomingues/c3po"
)

type Func func(ctx *Ctx)

type Schema any

type Meth struct {
	Func
	Schema
	schemaFielder *c3po.Fielder
}

type MapCtrl map[string]*Meth

type Route struct {
	Url  string
	Name string

	Func    Func
	Methods []string

	Cors        *Cors
	Schema      any
	MapCtrl     MapCtrl
	Middlewares Middlewares

	router     *Router
	simpleName string

	simpleUrl   string
	urlRegex    []*regexp.Regexp
	isUrlPrefix bool

	parsed,
	_isStatic bool
	hasSufix bool
}

func (r *Route) compileUrl() {
	uri := r.Url
	isPrefix := false
	r.hasSufix = strings.HasSuffix(r.simpleUrl, "/")
	if uri != "" && uri != "/" {
		uri = strings.TrimPrefix(strings.TrimSuffix(uri, "/"), "/")
		strs := strings.Split(uri, "/")

		for _, str := range strs {
			if str == "" {
				continue
			}
			if isPrefix {
				log.Panicf("Url Variable Invalid: '%s'", str)
			}
			if re.all.MatchString(str) {
				isPrefix = true
			}
			if re.dot2.MatchString(str) {
				str = re.dot2.ReplaceAllString(str, "/")
			}
			if re.slash2.MatchString(str) {
				str = re.slash2.ReplaceAllString(str, "/")
			}

			if re.isVar.MatchString(str) {
				str = re.str.ReplaceAllString(str, `(([\x00-\x7F]+)([^\\\/\s]+)|\d+)`)
				str = re.digit.ReplaceAllString(str, `(\d+)`)
			}
			if !isPrefix {
				r.urlRegex = append(r.urlRegex, regexp.MustCompile(fmt.Sprintf("^%s$", str)))
			}
		}
	}
	if r.hasSufix {
		r.Url = r.Url + "/"
	}
	r.isUrlPrefix = isPrefix
}

func (r *Route) compileMethods() {
	ctrl := MapCtrl{"OPTIONS": &Meth{}}

	// allow Route{URL:"/",Name:"index",Func:func()} with method default "GET"
	if len(r.MapCtrl) == 0 && len(r.Methods) == 0 {
		r.Methods = []string{"GET"}
	}

	for verb, m := range r.MapCtrl {
		v := strings.ToUpper(verb)
		if !reMethods.MatchString(v) {
			l.err.Fatalf("route '%s' has invalid Request Method: '%s'", r.Name, verb)
		}
		if m.Schema != nil {
			sch := c3po.ParseSchemaWithTag("slow", m.Schema)
			m.schemaFielder = sch
			m.Schema = nil

		}
		ctrl[v] = m
	}

	r.MapCtrl = ctrl

	for _, verb := range r.Methods {
		v := strings.ToUpper(verb)
		if !reMethods.MatchString(v) {
			l.err.Fatalf("route '%s' has invalid Request Method: '%s'", r.Name, verb)
		}

		if _, ok := r.MapCtrl[v]; !ok {
			r.MapCtrl[v] = &Meth{
				Func: r.Func,
			}

			if r.Schema != nil {
				sch := c3po.ParseSchemaWithTag("slow", r.Schema)
				r.MapCtrl[v].schemaFielder = sch
			}
		}
	}

	r.Methods = []string{}
	for verb := range r.MapCtrl {
		r.Methods = append(r.Methods, verb)
		if verb == "GET" {
			if _, ok := r.MapCtrl["HEAD"]; !ok {
				r.Methods = append(r.Methods, "HEAD")
			}
		}
	}

	if len(r.Methods) <= 1 {
		r.Methods = []string{"GET", "HEAD"}
		r.MapCtrl["GET"] = &Meth{
			Func: r.Func,
		}
	}
}

func (r *Route) parse() {
	if r.Func == nil && r.MapCtrl == nil {
		l.err.Fatalf("Route '%s' need a Func or Ctrl\n", r.Name)
	}

	r.compileUrl()
	r.compileMethods()

	if r.Cors != nil {
		r.Cors.AllowMethods = strings.Join(r.Methods, ", ")
	} else {
		r.Cors = &Cors{AllowMethods: strings.Join(r.Methods, ", ")}
	}
}

func (r *Route) matchURL(ctx *Ctx, url string) bool {
	// if url == /  and route.Url == /
	if url == "/" && url == r.Url {
		return true
	}

	nurl := strings.TrimPrefix(url, "/")
	nurl = strings.TrimSuffix(nurl, "/")
	urlSplit := strings.Split(nurl, "/")

	lSplit := len(urlSplit)
	lRegex := len(r.urlRegex)

	if lSplit != lRegex {
		if ctx.App.EnableStatic && r._isStatic {
			if strings.HasPrefix(ctx.Request.URL.Path, ctx.App.StaticUrlPath) {
				return true
			}
		}
		if r.isUrlPrefix {
			if lRegex < lSplit {
				for i, uRe := range r.urlRegex {
					str := urlSplit[i]
					if !uRe.MatchString(str) {
						return false
					}
				}
				return true
			}
		}
		return false
	}

	for i, uRe := range r.urlRegex {
		str := urlSplit[i]
		if !uRe.MatchString(str) {
			return false
		}
	}

	if ctx.App.StrictSlash {
		last := string(url[len(url)-1])
		if r.hasSufix && last == "/" {
			return true
		} else if !r.hasSufix && last != "/" {
			return true
		}
		return false
	}
	return true
}

func (r *Route) match(ctx *Ctx) bool {
	mi := ctx.MatchInfo
	rq := ctx.Request
	m := rq.Method
	url := rq.URL.Path

	if !r.matchURL(ctx, url) {
		return false
	}
	if m == "HEAD" {
		m = "GET"
	}

	if meth, ok := r.MapCtrl[m]; ok {
		mi.MethodNotAllowed = nil
		if meth.Func != nil {
			mi.Func = meth.Func
		}
		mi.Match = true
		mi.Route = r
		ctx.SchemaFielder = meth.schemaFielder
		return true
	}

	mi.Route = nil
	mi.MethodNotAllowed = ErrorMethodMismatch
	return false
}

func (r *Route) GetRouter() *Router {
	if r.router == nil {
		l.err.Panicf("unregistered route: '%s'", r.Name)
	}
	return r.router
}

/*


 */

func GET(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"GET"},
	}
}

func HEAD(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"HEAD"},
	}
}

func POST(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"POST"},
	}
}

func PUT(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PUT"},
	}
}

func DELETE(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"DELETE"},
	}
}

func CONNECT(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"CONNECT"},
	}
}

func OPTIONS(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"OPTIONS"},
	}
}

func TRACE(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"TRACE"},
	}
}

func PATCH(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PATCH"},
	}
}

func Handler(h http.Handler) Func {
	if h == nil {
		panic("handler is nil")
	}
	return func(ctx *Ctx) { h.ServeHTTP(ctx, ctx.Request.raw) }
}
