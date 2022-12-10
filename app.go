package slow

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

var (
	l = newLogger("")

	servername   string
	listenInAll  bool
	localAddress = getOutboundIP()
	allowEnv     = map[string]string{
		"dev":         "development",
		"development": "development",
		"prod":        "production",
		"production":  "production",
	}
)

// Returns a new app with a default settings
func NewApp() *App {
	router := &Router{
		Name:         "",
		Routes:       []*Route{},
		is_main:      true,
		routesByName: map[string]*Route{},
	}
	app := &App{
		Router:         router,
		routers:        []*Router{router},
		routerByName:   map[string]*Router{"": router},
		StaticFolder:   "./assets",
		TemplateFolder: "./templates",
		StaticUrlPath:  "/assets",
		EnableStatic:   true,
	}
	return app
}

type App struct {
	*Router

	SecretKey, // for sign session
	Servername, // for build url routes and route match
	StaticFolder, // for serve static files
	TemplateFolder, // for render template (html) files
	StaticUrlPath, // url uf request static file
	LogFile, // save log info in file
	Env string // environmnt

	Silent       bool //don't print logs
	EnableStatic bool // enable static endpoint for serving static files

	routers      []*Router
	routerByName map[string]*Router

	BeforeRequest, // exec before each request
	AfterRequest, // exec after each request (if the application dont crash)
	TearDownRequest Func // exec after each request, after send to cleint ( this dont has effect in response)

	built bool
	srv   *http.Server
}

// Parse the router and your routes
func (app *App) build() {
	servername = app.Servername
	if app.built {
		return
	}
	if app.EnableStatic {
		app.GET("/assets/{filepath:filepath}", serveFile)
	}
	for _, router := range app.routers {
		router.parse()
		if router != app.Router {
			maps.Copy(app.routesByName, router.routesByName)
		}
	}
	if app.Router.Name != "" {
		app.routerByName[app.Router.Name] = app.Router
	}
}

// exec route and handle errors of application
func (app *App) execRoute(ctx *Ctx) {
	rsp := ctx.Response
	rq := ctx.Request

	defer func() {
		err := recover()
		if err == ErrHttpAbort || err == nil {
			// if raise a error in any mid or in route func, app.AfterRequest not is executed.
			if app.AfterRequest != nil {
				app.AfterRequest(ctx)
			}

			if ctx.Session.changed {
				rsp.SetCookie(
					ctx.Session.save(),
				)
			}
			if rq.Method != "OPTIONS" {
				rsp.parseHeaders()
				rsp.Headers.Save(rsp.raw)
			}
			rsp.raw.WriteHeader(rsp.StatusCode)
			fmt.Fprint(rsp.raw, rsp.Body.String())
		} else {
			rsp.StatusCode = 200
			statusText := ""
			errStr, ok := err.(string)
			if ok && strings.HasPrefix(errStr, "abort:") {
				strCode := strings.TrimPrefix(errStr, "abort:")
				code, err := strconv.Atoi(strCode)
				if err != nil {
					panic(err)
				}
				rsp.StatusCode = code
				statusText = strCode + " " + http.StatusText(code)

			} else {
				rsp.StatusCode = 500
				statusText = "500 Internal Server Error"
				l.Error(err)
			}
			rsp.raw.WriteHeader(rsp.StatusCode)
			fmt.Fprint(rsp.raw, statusText)
		}
	}()
	if ctx.MatchInfo.Func == nil && rq.Method == "OPTIONS" {
		optionsHandler(ctx)
	} else {
		if app.BeforeRequest != nil {
			app.BeforeRequest(ctx)
		}

		for _, mid := range ctx.MatchInfo.Router.Middlewares {
			mid(ctx)
		}

		for _, mid := range ctx.MatchInfo.Route.Middlewares {
			mid(ctx)
		}

		// if raise a error in any mid, Route.Func not is executed.
		ctx.MatchInfo.Func(ctx)
	}
}

