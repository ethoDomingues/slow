package slow

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func ServeFile(ctx *Ctx) {
	rsp := ctx.Response
	rq := ctx.Request

	uri := rq.Raw.URL.Path
	static := ctx.App.StaticUrlPath

	pathToFile := strings.TrimPrefix(uri, static)

	dir, file := filepath.Split(ctx.App.StaticFolder + pathToFile)
	d := http.Dir(dir)
	if f, err := d.Open(file); err == nil {
		defer f.Close()
		io.Copy(rsp.Body, f)
		ext := file
		splitFilename := strings.Split(file, ".")
		if len(splitFilename) > 1 {
			ext = splitFilename[len(splitFilename)-1]
		}
		ct, ok := TypeByExtension[ext]
		if !ok {
			ct = getTypebyFilename(file)
		}

		rsp.Headers.Set("Content-Type", ct)
		rsp.Close()
	} else {
		rsp.NotFound()
	}
}
