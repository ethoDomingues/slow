# Slow

## Simple Example
```go
package main

import "github.com/ethodomingues/slow"

func main() {
	app := slow.NewApp()
	app.Get("/hello", helloWorld)
	app.Get("/hello/{user}", helloUser) // user is any string
	app.Get("/hello/{userID:int}", helloUser) // userID is only int

	app.Listen()
}

func helloWorld(ctx *slow.Ctx) {
	rsp := ctx.Response
	rsp.JSON(map[string]any{"Hello": "World"}, 200)
}

func helloUser(ctx *slow.Ctx) {
	rq := ctx.Request   // current Request
	rsp := ctx.Response // current Response

	user := rq.Args["user"]
	rsp.HTML("Hello "+user, 200)
}

```