package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func hasSavedCookies() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(filepath.Dir(exePath), ".env"))
	return err == nil
}

// OpenNativeFolderPicker executes OS level terminal automation to open a real window prompt
func OpenNativeFolderPicker() string {
	switch runtime.GOOS {
	case "darwin": // macOS AppleScript engine execution
		cmd := exec.Command("osascript", "-e", `POSIX path of (choose folder with prompt "Select Base Download Destination:")`)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			return strings.TrimSpace(out.String())
		}
	case "windows": // Windows PowerShell System UI automation thread execution
		psCmd := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; if($f.ShowDialog() -eq 'OK') { $f.SelectedPath }`
		cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			return strings.TrimSpace(out.String())
		}
	}
	return "" // Return fallback empty if cancelled or operating system is headless Linux
}

func getHTMLPage() string {
	authenticatedStyle := "display: none;"
	if hasSavedCookies() {
		authenticatedStyle = "display: block;"
	}

	return `<!DOCTYPE html>
<html>
<head>
	<title>ig-downloader UI</title>
	<style>
		body { font-family: sans-serif; max-width: 500px; margin: 40px auto; padding: 20px; background: #fafafa; color: #333; }
		.box { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); position: relative; }
		.top-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
		.top-row h2 { margin: 0; }
		.settings-btn { background: none; border: none; font-size: 22px; cursor: pointer; padding: 5px; line-height: 1; transition: transform 0.2s; }
		.settings-btn:hover { transform: rotate(45deg); }
		input[type="text"] { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; border: 1px solid #ccc; border-radius: 4px; }
		.action-btn { background: #0095f6; color: white; border: none; padding: 10px 15px; border-radius: 4px; cursor: pointer; width: 100%; font-weight: bold; }
		.row { display: flex; gap: 15px; margin: 15px 0; flex-wrap: wrap; }
		.progress-container { margin-top: 20px; display: none; }
		.bar-wrapper { margin-bottom: 15px; }
		.bar-label { display: flex; justify-content: space-between; font-size: 14px; margin-bottom: 5px; text-transform: capitalize; }
		progress { width: 100%; height: 20px; }
		
		.modal { display: none; position: fixed; z-index: 1000; left: 0; top: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.4); }
		.modal-content { background: white; margin: 10% auto; padding: 25px; border-radius: 8px; width: 360px; box-shadow: 0 4px 12px rgba(0,0,0,0.2); position: relative; }
		.modal-content h3 { margin-top: 0; margin-bottom: 15px; border-bottom: 1px solid #eee; padding-bottom: 10px; }
		.setting-section { margin-bottom: 20px; }
		.setting-section label { display: block; font-size: 13px; font-weight: bold; margin-bottom: 6px; color: #555; }
		.dir-input-container { display: flex; gap: 8px; align-items: center; width: 100%; }
		.dir-text-input { flex-grow: 1; padding: 8px; font-size: 12px; border-radius: 4px; border: 1px solid #ccc; box-sizing: border-box; }
		.browse-btn { background: #e1f5fe; color: #039be5; padding: 8px 12px; border-radius: 4px; cursor: pointer; font-size: 12px; font-weight: bold; border: 1px solid #b3e5fc; border: none; }
		.modal-content textarea { width: 100%; height: 100px; box-sizing: border-box; border: 1px solid #ccc; border-radius: 4px; resize: vertical; font-size: 11px; padding: 6px; }
		.modal-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 15px; border-top: 1px solid #eee; padding-top: 15px; }
		.btn-close { background: #ef4444; color: white; border: none; padding: 8px 14px; border-radius: 4px; cursor: pointer; font-weight: bold; }
		.btn-save { background: #22c55e; color: white; border: none; padding: 8px 14px; border-radius: 4px; cursor: pointer; font-weight: bold; }
	</style>
</head>
<body>
	<div id="settingsModal" class="modal">
		<div class="modal-content">
			<h3>Engine Configurations</h3>
			
			<div class="setting-section">
				<label>Base Download Location</label>
				<div class="dir-input-container">
					<input type="text" id="dirPathDisplay" class="dir-text-input" value="." placeholder="e.g., /Users/name/Downloads">
					<button class="browse-btn" onclick="triggerNativePicker()">Browse</button>
				</div>
			</div>

			<div class="setting-section">
				<label>Instagram Session Cookies</label>
				<textarea id="cookieInput" placeholder="Paste your raw browser headers text string context or JSON array maps right here..."></textarea>
			</div>

			<div class="modal-actions">
				<button class="btn-close" onclick="closeModal()">Cancel</button>
				<button class="btn-save" onclick="commitGlobalSettings()">Save</button>
			</div>
		</div>
	</div>

	<div class="box">
		<div class="top-row">
			<h2>Instagram Downloader</h2>
			<button class="settings-btn" onclick="openModal()" title="Open System Settings">⚙️</button>
		</div>
		
		<input type="text" id="url" placeholder="Paste instagram profile link here...">
		<div class="row">
			<label><input type="checkbox" id="posts" checked> Posts</label>
			<label><input type="checkbox" id="highlights" checked> Highlights</label>
			<label class="auth-label" style="` + authenticatedStyle + `"><input type="checkbox" id="reels" checked> Reels</label>
			<label class="auth-label" style="` + authenticatedStyle + `"><input type="checkbox" id="stories" checked> Stories</label>
		</div>
		<button class="action-btn" id="startBtn" onclick="startDownload()">Download</button>

		<div id="progressSection" class="progress-container">
			<h3>Download Progress</h3>
			<div id="bars"></div>
		</div>
	</div>

	<script>
		let eventSource = null;
		let progressData = {};

		function openModal() { document.getElementById('settingsModal').style.display = 'block'; }
		function closeModal() { document.getElementById('settingsModal').style.display = 'none'; }

		function triggerNativePicker() {
			fetch('/select-directory')
			.then(res => res.text())
			.then(absolutePath => {
				if(absolutePath.trim()) {
					document.getElementById('dirPathDisplay').value = absolutePath;
				}
			});
		}

		function commitGlobalSettings() {
			const rawData = document.getElementById('cookieInput').value.trim();
			const selectedCustomPath = document.getElementById('dirPathDisplay').value.trim() || ".";
			
			fetch('/update-settings', {
				method: 'POST',
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				body: "dir=" + encodeURIComponent(selectedCustomPath) + "&cookies=" + encodeURIComponent(rawData)
			})
			.then(res => {
				if(rawData && res.ok) {
					const authLabels = document.querySelectorAll('.auth-label');
					authLabels.forEach(el => el.style.display = "block");
				}
				document.getElementById('cookieInput').value = '';
				closeModal(); // Dialogue popups removed entirely as requested
			});
		}

		function startDownload() {
			const url = document.getElementById('url').value;
			const runPosts = document.getElementById('posts').checked;
			const runHighlights = document.getElementById('highlights').checked;
			
			const reelsEl = document.getElementById('reels');
			const runReels = reelsEl ? reelsEl.checked : false;
			
			const storiesEl = document.getElementById('stories');
			const runStories = storiesEl ? storiesEl.checked : false;

			if(!url) return alert("Please specify a profile link");

			document.getElementById('bars').innerHTML = '';
			document.getElementById('progressSection').style.display = 'block';
			progressData = {};

			if(runPosts) createProgressTrack('posts');
			if(runHighlights) createProgressTrack('highlights');
			if(runReels) createProgressTrack('reels');
			if(runStories) createProgressTrack('stories');

			if(eventSource) eventSource.close();
			eventSource = new EventSource('/stream');

			eventSource.onmessage = function(event) {
				const data = JSON.parse(event.data);
				const cat = data.category;
				
				if(!progressData[cat]) createProgressTrack(cat);
				const progObj = progressData[cat];

				if (data.type === 'init_update' || data.type === 'init') {
					if (data.value > 0 && data.value > progObj.total) {
						progObj.total = data.value;
						progObj.el.max = data.value;
					}
					if (progObj.total > 0) {
						progObj.label.innerText = cat + ": Found " + progObj.total + " assets...";
					} else {
						progObj.label.innerText = cat + ": Scanning profile metrics...";
					}
				} else if(data.type === 'progress') {
					progObj.current += data.value;
					progObj.el.value = progObj.current;
					const visualTotal = progObj.total > 0 ? progObj.total : progObj.current;
					progObj.label.innerText = cat + ": " + progObj.current + " / " + visualTotal;
				}
			};

			fetch('/download', {
				method: 'POST',
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				body: "url=" + encodeURIComponent(url) + "&posts=" + runPosts + "&highlights=" + runHighlights + "&reels=" + runReels + "&stories=" + runStories
			});
		}

		function createProgressTrack(cat) {
			const wrapper = document.createElement('div');
			wrapper.className = 'bar-wrapper';
			wrapper.innerHTML = "<div class='bar-label' id='label-" + cat + "'>" + cat + ": Finding assets...</div><progress id='bar-" + cat + "' value='0' max='100'></progress>";
			document.getElementById('bars').appendChild(wrapper);

			progressData[cat] = {
				current: 0, total: 0,
				el: document.getElementById("bar-" + cat),
				label: document.getElementById("label-" + cat)
			};
		}
	</script>
</body>
</html>`
}

var currentWebOutputDir = "."

func StartLocalWebServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, getHTMLPage())
	})

	// Bridge link handler to securely pass back exact native file explorer location path
	http.HandleFunc("/select-directory", func(w http.ResponseWriter, r *http.Request) {
		chosenPath := OpenNativeFolderPicker()
		fmt.Fprint(w, chosenPath)
	})

	http.HandleFunc("/update-settings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { return }
		_ = r.ParseForm()

		dirParam := r.FormValue("dir")
		cookieParam := r.FormValue("cookies")

		if dirParam != "" {
			currentWebOutputDir = dirParam
		}

		if cookieParam != "" {
			_ = ParseAndSaveCookies(cookieParam)
		}
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { return }
		_ = r.ParseForm()

		rawURL := r.FormValue("url")
		u, _ := url.Parse(rawURL)
		segments := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(segments) == 0 || segments[0] == "" { return }

		cfg := DownloadConfig{
			Username:           segments[0],
			DownloadPosts:      r.FormValue("posts") == "true",
			DownloadHighlights: r.FormValue("highlights") == "true",
			DownloadReels:      r.FormValue("reels") == "true",
			DownloadStories:    r.FormValue("stories") == "true",
			Concurrency:        10,
			OutputDir:          currentWebOutputDir,
		}

		go runWebDownloadPipeline(cfg)
	})

	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		for event := range WebProgressChan {
			bytes, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", string(bytes))
			if f, ok := w.(http.Flusher); ok { f.Flush() }
		}
	})

	fmt.Println("[*] Web UI interface initialized. Server listening on http://localhost:8080")
	_ = http.ListenAndServe(":8080", nil)
}

func runWebDownloadPipeline(config DownloadConfig) {
	client := NewHTTPClient()
	var queue []UniversalDownloadAsset

	if config.DownloadPosts {
		assets, err := GatherAndStructurePosts(client, config.Username)
		if err == nil {
			for _, a := range assets {
				queue = append(queue, UniversalDownloadAsset{
					DownloadURL: a.DownloadURL, 
					LocalPath:   filepath.Join(config.OutputDir, a.LocalPath), 
					Category:    "posts",
				})
			}
		}
	}

	if config.DownloadHighlights {
		assets, err := GatherAndStructureHighlights(client, config.Username)
		if err == nil {
			for _, a := range assets {
				queue = append(queue, UniversalDownloadAsset{
					DownloadURL: a.DownloadURL, 
					LocalPath:   filepath.Join(config.OutputDir, a.LocalPath), 
					Category:    "highlights",
				})
			}
		}
	}

	if config.DownloadReels && hasSavedCookies() {
		assets, err := GatherAndStructureReels(client, config.Username)
		if err == nil {
			for _, a := range assets {
				queue = append(queue, UniversalDownloadAsset{
					DownloadURL: a.DownloadURL, 
					LocalPath:   filepath.Join(config.OutputDir, a.LocalPath), 
					Category:    "reels",
				})
			}
		}
	}

	if config.DownloadStories && hasSavedCookies() {
		assets, err := GatherAndStructureStories(client, config.Username)
		if err == nil {
			for _, a := range assets {
				queue = append(queue, UniversalDownloadAsset{
					DownloadURL: a.DownloadURL, 
					LocalPath:   filepath.Join(config.OutputDir, a.LocalPath), 
					Category:    "stories",
				})
			}
		}
	}

	if len(queue) > 0 {
		ConcurrentDownloadPool(client, queue, config.Concurrency)
	}
}