package slow

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrorMethodMismatch = errors.New("405 Method Not Allowed")
	ErrorNotFound       = errors.New("404 Not Found")
)

type MatchInfo struct {
	MethodNotAllowed error
	Match            bool
	Func
	Route  *Route
	Router *Router
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
	isStr      = regexp.MustCompile(`{\w+(:str)?[?]?}`)
	isVar      = regexp.MustCompile(`{\w+(\:(int|str|filepath))?[?]?}`)
	isVarOpt   = regexp.MustCompile(`{\w+(\:(int|str))?[\?]}`)
	isDigit    = regexp.MustCompile(`{\w+:int[?]?}`)
	isFilepath = regexp.MustCompile(`{\w+:filepath}`)

	dot2      = regexp.MustCompile(`[.]{2,}`)
	isIP      = regexp.MustCompile(`^([0,255][.][0,255][.][0,255][.][0,255])((:[0,65535])?)$`)
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
	return str
}

func (r _re) getUrlValues(url, requestUrl string) map[string]string {
	req := strings.Split(requestUrl, "/")
	kv := map[string]string{}
	for i, str := range strings.Split(url, "/") {
		if re.isVar.MatchString(str) {
			kv[re.getVarName(str)] = req[i]
		}
	}
	return kv
}
