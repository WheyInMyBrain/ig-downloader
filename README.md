# Instagram Downloader (ig-downloader)

A zero-dependency, standalone tool written in Go to download timeline posts, story highlights, reels, and live active stories via a command-line interface (CLI) or an interactive Web UI.

## Target Binaries

* **macOS (Apple Silicon):** `ig-downloader-mac-arm64`
* **macOS (Intel):** `ig-downloader-mac-amd64`
* **Windows (64-bit standard):** `ig-downloader-windows-amd64.exe`
* **Windows (ARM standard):** `ig-downloader-windows-arm64.exe`

---

## How to Run

### On macOS

1. Open your **Terminal** app.
2. Navigate to the folder containing your binary and grant execution permissions:

```bash
chmod +x ./ig-downloader-mac-arm64

```

3. Run the binary against a target profile link:

```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/)

```

### On Windows

1. Open **Command Prompt** or **PowerShell**.
2. Navigate to the folder containing your executable and run it against a target profile link:

```powershell
.\ig-downloader-windows-amd64.exe [https://www.instagram.com/username/](https://www.instagram.com/username/)

```

---

## Available Command Options

* **Download Everything (Default CLI Mode):**
* *Note: Reels and active Stories extraction will run automatically if authenticated session cookies are found on your disk.*

```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/)

```

* **Download Posts Only (`--p`):**

```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/) --p

```

* **Download Highlights Only (`--h`):**

```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/) --h

```

* **Download Reels Only (`--r`):**
* *Requires an active authenticated session profile.*

```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/) --r

```

* **Download Active Stories Only (`--s`):**
* *Requires an active authenticated session profile.*

```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/) --s

```

* **Set Download Output Directory (`--dir`):**
* *Redirects the download location to a custom absolute or relative drive path instead of the default current binary directory folder.*

```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/) --dir "/Users/yourname/Downloads/Instagram"

```

* **Save Session Configuration Standalone (`--cookies`):**
* *Saves raw browser payload network headers text strings or a JSON cookie array from your browser straight to your disk workspace (`.env`), then exits cleanly.*

```bash
./ig-downloader-mac-arm64 --cookies "your_raw_cookie_or_header_string_here"

```

* **Change Concurrent Workers (`--workers`):**
* *Modifies the parallel worker pool limit size to control network download throttling speeds.*

```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/) --workers 5

```

* **Start Local Web UI Server (`--serve`):**
* *Launches the local lightweight web instance hosting the system controller control panel dashboard.*

```bash
./ig-downloader-mac-arm64 --serve

```

*Once started, open your web browser and navigate to:* `http://localhost:8080`

---

## Running with Docker

Running the application inside an isolated Docker container prevents local environment configuration clutter. Because containers run in sandboxed filesystems, you must expose an external folder mount to direct media downloads directly onto your computer's main storage workspace.

1. **Build the container image:**

```bash
docker build -t ig-downloader .

```

2. **Run the container instance using an External Storage Volume Mount:**

Choose the command matching your operating shell type below to link a folder named `instagram_downloads` on your computer directly to the container endpoint pipeline:

#### On Mac / Linux Terminal:

```bash
docker run -d \
  -p 8080:8080 \
  -v "$(pwd)/instagram_downloads:/downloads" \
  --name ig-dl-server \
  ig-downloader

```

#### On Windows PowerShell:

```powershell
docker run -d `
  -p 8080:8080 `
  -v "${PWD}/instagram_downloads:/downloads" `
  --name ig-dl-server `
  ig-downloader

```

3. Open `http://localhost:8080` in your web browser.

*Note: When using the Web UI inside a Docker environment, keep the **Base Download Location** path value in the engine configuration settings window set to its default (`.`) or explicitly typed as `/downloads`. The assets will automatically route across the container volume boundary layer directly onto your local desktop computer host storage drive.*
