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

type MapCtrl map[string]*Meth

type Route struct {
	Url     string
	Name    string
	Func    Func
	MapCtrl MapCtrl
	// Ctrl_ any  // struct{}.GET(ctx *Ctx)

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

	for i, str := range strs {
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
			if re.isVarOpt.MatchString(str) {
				if i != len(strs)-1 {
					l.err.Fatal("optional url var must be last")
				}
				str = re.str.ReplaceAllString(str, `(([\x00-\x7F]+)([^\\\/\s]+)|\d+)?`)
				str = re.digit.ReplaceAllString(str, `(\d+)?`)
				str = re.filepath.ReplaceAllString(str, `([\/\w+.-]+)?`)
			} else {
				str = re.str.ReplaceAllString(str, `(([\x00-\x7F]+)([^\\\/\s]+)|\d+)`)
				str = re.digit.ReplaceAllString(str, `(\d+)`)
				str = re.filepath.ReplaceAllString(str, `([\/\w+.-]+)`)
			}
		}
		r.urlRegex = append(r.urlRegex, regexp.MustCompile(fmt.Sprintf("^%s$", str)))
	}
	if strings.HasSuffix(r.fullUrl, "/") {
		r.urlRegex = append(r.urlRegex, regexp.MustCompile(`^(\/)?$`))
	}
}

func (r *Route) parse() {
	if r.Func == nil && r.MapCtrl == nil {
		l.err.Fatalf("Route '%s' need a func or Ctrl\n", r.fullName)
	}

	r.compileUrl()
	ctrl := MapCtrl{
		"OPTIONS": &Meth{
			Method: "OPTIONS",
		},
	}

	strMth := map[string]any{}
	for verb, meth := range r.MapCtrl {
		v := strings.ToUpper(verb)
		if !reMethods.MatchString(v) {
			panic(fmt.Errorf("route '%s' has invalid Request Method: '%s'", r.fullName, verb))
		}
		ctrl[v] = meth
		strMth[v] = nil
	}
	r.MapCtrl = ctrl

	for _, verb := range r.Methods {
		v := strings.ToUpper(verb)
		if !reMethods.MatchString(v) {
			panic(fmt.Errorf("route '%s' has invalid Request Method: '%s'", r.fullName, verb))
		}
		if _, ok := r.MapCtrl[v]; !ok {
			r.MapCtrl[v] = &Meth{
				Func:   r.Func,
				Schema: r.Schema,
				Method: v,
			}
		}
		strMth[v] = nil
	}
	if _, ok := r.MapCtrl["HEAD"]; !ok {
		if _, ok := r.MapCtrl["GET"]; ok {
			r.MapCtrl["HEAD"] = r.MapCtrl["GET"]
			strMth["HEAD"] = nil
		}
	}

	lmths := []string{"OPTIONS"}
	for v := range strMth {
		lmths = append(lmths, v)
	}
	r.Methods = lmths

	if r.Cors != nil {
		r.Cors.AllowMethods = lmths
	} else {
		r.Cors = &Cors{AllowMethods: lmths}
	}

}

func (r *Route) matchURL(ctx *Ctx, url string) bool {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	urlSplit := strings.Split(url, "/")

	lSplit := len(urlSplit)
	lRegex := len(r.urlRegex)

	if lSplit != lRegex {
		if !re.isVarOpt.MatchString(r.Url) && lSplit != lRegex-1 {
			return false
		}
	}
	for i, uRe := range r.urlRegex {
		str := ""
		if i < lSplit {
			str = urlSplit[i]
		}
		if !uRe.MatchString(str) {
			return false
		}
	}
	return true
}

func (r *Route) Match(ctx *Ctx) bool {
	mi := ctx.MatchInfo

	rq := ctx.Request
	m := rq.Method

	if !r.matchURL(ctx, rq.Raw.URL.Path) {
		if !re.filepath.MatchString(r.fullUrl) {
			return false
		}
		if !strings.HasPrefix(rq.Raw.URL.Path, ctx.App.StaticUrlPath) {
			return false
		}
	}

	if meth, ok := r.MapCtrl[m]; ok {
		mi.MethodNotAllowed = nil

		if meth.Func != nil {
			mi.Func = meth.Func
		}
		mi.Match = true
		mi.route = r.fullName

		return true
	}
	mi.MethodNotAllowed = ErrorMethodMismatch
	mi.route = ""
	return false
}

func ALL(url string, f Func) *Route {
	return &Route{
		Url:  url,
		Func: f,
		Name: getFunctionName(f),
		Methods: []string{
			"GET", "HEAD", "POST",
			"PUT", "DELETE", "CONNECT",
			"OPTIONS", "TRACE", "PATCH"},
	}
}

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
