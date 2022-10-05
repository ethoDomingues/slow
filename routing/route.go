package routing

import (
	"net/http"
	"regexp"
)

type Func func(*http.Request, http.ResponseWriter)

type Route struct {
	Url      string
	reUrl    *regexp.Regexp
	mapMeths map[string]any
	Name     string
	Methods  []string
	Func
}

func (r *Route) match(req *http.Request, mi *MatchInfo) bool {
	if r.reUrl.MatchString(req.URL.Path) {
		if _, ok := r.mapMeths[req.Method]; ok {
			mi.MethodNotAllowed = nil
			mi.Route = r
			return true
		}
		mi.MethodNotAllowed = ErrorMethodMismatch
	}
	mi.Router = nil
	mi.Route = nil

	return false
}

func (r *Route) parse() {
	var err error
	r.reUrl, err = regexp.Compile(`^` + r.Url + `$`)
	if err != nil {
		panic(err)
	}
}
