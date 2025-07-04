# Obsidian Publish Downloader 

Obsidian Publish downloader in Go, no external dependencies and simple.

## Usage

1. Download the binary from the [releases](https://github.com/Sudo-Ivan/obsidian-publish-downloader/releases) page.

or use Go:

```bash
go install github.com/Sudo-Ivan/obsidian-publish-downloader@latest
```

or build from source:

```bash
go build -o obsidian-downloader main.go
./obsidian-downloader URL FOLDER
```

## Example

```bash
./obsidian-downloader https://your-site.obsidian.md/ downloads/
```

## Features

- Zero external dependencies (uses only Go standard library)
- Downloads all files from an Obsidian Publish site
- Creates necessary directories automatically
- Progress tracking during download
- Error handling and cleanup

## How it works

1. Fetches the main page and extracts site information
2. Downloads the cache data containing file metadata
3. Downloads each file to the specified folder
4. Creates parent directories as needed 