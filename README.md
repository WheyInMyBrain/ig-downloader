
---

# Instagram Downloader (ig-downloader)

A zero-dependency, standalone tool to download timeline posts, story highlights, and reels via CLI or Web UI.

## Target Binaries

* **macOS (Apple Silicon):** `ig-downloader-mac-arm64`
* **macOS (Intel):** `ig-downloader-mac-amd64`
* **Windows (64-bit standard):** `ig-downloader-windows-amd64.exe`
* **Windows (ARM standard):** `ig-downloader-windows-arm64.exe`

---

## How to Run

### On macOS

1. Open your **Terminal** app.
2. Grant execution permissions to the downloaded file:

```bash
chmod +x ./ig-downloader-mac-arm64

```

3. Run the binary against a profile link:

```bash
./ig-downloader-mac-arm64 https://www.instagram.com/username/

```

### On Windows

1. Open **Command Prompt** or **PowerShell**.

2. Run the executable against a profile link:

```powershell
.\ig-downloader-windows-amd64.exe https://www.instagram.com/username/

```

---

## Available Command Options

* **Download Everything (Default CLI):**
* *Note: Reels extraction will trigger automatically if an authenticated `.env` configuration file is present on disk.*



```bash
./ig-downloader-mac-arm64 https://www.instagram.com/username/

```

* **Download Posts Only (`--p`):**

```bash
./ig-downloader-mac-arm64 https://www.instagram.com/username/ --p

```

* **Download Highlights Only (`--h`):**

```bash
./ig-downloader-mac-arm64 https://www.instagram.com/username/ --h

```

* **Download Reels Only (`--r`):**
* *Requires a valid session configuration file to access metadata graphs.*



```bash
./ig-downloader-mac-arm64 https://www.instagram.com/username/ --r

```

* **Save Session Configuration Standalone (`--cookies`):**
* *Saves raw headers text or a JSON cookie array from your browser straight to your disk environment, then exits cleanly.*



```bash
./ig-downloader-mac-arm64 --cookies "your_raw_cookie_string_here"

```

* **Change Concurrent Workers (`--workers`):**

```bash
./ig-downloader-mac-arm64 https://www.instagram.com/username/ --workers 5

```

* **Start Local Web UI Server (`--serve`):**

```bash
./ig-downloader-mac-arm64 --serve

```

*Once started, open your browser and navigate to:* `http://localhost:8080`

---

## Running with Docker

If you prefer to run the Web UI inside an isolated Docker container:

1. **Build the container image:**

```bash
docker build -t ig-downloader .

```

2. **Run the container instance:**

```bash
docker run -d -p 8080:8080 --name ig-dl-server ig-downloader

```

3. Open `http://localhost:8080` in your web browser to download assets.