// Show All Routes ( internal )
func (app *App) listRoutes() {
	app.build()
	nameLen := 0
	methLen := 0
	pathLen := 0

	listRouteName := []string{}
	for _, r := range app.routesByName {
		listRouteName = append(listRouteName, r.fullName)
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
	sort.Strings(listRouteName)

	line1 := strings.Repeat("-", nameLen)
	line2 := strings.Repeat("-", methLen)
	line3 := strings.Repeat("-", pathLen)

	routeN := "ROUTES" + strings.Repeat(" ", nameLen-6)
	methodsN := "METHODS" + strings.Repeat(" ", methLen-7)
	endpointN := "ENDPOINTS" + strings.Repeat(" ", pathLen-9)

	fmt.Printf("+-%s-+-%s-+-%s-+\n", line1, line2, line3)
	fmt.Printf("| %s | %s | %s |\n", routeN, methodsN, endpointN)
	fmt.Printf("+-%s-+-%s-+-%s-+\n", line1, line2, line3)
	for _, rName := range listRouteName {
		r := app.routesByName[rName]
		mths_ := strings.Join(r.Methods, " ")
		space1 := nameLen - len(rName)
		space2 := methLen - len(mths_)
		space3 := pathLen - len(r.fullUrl)

		endpoint := r.fullName + strings.Repeat(" ", space1)
		mths := mths_ + strings.Repeat(" ", space2)
		path := r.fullUrl + strings.Repeat(" ", space3)
		fmt.Printf("| %s | %s | %s |\n", endpoint, mths, path)
	}
	fmt.Printf("+-%s-+-%s-+-%s-+\n", line1, line2, line3)
}

func (app *App) Match(ctx *Ctx) bool {
	rq := ctx.Request
	rqUrl := rq.Raw.Host
	hasServername := servername != ""
	validRequestHost := strings.Contains(rqUrl, servername)
	if hasServername && !validRequestHost {
		ctx.Request.Raw.Context().Done()
	}

	for _, router := range app.routers {
		if router.Match(ctx) {
			return true
		}
	}
	return false
}

// http.Handler func
func (app *App) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	ctx := newCtx(app)
	rsp := NewResponse(wr, ctx)
	rq := NewRequest(req, ctx)

	ctx.Request = rq
	ctx.Response = rsp

	rq.parseRequest()

	if app.SecretKey != "" {
		if c, ok := rq.Cookies["session"]; ok {
			ctx.Session.validate(c, app.SecretKey)
		}
	}

	mi := ctx.MatchInfo
	if app.Match(ctx) {
		rq.Query = req.URL.Query()
		rq.Args = re.getUrlValues(mi.Route.fullUrl, req.URL.Path)
		app.execRoute(ctx)
	} else if mi.MethodNotAllowed != nil {
		rsp.StatusCode = 405
		rsp.raw.WriteHeader(405)
		fmt.Fprint(rsp.raw, "405 Method Not Allowed")
	} else {
		rsp.StatusCode = 404
		rsp.raw.WriteHeader(404)
		fmt.Fprint(rsp.raw, "404 Not Found")
	}

	if app.TearDownRequest != nil {
		app.TearDownRequest(ctx)
	}
	l.LogRequest(ctx)
	// here, the request is finished
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
		l.err.Fatalf("vc esta tentando usar essa função fora de um contexto")
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
			panic(routerName + " is undefined")
		}
	} else {
		router = app.Router
	}

	if route == nil {
		l.err.Panicln("route '" + name + "' is not found")
	}

	// Pre Build
	var sUrl = strings.Split(route.fullUrl, "/")
	var urlBuf strings.Builder

	// Build Host
	if external {
		schema := "http://"
		if len(app.srv.TLSConfig.Certificates) > 0 {
			schema = "https://"
		}
		if router.Subdomain != "" {
			host = schema + router.Subdomain + "." + servername
		} else {
			if servername == "" {
				_, p, _ := net.SplitHostPort(app.srv.Addr)
				h := net.JoinHostPort(localAddress.String(), p)
				host = schema + h
			} else {
				host = schema + servername
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
Build app & server, but not start serve

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
		address = addr[0]
	} else if !flag.Parsed() {
		flag.StringVar(&address, "address", "127.0.0.1:5000", "address of server listen. default: localhost")
		flag.Parse()
	} else {
		l.warn.Println("you are using more than one application and you are trying to use the same address for both.")
		p := getFreePort()
		address = "127.0.0.1:" + p
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

// Build a app and starter Server
func (app *App) Listen(host ...string) error {
	app.Build(host...)
	_, port, err := net.SplitHostPort(app.srv.Addr)
	if err != nil {
		l.err.Fatal(err)
	}

	if !app.Silent {
		if listenInAll {
			l.Default("Server is listening in all address")
			l.info.Printf("          listening in: http://%s:%s", localAddress, port)
			l.info.Printf("          listening in: http://127.0.0.1:%s", port)
		} else {
			l.Default("Server is linsten in", app.srv.Addr)
		}
	}
	return app.srv.ListenAndServe()
}

// Build a app and starter Server
func (app *App) ListenTLS(certFile string, keyFile string, host ...string) error {
	app.Build(host...)
	port := strings.Split(app.srv.Addr, ":")[1]
	if listenInAll {
		l.Default("Server is listening in all address")
		l.info.Printf("          listening in: https://%s:%s", localAddress, port)
		l.info.Printf("          listening in: https://127.0.0.1:%s", port)
	} else {
		l.Default("Server is linsten in", app.srv.Addr)
	}
	l.Default("environment: ", app.Env)
	return app.srv.ListenAndServeTLS(certFile, keyFile)
}
