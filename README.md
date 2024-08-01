# GoFast CLI

The Ultimate Foundation for High-Performance, Scalable Web Applications.

## Build

```bash
GOOS=linux GOARCH=amd64 go build -o gofast-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o gofast-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o gofast-windows-amd64.exe
```

## Installation

### Using Go

```sh
go install github.com/gofast-live/gofast-cli/cmd/gofast@latest
```

### Download the binary

Go to the [Releases](https://github.com/gofast-live/gofast-cli/releases) page and download the appropriate binary for your operating system.

### Install the binary

#### Linux

```sh
wget https://github.com/gofast-live/gofast-cli/releases/download/v0.0.3/gofast-linux-amd64 -O /usr/local/bin/gofast
chmod +x /usr/local/bin/gofast
```

#### macOS

```sh
wget https://github.com/gofast-live/gofast-cli/releases/download/v0.0.3/gofast-darwin-amd64 -O /usr/local/bin/gofast
chmod +x /usr/local/bin/gofast
```

#### Windows
```sh
curl -L -o gofast.exe https://github.com/gofast-live/gofast-cli/releases/download/v0.0.3/gofast-windows-amd64.exe
move gofast.exe C:\Windows\System32
```

