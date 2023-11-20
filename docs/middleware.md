# Middleware - Full Usage Example

```go
package main

import (
 "fmt"

 "github.com/ethoDomingues/slow"
)

func main() {
 app := slow.NewApp(nil)
 app.Middlewares = slow.Middlewares{middle1, middle2}

 app.AddRoute(&slow.Route{
  Url:         "/",
  Func:        home,
  Middlewares: slow.Middlewares{middle3, middle4},
 })

 app.AddRoute(&slow.Route{
  Url:  "/echo",
  Name: "echo",
  Func: echo,
 })
 app.Listen()
}

func home(ctx *slow.Ctx) {
 ctx.response.HTML("<h1>Hello</h1>",200)
}


func middle1(ctx *slow.Ctx) {
 fmt.Println("middle 1")
 ctx.Next()
}

func middle2(ctx *slow.Ctx) {
 fmt.Println("middle 2")
 ctx.Next()
}

func middle3(ctx *slow.Ctx) {
 fmt.Println("middle 3")
 ctx.Next()
}
func middle4(ctx *slow.Ctx) {
 fmt.Println("middle 4")
 ctx.Next()
}
```
