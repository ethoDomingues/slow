package slow

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func serveFile(ctx *Ctx) {
	rsp := ctx.Response
	rq := ctx.Request

	uri := rq.URL.Path
	static := ctx.App.StaticUrlPath
	if ctx.App.Prefix != "" {
		static = filepath.Join(ctx.App.Prefix, static)
	}

	pathToFile := strings.TrimPrefix(uri, static)
	p := GetFullPath()
	pathToFile = filepath.Join(p, ctx.App.StaticFolder, pathToFile)

	if f, err := os.Open(pathToFile); err == nil {

		_, file := filepath.Split(pathToFile)
		defer f.Close()
		if fStat, err := f.Stat(); err != nil || fStat.IsDir() {
			rsp.NotFound()
		}
		io.Copy(rsp.Body, f)

		ctype := mime.TypeByExtension(filepath.Ext(file))

		if ctype == "application/octet-stream" {
			ctype = http.DetectContentType(rsp.Body.Bytes())
		}

		rsp.Header.Set("Content-Type", ctype)
		rsp.Close()
	} else if ctx.App.Env != "prod" {
		rsp.NotFound(err)
	} else {
		rsp.NotFound()
	}
}

func optionsHandler(ctx *Ctx) {
	rsp := ctx.Response
	mi := ctx.MatchInfo

	rsp.StatusCode = 200
	strMeths := strings.Join(mi.Route.Cors.AllowMethods, ", ")
	if rsp.Header.Get("Access-Control-Allow-Methods") == "" {
		rsp.Header.Set("Access-Control-Allow-Methods", strMeths)
	}

	rsp.parseHeaders()
	rsp.Header.Save(rsp.raw)
}
