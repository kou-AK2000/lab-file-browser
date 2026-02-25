package server

import (
	"net/http"

	"linux-fs-viewer/api"
)

func registerRoutes(mux *http.ServeMux) {

	mux.HandleFunc("/api/list", api.ListHandler)
	mux.HandleFunc("/api/file", api.FileHandler)
	mux.HandleFunc("/api/system", api.SystemHandler)
	mux.HandleFunc("/api/monitor", api.MonitorHandler)
	mux.HandleFunc("/api/cpu", api.CPUHandler)
	mux.HandleFunc("/api/memory", api.MemoryHandler)
	mux.HandleFunc("/api/uptime", api.UptimeHandler)
	mux.HandleFunc("/api/disk", api.DiskHandler)
	mux.HandleFunc("/api/df", api.DFHandler)
	mux.HandleFunc("/api/process", api.ProcessHandler)
	mux.HandleFunc("/api/processes", api.ProcessesHandler)

	// static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)
}
