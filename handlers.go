package slow

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handlerServeFile(ctx *Ctx) {
	rsp := ctx.Response
	rq := ctx.Request

	uri := rq.URL.Path
	static := ctx.App.StaticUrlPath
	if ctx.App.Prefix != "" {
		static = filepath.Join(ctx.App.Prefix, static)
	}
	pathToFile := strings.TrimPrefix(uri, static)
	pathToFile = filepath.Join(ctx.App.StaticFolder, pathToFile)
	if f, err := os.Open(pathToFile); err == nil {

		_, file := filepath.Split(pathToFile)
		defer f.Close()
		if fStat, err := f.Stat(); err != nil || fStat.IsDir() {
			rsp.NotFound()
		}
		io.Copy(rsp, f)

		ctype := mime.TypeByExtension(filepath.Ext(file))

		if ctype == "application/octet-stream" {
			ctype = http.DetectContentType(rsp.Bytes())
		}

		rsp.Header.Set("Content-Type", ctype)
		rsp.Close()
	} else {
		rsp.TEXT(err, 404)
	}

}

func ServeFile(ctx *Ctx, pathToFile ...string) {
	rsp := ctx.Response
	path := filepath.Join(pathToFile...)

	if f, err := os.Open(path); err == nil {
		_, file := filepath.Split(path)
		defer f.Close()
		if fStat, err := f.Stat(); err != nil || fStat.IsDir() {
			rsp.NotFound()
		}
		io.Copy(rsp, f)
		ctype := mime.TypeByExtension(filepath.Ext(file))
		if ctype == "application/octet-stream" {
			ctype = http.DetectContentType(rsp.Bytes())
		}
		rsp.Header.Set("Content-Type", ctype)
		rsp.Close()
	} else if ctx.App.Env != "prod" {
		rsp.checkErrByEnv(err)
	} else {
		rsp.NotFound()
	}
}

func optionsHandler(ctx *Ctx) {
	rsp := ctx.Response
	mi := ctx.MatchInfo

	rsp.StatusCode = 200
	strMeths := mi.Route.Cors.AllowMethods
	if rsp.Header.Get("Access-Control-Allow-Methods") == "" {
		rsp.Header.Set("Access-Control-Allow-Methods", strMeths)
	}

	rsp.parseHeaders()
	rsp.Header.Save(rsp.raw)
}
