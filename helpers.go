package slow

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
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
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "0.0.0.0"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

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
