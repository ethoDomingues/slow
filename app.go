package slow

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

var (
	l = newLogger()

	servername    string
	_secretKey    string
	contextsNamed map[string]*Ctx
)

func NewApp() *App {

	router := &Router{
		_main:        true,
		Name:         "",
		Routes:       []*Route{},
		routesByName: map[string]*Route{},
	}
	app := &App{
		Router:         router,
		routers:        []*Router{router},
		routerByName:   map[string]*Router{"": router},
		StaticFolder:   "./assets",
		TemplateFolder: "./templates",
		StaticUrlPath:  "/assets",
	}
	router.Get("/assets/{filepath:filepath}", ServeFile)
	return app
}

type App struct {
	*Router

	SecretKey,
	Servername,
	StaticFolder,
	TemplateFolder,
	StaticUrlPath string

	routers      []*Router
	routerByName map[string]*Router

	BeforeRequest   Func
	AfterRequest    Func
	TearDownRequest Func

	building bool
}

func (app *App) build() {
	if app.building {
		return
	}
	if app.Servername != "" {
		servername = app.Servername
	} else {
		servername = "localhost:5000"
	}
	for _, router := range app.routers {
		router.parse()
		if router.Name != "" {
			maps.Copy(app.routesByName, router.routesByName)
		}
	}
	app.building = true
}

func (app *App) Mount(routes ...*Router) {
	for _, route := range routes {
		if route.Name == "" {
			panic(fmt.Errorf("the routers must be named"))
		} else if _, ok := app.routerByName[route.Name]; ok {
			panic(fmt.Errorf("router '%s' already regitered", route.Name))
		}
		app.routers = append(app.routers, route)
	}
}

func (app *App) UrlFor(name string, external bool, params map[string]string) string {
	var (
		host   = ""
		route  *Route
		router *Router
	)
	if r, ok := app.routesByName[name]; ok {
		route = r
	}
	if strings.Contains(name, ".") {
		routerName := strings.Split(name, ",")[0]
		router = app.routerByName[routerName]
	} else {
		router = app.Router
	}

	if route == nil {
		panic(errors.New("route '" + name + "' is not found"))
	}
	var sUrl = strings.Split(route.fullUrl, "/")
	var urlBuf strings.Builder

	if external {
		if router.Subdomain != "" {
			host = "http://" + router.Subdomain + "." + servername
		} else {
			host = "http://" + servername
		}
	}
	for _, str := range sUrl {
		if re.isVar.MatchString(str) {
			fname := re.getVarName(str)
			value, ok := params[fname]
			if !ok {
				panic(errors.New("Route '" + name + "' needs parameter '" + str + "' but not passed"))
			}
			urlBuf.WriteString("/" + value)
			delete(params, fname)
		} else {
			urlBuf.WriteString("/" + str)
		}
	}
	if len(params) > 0 {
		urlBuf.WriteString("?")
		for k, v := range params {
			urlBuf.WriteString(k + "=" + v + "&")
		}
	}
	url := strings.TrimSuffix(urlBuf.String(), "&")
	url = re.slash2.ReplaceAllString(url, "/")
	url = re.dot2.ReplaceAllString(url, ".")
	return strings.TrimSuffix(host, "/") + url
}

func (app *App) execRoute(ctx *Ctx) {
	rsp := ctx.Response
	rq := ctx.Request
	if app.BeforeRequest != nil {
		app.BeforeRequest(ctx)
	}
	defer func() {
		err := recover()
		if err == HttpAbort || err == nil {
			if len(ctx.Session.jwt.Payload) > 0 {
				s := ctx.Session.Save()
				rsp.SetCookie(s)
			}
			rsp._afterRequest()
			rsp.Headers.Save(rsp.raw)
			rsp.raw.WriteHeader(rsp.StatusCode)
			fmt.Fprint(rsp.raw, rsp.Body.String())
		} else {
			rsp.StatusCode = 200
			statusText := ""
			if TypeOf(err) == "string" {
				str := err.(string)
				if strings.HasPrefix(str, "abort:") {
					strCode := strings.TrimPrefix(str, "abort:")
					code, err := strconv.Atoi(strCode)
					if err != nil {
						panic(err)
					}
					rsp.StatusCode = code
					statusText = strCode + " " + http.StatusText(code)
				}
			} else {
				rsp.StatusCode = 500
				statusText = "500 Internal Server Error"
				newLogger().Error(err)
			}
			rsp.raw.WriteHeader(rsp.StatusCode)
			fmt.Fprint(rsp.raw, statusText)
			if app.TearDownRequest != nil {
				app.TearDownRequest(ctx)
			}
		}
	}()
	for _, mid := range rq.MatchInfo.Router.Middlewares {
		mid(ctx)
	}
	for _, mid := range rq.MatchInfo.Route.Middlewares {
		mid(ctx)
	}
	// if raise a error in any mid, Route.Func not is executed.
	rq.MatchInfo.Func(ctx)
	if app.AfterRequest != nil {
		app.AfterRequest(ctx)
	}
}

