// WarDogs Keybind Editor - a small local app that edits WarDogs' GameUserSettings.ini.
// It serves the editor page on 127.0.0.1 only and opens it in your default browser.
// It reads and writes exactly one file: %LOCALAPPDATA%\Wardogs\Saved\Config\WindowsClient\GameUserSettings.ini
// (or ...\Config\Windows\ if that's where yours is), and keeps backups next to it.
package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed index.html
var page string

const version = "1.2"

var token string

func configPath() string {
	base := os.Getenv("LOCALAPPDATA")
	for _, sub := range []string{"WindowsClient", "Windows"} {
		p := filepath.Join(base, "Wardogs", "Saved", "Config", sub, "GameUserSettings.ini")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(base, "Wardogs", "Saved", "Config", "WindowsClient", "GameUserSettings.ini")
}

// gameRunning looks for a running process whose name contains "wardogs" (not this program).
func gameRunning() bool {
	out, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	if err != nil {
		return false
	}
	self, _ := os.Executable()
	selfName := strings.ToLower(filepath.Base(self))
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.SplitN(line, ",", 2)
		name := strings.ToLower(strings.Trim(f[0], "\" \r"))
		if strings.Contains(name, "wardogs") && name != selfName {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func authed(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("X-Token") != token {
		writeJSON(w, 403, map[string]string{"error": "bad token"})
		return false
	}
	return true
}

func main() {
	b := make([]byte, 16)
	rand.Read(b)
	token = hex.EncodeToString(b)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		io.WriteString(w, strings.Replace(page, "/*DESKTOP*/", "window.WD_DESKTOP={token:"+jsString(token)+",version:"+jsString(version)+"};", 1))
	})
	mux.HandleFunc("/api/load", func(w http.ResponseWriter, r *http.Request) {
		if !authed(w, r) {
			return
		}
		p := configPath()
		data, err := os.ReadFile(p)
		if err != nil {
			writeJSON(w, 404, map[string]string{"error": "Couldn't find GameUserSettings.ini", "path": p})
			return
		}
		writeJSON(w, 200, map[string]any{"path": p, "text": string(data), "running": gameRunning()})
	})
	mux.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		if !authed(w, r) || r.Method != http.MethodPost {
			return
		}
		var req struct {
			Text  string `json:"text"`
			Force bool   `json:"force"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&req); err != nil || len(req.Text) < 20 {
			writeJSON(w, 400, map[string]string{"error": "Nothing to save"})
			return
		}
		if !req.Force && gameRunning() {
			writeJSON(w, 409, map[string]any{"running": true})
			return
		}
		p := configPath()
		dir := filepath.Dir(p)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		var backup string
		if old, err := os.ReadFile(p); err == nil {
			orig := filepath.Join(dir, "GameUserSettings.original-backup.ini")
			if _, err := os.Stat(orig); os.IsNotExist(err) {
				os.WriteFile(orig, old, 0o644)
			}
			backup = filepath.Join(dir, "GameUserSettings.backup-"+time.Now().Format("2006-01-02_150405")+".ini")
			if err := os.WriteFile(backup, old, 0o644); err != nil {
				writeJSON(w, 500, map[string]string{"error": "Couldn't write backup: " + err.Error()})
				return
			}
		}
		tmp := p + ".tmp"
		if err := os.WriteFile(tmp, []byte(req.Text), 0o644); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		if err := os.Rename(tmp, p); err != nil {
			os.Remove(tmp)
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"path": p, "backup": backup})
	})
	mux.HandleFunc("/api/folder", func(w http.ResponseWriter, r *http.Request) {
		if !authed(w, r) || r.Method != http.MethodPost {
			return
		}
		exec.Command("explorer", filepath.Dir(configPath())).Start()
		writeJSON(w, 200, map[string]bool{"ok": true})
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("Couldn't start:", err)
		pause()
		return
	}
	url := "http://" + ln.Addr().String() + "/"
	fmt.Println("WarDogs Keybind Editor " + version)
	fmt.Println()
	fmt.Println("The editor is opening in your browser:")
	fmt.Println("  " + url)
	fmt.Println()
	fmt.Println("Settings file: " + configPath())
	fmt.Println()
	fmt.Println("Keep this window open while you edit. Close it when you're done.")
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	if err := http.Serve(ln, mux); err != nil {
		fmt.Println(err)
		pause()
	}
}

func jsString(s string) string { b, _ := json.Marshal(s); return string(b) }

func pause() {
	fmt.Println("Press Enter to close.")
	fmt.Scanln()
}
