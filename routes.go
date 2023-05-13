package slow

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ethodomingues/c3po"
)

type Func func(*Ctx)
type Methods []string
type Schema any

type Meth struct {
	Func
	Schema
	schemaFielder *c3po.Fielder
}

type MapCtrl map[string]*Meth

type Route struct {
	Url     string
	Name    string
	Func    Func
	MapCtrl MapCtrl

	Cors        *Cors
	Schema      any
	Methods     Methods
	Middlewares Middlewares

	fullName string
	fullUrl  string
	urlRegex []*regexp.Regexp
}

func (r *Route) compileUrl() {
	url := strings.TrimPrefix(r.fullUrl, "/")
	url = strings.TrimSuffix(url, "/")
	strs := strings.Split(url, "/")

	for _, str := range strs {
		if str == "" {
			continue
		}
		if re.dot2.MatchString(str) {
			re.dot2.ReplaceAllString(str, "/")
		}
		if re.slash2.MatchString(str) {
			re.slash2.ReplaceAllString(str, "/")
		}
		if re.isVar.MatchString(str) {
			str = re.str.ReplaceAllString(str, `(([\x00-\x7F]+)([^\\\/\s]+)|\d+)`)
			str = re.digit.ReplaceAllString(str, `(\d+)`)
			str = re.filepath.ReplaceAllString(str, `(.{0,})`)
		}
		r.urlRegex = append(r.urlRegex, regexp.MustCompile(fmt.Sprintf("^%s$", str)))
	}
	if strings.HasSuffix(r.fullUrl, "/") {
		r.urlRegex = append(r.urlRegex, regexp.MustCompile(`^(\/)?$`))
	}
}

func (r *Route) compileMethods() {
	ctrl := MapCtrl{"OPTIONS": &Meth{}}

	for verb, meth := range r.MapCtrl {
		v := strings.ToUpper(verb)
		if !reMethods.MatchString(v) {
			l.err.Fatalf("route '%s' has invalid Request Method: '%s'", r.fullName, verb)
		}
		if meth.Schema != nil {
			meth.schemaFielder = c3po.ParseSchemaWithTag("slow", meth.Schema)
			meth.Schema = nil
		}
		ctrl[v] = meth
	}
	if r.MapCtrl == nil {
		r.MapCtrl = MapCtrl{}
	}
	for _, verb := range r.Methods {
		v := strings.ToUpper(verb)
		if !reMethods.MatchString(v) {
			l.err.Fatalf("route '%s' has invalid Request Method: '%s'", r.fullName, verb)
		}
		if meth, ok := r.MapCtrl[v]; !ok {
			r.MapCtrl[v] = &Meth{
				Func:   r.Func,
				Schema: r.Schema,
			}
		} else if r.Schema != nil {
			meth.schemaFielder = c3po.ParseSchemaWithTag("slow", r.Schema)
		}
	}
	r.MapCtrl = ctrl
}

func (r *Route) parse() {
	if r.Func == nil && r.MapCtrl == nil {
		l.err.Fatalf("Route '%s' need a func or Ctrl\n", r.fullName)
	}
	r.compileUrl()
	r.compileMethods()

	if r.Cors != nil {
		r.Cors.AllowMethods = r.Methods
	} else {
		r.Cors = &Cors{AllowMethods: r.Methods}
	}
}

func (r *Route) matchURL(ctx *Ctx, url string) bool {
	nurl := strings.TrimPrefix(url, "/")
	nurl = strings.TrimSuffix(nurl, "/")
	urlSplit := strings.Split(nurl, "/")

	lSplit := len(urlSplit)
	lRegex := len(r.urlRegex)

	if lSplit != lRegex {
		if strings.HasPrefix(url, ctx.App.StaticUrlPath) {
			return true
		}
	}

	for i, uRe := range r.urlRegex {
		var str string
		if i < lSplit {
			str = urlSplit[i]
		}
		if !uRe.MatchString(str) {
			return false
		}
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
		mi.ctx.SchemaFielder = meth.schemaFielder
		return true
	}
	mi.Route = nil
	mi.MethodNotAllowed = ErrorMethodMismatch
	return false
}

func Get(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"GET"},
	}
}

func Head(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"HEAD"},
	}
}

func Post(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"POST"},
	}
}

func Put(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PUT"},
	}
}

func Delete(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"DELETE"},
	}
}

func Connect(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"CONNECT"},
	}
}

func Options(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"OPTIONS"},
	}
}

func Trace(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"TRACE"},
	}
}

func Patch(url string, f Func) *Route {
	return &Route{
		Url:     url,
		Func:    f,
		Name:    getFunctionName(f),
		Methods: []string{"PATCH"},
	}
}
