package api

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
)

/* =============================
   Disk使用率
============================= */

func DiskHandler(w http.ResponseWriter, r *http.Request) {

	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	usage := float64(used) / float64(total) * 100

	json.NewEncoder(w).Encode(map[string]float64{
		"disk": usage,
	})
}

func DFHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	lines := strings.Split(string(output), "\n")
	var data []map[string]string

	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		data = append(data, map[string]string{
			"filesystem": fields[0],
			"size":       fields[1],
			"used":       fields[2],
			"avail":      fields[3],
			"usePercent": fields[4],
			"mounted":    fields[5],
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
