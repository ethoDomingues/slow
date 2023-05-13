package slow

import (
	"errors"
	"flag"
	"fmt"
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
	l = newLogger("")

	servername   string
	listenInAll  bool
	localAddress = getOutboundIP()
	allowEnv     = map[string]string{
		"test":        "test",
		"dev":         "development",
		"development": "development",
		"prod":        "production",
		"production":  "production",
	}
)

// Returns a new app with a default settings
func NewApp(c *Config) *App {
	router := &Router{
		Name:         "",
		Routes:       []*Route{},
		routesByName: map[string]*Route{},
		is_main:      true,
	}

	nC := NewConfig()
	nC.Update(c)

	return &App{
		Router:       router,
		routers:      []*Router{router},
		routerByName: map[string]*Router{"": router},
		Config:       nC,
	}
}

type App struct {
	*Router
	*Config

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
	if app.Servername != "" {
		srv := app.Servername
		srv = strings.TrimPrefix(srv, ".")
		srv = strings.TrimSuffix(srv, "/")

		servername = srv
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
		app.Get(path, serveFile)
	}
	// se o usuario mudar o router principal,
	// isso evita alguns erro
	if !app.is_main {
		if app.Router.routesByName == nil {
			app.Router.routesByName = map[string]*Route{}
		}
		if app.Router.Cors == nil {
			app.Router.Cors = &Cors{}
		}
		if app.Router.Cors == nil {
			app.Router.Cors = &Cors{}
		}
		if app.Router.Middlewares == nil {
			app.Router.Middlewares = NewMiddleware(nil)
		}
		if app.Router.Routes == nil {
			app.Router.Routes = []*Route{}
		}

		app.is_main = true
	} else {
		l.Error("sign of App.Router is invalid")
	}
	for _, router := range app.routers {
		router.parse()
		if router != app.Router {
			maps.Copy(app.routesByName, router.routesByName)
		}
	}
	app.routerByName[app.Router.Name] = app.Router
}

func (app *App) closeConn(ctx *Ctx) {
	rsp := ctx.Response
	err := recover()
	defer l.LogRequest(ctx)
	if err == ErrHttpAbort || err == nil {
		mi := ctx.MatchInfo
		if mi.Match {
			if ctx.Session.changed {
				rsp.SetCookie(
					ctx.Session.save(),
				)
			}
			rsp.parseHeaders()
			rsp.Header.Save(rsp.raw)
		} else {
			if mi.MethodNotAllowed != nil {
				rsp.StatusCode = 405
				rsp.Body.WriteString("405 Method Not Allowed")
			} else {
				rsp.StatusCode = 404
				rsp.Body.WriteString("404 Not Found")
			}
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

	if app.TearDownRequest != nil {
		app.TearDownRequest(ctx)
	}
}

// exec route and handle errors of application
func (app *App) execRoute(ctx *Ctx) {
	rq := ctx.Request
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
		ctx.MatchInfo.Func(ctx)
		if app.AfterRequest != nil {
			app.AfterRequest(ctx)
		}
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

func (app *App) match(ctx *Ctx) bool {
	rq := ctx.Request

	if servername != "" {
		rqUrl := rq.URL.Host
		if !strings.Contains(rqUrl, servername) {
			return false
		}
		if net.ParseIP(rqUrl) != nil {
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
	rsp := NewResponse(wr, ctx)
	rq := NewRequest(req, ctx)

	ctx.Request = rq
	ctx.Response = rsp

	defer app.closeConn(ctx)
	if app.SecretKey != "" {
		if c, ok := rq.Cookies["session"]; ok {
			ctx.Session.validate(c, app.SecretKey)
		}
	}

	if app.match(ctx) {
		rq.parseRequest()
		app.execRoute(ctx)
	} else if app.TearDownRequest != nil {
		app.TearDownRequest(ctx)
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
		fmt.Println("def?")
		address = addr[0]
	} else if !flag.Parsed() {
		flag.StringVar(&address, "address", "127.0.0.1:5000", "address of server listen. default: localhost")
		flag.Parse()
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

func (app *App) parseListener() {
	_, port, err := net.SplitHostPort(app.srv.Addr)
	if err != nil {
		l.err.Fatal(err)
	}

	localAddress := getOutboundIP()

	if !app.Silent {
		env := allowEnv[strings.ToLower(app.Env)]
		envDev := env == "" || env == "development"
		if listenInAll || envDev {
			if envDev {
				l.Defaultf("Server is listening on all address in %sdevelopment mode%s", _RED, _RESET)
			} else {
				l.Default("Server is listening on all address")
			}
			l.info.Printf("          listening on: http://%s:%s", localAddress.IP, port)
			l.info.Printf("          listening on: http://127.0.0.1:%s", port)
		} else {
			l.Default("Server is linsten in", app.srv.Addr)
		}
	}
}

// Build a app and starter Server
func (app *App) Listen(host ...string) error {
	app.Build(host...)
	app.parseListener()
	err := app.srv.ListenAndServe()
	l.Error(err)
	return err
}
