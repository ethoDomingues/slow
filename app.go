package slow

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

var (
	l = newLogger()

	appStack      []*App
	servername    string
	localAddress  = GetOutboundIP()
	contextsNamed map[string]*Ctx
)

// Retur a new app with a default settings
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

// Parse the router and your routes
func (app *App) build() {
	if app.building {
		return
	}
	appStack = append(appStack, app)
	go app.parseHosts()

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

// read hosts file and create a regex of hosts and addrss
func (app *App) parseHosts() {
	_hosts := map[string]any{}
	f, err := os.ReadFile("/etc/hosts")
	if err != nil {
		panic(err)
	}
	fStr := string(f)
	for _, str := range strings.Split(fStr, "\n") {
		if str == "" {
			continue
		}
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, "#") {
			continue
		}
		var h string
		var h2 string

		if strings.Contains(str, "\t") {
			hs := strings.Split(str, "\t")
			h = hs[0]
			if len(hs) > 1 {
				h2 = hs[1]
			}
		} else {
			hs := strings.Split(str, " ")
			h = hs[0]
			if len(hs) > 1 {
				h2 = hs[1]
			}
		}
		if h != "" {
			_hosts[h] = nil
		}
		if h2 != "" {
			_hosts[h2] = nil
		}
	}
	if _, ok := _hosts["0.0.0.0"]; !ok {
		_hosts["0.0.0.0"] = nil
	}

	hSlice := []string{}
	for k := range _hosts {
		hSlice = append(hSlice, k)
	}
	strH := strings.Join(hSlice, "|")
	hosts = regexp.MustCompile(`^(` + strH + `)((\:\d+)?)$`)
}

// register the rouuter in app
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

func (app *App) execRoute(ctx *Ctx) {
	rsp := ctx.Response
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
	if app.BeforeRequest != nil {
		app.BeforeRequest(ctx)
	}
	for _, mid := range ctx.MatchInfo.Router().Middlewares {
		mid(ctx)
	}
	for _, mid := range ctx.MatchInfo.Route().Middlewares {
		mid(ctx)
	}
	// if raise a error in any mid, Route.Func not is executed.
	ctx.MatchInfo.Func(ctx)

	// if raise a error in any mid or in route func, app.AfterRequest not is executed.
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

	mi := ctx.MatchInfo
	for _, router := range app.routers {
		if router.Match(ctx) {
			break
		}
	}
	if mi.Match {
		if rq.Raw.Method == "OPTIONS" {
			rsp.StatusCode = 200
			strMeths := strings.Join(mi.Route().Cors.AllowMethods, ", ")
			rsp.Headers.Set("Access-Control-Allow-Methods", strMeths)

			rsp._afterRequest()

			rsp.Headers.Save(rsp.raw)
			fmt.Fprint(rsp.raw, "")
		} else {
			rq.Query = req.URL.Query()
			rq.Args = re.getUrlValues(mi.Route().fullUrl, req.URL.Path)
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
