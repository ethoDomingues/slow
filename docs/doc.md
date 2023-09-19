# Documentation

---

> **Table Of Content**

- [Documentation](#documentation)
  - [**_App_**](#app)
  - [Cors](#cors)
  - [Ctx](#ctx)
  - [File](#file)
  - [Func](#func)
  - [Header](#header)
  - [JWT](#jwt)
  - [MapCtrl](#mapctrl)
  - [MatchInfo](#matchinfo)
  - [Meth](#meth)
  - [Methods](#methods)
  - [Middlewares](#middlewares)
  - [Request](#request)
  - [Response](#response)
  - [Route](#route)
  - [Router](#router)
  - [URL](#url)
  - [Schema](#schema)
  - [Session](#session)

---

## **_App_**

- ### App Attributes

  - *[Router](#router)
    > ...
  - **Env** _string_

            - development (default)
            - testing
            - production

  - **LogFile** _string_
      > path to log file

  - **SecretKey** _string_
      > a secret of session cript

  - **Servername** _string_
      > current hostname ( **example.com**,  **<www.hostname.org>**, etc... )

  - **_StaticFolder_** _string_
      > path to assets folder of app.

  - **StaticUrlPath** _string_
      > url path for access assets files (images, files, ...)

  - **TemplateFolder** _string_
      > path to func **Render_template()**

  - **Silent** _bool_
      > if is true, hide console log

  - **EnableStatic** _bool_
      > enable route of server file

  - **BeforeRequest( )** [Func](#func)
      > is exec before each handler
      >
      ```go
      app.BeforeRequest = func(ctx *slow.Ctx) {
            db := DBSessionExample()
            user := db.Find(&User{})
            ctx.Global["db"] = db
            ctx.Global["current_user"] = user
            // do anything
      }
      ```

  - **AfterRequest( )** [Func](#func)
      > is exec after each handler ( 'if not raise a error' )

      ```go
      app.AfterRequest = func(ctx *slow.Ctx) {
            // do anything here
            ctx.Response.Header.Add("X-Foo","Bar")
      }
      ```

  - **TearDownRequest( )** [Func](#func)
      > is exec after each handler close response (no has effect in response)

       ```go
      app.TearDownRequest = func(ctx *slow.Ctx) {
            db := ctx.Global["db"].(*Database)
            db.CloseConn()
      }
      ```

- ### App Methods

  - **NewApp( ) ->** _\*App_
      > return a new *App with defaults settings

  - _(app \*App)_ **Build(** _addr ...string_ **)**
      > build the current app, but dont start the server

  - _(app \*App)_ **Listen(** _host ...string_ **) ->** _error_
      > Build a app  and starter Server

  - _(app \*App)_ **ListenTLS(** _path_certFile string, path_keyFile string, host ...string_ **) ->** _error_
      > Build a app and starter Server

  - _(app \*App)_ **Mount**( _routers ...\*[Router](#router)_ **)**
      > Register the routers in app

    ```go
      func main() {
            api := NewRouter("api")
            api.Subdomain = "api"
            api.Prefix = "/v1"
            api.post("/products")
            api.get("/products/{productID:int}")

            app := NewApp()

            // This Function
            app.Mount(getApiRouter)

            app.Listen()
      }
    ```

  - _(app \*App)_ **ServeHTTP(** _http.ResponseWriter, http.Request_ **)**
      > is the http.Handler func

  - _(app *App)_ **ShowRoutes( )**
      > List all Routes in Terminal or Console

  - _(app *App)_ **UrlFor(** _name string, external bool, args ...string_ **) ->** _string_
      > Url Builder

      ```go
        app.GET("/users/{userID:int}", index)

        app.UrlFor("index", false, "userID", "1")
        // results: /users/1

        app.UrlFor("index", true, "userID", "1")
        // results: http://yourAddress/users/1
      ```

- ### App Example

    [**Full example here**](https://github.com/ethodomingues/slow_example)

    ```go
    package main

    import "github.com/ethodomingues/slow"

    func main() {
        app := slow.NewApp()
        app.GET("/",index)
        app.GET("/{name}",dynamicRoute)
        app.Listen()
    }

    func index(ctx *slow.Ctx) {
        ctx.Response.HTML("<h1>Hello World!</h1>")
    }

    func dynamicRoute(ctx *slow.Ctx) {
        name := ctx.Request.Args["name"]
        ctx.Response.HTML("<h1>Hello, "+name+"!</h1>")
    }
    ```

## Cors
>
> If present on route or router, allows resource sharing between origins

- ### Cors Attributes
  >
  > [CORS documentation](https://developer.mozilla.org/pt-BR/docs/Web/HTTP/CORS)
  - **MaxAge** _string_
      > Access-Control-Max-Age

  - **AllowOrigin** _string_
      > Access-Control-Allow-Origin

  - **AllowMethods** _[ ]string_
      > Access-Control-Allow-Methods

  - **AllowHeaders** _[ ]string_
      > Access-Control-Allow-Headers

  - **ExposeHeaders** _[ ]string_
      > Access-Control-Expose-Headers

  - **RequestMethod** _string_
      > Access-Control-Request-Method

  - **AllowCredentials** _bool_
      > Access-Control-Allow-Credentials

- ### CORS Example

    ```go
    func main() {
        router := slow.Router{
            Name:"api",
            Prefix:"/v1",
            Subdomain:"api",
            Cors: &slow.Cors{
                MaxAge:"36000",
                AllowOrigin:"*",
                AllowMethods:[]string{"GET","POST"},      // automatic answer
                AllowHeaders:[]string{"Authorization"},   // automatic answer
                ExposeHeaders:[]string{"Authorization"},
                RequestMethod:"GET",                      // automatic answer
                AllowCredentials: true,
            },
        }
    }
    ```

## Ctx

- ### Ctx Attributes

  - **App** _*[App](#app)_
      > Current App

  - **Global** _map\[string]any_
      > a map of storage any data

      ```go
      func beforeRequest(ctx *Ctx) {
            db := DBSessionExample()
            user := db.Find(&User{})
            ctx.Global["db"] = db
            ctx.Global["current_user"] = user
            // do anything 
        }

        func routeExample(ctx *Ctx) {
            db := ctx.Global["db"].(*Database)
            user := ctx.Global["current_user"]
            // do more anything 
        }
      ```

  - **Request** _*[Request](#request)_
      > Current Request

  - **Response** _*[Response](#response)_
      > Current Response

  - **Session** _*[Session](#session)_
      > ...

  - **MatchInfo** _*[MatchInfo](#matchinfo)_
      > Contains information about the current request and the route

- ### Ctx Methods

  - _(ctx *Ctx)_ **UrlFor(** _name string, external bool, args ...string_ **) ->** _string_
      > ...

- ### Ctx Example

    ```go

    ```

## File

- ### File Attributes

  - **Filename** _string_
      > name of file

  - **ContentType** _string_
      > Content-Type of file

  - **ContentLeght** _int_
      > Content-Length of file

  - **Stream** _*bytes.Buffer_
      > buffer with the file

- ### File Methods

  - **NewFile(** _*multipart.Part_  **) ->** _*File_
      > ..

- ### File Example

      ...

## Func
>
> **func( \*[Ctx](#ctx)** )

```go
func index(ctx *slow.Ctx) {
    body := map[string]string{"hello":"world"}
    ctx.Response.JSON(, 200)
}
```

## Header
 >
 > **http.Header**

- ### Header Methods

  - _(h *Header)_ **Add(** _key string, value string_ **)**
      > Add value in a Header Key. If the key does not exist, it is created

  - _(h *Header)_ **Del(** _key string_ **)**
      > Delete a value in Header

  - _(h *Header)_ **Get(** _key string_ **) ->** _string_
      > Return a value of Header Key. If the key does not exist, return a empty string

  - _(h *Header)_ **Save(** _w http.ResponseWriter_ **)**
      > Write the headers in the response

  - _(h *Header)_ **Set(** _key string, value string_ **)**
      > Set a Header. (if the key exists, it is overwritten )

  - _(h_Header)_**SetCookie(**_cookie http.Cookie_ **)**
      > Set a Cookie. Has the same effect as 'Response.SetCookie'

- ### Header Example

```go
func index(ctx *slow.Ctx) {
    body := map[string]string{"hello":"world"}

    ctx.Response.Header.Set("X-Token","tokenValue")
    ctx.Response.Header.Set("X-Header","value")
    ctx.Response.JSON(, 200)
}
```

## JWT

- ### JWT Attributes

  - **Headers** _map[string]string_
      > ...

  - **Payload** _map[string]string_
      > ...

  - **Secret**  _string_
      > ...

- ### JWT Methods

  - **NewJWT(** _secret string_ **) ->** _*JWT_
      > ...

  - **ValidJWT(** _jwt string, secret string_ **)->** _(*JWT, bool)_
      > ...

  - _(j *JWT)_ **Sign( ) ->** _string_
      > ...

- ### JWT Example

## MapCtrl

> map[string]*[Meth](#meth)

- ### MapCtrl Example

```go
var Routes = []*slow.Route{
    {
        Url:"/login",
        Name:"loginRoute",
        MapCtrl: slow.MapCtrl{
            "GET":{Func:func1},
            "POST":{Func:func2},
            "PUT":{Func:func3},
        },
    },
}
```

## MatchInfo

- ### MatchInfo Attibutes

  - **Func** _*[Func](#func)_
  
  - **Match** _bool_

  - **MethodNotAllowed** _error_

  - **Route**  _*[Route](#route)_

  - **Router** _*[Router](#router)_

## Meth

- ### Meth Attibutes

  - **[Func](#func)** _*Func_

  - **Method** _string_

  - **[Schema](#schema)** _[Schema](#schema)_

## Methods

> [ ]string

## Middlewares

> \[ ][Func](#func)

- ### Middlewares Methods

  - **NewMiddleware(** _f ...[Func](#func)_ **) ->** _Middlewares_

- ### Middlewares Example

```go
func mid1(ctx *slow.Ctx) {
    // is executed fist, after "app.BeforeRequest"
    return
}

func mid2(ctx *slow.Ctx) {
    // is executed 2°
    return
}

func index(ctx *slow.Ctx) {
    // runs after all "mids", if not aborted or has an error
}
func main(){
    ...
    app.Middlewares = NewMiddleware(mid1,mid2)
    ...
}
```

## Request

- ### Request Atrtributes

  - **Header** [_Header_](#header)

  - **Body** _string_
      > Raw Request Body

  - **Method** _string_
      > request Method

  - **RemoteAddr** _string_
      > possible Remote address

  - **RequestURI** _string_
      > URl request

  - **ContentType** _string_
      > Request Content-Type

  - **ContentLength** _int_
      > Request Body Content Length

  - **URL** _*url.URL_
      > URL object from current request

  - **Form** map\[string]any
      > form with data from request (JSON, XML, POST, MULTIPART-FORM,... )

  - **Args** map\[string]string
      > Args from url (example: "/user/{name}", value from variable name)

  - **Mime**  map\[string]string
      > Mimes from request Header

  - **Query** map\[string][]string
      > Query args from URL

  - **Files** map\[string][]*[File](#file)
      > Files from Multipart Form Data

  - **Cookies** map\[string]*http.Cookie
      > all cookies from request

  - **TransferEncoding** _[ ]string_
      > ...

  - **Proto** _string_
      > "HTTP/1.0"

  - **ProtoMajor** _int_
      > 1

  - **ProtoMinor** _int_
      > 0

- ### Request Methods

  - NewRequest(req _http.Request, ctx_Ctx) *Request
      > Create a new Request Object

  - (*Request) BasicAuth() (username, password string, ok bool)
      > Get basic user:email from authorization tag

  - (*Request) Cancel()
      > Abort the current request. Server does not respond to client

  - (*Request) Clone(ctx context.Context)*Request
      > Clone the context

  - (*Request) Context() context.Context
      > Returns a 'context.Context' of the current request

  - (*Request) Ctx()*Ctx
      > Returns a '*Ctx' of the current request

  - (*Request) ProtoAtLeast(major, minor int) bool
      > ...

  - (*Request) RawRequest()*http.Request
      > return a *http.Request

  - (*Request) Referer() string
      > previous request address

  - (*Request) RequestURL() string
      > Build e return the current request URL

  - (*Request) UrlFor(name string, external bool, args ...string) string
      > alias of App.UrlFor

  - (*Request) UserAgent() string
      > ...

  - (*Request) WithContext(ctx context.Context)*Request
      > ...

- ### Request Example

```go
// route url -> /api/user/{userID:int}
func index(ctx *slow.Ctx) {
    req := ctx.Request
    userID := req.Args["userID"]

    if req.Method == "POST" {
        data := req.Form
        username := req.Form["username"]
        user, passwd, ok := req.BasicAuth()
        ...
    }
    ...
}
```

## Response

- ### Response Attributes

  - **StatusCode** _int_
      > status code that will be written in the response

  - **Body**   _*bytes.Buffer_
      > response body.
      >
      > you can write directly or use some methods as you like

  - **Header** _[Header](#header)_
      > header that will be written in the response

- ### Response Methods

  - **NewResponse(** _wr http.ResponseWriter, ctx *Ctx_ **) ->** _*Response_
      > create a new Response with defaults settings

  - _(*Response)_ **BadRequest(** _body ...any_ **)**
      > Sends a BadRequest ( Status and Text )

  - _(*Response)_ **Close( )**
      > Halts execution and closes the "response". This does not clear the response body

  - _(*Response)_ **Forbidden(** _body ...any_ **)**
      > Sends a StatusForbidden (403) - Status and Text

  - _(*Response)_ **HTML(** _body string, code int_ **)**
      > Sends an HTML document with its corresponding body data and status code

  - _(*Response)_ **ImATaerpot(** _body ...any_ **)**
      > Sends a StatusImATaerpot (418) - Status and Text

  - _(*Response)_ **InternalServerError(** _body ...any_ **)**
      > Sends a StatusInternalServerError (500) - Status and Text

  - _(*Response)_ **JSON(** _body any, code int_ **)**
      > sends an JSON document with its corresponding body data and status code

  - _(*Response)_ **MethodNotAllowed(** _body ...any_ **)**
      > Sends a StatusMethodNotAllowed (405) - Status and Text

  - _(*Response)_ **NotFound(** _body ...any_ **)**
      > Sends a StatusNotFound (404) - Status and Text

  - _(*Response)_ **Ok(** _body ...any_ **)**
      > Sends a StatusOk (200) - Status and Text

  - _(*Response)_ **Redirect(** _url string_ **)**
      > Redirects to the following url

  - _(*Response)_ **RenderTemplate(** _pathToFile string, data ...any_ **)**
      > Parse Html file and sends to client

  - _(*Response)_ **SetCookie(** _cookie http.Cookie_ **)**
      > Set a cookie in the Headers of Response

  - _(*Response)_ **TEXT(** _body string, code int_ **)**
      > Sends an TEXT document with its corresponding body data and status code

  - _(*Response)_ **Unauthorized(** _body ...any_ **)**
      > Sends a Unauthorized (401) - Status and Text

- ### Response Example

        ...

## Route

- ### Route Attributes

  - **[Url](#url)** _string_

  - **Name** _string_

  - **Func** _[Func](#func)_

  - **MapCtrl** _[MapCtrl](#mapctrl)_

  - **Cors** *_[Cors](#cors)_

  - **Schema** _[Schema](#schema)_

  - **Methods** _[Methods](#methods)_

  - **Middlewares** _[Middlewares](#middlewares)_

- ### Route Methods

  - **Connect(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "CONNECT" method

  - **Delete(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "DELETE" method

  - **Get(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "GET" method

  - **Head(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "HEAD" method

  - **Options(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "OPTIONS" method

  - **Patch(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "PATCH" method

  - **Post(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "POST" method

  - **Put(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "PUT" method

  - **Trace(** _[url](#url) string, f [Func](#func)_ **) ->** _*Route_
      > return a route that matches the "TRACE" method

- ### Route Example

    ```go
        var routes := []*slow.Route{
            slow.Post("/users",postUsersFunc),
            slow.Get("/users/{userID:int}",getUserFunc),
            slow.Put("/users/{userID:int}",putUserFunc),
            slow.Delete("/users/{userID:int}",delUserFunc),

            slow.Post("/products",postProductsFunc),
            {
                Url:  "/products/{productsID:int}",
                Name: "products",
                MapCtrl: slow.MapCtrl{
                    "GET": {
                        Func: getProducts,
                    },
                    "POST": {
                        Func: postProducts,
                    },
                    "PUT": {
                        Func: putProducts,
                    }
                },
            }
        }
        func main() {
            ...
            router.AddAll(routes)
            app.Mount(router)
            ...
        }
    ```

## Router

- ### Router Attributes

  - **Name** _string_
      > Router Name

  - **Prefix** _string_
      > prefix of url route. example:
      >
      > route url -> "/user/photos"
      >
      > router prefix> "/api/v1"
      >
      > final url -> "/api/v1/user/photos"

  - **Subdomain** _string_
      > router subdomain. example.
      >
      > subdomain -> "api"
      >
      > app.Servername -> "example.com"
      >
      > the app will listen on "api.example.com"

  - **Cors**        *[Cors](#cors)
      > ...

  - **Routes**      \[ ]\*[Route](#route)
      > ...

  - **Middlewares** [Middlewares](#middlewares)
      > ...

- ### Router Methods

  - **NewRouter(** _name string_ **) ->** _*Router_

  - _(r *Router)_ **Add(** _[url](#url) string, name string, f [Func](#func), meths []string_ **)**
      > example:

      ```go
      router := &slow.Router{}
      router.Add("/", "index", func1, []string{"GET"})
      ```

  - _(r *Router)_ **AddAll(** _routes ...*Route_ **)**
      > example

      ```go
      var routes = []*slow.Route{}

      var router = &slow.Router{}
      router.AddAll(routes...)
            
      ```

  - _(r *Router)_ **Connect(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "CONNECT" method

  - _(r *Router)_ **Delete(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "DELETE" method

  - _(r *Router)_ **Get(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "GET" method

  - _(r *Router)_ **Head(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "HEAD" method

  - _(r *Router)_ **Options(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "OPTIONS" method

  - _(r *Router)_ **Patch(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "PATCH" method

  - _(r *Router)_ **Post(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "POST" method

  - _(r *Router)_ **Put(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "PUT" method

  - _(r *Router)_ **Trace(** _[url](#url) string, f [Func](#func)_ **)**
      > adds a new route that corresponds to the http "TRACE" method

- ### Router Example

    ```go
    func main() {
        router := slow.Router{
            Name:"api",
            Prefix:"/v1",
            Subdomain:"api",
            Cors: &slow.Cors{
                AllowOrigin: "*",
            },
        }
        router.Get("/user/{userID:int}",anyFunc1)
        router.Get("/user/{userID:int}/profile",otherFunc1)
        router.Put("/user/{userID:int}/profile",otherFunc2)
        router.Head("/user/{userID:int}/profile",anyFunc2)
        router.Post("/user/{userID:int}/profile",otherFunc3)

        app.Mount(router)
        ...
    }
    ```

## URL

> _url_ : _string_

- usage:

  - literal path
    >
    > - "/foo", "/bar/", "/foo/bar", ...

  - dynamic path
    >
    >- "/foo/{dynamic}" -> any string
    >- "/foo/{dynamic:str}" -> any string
    >- "/foo/{dynamic:int}" -> only int (0-9+)
    >- "/foo/{dynamic:path}" -> any path ("/path/to/filename")

## Schema

> any value of data validation frm request form

## Session

- ### Session Attributes

  - **Permanent** *_bool_
      > if true, session validity is one year, otherwise half hour

- ### Session Methods

  - _(s *Session)_ **Del(** _key string_ **)**
      > Delete a Value from Session

  - _(s *Session)_ **Get(** _key string_ **) ->** _(string, bool)_
      > Returns a session value based on the key. If key does not exist, returns an empty string

  - _(s *Session)_ **GetSign( ) ->** _string_
      > Returns a JWT Token from session data

  - _(s *Session)_ **Set(** _key, value string_ **)**
      > This inserts a value into the session

- ### Session Example

```go
func login(ctx *slow.Ctx) {
    ...
    if user.Authenticate() {
        ctx.Session.Set("user",user.ID)
        ...
    }
    ...
}
```
