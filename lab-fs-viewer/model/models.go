package model

type FileInfo struct {
	Name     string `json:"name"`
	IsDir    bool   `json:"is_dir"`
	Size     int64  `json:"size"`
	Perm     string `json:"perm"`
	Mode     string `json:"mode"`
	Nlink    uint32 `json:"nlink"`
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
