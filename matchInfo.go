package slow

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrHttpAbort        = errors.New("aborted")
	ErrorNotFound       = errors.New("404 Not Found")
	ErrorMethodMismatch = errors.New("405 Method Not Allowed")
)

type MatchInfo struct {
	Func

	Match            bool
	MethodNotAllowed error

	ctx    string
	route  string
	router string
}

func (m *MatchInfo) Ctx() *Ctx { return contextsNamed[m.ctx] }

func (m *MatchInfo) Router() *Router {
	return m.Ctx().App.routerByName[m.router]
}

func (m *MatchInfo) Route() *Route {
	return m.Ctx().App.routesByName[m.route]
}

type _re struct {
	str      *regexp.Regexp
	digit    *regexp.Regexp
	filepath *regexp.Regexp

	isVar    *regexp.Regexp
	isVarOpt *regexp.Regexp
	dot2     *regexp.Regexp
	slash2   *regexp.Regexp
}

var (
	// hosts      *regexp.Regexp
	isStr      = regexp.MustCompile(`{\w+(:str)?[?]?}`)
	isVar      = regexp.MustCompile(`{\w+(\:(int|str|filepath))?[?]?}`)
	isVarOpt   = regexp.MustCompile(`{\w+(\:(int|str))?[\?]}`)
	isDigit    = regexp.MustCompile(`{\w+:int[?]?}`)
	isFilepath = regexp.MustCompile(`{\w+:filepath}`)

	dot2      = regexp.MustCompile(`[.]{2,}`)
	slash2    = regexp.MustCompile(`[\/]{2,}`)
	reMethods = regexp.MustCompile("^(?i)(GET|PUT|HEAD|POST|TRACE|PATCH|DELETE|CONNECT|OPTIONS)$")

	re = _re{
		str:      isStr,
		isVar:    isVar,
		isVarOpt: isVarOpt,
		digit:    isDigit,
		filepath: isFilepath,
		dot2:     dot2,
		slash2:   slash2,
	}
)

// if 'str' is a var (example: {id:int} ), return 'id', else return str
func (r *_re) getVarName(str string) string {
	if r.isVar.MatchString(str) {
		str = strings.Replace(str, "{", "", -1)
		str = strings.Replace(str, "}", "", -1)
		str = strings.Split(str, ":")[0]
	}
	if strings.HasSuffix(str, "?") {
		return strings.TrimSuffix(str, "?")
	}
	return str
}

func (r _re) getUrlValues(url, requestUrl string) map[string]string {
	req := strings.Split(requestUrl, "/")
	kv := map[string]string{}

	for i, str := range strings.Split(url, "/") {
		if i < len(req) {
			if re.isVar.MatchString(str) {
				kv[re.getVarName(str)] = req[i]
			}
		}
	}
	return kv
}
