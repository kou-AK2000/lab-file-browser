package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

/* =============================
   構造体
============================= */

type FileInfo struct {
	Name     string `json:"name"`
	IsDir    bool   `json:"is_dir"`
	Size     int64  `json:"size"`
	Perm     string `json:"perm"`
	Mode     string `json:"mode"`
	Nlink    uint32 `json:"nlink"` // ← これを uint32 にする
	Owner    string `json:"owner"`
	Group    string `json:"group"`
	ModTime  string `json:"mod_time"`
	Readable bool   `json:"readable"`
}

type SystemInfo struct {
	Hostname     string `json:"hostname"`
	Distribution string `json:"distribution"`
	Kernel       string `json:"kernel"`
	Arch         string `json:"arch"`
}

type ProcessInfo struct {
	PID     int    `json:"pid"`
	User    string `json:"user"`
	RSS     uint64 `json:"rss"`
	VSZ     uint64 `json:"vsz"`
	State   string `json:"state"`
	Command string `json:"command"`
}

type DiskInfo struct {
	Filesystem string  `json:"filesystem"`
	Size       uint64  `json:"size"`
	Used       uint64  `json:"used"`
	Avail      uint64  `json:"avail"`
	UsePercent float64 `json:"use_percent"`
	MountedOn  string  `json:"mounted_on"`
}

/* =============================
   ファイル一覧
============================= */

func listHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	// 🔹 追加：隠しファイル表示フラグ取得
	showHidden := r.URL.Query().Get("show_hidden") == "true"

	entries, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var result []FileInfo

	for _, e := range entries {

		// 🔹 追加：隠しファイルフィルタ
		if !showHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}

		fullPath := path + "/" + e.Name()
		info, err := e.Info()
		if err != nil {
			continue
		}

		var modeStr string
		var nlink uint32
		var ownerName string = "-"
		var groupName string = "-"
		readable := true

		// 特殊ファイル判定
		mode := info.Mode()
		if mode&os.ModeDevice != 0 ||
			mode&os.ModeSocket != 0 ||
			mode&os.ModeNamedPipe != 0 ||
			mode&os.ModeSymlink != 0 {
			readable = false
		}

		// 仮想FS拒否
		if strings.HasPrefix(fullPath, "/proc") ||
			strings.HasPrefix(fullPath, "/sys") ||
			strings.HasPrefix(fullPath, "/dev") {
			readable = false
		}

		if stat, ok := info.Sys().(*syscall.Stat_t); ok {

			uid := strconv.Itoa(int(stat.Uid))
			gid := strconv.Itoa(int(stat.Gid))

			if u, err := user.LookupId(uid); err == nil {
				ownerName = u.Username
			}

			if g, err := user.LookupGroupId(gid); err == nil {
				groupName = g.Name
			}

			nlink = stat.Nlink
			modeStr = info.Mode().String()
		}

		result = append(result, FileInfo{
			Name:     e.Name(),
			IsDir:    e.IsDir(),
			Size:     info.Size(),
			Perm:     fmt.Sprintf("%o", info.Mode().Perm()),
			Mode:     modeStr,
			Nlink:    nlink,
			Owner:    ownerName,
			Group:    groupName,
			ModTime:  info.ModTime().Format("2006-01-02 15:04:05"),
			Readable: readable,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* =============================
   ファイル閲覧
============================= */

func fileHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Query().Get("path")
	info, err := os.Lstat(path)
	if err != nil || info.IsDir() {
		http.Error(w, "invalid file", 400)
		return
	}

	// 特殊ファイル拒否
	mode := info.Mode()
	if mode&os.ModeDevice != 0 ||
		mode&os.ModeSocket != 0 ||
		mode&os.ModeNamedPipe != 0 ||
		mode&os.ModeSymlink != 0 {
		http.Error(w, "special file not allowed", 400)
		return
	}

	// 仮想FS拒否
	if strings.HasPrefix(path, "/proc") ||
		strings.HasPrefix(path, "/sys") ||
		strings.HasPrefix(path, "/dev") {
		http.Error(w, "virtual fs not allowed", 400)
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(data)
}

/* =============================
   システム情報
============================= */

func systemHandler(w http.ResponseWriter, r *http.Request) {

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

	json.NewEncoder(w).Encode(SystemInfo{
		Hostname:     host,
		Distribution: dist,
		Kernel:       kernel,
		Arch:         runtime.GOARCH,
	})
}

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

func cpuHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]float64{
		"cpu": cpuUsage(),
	})
}

/* =============================
   メモリ使用率
============================= */

func memoryHandler(w http.ResponseWriter, r *http.Request) {

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

func uptimeHandler(w http.ResponseWriter, r *http.Request) {

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

/* =============================
   Disk使用率
============================= */

func diskHandler(w http.ResponseWriter, r *http.Request) {

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

func monitorHandler(w http.ResponseWriter, r *http.Request) {

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
	var processes []ProcessInfo

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

		processes = append(processes, ProcessInfo{
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

/* =============================
   プロセス一覧
============================= */

func processHandler(w http.ResponseWriter, r *http.Request) {

	entries, _ := os.ReadDir("/proc")
	var list []ProcessInfo

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

		list = append(list, ProcessInfo{
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

func dfHandler(w http.ResponseWriter, r *http.Request) {
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

type Process struct {
	User    string `json:"user"`
	Pid     string `json:"pid"`
	Cpu     string `json:"cpu"`
	Mem     string `json:"mem"`
	Command string `json:"command"`
}

func processesHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("ps", "aux", "--sort=-%cpu")
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "ps failed", http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(output), "\n")

	var processes []Process

	// 1行目はヘッダなのでスキップ
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		processes = append(processes, Process{
			User:    fields[0],
			Pid:     fields[1],
			Cpu:     fields[2],
			Mem:     fields[3],
			Command: strings.Join(fields[10:], " "),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(processes)
}

/* =============================
   main
============================= */

func main() {

	if os.Geteuid() == 0 {
		fmt.Println("Do not run as root")
		return
	}

	http.HandleFunc("/api/list", listHandler)
	http.HandleFunc("/api/file", fileHandler)
	http.HandleFunc("/api/system", systemHandler)
	http.HandleFunc("/api/monitor", monitorHandler)
	http.HandleFunc("/api/cpu", cpuHandler)
	http.HandleFunc("/api/memory", memoryHandler)
	http.HandleFunc("/api/uptime", uptimeHandler)
	http.HandleFunc("/api/disk", diskHandler)
	http.HandleFunc("/api/df", dfHandler)
	http.HandleFunc("/api/process", processHandler)
	http.HandleFunc("/api/processes", processesHandler)

	// static 配信はこれだけでOK
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	fmt.Println("Running at http://127.0.0.1:8080")
	http.ListenAndServe("127.0.0.1:8080", nil)
}
