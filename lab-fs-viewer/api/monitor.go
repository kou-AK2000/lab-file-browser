package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"linux-fs-viewer/model"
)

/* =============================
   CPU使用率
============================= */

var lastIdle, lastTotal uint64

func cpuUsage() float64 {

	data, _ := os.ReadFile("/proc/stat")
	fields := strings.Fields(strings.Split(string(data), "\n")[0])

	var idle, total uint64

	for i, v := range fields[1:] {
		val, _ := strconv.ParseUint(v, 10, 64)
		total += val
		if i == 3 {
			idle = val
		}
	}

	diffIdle := idle - lastIdle
	diffTotal := total - lastTotal

	lastIdle = idle
	lastTotal = total

	if diffTotal == 0 {
		return 0
	}

	return 100 * (1 - float64(diffIdle)/float64(diffTotal))
}

func CPUHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]float64{
		"cpu": cpuUsage(),
	})
}

/* =============================
   メモリ使用率
============================= */

func MemoryHandler(w http.ResponseWriter, r *http.Request) {

	data, _ := os.ReadFile("/proc/meminfo")
	lines := strings.Split(string(data), "\n")

	var total, avail float64

	for _, l := range lines {
		if strings.HasPrefix(l, "MemTotal:") {
			fmt.Sscanf(l, "MemTotal: %f kB", &total)
		}
		if strings.HasPrefix(l, "MemAvailable:") {
			fmt.Sscanf(l, "MemAvailable: %f kB", &avail)
		}
	}

	used := total - avail
	json.NewEncoder(w).Encode(map[string]float64{
		"memory": used / total * 100,
	})
}

/* =============================
   Uptime(稼働時間)
============================= */

func UptimeHandler(w http.ResponseWriter, r *http.Request) {

	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		http.Error(w, "invalid uptime", 500)
		return
	}

	seconds, _ := strconv.ParseFloat(fields[0], 64)

	json.NewEncoder(w).Encode(map[string]float64{
		"uptime": seconds,
	})
}



func MonitorHandler(w http.ResponseWriter, r *http.Request) {

	cpu := cpuUsage()

	// uptime
	uptimeData, _ := os.ReadFile("/proc/uptime")
	uptimeFields := strings.Fields(string(uptimeData))
	uptimeSeconds, _ := strconv.ParseFloat(uptimeFields[0], 64)

	// memory
	memData, _ := os.ReadFile("/proc/meminfo")
	lines := strings.Split(string(memData), "\n")

	var memTotal, memAvail float64
	for _, l := range lines {
		if strings.HasPrefix(l, "MemTotal:") {
			fmt.Sscanf(l, "MemTotal: %f kB", &memTotal)
		}
		if strings.HasPrefix(l, "MemAvailable:") {
			fmt.Sscanf(l, "MemAvailable: %f kB", &memAvail)
		}
	}

	memUsed := memTotal - memAvail

	// disk
	var stat syscall.Statfs_t
	syscall.Statfs("/", &stat)
	diskTotal := stat.Blocks * uint64(stat.Bsize)
	diskFree := stat.Bfree * uint64(stat.Bsize)
	diskUsed := diskTotal - diskFree

	// process
	entries, _ := os.ReadDir("/proc")
	var processes []model.ProcessInfo

	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}

		status, err := os.ReadFile("/proc/" + e.Name() + "/status")
		if err != nil {
			continue
		}

		var rss uint64
		var uid int

		lines := strings.Split(string(status), "\n")
		for _, l := range lines {
			if strings.HasPrefix(l, "VmRSS:") {
				fmt.Sscanf(l, "VmRSS: %d kB", &rss)
			}
			if strings.HasPrefix(l, "Uid:") {
				fmt.Sscanf(l, "Uid: %d", &uid)
			}
		}

		username := "-"
		if u, err := user.LookupId(strconv.Itoa(uid)); err == nil {
			username = u.Username
		}

		cmd, _ := os.ReadFile("/proc/" + e.Name() + "/cmdline")

		processes = append(processes, model.ProcessInfo{
			PID:     pid,
			User:    username,
			RSS:     rss * 1024,
			Command: strings.ReplaceAll(string(cmd), "\x00", " "),
		})
	}

	response := map[string]interface{}{
		"system": map[string]interface{}{
			"cpu_usage":  cpu,
			"uptime":     uptimeSeconds,
			"mem_total":  memTotal * 1024,
			"mem_used":   memUsed * 1024,
			"mem_free":   memAvail * 1024,
			"rss_total":  0,
			"disk_total": diskTotal,
			"disk_used":  diskUsed,
			"load_avg":   []float64{0, 0, 0},
			"cpu_cores":  []float64{},
		},
		"processes": processes,
	}

	json.NewEncoder(w).Encode(response)
}
