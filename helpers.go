package slow

import (
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime"
	"strings"
)

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

func getFunctionName(i interface{}) string {
	splitName := strings.Split(
		runtime.FuncForPC(
			reflect.ValueOf(i).Pointer(),
		).Name(), ".",
	)
	return splitName[len(splitName)-1]
}

func HtmlEscape(s string) string { return htmlReplacer.Replace(s) }

// Alias of 'fmt.Sprintf("%T", obj)'
func TypeOf(obj any) string { return fmt.Sprintf("%T", obj) }

/*
build a URL of a route

	app.Get("/users/{userID:int}", index)

	slow.UrlFor("index", false, map[string]string{"userID": 1})
	// result in: /users/1

	slow.UrlFor("index", true, map[string]string{"userID": 1})
	// result in: http://yourAddress/users/1
*/
func UrlFor(name string, external bool, params map[string]string) string {
	var (
		host   = ""
		route  *Route
		router *Router
		app    *App
	)

	// Get Route & Router
	lenStack := len(appStack)
	if lenStack > 0 {
		app = appStack[lenStack-1]
	} else {
		l.err.Fatalf("vc esta tentando usar essa função fora de um contexto")
	}

	if r, ok := app.routesByName[name]; ok {
		route = r
	}
	if strings.Contains(name, ".") {
		routerName := strings.Split(name, ".")[0]
		router = app.routerByName[routerName]
	} else {
		router = app.Router
	}

	if route == nil {
		panic(errors.New("route '" + name + "' is not found"))
	}

	// Pre Build
	var sUrl = strings.Split(route.fullUrl, "/")
	var urlBuf strings.Builder

	// Build Host
	if external {
		if router.Subdomain != "" {
			host = "http://" + router.Subdomain + "." + servername
		} else {
			host = "http://" + localAddress.String() + ":" + strings.Split(app.srv.Addr, ":")[1]
		}
	}

	// Build path
	for _, str := range sUrl {
		if re.isVar.MatchString(str) {
			fname := re.getVarName(str)
			value, ok := params[fname]
			if !ok {
				if !re.isVarOpt.MatchString(str) {
					panic(errors.New("Route '" + name + "' needs parameter '" + str + "' but not passed"))
				}
			} else {
				urlBuf.WriteString("/" + value)
				delete(params, fname)
			}
		} else {
			urlBuf.WriteString("/" + str)
		}
	}

	// Build Query
	var query strings.Builder
	if len(params) > 0 {
		urlBuf.WriteString("?")
		for k, v := range params {
			query.WriteString(k + "=" + v + "&")
		}
	}
	url := urlBuf.String()
	url = re.slash2.ReplaceAllString(url, "/")
	url = re.dot2.ReplaceAllString(url, ".")
	fmt.Println(host)
	fmt.Println(url)
	fmt.Println(query.String())
	if len(params) > 0 {
		return host + url + strings.TrimSuffix(query.String(), "&")
	}
	return host + url
}

// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
