package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const HTMLPage = `<!DOCTYPE html>
<html>
<head>
    <title>ig-downloader UI</title>
    <style>
        body { font-family: sans-serif; max-width: 500px; margin: 40px auto; padding: 20px; background: #fafafa; color: #333; }
        .box { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        input[type="text"] { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; border: 1px solid #ccc; border-radius: 4px; }
        button { background: #0095f6; color: white; border: none; padding: 10px 15px; border-radius: 4px; cursor: pointer; width: 100%; font-weight: bold; }
        .row { display: flex; gap: 20px; margin: 15px 0; }
        .progress-container { margin-top: 20px; display: none; }
        .bar-wrapper { margin-bottom: 15px; }
        .bar-label { display: flex; justify-content: space-between; font-size: 14px; margin-bottom: 5px; text-transform: capitalize; }
        progress { width: 100%; height: 20px; }
    </style>
</head>
<body>
    <div class="box">
        <h2>Instagram Downloader</h2>
        <input type="text" id="url" placeholder="Paste instagram profile link here...">
        <div class="row">
            <label><input type="checkbox" id="posts" checked> Posts</label>
            <label><input type="checkbox" id="highlights" checked> Highlights</label>
        </div>
        <button id="startBtn" onclick="startDownload()">Download</button>

        <div id="progressSection" class="progress-container">
            <h3>Download Progress</h3>
            <div id="bars"></div>
        </div>
    </div>

    <script>
        let eventSource = null;
        let progressData = {};

        function startDownload() {
            const url = document.getElementById('url').value;
            const runPosts = document.getElementById('posts').checked;
            const runHighlights = document.getElementById('highlights').checked;

            if(!url) return alert("Please specify a profile link");

            document.getElementById('bars').innerHTML = '';
            document.getElementById('progressSection').style.display = 'block';
            progressData = {};

            if(runPosts) createProgressTrack('posts');
            if(runHighlights) createProgressTrack('highlights');

            if(eventSource) eventSource.close();
            eventSource = new EventSource('/stream');

            eventSource.onmessage = function(event) {
                const data = JSON.parse(event.data);
                const cat = data.category;
                
                if(!progressData[cat]) createProgressTrack(cat);
                const progObj = progressData[cat];

                if (data.type === 'init_update' || data.type === 'init') {
                    if (data.value > progObj.total) {
                        progObj.total = data.value;
                        progObj.el.max = data.value;
                    }
                    if (progObj.current === 0) {
                        progObj.label.innerText = cat + ": Found " + progObj.total + " assets...";
                    }
                } else if(data.type === 'progress') {
                    progObj.current += data.value;
                    progObj.el.value = progObj.current;
                    progObj.label.innerText = cat + ": " + progObj.current + " / " + progObj.total;
                }
            };

            // Fire download independently to avoid EventSource open-handshake locking bugs
            fetch('/download', {
                method: 'POST',
                headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                body: "url=" + encodeURIComponent(url) + "&posts=" + runPosts + "&highlights=" + runHighlights
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

func StartLocalWebServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, HTMLPage)
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
			Concurrency:        10,
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
			WebProgressChan <- ProgressEvent{Category: "posts", Type: "init", Value: len(assets)}
			for _, a := range assets {
				queue = append(queue, UniversalDownloadAsset{DownloadURL: a.DownloadURL, LocalPath: a.LocalPath, Category: "posts"})
			}
		}
	}

	if config.DownloadHighlights {
		assets, err := GatherAndStructureHighlights(client, config.Username)
		if err == nil {
			WebProgressChan <- ProgressEvent{Category: "highlights", Type: "init", Value: len(assets)}
			for _, a := range assets {
				queue = append(queue, UniversalDownloadAsset{DownloadURL: a.DownloadURL, LocalPath: a.LocalPath, Category: "highlights"})
			}
		}
	}

	if len(queue) > 0 {
		ConcurrentDownloadPool(client, queue, config.Concurrency)
	}
}