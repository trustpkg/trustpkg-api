# trustpkg-api

API for tracking vulnerability occurrence frequencies.

## How to start

1. Initialize Git hooks

```bash
make init
2. Install Go
```

This step depends on your operating system.

For Fedora:

```bash
sudo dnf install golang
```

Verify the installation:

```bash
go version
```
Configure Go tools

If you use Bash, add the following to ~/.bashrc:

```bash
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
export PATH="$PATH:$GOBIN"
```

Then reload your shell configuration:

```bash
source ~/.bashrc
```

Install Air for automatic application reloading during development:

```bash
go install github.com/air-verse/air@latest
```

3. Run the service

```bash
make replicator-service
```
