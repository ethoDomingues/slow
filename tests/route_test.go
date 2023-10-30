package tests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ethoDomingues/slow"
)

func handlerTest1(ctx *slow.Ctx)        {}
func handlerTest2(ctx *slow.Ctx)        {}
func handlerGetRouteName(ctx *slow.Ctx) { ctx.Response.TEXT(ctx.MatchInfo.Route.Name, 200) }

func TestRouteNames_Mode1(t *testing.T) {
	app := slow.NewApp(nil)
	app.AddRoute([]*slow.Route{
		{
			Url:  "/",
			Name: "route1",
			Func: handlerTest1,
		},
		{
			Url:  "/",
			Func: handlerTest1,
		},
		slow.GET("/", handlerTest2),
	}...)
	app.Build()
	names := []string{"route1", "handlerTest1", "handlerTest2"}
	for i, route := range app.Routes {
		if route.Name != names[i] {
			t.Fatalf("Route Name Unmatch: Want -> '%s', recv -> '%s'", names[i], route.Name)
		}
	}
}

func TestRouteNames_Mode2(t *testing.T) {
	app := slow.NewApp(nil)
	app.Name = "api"
	app.AddRoute([]*slow.Route{
		{
			Url:  "/",
			Name: "route1",
			Func: handlerTest1,
		},
		{
			Url:  "/",
			Func: handlerTest1,
		},
		slow.GET("/", handlerTest2),
	}...)
	app.Build()
	names := []string{"api.route1", "api.handlerTest1", "api.handlerTest2"}
	for i, route := range app.Routes {
		if route.Name != names[i] {
			t.Fatalf("Route Name Unmatch: Want -> '%s', recv -> '%s'", names[i], route.Name)
		}
	}
}

func TestRouteURLs_mode1(t *testing.T) {
	app := slow.NewApp(nil)
	app.AddRoute([]*slow.Route{
		{
			Url:  "/",
			Func: handlerTest1,
		},
		{
			Func: handlerTest1,
		},
		slow.GET("/", handlerTest2),
	}...)
	app.Build()
	for _, route := range app.Routes {
		if route.Url != "/" {
			t.Fatalf("Route URL Unmatch: Want -> '/', recv -> '%s'", route.Url)
		}
	}
}

func TestRouteURLs_mode(t *testing.T) {
	app := slow.NewApp(nil)
	app.Prefix = "/api/v1"
	app.StrictSlash = true
	app.AddRoute([]*slow.Route{
		{
			Url:  "//",
			Func: handlerTest1,
		},
		{
			Func: handlerTest1,
		},
		slow.GET("/", handlerTest2),
	}...)
	app.Build()
	for _, route := range app.Routes {
		if route.Url != "/api/v1" {
			t.Fatalf("Route URL Unmatch: Want -> '/api/v1/', recv -> '%s'", route.Url)
		}
	}
}

func TestUrlMatch(t *testing.T) {
	app := slow.NewApp(nil)
	app.Name = "main"
	app.Silent = true

	app.AddRoute(&slow.Route{
		Url:  "/varString/{v:str}",
		Name: "str",
		Func: handlerGetRouteName,
	})
	app.AddRoute(&slow.Route{
		Url:  "/varInt/{v:int}",
		Name: "int",
		Func: handlerGetRouteName,
	})

	go app.Listen(":5001")
	time.Sleep(time.Second)
	buf := bytes.NewBufferString("")

	// t1
	r, _ := http.NewRequest("GET", "http://127.0.0.1:5001/varString/test", nil)
	rsp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal("http.Request 1 failed with error: ", err)
	}

	io.Copy(buf, rsp.Body)
	if buf.String() != "main.str" {
		t.Fatalf("Unmatch Route: want -> '%s', recv -> '%s'", "main.str", buf.String())
	}
	buf.Reset()

	// t2
	r, _ = http.NewRequest("GET", "http://127.0.0.1:5001/varString/1", nil)
	rsp, err = http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal("http.Request 2 failed with error: ", err)
	}

	io.Copy(buf, rsp.Body)
	if buf.String() != "main.str" {
		t.Fatalf("Unmatch Route: want -> '%s', recv -> '%s'", "main.str", buf.String())
	}
	buf.Reset()

	// t3
	r, _ = http.NewRequest("GET", "http://127.0.0.1:5001/varInt/1", nil)
	rsp, err = http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal("http.Request 3 failed with error: ", err)
	}

	io.Copy(buf, rsp.Body)
	if buf.String() != "main.int" {
		t.Fatalf("Unmatch Route: want -> '%s', recv -> '%s'", "main.int", buf.String())
	}
	buf.Reset()

	// t4
	r, _ = http.NewRequest("GET", "http://127.0.0.1:5001/varInt/test", nil)
	rsp, err = http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal("http.Request 4 failed with error: ", err)
	}

	io.Copy(buf, rsp.Body)
	if rsp.Status != "404 Not Found" {
		fmt.Println(buf)
		t.Fatalf("Unmatch Route: want -> '%s', recv -> '%s'", "404 Not Found", rsp.Status)
	}
	buf.Reset()
}

func TestRouteMapCtrl(t *testing.T) {
	app := slow.NewApp(nil)

	app.AddRoute(&slow.Route{
		Url: "",
		MapCtrl: slow.MapCtrl{
			"GET":    {Func: func(c *slow.Ctx) {}},
			"POST":   {Func: func(c *slow.Ctx) {}},
			"PUT":    {Func: func(c *slow.Ctx) {}},
			"DELETE": {Func: func(c *slow.Ctx) {}},
		},
	})
	app.Clone()
}
