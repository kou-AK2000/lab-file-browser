package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"linux-fs-viewer/model"
)

func ProcessHandler(w http.ResponseWriter, r *http.Request) {

	entries, _ := os.ReadDir("/proc")
	var list []model.ProcessInfo

	for _, e := range entries {

		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}

		status, err := os.ReadFile("/proc/" + e.Name() + "/status")
		if err != nil {
			continue
		}

		var rss, vsz uint64
		var state string
		var uid int

		lines := strings.Split(string(status), "\n")
		for _, l := range lines {
			if strings.HasPrefix(l, "VmRSS:") {
				fmt.Sscanf(l, "VmRSS: %d kB", &rss)
			}
			if strings.HasPrefix(l, "VmSize:") {
				fmt.Sscanf(l, "VmSize: %d kB", &vsz)
			}
			if strings.HasPrefix(l, "State:") {
				state = l
			}
			if strings.HasPrefix(l, "Uid:") {
				fmt.Sscanf(l, "Uid: %d", &uid)
			}
		}

		u, _ := user.LookupId(strconv.Itoa(uid))
		cmd, _ := os.ReadFile("/proc/" + e.Name() + "/cmdline")

		list = append(list, model.ProcessInfo{
			PID:     pid,
			User:    u.Username,
			RSS:     rss,
			VSZ:     vsz,
			State:   state,
			Command: strings.ReplaceAll(string(cmd), "\x00", " "),
		})
	}

	json.NewEncoder(w).Encode(list)
}



func ProcessesHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("ps", "aux", "--sort=-%cpu")
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "ps failed", http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(output), "\n")

	var processes []model.ProcessInfo

	// 1行目はヘッダなのでスキップ
	for _, line := range lines {

    	fields := strings.Fields(line)
     	if len(fields) < 6 {
        	continue
      	}

       	pid, _ := strconv.Atoi(fields[0])
        username := fields[1]
        rss, _ := strconv.ParseUint(fields[2], 10, 64)
        vsz, _ := strconv.ParseUint(fields[3], 10, 64)
        state := fields[4]
        command := strings.Join(fields[5:], " ")

        processes = append(processes, model.ProcessInfo{
        	PID:     pid,
        	User:    username,
        	RSS:     rss,
        	VSZ:     vsz,
        	State:   state,
        	Command: command,
        })
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(processes)
}
