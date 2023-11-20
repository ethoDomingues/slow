# Slow

## [See the Documentation](https://github.com/ethoDomingues/slow/blob/main/docs/doc.md)

## Simple Example

### [With a correctly configured Go toolchain:](https://go.dev/doc/install)

```sh
go get github.com/ethoDomingues/slow
```

> _main.go_

```go
package main

import "github.com/ethoDomingues/slow"

func main() {
 app := slow.NewApp()
 app.GET("/hello", helloWorld)
 app.GET("/hello/{name}", helloUser) // 'name' is any string
 app.GET("/hello/{userID:int}", userByID) // 'userID' is only int

 fmt.Println(app.Listen())
}

func helloWorld(ctx *slow.Ctx) {
 hello := map[string]any{"Hello": "World"}
 ctx.JSON(hello, 200)
}

func helloUser(ctx *slow.Ctx) {
 rq := ctx.Request   // current Request
 name := rq.Args["name"]
 ctx.HTML("<h1>Hello "+name+"</h1>", 200)
}

func userByID(ctx *slow.Ctx) {
 rq := ctx.Request   // current Request
 id := rq.Args["userID"]
 user := AnyQuery(id)
 ctx.JSON(user, 200)
}
```
