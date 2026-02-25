package api

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strings"
	"syscall"

	"linux-fs-viewer/model"
)

/* =============================
   システム情報
============================= */

func SystemHandler(w http.ResponseWriter, r *http.Request) {

	host, _ := os.Hostname()

	var uname syscall.Utsname
	syscall.Uname(&uname)

	kernel := ""
	for _, c := range uname.Release {
		if c == 0 {
			break
		}
		kernel += string(byte(c))
	}

	// Distribution取得
	dist := "Unknown"
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, l := range lines {
			if strings.HasPrefix(l, "PRETTY_NAME=") {
				dist = strings.Trim(l[13:], `"`)
				break
			}
		}
	}

	json.NewEncoder(w).Encode(model.SystemInfo{
		Hostname:     host,
		Distribution: dist,
		Kernel:       kernel,
		Arch:         runtime.GOARCH,
	})
}
