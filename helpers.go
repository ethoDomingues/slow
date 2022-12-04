package slow

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime"
	"strings"
	"time"
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

func HtmlEscape(s string) string { return htmlReplacer.Replace(s) }

// Alias of 'fmt.Sprintf("%T", obj)'
func TypeOf(obj any) string { return fmt.Sprintf("%T", obj) }

// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func getFreePort() (port string) {
	for i := 1000; i < 64000; i++ {
		addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(i))
		conn, err := net.DialTimeout("tcp", addr, time.Millisecond*200)
		if err != nil {
			continue
		}
		defer conn.Close()
		_, port, _ = net.SplitHostPort(conn.LocalAddr().String())
		break
	}
	fmt.Println(port)
	return port
}
