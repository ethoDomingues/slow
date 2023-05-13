package slow

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

func getFunctionName(i interface{}) string {
	splitName := strings.Split(
		runtime.FuncForPC(
			reflect.ValueOf(i).Pointer(),
		).Name(), ".",
	)
	return splitName[len(splitName)-1]
}

// Get preferred outbound ip of this machine
func getOutboundIP() *net.UDPAddr {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr)
}

func HtmlEscape(s string) string { return htmlReplacer.Replace(s) }

// Alias of 'fmt.Sprintf("%T", obj)'
func TypeOf(obj any) string { return fmt.Sprintf("%T", obj) }

func GetFullPath() string {
	p, _ := os.Executable()
	if p == "" || strings.HasPrefix(p, "/tmp") {
		p, _ = os.Getwd()
		return p
	} else {
		ps := strings.Split(p, "/")
		ps[len(ps)-1] = ""
		p = filepath.Join(ps...)
	}
	return filepath.Join("/", p)
}
