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

Make sure that your `PATH` includes the Go bin directory. You can add it to your `~/.bashrc` or `~/.zshrc` file.

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

### Download the binary

Go to the [Releases](https://github.com/gofast-live/gofast-cli/releases) page and download the appropriate binary for your operating system.

### Install the binary

#### Linux

```bash
wget https://github.com/gofast-live/gofast-cli/releases/download/v0.12.0/gofast-linux-amd64 -O /usr/local/bin/gofast
chmod +x /usr/local/bin/gofast
```

#### macOS

```bash
wget https://github.com/gofast-live/gofast-cli/releases/download/v0.12.0/gofast-darwin-amd64 -O /usr/local/bin/gofast
chmod +x /usr/local/bin/gofast
```

#### Windows

```bash
curl -L -o gofast.exe https://github.com/gofast-live/gofast-cli/releases/download/v0.12.0/gofast-windows-amd64.exe
move gofast.exe C:\Windows\System32
```
