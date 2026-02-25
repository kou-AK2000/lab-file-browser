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
   ファイル一覧
============================= */

func ListHandler(w http.ResponseWriter, r *http.Request) {

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

	var result []model.FileInfo

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

		result = append(result, model.FileInfo{
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

func FileHandler(w http.ResponseWriter, r *http.Request) {

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
