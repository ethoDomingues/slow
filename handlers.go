package slow

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

func serveFile(ctx *Ctx) {
	rsp := ctx.Response
	rq := ctx.Request

	uri := rq.Raw.URL.Path
	static := ctx.App.StaticUrlPath

	pathToFile := strings.TrimPrefix(uri, static)

	dir, file := filepath.Split(ctx.App.StaticFolder + pathToFile)
	d := http.Dir(dir)
	if f, err := d.Open(file); err == nil {
		defer f.Close()
		if fStat, err := f.Stat(); err != nil || fStat.IsDir() {
			rsp.NotFound()
		}
		io.Copy(rsp.Body, f)

		ctype := mime.TypeByExtension(filepath.Ext(file))

		if ctype == "application/octet-stream" {
			ctype = http.DetectContentType(rsp.Body.Bytes())
		}

		rsp.Headers.Set("Content-Type", ctype)
		rsp.Close()
	} else {
		rsp.NotFound()
	}
}

func optionsHandler(ctx *Ctx) {
	rsp := ctx.Response
	mi := ctx.MatchInfo

	rsp.StatusCode = 200
	strMeths := strings.Join(mi.Route().Cors.AllowMethods, ", ")
	rsp.Headers.Set("Access-Control-Allow-Methods", strMeths)

	rsp.parseHeaders()
	rsp.Headers.Save(rsp.raw)
}
