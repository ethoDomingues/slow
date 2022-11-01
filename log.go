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

func newLogger() *Logger {
	info := log.New(os.Stdout, _GREEN+"info: "+_RESET, log.Ldate|log.Ltime)
	warn := log.New(os.Stdout, _YELLOW+"warn: "+_RESET, log.Ldate|log.Ltime)
	erro := log.New(os.Stdout, _RED+"error: "+_RESET, 0)
	return &Logger{
		info: info,
		warn: warn,
		err:  erro,
	}
}

type Logger struct {
	info *log.Logger
	warn *log.Logger
	err  *log.Logger
}

func (l *Logger) Deafault(v ...any) {
	l.info.Println(v...)
}

func (l *Logger) Error(v ...any) {
	l.err.Println(v...)
	for i := 4; i < 10; i++ {
		_, file, line, _ := runtime.Caller(i)
		fmt.Printf("\t%s:%d\n", file, line)
	}
}

func (l *Logger) LogRequest(ctxID string) {
	rq := contextsNamed[ctxID].Request
	rsp := contextsNamed[ctxID].Response

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
	l.Deafault(color, rsp.StatusCode, _RESET, "-> ", rq.Raw.Method, rq.Raw.URL.Path)
}
