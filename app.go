package slow

import (
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

var (
	allowEnv = map[string]string{
		"test":        "test",
		"dev":         "development",
		"development": "development",
		"prod":        "production",
		"production":  "production",
	}
	l             = newLogger("")
	listenInAll   bool
	localAddress  = getOutboundIP()
	htmlTemplates map[string]*template.Template
)

// Returns a new app with a default settings
func NewApp(c *Config) *App {
	router := &Router{
		Name:         "",
		Routes:       []*Route{},
		routesByName: map[string]*Route{},
		is_main:      true,
	}

	cfg := NewConfig()
	if c != nil {
		cfg.Update(c)
	}

	return &App{
		Config:       cfg,
		Router:       router,
		routers:      []*Router{router},
		routerByName: map[string]*Router{"": router},
	}
}

type App struct {
	*Router
	*Config
	// TODO -> add testConfig, ProdConfig
	AfterRequest, // exec after each request (if the application dont crash)
	BeforeRequest, // exec before each request
	TearDownRequest Func // exec after each request, after send to cleint ( this dont has effect in response)

	routers      []*Router
	routerByName map[string]*Router

	srv   *http.Server
	built bool
}

// Parse the router and your routes
func (app *App) build() {
	if app.Servername != "" {
		srv := app.Servername
		srv = strings.TrimPrefix(srv, ".")
		srv = strings.TrimSuffix(srv, "/")
		app.Servername = srv
	}
	if app.built {
		return
	}

	if app.EnableStatic {
		staticUrl := "/assets"
		fp := "/{filepath:path}"
		if app.StaticUrlPath != "" {
			staticUrl = app.StaticUrlPath
		}
		path := filepath.Join(staticUrl, fp)
		app.AddRoute(&Route{
			Url:       path,
			Func:      serveFileHandler,
			Name:      "assets",
			_isStatic: true,
		})
	}

	// se o usuario mudar o router principal,
	// isso evita alguns erro
	if !app.is_main {
		app.is_main = true

		if app.Router.Routes == nil {
			app.Router.Routes = []*Route{}
		}
		if app.Router.routesByName == nil {
			app.Router.routesByName = map[string]*Route{}
		}
		if app.Router.Cors == nil {
			app.Router.Cors = &Cors{}
		}
		if app.Router.Middlewares == nil {
			app.Router.Middlewares = []Func{}
		}

	}
	for _, router := range app.routers {
		router.parse(app.Servername)
		if router != app.Router {
			maps.Copy(app.routesByName, router.routesByName)
		}
	}
	app.routerByName[app.Router.Name] = app.Router
	app.built = true
}

func (app *App) closeConn(ctx *Ctx) {
	rsp := ctx.Response
	err := recover()
	defer l.LogRequest(ctx)
	defer execTeardown(ctx)

	if err == ErrHttpAbort || err == nil {
		mi := ctx.MatchInfo
		if mi.Match {
			if ctx.Request.isWebsocket {
				return
			}
			if ctx.Session.changed {
				rsp.SetCookie(ctx.Session.save())
			}
			rsp.parseHeaders()
			rsp.Header.Save(rsp.raw)
		} else {
			if mi.MethodNotAllowed != nil {
				rsp.StatusCode = 405
				rsp.WriteString("405 Method Not Allowed")
			} else {
				rsp.StatusCode = 404
				rsp.WriteString("404 Not Found")
			}
		}
		rsp.raw.WriteHeader(rsp.StatusCode)
		fmt.Fprint(rsp.raw, rsp.String())
	} else {
		statusText := ""
		errStr, ok := err.(string)
		if ok && strings.HasPrefix(errStr, "abort:") {
			if ctx.App.AfterRequest != nil {
				ctx.App.AfterRequest(ctx)
			}
			strCode := strings.TrimPrefix(errStr, "abort:")
			code, err := strconv.Atoi(strCode)
			if err != nil {
				panic(err)
			}
			rsp.StatusCode = code
			statusText = rsp.String()

		} else {
			rsp.StatusCode = 500
			statusText = "500 Internal Server Error"
			l.Error(err)
		}
		rsp.raw.WriteHeader(rsp.StatusCode)
		fmt.Fprint(rsp.raw, statusText)
	}
}

func execTeardown(ctx *Ctx) {
	if ctx.App.TearDownRequest != nil {
		go ctx.App.TearDownRequest(ctx)
	}
}

// exec route and handle errors of application
func (app *App) execRoute(ctx *Ctx) {
	rq := ctx.Request
	mi := ctx.MatchInfo
	if mi.Func == nil && rq.Method == "OPTIONS" {
		optionsHandler(ctx)
	} else {
		rq.parse()
		ctx.parseMids()
		if app.BeforeRequest != nil {
			app.BeforeRequest(ctx)
		}
		ctx.Next()
	}
}

// Show All Routes ( internal )
func (app *App) listRoutes() {
	app.Build()
	nameLen := 0
	methLen := 0
	pathLen := 0
	subDoLen := 0

	listRouteName := []string{}
	for _, r := range app.routesByName {
		listRouteName = append(listRouteName, r.fullName)
		if nl := len(r.fullName); nl > nameLen {
			nameLen = nl
		}
		if ml := len(strings.Join(r.Methods, ",")); ml > methLen {
			methLen = ml
		}
		if pl := len(r.fullUrl); pl > pathLen {
			pathLen = pl
		}
		if sName := strings.Split(r.fullName, "."); len(sName) == 2 {
			router := app.routerByName[sName[0]]
			if router != nil && router.Subdomain != "" {
				if l := len(router.Subdomain); l > subDoLen {
					subDoLen = l
				}
			}
		}
	}
	sort.Strings(listRouteName)

	if nameLen < 6 {
		nameLen = 6
	}
	if methLen < 7 {
		methLen = 7
	}
	if pathLen < 9 {
		pathLen = 9
	}
	if subDoLen < 10 && subDoLen != 0 {
		subDoLen = 10
	}

	line1 := strings.Repeat("-", nameLen+1)
	line2 := strings.Repeat("-", methLen+1)
	line3 := strings.Repeat("-", pathLen+1)
	line4 := strings.Repeat("-", subDoLen+1)

	routeN := "ROUTES " + strings.Repeat(" ", nameLen-6)
	methodsN := "METHODS " + strings.Repeat(" ", methLen-7)
	endpointN := "ENDPOINTS " + strings.Repeat(" ", pathLen-9)

	if subDoLen > 0 {
		subdomainN := "SUBDOMAINS " + strings.Repeat(" ", subDoLen-10)

		fmt.Printf("+-%s+-%s+-%s+-%s+\n", line1, line2, line3, line4)
		fmt.Printf("| %s| %s| %s| %s|\n", routeN, methodsN, endpointN, subdomainN)
		fmt.Printf("+-%s+-%s+-%s+-%s+\n", line1, line2, line3, line4)
		for _, rName := range listRouteName {
			r := app.routesByName[rName]
			mths_ := strings.Join(r.Methods, ",")
			space1 := nameLen - len(rName)
			space2 := methLen - len(mths_)
			space3 := pathLen - len(r.fullUrl)
			space4 := subDoLen - len(r.GetRouter().Subdomain)

			endpoint := r.fullName + strings.Repeat(" ", space1)
			mths := mths_ + strings.Repeat(" ", space2)
			path := r.fullUrl + strings.Repeat(" ", space3)
			sub := r.GetRouter().Subdomain + strings.Repeat(" ", space4)
			fmt.Printf("| %s | %s | %s | %s |\n", endpoint, mths, path, sub)
		}
		fmt.Printf("+-%s+-%s+-%s+-%s+\n", line1, line2, line3, line4)
	} else {
		fmt.Printf("+-%s+-%s+-%s+\n", line1, line2, line3)
		fmt.Printf("| %s| %s| %s|\n", routeN, methodsN, endpointN)
		fmt.Printf("+-%s+-%s+-%s+\n", line1, line2, line3)
		for _, rName := range listRouteName {
			r := app.routesByName[rName]
			mths_ := strings.Join(r.Methods, ",")
			space1 := nameLen - len(rName)
			space2 := methLen - len(mths_)
			space3 := pathLen - len(r.fullUrl)

			endpoint := r.fullName + strings.Repeat(" ", space1)
			mths := mths_ + strings.Repeat(" ", space2)
			path := r.fullUrl + strings.Repeat(" ", space3)
			fmt.Printf("| %s | %s | %s |\n", endpoint, mths, path)
		}
		fmt.Printf("+-%s+-%s+-%s+\n", line1, line2, line3)
	}
}

func (app *App) match(ctx *Ctx) bool {
	rq := ctx.Request
	if app.Servername != "" {
		rqUrl := rq.URL.Host
		if net.ParseIP(rqUrl) != nil {
			return false
		}
		if !strings.Contains(rqUrl, app.Servername) {
			return false
		}
	}

	for _, router := range app.routers {
		if router.match(ctx) {
			return true
		}
	}
	return false
}

// http.Handler func
func (app *App) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	ctx := newCtx(app)
	ctx.Request = NewRequest(req, ctx)
	ctx.Response = NewResponse(wr, ctx)
	defer app.closeConn(ctx)
	if app.match(ctx) {
		app.execRoute(ctx)
	}
}

