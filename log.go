package slow

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

const (
	_RED            = "\033[31m"
	_BLUE           = "\033[34m"
	_CYAN           = "\033[36m"
	_BLACK          = "\033[30m"
	_GREEN          = "\033[32m"
	_WHITE          = "\033[37m"
	_YELLOW         = "\033[33m"
	_MAGENTA        = "\033[35m"
	_BRIGHT_RED     = "\033[91m"
	_BRIGHT_BLUE    = "\033[94m"
	_BRIGHT_CYAN    = "\033[96m"
	_BRIGHT_BLACK   = "\033[90m"
	_BRIGHT_GREEN   = "\033[92m"
	_BRIGHT_WHITE   = "\033[97m"
	_BRIGHT_YELLOW  = "\033[93m"
	_BRIGHT_MAGENTA = "\033[95m"
	_RESET          = "\033[m"
)

func newLogger(logFile string) *logger {
	var lFile *log.Logger
	if logFile != "" {
		var f *os.File
		_, err := os.Stat(logFile)
		if err != nil {
			f, err = os.Create(logFile)
			if err != nil {
				panic(err)
			}
		} else {
			f, err = os.Open(logFile)
			if err != nil {
				panic(err)
			}
		}
		lFile = log.New(f, "", log.Ldate|log.Ltime)
	}

	return &logger{
		err:     log.New(os.Stdout, _RED+"error: "+_RESET, 0),
		warn:    log.New(os.Stdout, _YELLOW+"warn: "+_RESET, log.Ldate|log.Ltime),
		info:    log.New(os.Stdout, _GREEN+"info: "+_RESET, log.Ldate|log.Ltime),
		logFile: lFile,
	}
}

type logger struct {
	info    *log.Logger
	warn    *log.Logger
	err     *log.Logger
	logFile *log.Logger
}

func (l *logger) Default(v ...any) {
	l.info.Println(v...)
	if l.logFile != nil {
		l.logFile.Println(v...)
	}
}

func (l *logger) Defaultf(formatString string, v ...any) {
	l.info.Printf(formatString, v...)
	if l.logFile != nil {
		l.logFile.Printf(formatString, v...)
	}
}

func (l *logger) Error(v ...any) {
	l.err.Println(v...)
	if l.logFile != nil {
		l.logFile.Println(v...)
	}
	for i := 0; i < 20; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			return
		}
		fmt.Printf("\t%s:%d\n", file, line)
	}
}

func (l *logger) LogRequest(ctx *Ctx) {
	rq := ctx.Request
	rsp := ctx.Response

	color := ""
	switch {
	case rsp.StatusCode >= 500:
		color = _RED
	case rsp.StatusCode >= 400:
		color = _YELLOW
	case rsp.StatusCode >= 300:
		color = _CYAN
	case rsp.StatusCode >= 200:
		color = _GREEN
	default:
		color = _WHITE
	}
	addr := ""
	if ctx.MatchInfo.Router != nil {
		addr = ctx.MatchInfo.Router.Subdomain
	}
	if addr != "" {
		addr = addr + ".[...]" + rq.URL.Path
	} else {
		addr = rq.URL.Path
	}

	rd := rq.RemoteAddr
	if rd != "" {
		rd = "[" + rd + "]"
	}

	if l.logFile != nil {
		l.logFile.Printf("%s %d -> %s -> %s", rd, rsp.StatusCode, rq.Method, addr)
	}
	if ctx.App.Silent {
		return
	}

	appName := ""
	if ctx.App.Name != "" {
		appName = ctx.App.Name + ": "
	}
	l.info.Printf("%s%s %s%d%s -> %s -> %s", appName, rd, color, rsp.StatusCode, _RESET, rq.Method, addr)
}
