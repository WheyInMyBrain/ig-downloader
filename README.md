# Instagram Downloader (ig-downloader)

A zero-dependency, standalone CLI tool to download timeline posts and story highlights.

## Target Binaries
* **macOS (Apple Silicon M1/M2/M3/M4):** `ig-downloader-mac-arm64`
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
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/)

```



### On Windows

1. Open **Command Prompt** or **PowerShell**.
2. Run the executable against a profile link:
```powershell
.\ig-downloader-windows-amd64.exe [https://www.instagram.com/username/](https://www.instagram.com/username/)

```


---

## Available Command Options

* **Download Everything (Default):**
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


* **Change Concurrent Workers (`--workers`):**
```bash
./ig-downloader-mac-arm64 [https://www.instagram.com/username/](https://www.instagram.com/username/) --workers 5

```