/*
Register the router in app

	func main() {
		api := slow.NewRouter("api")
		api.Subdomain = "api"
		api.Prefix = "/v1"
		api.post("/products")
		api.get("/products/{productID:int}")


		app := slow.NewApp()

		// This Function
		app.Mount(getApiRouter)

		app.Listen()
	}
*/
func (app *App) Mount(routers ...*Router) {
	for _, router := range routers {
		if router.Name == "" {
			panic(fmt.Errorf("the routers must be named"))
		} else if _, ok := app.routerByName[router.Name]; ok {
			panic(fmt.Errorf("router '%s' already regitered", router.Name))
		}
		router.parse(app.Servername)
		app.routerByName[router.Name] = router
		app.routers = append(app.routers, router)
	}
}

/*
Url Builder

	app.GET("/users/{userID:int}", index)

	app.UrlFor("index", false, "userID", "1"})
	// results: /users/1

	app.UrlFor("index", true, "userID", "1"})
	// results: http://yourAddress/users/1
*/
func (app *App) UrlFor(name string, external bool, args ...string) string {
	var (
		host   = ""
		route  *Route
		router *Router
	)
	if app.srv == nil {
		l.err.Fatalf("you are trying to use this function outside of a context")
	}
	if len(args)%2 != 0 {
		l.err.Fatalf("numer of args of build url, is invalid: UrlFor only accept pairs of args ")
	}
	params := map[string]string{}
	c := len(args)
	for i := 0; i < c; i++ {
		if i%2 != 0 {
			continue
		}
		params[args[i]] = args[i+1]
	}
	if r, ok := app.routesByName[name]; ok {
		route = r
	}
	if strings.Contains(name, ".") {
		routerName := strings.Split(name, ".")[0]
		router = app.routerByName[routerName]
		if router == nil {
			panic(fmt.Sprintf("Router '%s' is undefined \n", routerName))
		}
	} else {
		router = app.Router
	}
	if route == nil {
		panic(fmt.Sprintf("Route '%s' is undefined \n", name))
	}
	// Pre Build
	var sUrl = strings.Split(route.fullUrl, "/")
	var urlBuf strings.Builder
	// Build Host
	if external {
		schema := "http://"
		if app.ListeningInTLS || len(app.srv.TLSConfig.Certificates) > 0 {
			schema = "https://"
		}
		if router.Subdomain != "" {
			host = schema + router.Subdomain + "." + app.Servername
		} else {
			if app.Servername == "" {
				_, p, _ := net.SplitHostPort(app.srv.Addr)
				h := net.JoinHostPort(localAddress, p)
				host = schema + h
			} else {
				host = schema + app.Servername
			}
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
	if len(params) > 0 {
		return host + url + strings.TrimSuffix(query.String(), "&")
	}
	return host + url
}

// Show All Routes
func (app *App) ShowRoutes() { app.listRoutes() }

/*
Build the App, but not start serve

example:

	func index(ctx slow.Ctx){}

	// it's work
	func main() {
		app := slow.NewApp()
		app.GET("/",index)
		app.Build(":5000")
		app.UrlFor("index",true)
	}
	// it's don't work
	func main() {
		app := slow.NewApp()
		app.GET("/",index)
		app.UrlFor("index",true)
	}
*/
func (app *App) Build(addr ...string) {

	var address string
	if app.Env == "" {
		app.Env = "development"
	} else if _, ok := allowEnv[app.Env]; !ok {
		l.err.Fatalf("environment '%s' is not valid", app.Env)
	}
	l = newLogger(app.LogFile)

	if len(addr) > 0 {
		a_ := addr[0]
		if a_ != "" {
			_, _, err := net.SplitHostPort(a_)
			if err == nil {
				address = a_
			}
		}
	}
	if address == "" {
		address = "127.0.0.1:5000"
	}

	app.build()

	if strings.Contains(address, "0.0.0.0") {
		listenInAll = true
	}

	app.srv = &http.Server{
		Addr:           address,
		Handler:        app,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	app.built = true
}

func (app *App) logStarterListener() {
	addr, port, err := net.SplitHostPort(app.srv.Addr)
	if err != nil {
		l.err.Panic(err)
	}
	envDev := app.Env == "development"
	if listenInAll {
		app.srv.Addr = localAddress
		if envDev {
			l.Defaultf("Server is listening on all address in %sdevelopment mode%s", _RED, _RESET)
		} else {
			l.Default("Server is listening on all address")
		}
		l.info.Printf("          listening on: http://%s:%s", getOutboundIP(), port)
		l.info.Printf("          listening on: http://0.0.0.0:%s", port)
	} else {
		if envDev {
			l.Defaultf("Server is listening in %sdevelopment mode%s", _RED, _RESET)
		} else {
			l.Default("Server is linsten")
		}
		if addr == "" {
			addr = "localhost"
		}
		l.info.Printf("          listening on: http://%s:%s", addr, port)
	}
}

// Build a app and starter Server
func (app *App) Listen(host ...string) {
	app.Build(host...)
	app.logStarterListener()
	app.srv.ListenAndServe()
}