func (app *App) ServeHTTP(wr http.ResponseWriter, req *http.Request) {

	ctx := NewCtx(app, req.Context())

	rsp := NewResponse(wr, ctx.id)
	rq := NewRequest(req, ctx.id)

	ctx.Request = rq
	ctx.Response = rsp

	contextsNamed[ctx.id] = ctx
	defer delete(contextsNamed, ctx.id)

	rq.Parse()

	if app.SecretKey != "" {
		if c, ok := rq.Cookies["session"]; ok {
			ctx.Session.validate(c, app.SecretKey)
		}
	}

	mi := rq.MatchInfo
	for _, router := range app.routers {
		if router.Match(ctx) {
			break
		}
	}
	if mi.Match {
		if rq.Raw.Method == "OPTIONS" {
			rsp.StatusCode = 200
			strMeths := strings.Join(mi.Route.Methods, ", ")
			rsp.Headers.Set("Access-Control-Allow-Methods", strMeths)

			rsp._afterRequest()

			rsp.Headers.Save(rsp.raw)
			fmt.Fprint(rsp.raw, "")
		} else {
			rq.Query = req.URL.Query()
			rq.Args = re.getUrlValues(mi.Route.fullUrl, req.URL.Path)
			app.execRoute(ctx)
		}
	} else if mi.MethodNotAllowed != nil {
		rsp.StatusCode = 405
		rsp.raw.WriteHeader(405)
		fmt.Fprint(rsp.raw, "405 Method Not Allowed")
	} else {
		rsp.StatusCode = 404
		rsp.raw.WriteHeader(404)
		fmt.Fprint(rsp.raw, "404 Not Found")
	}
	l.LogRequest(ctx.id)
}

func (app *App) Listen() {
	app.build()
	srv := &http.Server{
		Addr:           ":5000",
		Handler:        app,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if contextsNamed == nil {
		contextsNamed = map[string]*Ctx{}
	}
	l.Deafault("Server is linsten in", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func (app *App) listRoutes() {
	app.build()
	nameLen := 0
	methLen := 0
	pathLen := 0
	for _, r := range app.routesByName {
		if nl := len(r.fullName); nl > nameLen {
			nameLen = nl + 1
		}
		if ml := len(strings.Join(r.Methods, " ")); ml > methLen {
			methLen = ml + 1
		}
		if pl := len(r.fullUrl); pl > pathLen {
			pathLen = pl + 1
		}
	}

	line1 := strings.Repeat("-", nameLen)
	line2 := strings.Repeat("-", methLen)
	line3 := strings.Repeat("-", pathLen)

	routeN := "ROUTES" + strings.Repeat(" ", nameLen-6)
	methodsN := "METHODS" + strings.Repeat(" ", methLen-7)
	endpointN := "ENDPOINTS" + strings.Repeat(" ", pathLen-9)

	fmt.Printf("+-%s-+-%s-+-%s-+\n", line1, line2, line3)
	fmt.Printf("| %s | %s | %s |\n", routeN, methodsN, endpointN)
	fmt.Printf("+-%s-+-%s-+-%s-+\n", line1, line2, line3)
	for _, r := range app.routesByName {
		mths_ := strings.Join(r.Methods, " ")
		space1 := nameLen - len(r.fullName)
		space2 := methLen - len(mths_)
		space3 := pathLen - len(r.fullUrl)

		endpoint := r.fullName + strings.Repeat(" ", space1)
		mths := mths_ + strings.Repeat(" ", space2)
		path := r.fullUrl + strings.Repeat(" ", space3)
		fmt.Printf("| %s | %s | %s |\n", endpoint, mths, path)
	}
	fmt.Printf("+-%s-+-%s-+-%s-+\n", line1, line2, line3)
}

func (app *App) ShowRoutes() { app.listRoutes() }
