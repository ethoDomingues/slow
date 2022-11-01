# Slow

## Simple Example
```go
package main

import "github.com/ethodomingues/slow"

func main() {
	app := slow.NewApp()
	app.Get("/hello", helloWorld)
	app.Get("/hello/{name}", helloUser) // 'name' is any string
	app.Get("/hello/{userID:int}", userByID) // 'userID' is only int

	app.Listen()
}

func helloWorld(ctx *slow.Ctx) {
	rsp := ctx.Response
	rsp.JSON(map[string]any{"Hello": "World"}, 200)
}

func helloUser(ctx *slow.Ctx) {
	rq := ctx.Request   // current Request
	rsp := ctx.Response // current Response

	name := rq.Args["name"]
	rsp.HTML("Hello "+name, 200)
}
func userByID(ctx *slow.Ctx) {
	rq := ctx.Request   // current Request
	rsp := ctx.Response // current Response

	id := rq.Args["userID"]
	user := AnyQuery(id)
	rsp.JSON(user, 200)
}

```