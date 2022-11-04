package slow

import (
	"fmt"
	"regexp"
	"strings"
)

type Func func(*Ctx)
type Methods []string
type Schema any
type Meth struct {
	Func
	Method string
	Schema
}

type Ctrl map[string]*Meth

type Route struct {
	Func
	Methods

	Url    string
	Name   string
	Schema any
	Ctrl
	*Cors
	Middlewares

	fullName string
	fullUrl  string
	urlRegex []*regexp.Regexp
	Router   *Router
}

func (r *Route) compileUrl() {
	for _, str := range strings.Split(r.fullUrl, "/") {
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
			str = re.str.ReplaceAllString(str, `([\x00-\x7F]+[^\\\/\s]+)`)
			str = re.digit.ReplaceAllString(str, `(\d+)`)
			str = re.filepath.ReplaceAllString(str, `([\/\w+.-]+)`)
		}
		r.urlRegex = append(r.urlRegex, regexp.MustCompile(fmt.Sprintf("^%s$", str)))
	}
	if strings.HasSuffix(r.fullUrl, "/") {
		r.urlRegex = append(r.urlRegex, regexp.MustCompile(`^(\/)?$`))
	}
}

func (r *Route) parse() {
	if r.Func == nil && r.Ctrl == nil {
		l.err.Fatalf("Route '%s' need a func or Ctrl\n", r.fullName)
	}
	r.compileUrl()
	ctrl := Ctrl{"OPTIONS": nil}
	strMth := map[string]any{}
	for verb, meth := range r.Ctrl {
		v := strings.ToUpper(verb)
		if !reMethods.MatchString(v) {
			panic(fmt.Errorf("route '%s' has invalid Request Method: '%s'", r.fullName, verb))
		}
		ctrl[v] = meth
		strMth[v] = nil
	}
	r.Ctrl = ctrl
	for _, verb := range r.Methods {
		v := strings.ToUpper(verb)
		if !reMethods.MatchString(v) {
			panic(fmt.Errorf("route '%s' has invalid Request Method: '%s'", r.fullName, verb))
		}
		if _, ok := r.Ctrl[v]; !ok {
			r.Ctrl[v] = &Meth{
				Func:   r.Func,
				Schema: r.Schema,
				Method: v,
			}
		}
		strMth[v] = nil
	}
	if _, ok := r.Ctrl["GET"]; ok {
		r.Ctrl["HEAD"] = r.Ctrl["GET"]
		strMth["HEAD"] = nil
	}
	lmths := []string{"OPTIONS"}
	for v := range strMth {
		lmths = append(lmths, v)
	}
	r.Methods = lmths
	if r.Cors != nil {
		r.Cors.AllowMethods = lmths
	}
}

func (r *Route) matchURL(ctx *Ctx, url string) bool {
	urlSplit := strings.Split(strings.TrimPrefix(url, "/"), "/")
	if re.filepath.MatchString(r.fullUrl) {
		if strings.HasPrefix(url, ctx.App.StaticUrlPath) {
			return true
		}
		return false
	}
	if len(urlSplit) != len(r.urlRegex) {
		return false
	}
	for i, uRe := range r.urlRegex {
		str := urlSplit[i]
		if !uRe.MatchString(str) {
			return false
		}
	}
	return true
}

func (r *Route) Match(ctx *Ctx) bool {
	rq := ctx.Request
	mi := rq.MatchInfo
	m := rq.Method
	if r.matchURL(ctx, rq.Raw.URL.Path) {
		if meth, ok := r.Ctrl[m]; ok {
			mi.MethodNotAllowed = nil
			mi.Match = true
			mi.Route = r
			if m != "OPTIONS" {
				mi.Func = meth.Func
			}
			return true
		}
		mi.MethodNotAllowed = ErrorMethodMismatch
	}
	mi.Route = nil
	return false
}
